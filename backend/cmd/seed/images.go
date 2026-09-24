package main

import (
	"context"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"

	"backend/internal/database"
	"backend/internal/shared/types"
)

// รูปเดโมของแต่ละสาขาอยู่ใน uploads/<slug>/ ของ repo และถูกเสิร์ฟผ่าน route /uploads/*
// ที่มีอยู่แล้ว seed จึงเขียนแค่ URL ลงฐานข้อมูล ไม่ต้องอัปโหลดเป็น blob เข้าตาราง assets
//
// ผลคือรูปเดโมรอดจาก docker compose down -v เพราะไฟล์อยู่ใน repo ไม่ใช่ใน volume ของฐานข้อมูล
// แลกกับว่าต้องรัน seed ใหม่เมื่อเปลี่ยน PUBLIC_BASE_URL เพราะ URL ถูกเก็บแบบเต็ม
//
// seed แยกรูปของตัวเองออกจากรูปที่แอดมินอัปเองด้วย prefix ของ URL:
// รูปเดโมขึ้นต้นด้วย {baseURL}/uploads/ ส่วนรูปที่อัปผ่าน API ขึ้นต้นด้วย {baseURL}/files/
// จึงลบหรือทับได้เฉพาะของตัวเอง ไม่ไปแตะรูปที่แอดมินตั้งไว้บนระบบที่ใช้งานจริง

// imageExts คือชนิดไฟล์ที่นับเป็นรูป ตรงกับที่ API รับอัปโหลด (JPG / PNG / WEBP)
var imageExts = map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true}

// captionFile คือไฟล์คำบรรยายของทุกรูปในโฟลเดอร์เดียวกัน (UTF-8)
const captionFile = "caption.txt"

// demoImage คือรูปหนึ่งไฟล์ใน uploads/<slug>/
type demoImage struct {
	relPath string // path สัมพัทธ์กับ uploads/<slug>/ คั่นด้วย / เสมอ ไม่ว่ารันบน OS ไหน
	caption string
}

// demoImages บอกว่าจะอ่านรูปเดโมจากโฟลเดอร์ไหน และจะประกอบ URL ด้วย origin อะไร
type demoImages struct {
	dir     string
	baseURL string
}

// seedBranch แทนที่รูปเดโมในแกลเลอรีของสาขาด้วยไฟล์ที่มีอยู่ตอนนี้
// ไฟล์ที่ถูกลบออกจาก repo จะหายจากแกลเลอรีด้วย ส่วนรูปที่แอดมินอัปเองอยู่ครบ
func (d demoImages) seedBranch(ctx context.Context, db database.Executor, branchID types.BranchID, slug string) error {
	images, err := scanBranchImages(d.dir, slug)
	if err != nil {
		return err
	}

	prefix := d.baseURL + "/uploads/" + slug + "/"
	if _, err := db.Exec(ctx,
		`DELETE FROM branch_images WHERE branch_id = $1 AND starts_with(image_url, $2)`,
		branchID, prefix); err != nil {
		return fmt.Errorf("ล้างรูปเดโมสาขา %s: %w", slug, err)
	}

	for i, img := range images {
		if _, err := db.Exec(ctx,
			`INSERT INTO branch_images (branch_id, image_url, caption, sort_order) VALUES ($1, $2, $3, $4)`,
			branchID, uploadURL(d.baseURL, slug, img.relPath), img.caption, i); err != nil {
			return fmt.Errorf("เพิ่มรูป %s/%s: %w", slug, img.relPath, err)
		}
	}
	return nil
}

// setRoomCover ตั้งรูปปกห้องจากไฟล์ใน uploads/<slug>/<relPath>
//
// ทับได้เฉพาะห้องที่ยังไม่มีรูปปก หรือรูปปกเดิมเป็นรูปเดโมเหมือนกัน
// รูปที่แอดมินอัปเองผ่านหน้าจัดการห้องพักไม่ถูกแตะ
func (d demoImages) setRoomCover(ctx context.Context, db database.Executor, roomID types.RoomID, slug, relPath string) error {
	file := filepath.Join(d.dir, slug, filepath.FromSlash(relPath))
	if _, err := os.Stat(file); err != nil {
		return fmt.Errorf("รูปปกห้อง %s: %w", file, err)
	}

	_, err := db.Exec(ctx,
		`UPDATE rooms SET image_url = $2, updated_at = now()
		 WHERE id = $1 AND (image_url = '' OR starts_with(image_url, $3))`,
		roomID, uploadURL(d.baseURL, slug, relPath), d.baseURL+"/uploads/")
	return err
}

// scanBranchImages คืนไฟล์รูปทั้งหมดใต้ uploads/<slug>/ เรียงตาม path
// สาขาที่ไม่มีโฟลเดอร์คืนรายการว่าง ไม่ใช่ error
func scanBranchImages(uploadDir, slug string) ([]demoImage, error) {
	root := filepath.Join(uploadDir, slug)
	if _, err := os.Stat(root); os.IsNotExist(err) {
		return nil, nil
	}

	captions := map[string]string{} // โฟลเดอร์ -> คำบรรยาย อ่านไฟล์ละครั้ง
	var out []demoImage

	err := filepath.WalkDir(root, func(p string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		if !imageExts[strings.ToLower(filepath.Ext(p))] {
			return nil
		}

		dir := filepath.Dir(p)
		caption, seen := captions[dir]
		if !seen {
			if caption, err = readCaption(dir); err != nil {
				return err
			}
			captions[dir] = caption
		}

		rel, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}
		out = append(out, demoImage{relPath: filepath.ToSlash(rel), caption: caption})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("อ่านรูปสาขา %s: %w", slug, err)
	}

	sort.Slice(out, func(i, j int) bool { return out[i].relPath < out[j].relPath })
	return out, nil
}

// readCaption อ่าน caption.txt ในโฟลเดอร์ ไม่มีไฟล์ = ไม่มีคำบรรยาย
func readCaption(dir string) (string, error) {
	raw, err := os.ReadFile(filepath.Join(dir, captionFile))
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	// ข้อความที่ไม่ใช่ UTF-8 (เช่นบันทึกจาก Notepad เป็น ANSI/CP874) จะไปพังตอน INSERT
	// บอกชื่อไฟล์ตรงนี้เลยแก้ง่ายกว่า
	if !utf8.Valid(raw) {
		return "", fmt.Errorf("%s ต้องบันทึกเป็น UTF-8", filepath.Join(dir, captionFile))
	}
	text := strings.TrimPrefix(string(raw), string([]byte{0xEF, 0xBB, 0xBF})) // BOM ที่ Notepad ชอบใส่มา
	return strings.TrimSpace(text), nil
}

// uploadURL ประกอบ URL สาธารณะของไฟล์ใน uploads/<slug>/<relPath>
// escape ทีละท่อน เพราะชื่อไฟล์จริงมีช่องว่างและวงเล็บ เช่น "m-45-3 (1).jpg"
func uploadURL(baseURL, slug, relPath string) string {
	parts := append([]string{slug}, strings.Split(relPath, "/")...)
	for i, p := range parts {
		parts[i] = url.PathEscape(p)
	}
	return baseURL + path.Join("/uploads", strings.Join(parts, "/"))
}
