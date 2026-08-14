package httpx

import (
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ตัวช่วยอ่าน query string ที่ทุก module ใช้ร่วมกัน
// ทุกตัวคืน nil เมื่อไม่ได้ส่งค่ามา และคืน 400 พร้อมข้อความไทยเมื่อรูปแบบผิด

func QueryUUID(c *gin.Context, key string) (*uuid.UUID, error) {
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" {
		return nil, nil
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return nil, BadRequest("พารามิเตอร์ " + key + " ไม่ถูกต้อง").Wrap(err)
	}
	return &id, nil
}

func QueryDate(c *gin.Context, key string) (*time.Time, error) {
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return nil, BadRequest("พารามิเตอร์ " + key + " ต้องอยู่ในรูปแบบ YYYY-MM-DD").Wrap(err)
	}
	return &t, nil
}

func QueryFloat(c *gin.Context, key string) (*float64, error) {
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" {
		return nil, nil
	}
	f, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return nil, BadRequest("พารามิเตอร์ " + key + " ต้องเป็นตัวเลข").Wrap(err)
	}
	return &f, nil
}

func QueryString(c *gin.Context, key string) string {
	return strings.TrimSpace(c.Query(key))
}

// FormValue อ่านค่าข้อความจาก multipart form พร้อมตรวจว่าเป็น UTF-8 ที่ถูกต้อง
//
// ต่างจาก JSON ตรงที่ encoding/json แปลงไบต์เสียเป็น U+FFFD ให้อยู่แล้ว
// แต่ multipart ส่งไบต์ดิบมาตรง ๆ ถ้าปล่อยผ่านจะไปตายที่ PostgreSQL
// (invalid byte sequence for encoding UTF8) แล้วกลายเป็น 500 ทั้งที่เป็นความผิดของ client
func FormValue(c *gin.Context, key string) (string, error) {
	v := strings.TrimSpace(c.PostForm(key))
	if !utf8.ValidString(v) {
		return "", BadRequest("ข้อความในช่อง " + key + " ไม่ใช่ข้อความ UTF-8 ที่ถูกต้อง")
	}
	return v, nil
}

// ParseFlexibleTime รับได้ทั้ง RFC3339 และรูปแบบที่ฟอร์มฝั่งเว็บส่งมาตามปกติ
//
// รูปแบบที่ไม่มี timezone (เช่น "2026-08-06 14:30" ที่ผู้ใช้กรอกในฟอร์มแจ้งโอนเงิน)
// ถูกตีความเป็น "เวลาท้องถิ่นของเซิร์ฟเวอร์" ไม่ใช่ UTC — เพราะผู้ใช้กรอกเวลาไทย
// ถ้าใช้ time.Parse ตรง ๆ Go จะถือว่าเป็น UTC ทำให้เวลาบ่ายกลายเป็นเวลาอนาคต
// แล้วโดน validation ตีกลับทั้งที่ผู้ใช้กรอกถูก
//
// (container ตั้ง TZ=Asia/Bangkok ไว้แล้วใน Dockerfile จึงตรงกับผู้ใช้)
func ParseFlexibleTime(raw string) (time.Time, error) {
	raw = strings.TrimSpace(raw)

	// มี timezone มาด้วย — เชื่อค่าที่ client ส่งมา
	if t, err := time.Parse(time.RFC3339, raw); err == nil {
		return t, nil
	}

	localLayouts := []string{
		"2006-01-02T15:04:05",
		"2006-01-02T15:04",
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02",
	}
	var lastErr error
	for _, layout := range localLayouts {
		t, err := time.ParseInLocation(layout, raw, time.Local)
		if err == nil {
			return t, nil
		}
		lastErr = err
	}
	return time.Time{}, lastErr
}
