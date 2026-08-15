package storage

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"backend/internal/httpx"
)

// allowedTypes จำกัดชนิดไฟล์ตามที่ prototype ระบุไว้ (JPG / PNG)
var allowedTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

// sniffType ตรวจชนิดไฟล์จากเนื้อไฟล์จริง ไม่เชื่อ Content-Type ที่ client ส่งมา
// แล้ว seek กลับต้นไฟล์ให้ผู้เรียกอ่านเนื้อไฟล์ต่อได้ครบ
func sniffType(file multipart.File) (contentType, ext string, err error) {
	sniff := make([]byte, 512)
	n, err := file.Read(sniff)
	if err != nil && err != io.EOF {
		return "", "", httpx.BadRequest("อ่านไฟล์ไม่สำเร็จ").Wrap(err)
	}
	contentType = http.DetectContentType(sniff[:n])
	ext, ok := allowedTypes[contentType]
	if !ok {
		return "", "", httpx.BadRequest("รองรับเฉพาะไฟล์รูปภาพ JPG, PNG หรือ WEBP เท่านั้น")
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", "", httpx.ErrInternal.Wrap(err)
	}
	return contentType, ext, nil
}

func errTooLarge(maxBytes int64) error {
	return httpx.BadRequest(fmt.Sprintf("ไฟล์ต้องมีขนาดไม่เกิน %d MB", maxBytes/1024/1024))
}

// LocalStore เก็บไฟล์อัปโหลดลงดิสก์ แล้วเสิร์ฟผ่าน /uploads/*
type LocalStore struct {
	baseDir  string
	baseURL  string
	maxBytes int64
}

func NewLocalStore(baseDir, publicBaseURL string, maxBytes int64) (*LocalStore, error) {
	if err := os.MkdirAll(baseDir, 0o750); err != nil {
		return nil, fmt.Errorf("storage: สร้างโฟลเดอร์อัปโหลดไม่สำเร็จ: %w", err)
	}
	return &LocalStore{baseDir: baseDir, baseURL: publicBaseURL, maxBytes: maxBytes}, nil
}

func (s *LocalStore) BaseDir() string { return s.baseDir }
func (s *LocalStore) MaxBytes() int64 { return s.maxBytes }

// Save เขียนไฟล์ลง <baseDir>/<folder>/<uuid><ext> แล้วคืน URL สาธารณะ
// folder ถูกกำหนดจากโค้ดเท่านั้น (ไม่รับจาก client) จึงไม่มีช่องทาง path traversal
func (s *LocalStore) Save(folder string, file multipart.File, header *multipart.FileHeader) (string, error) {
	if header.Size > s.maxBytes {
		return "", errTooLarge(s.maxBytes)
	}

	_, ext, err := sniffType(file)
	if err != nil {
		return "", err
	}

	dir := filepath.Join(s.baseDir, folder, time.Now().Format("2006/01"))
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return "", httpx.ErrInternal.Wrap(err)
	}

	name := uuid.NewString() + ext
	dst, err := os.OpenFile(filepath.Join(dir, name), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o640)
	if err != nil {
		return "", httpx.ErrInternal.Wrap(err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, io.LimitReader(file, s.maxBytes)); err != nil {
		_ = os.Remove(dst.Name())
		return "", httpx.ErrInternal.Wrap(err)
	}

	rel := filepath.ToSlash(filepath.Join(folder, time.Now().Format("2006/01"), name))
	return s.baseURL + "/uploads/" + rel, nil
}

// SaveFromRequest ดึงไฟล์จาก multipart form field เดียวแล้วบันทึก
func (s *LocalStore) SaveFromRequest(c *gin.Context, field, folder string) (string, error) {
	header, err := c.FormFile(field)
	if err != nil {
		return "", httpx.BadRequest("กรุณาแนบไฟล์ในช่อง " + field).Wrap(err)
	}
	file, err := header.Open()
	if err != nil {
		return "", httpx.BadRequest("เปิดไฟล์ที่แนบมาไม่สำเร็จ").Wrap(err)
	}
	defer file.Close()
	return s.Save(folder, file, header)
}
