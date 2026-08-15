package storage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"backend/internal/database"
	"backend/internal/httpx"
	"backend/internal/shared/access"
)

// formOverhead คือส่วนเผื่อของ multipart body ที่ไม่ใช่ตัวไฟล์ (boundary, ชื่อฟิลด์, ฟิลด์ข้อความ)
const formOverhead = 1 << 20

// DBStore เก็บรูปเป็นเนื้อไฟล์จริงในตาราง assets แล้วเสิร์ฟผ่าน GET /files/:assetID
//
// ต่างจาก LocalStore ที่ยังใช้กับสลิปโอนเงิน: รูปสาขา/ห้องไม่ต้องพึ่ง volume
// สำรองฐานข้อมูลชุดเดียวก็ได้ทั้งข้อมูลและรูป
//
// ค่าที่คืนออกไปเก็บลงคอลัมน์ image_url/cover_image_url เหมือนเดิม
// ฝั่งที่เรียกใช้จึงไม่ต้องรู้ว่ารูปถูกเก็บที่ไหน
type DBStore struct {
	db       *database.TxManager
	baseURL  string
	maxBytes int64
}

func NewDBStore(db *database.TxManager, publicBaseURL string, maxBytes int64) *DBStore {
	return &DBStore{db: db, baseURL: publicBaseURL, maxBytes: maxBytes}
}

func (s *DBStore) MaxBytes() int64 { return s.maxBytes }

func (s *DBStore) URL(id uuid.UUID) string { return s.baseURL + "/files/" + id.String() }

// SaveFromRequest ดึงไฟล์จาก multipart form field เดียวแล้วเก็บลงฐานข้อมูล คืน URL สาธารณะ
//
// ต้องเรียกก่อนอ่านฟิลด์ข้อความอื่นใน form เพราะเป็นจุดที่พาร์ส multipart
func (s *DBStore) SaveFromRequest(c *gin.Context, field string) (string, error) {
	// จำกัดขนาดตั้งแต่ระดับ reader ไม่ให้ไฟล์ยักษ์ถูกอ่านเข้าหน่วยความจำก่อนจะได้ตรวจ
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, s.maxBytes+formOverhead)

	header, err := c.FormFile(field)
	if err != nil {
		return "", httpx.BadRequest("กรุณาแนบไฟล์ในช่อง " + field).Wrap(err)
	}
	file, err := header.Open()
	if err != nil {
		return "", httpx.BadRequest("เปิดไฟล์ที่แนบมาไม่สำเร็จ").Wrap(err)
	}
	defer file.Close()
	return s.Save(c.Request.Context(), file, header)
}

func (s *DBStore) Save(ctx context.Context, file multipart.File, header *multipart.FileHeader) (string, error) {
	if header.Size > s.maxBytes {
		return "", errTooLarge(s.maxBytes)
	}
	contentType, _, err := sniffType(file)
	if err != nil {
		return "", err
	}

	// อ่านเกินเพดาน 1 ไบต์เพื่อแยก "พอดีเพดาน" ออกจาก "เกินเพดานแต่ header โกหกขนาดมา"
	data, err := io.ReadAll(io.LimitReader(file, s.maxBytes+1))
	if err != nil {
		return "", httpx.BadRequest("อ่านไฟล์ไม่สำเร็จ").Wrap(err)
	}
	if int64(len(data)) > s.maxBytes {
		return "", errTooLarge(s.maxBytes)
	}

	sum := sha256.Sum256(data)
	checksum := hex.EncodeToString(sum[:])

	// ไฟล์เดิมที่อัปโหลดซ้ำใช้แถวเดิม ไม่เก็บ blob ซ้ำในฐานข้อมูล
	// ต้องเป็น DO UPDATE (ไม่ใช่ DO NOTHING) เพราะ DO NOTHING ไม่คืนแถวที่ชนให้ RETURNING
	const q = `
		INSERT INTO assets (content_type, size_bytes, checksum, data)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (checksum) DO UPDATE SET checksum = EXCLUDED.checksum
		RETURNING id`
	var id uuid.UUID
	if err := s.db.Executor(ctx).QueryRow(ctx, q, contentType, len(data), checksum, data).Scan(&id); err != nil {
		return "", httpx.ErrInternal.Wrap(err)
	}
	return s.URL(id), nil
}

// Serve คือ GET /files/:assetID — เปิดสาธารณะเหมือนไฟล์ใน /uploads
func (s *DBStore) Serve(c *gin.Context) {
	id, err := httpx.ParseUUID(c.Param("assetID"))
	if err != nil {
		httpx.Error(c, err)
		return
	}

	ctx := c.Request.Context()
	var contentType, checksum string
	var data []byte
	err = s.db.Executor(ctx).QueryRow(ctx,
		`SELECT content_type, checksum, data FROM assets WHERE id = $1`, id,
	).Scan(&contentType, &checksum, &data)
	if err != nil {
		httpx.Error(c, access.MapErr(database.NormalizeErr(err)))
		return
	}

	// เนื้อไฟล์ที่ต่างกันได้ id คนละตัวเสมอ (dedup ด้วย checksum) URL หนึ่งจึงชี้ของเดิมตลอดไป
	etag := `"` + checksum + `"`
	c.Header("ETag", etag)
	c.Header("Cache-Control", "public, max-age=31536000, immutable")
	c.Header("X-Content-Type-Options", "nosniff")
	if c.GetHeader("If-None-Match") == etag {
		c.Status(http.StatusNotModified)
		return
	}
	c.Data(http.StatusOK, contentType, data)
}
