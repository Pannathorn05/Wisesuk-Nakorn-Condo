package storage

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"backend/internal/httpx"
)

// ParseMultipart อ่าน multipart form พร้อมจำกัดขนาด body
//
// เผื่อ overhead ของ multipart boundary ไว้ 1MB นอกเหนือจากเพดานขนาดไฟล์
// เพื่อให้ไฟล์ที่ขนาดพอดีเพดานยังอัปโหลดผ่าน
func (s *LocalStore) ParseMultipart(c *gin.Context) error {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, s.maxBytes+1<<20)
	if err := c.Request.ParseMultipartForm(s.maxBytes); err != nil {
		return httpx.BadRequest("ไฟล์มีขนาดใหญ่เกินกำหนด หรือรูปแบบฟอร์มไม่ถูกต้อง").Wrap(err)
	}
	return nil
}
