package httpx

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ---------------------------------------------------------------- error type

// APIError คือ error ที่ปลอดภัยจะส่งกลับให้ client เห็น
type APIError struct {
	Status  int               `json:"-"`
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
	wrapped error
}

func (e *APIError) Error() string {
	if e.wrapped != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.wrapped)
	}
	return e.Message
}

func (e *APIError) Unwrap() error { return e.wrapped }

func (e *APIError) Wrap(err error) *APIError {
	clone := *e
	clone.wrapped = err
	return &clone
}

func NewError(status int, code, message string) *APIError {
	return &APIError{Status: status, Code: code, Message: message}
}

var (
	ErrUnauthorized = NewError(http.StatusUnauthorized, "unauthorized", "กรุณาเข้าสู่ระบบ")
	ErrForbidden    = NewError(http.StatusForbidden, "forbidden", "คุณไม่มีสิทธิ์เข้าถึงส่วนนี้")
	ErrNotFound     = NewError(http.StatusNotFound, "not_found", "ไม่พบข้อมูลที่ต้องการ")
	ErrConflict     = NewError(http.StatusConflict, "conflict", "ข้อมูลนี้มีอยู่ในระบบแล้ว")
	ErrInternal     = NewError(http.StatusInternalServerError, "internal_error", "เกิดข้อผิดพลาดภายในระบบ")
)

func BadRequest(message string) *APIError {
	return NewError(http.StatusBadRequest, "bad_request", message)
}

func ValidationFailed(fields map[string]string) *APIError {
	return &APIError{
		Status:  http.StatusUnprocessableEntity,
		Code:    "validation_failed",
		Message: "ข้อมูลที่กรอกไม่ถูกต้อง",
		Fields:  fields,
	}
}

// ---------------------------------------------------------------- responses

type envelope struct {
	Data any   `json:"data,omitempty"`
	Meta *Meta `json:"meta,omitempty"`
}

type Meta struct {
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	TotalItems int `json:"total_items"`
	TotalPages int `json:"total_pages"`
}

func JSON(c *gin.Context, status int, data any) {
	if data == nil {
		c.Status(status)
		return
	}
	c.JSON(status, data)
}

func OK(c *gin.Context, data any) { c.JSON(http.StatusOK, envelope{Data: data}) }

func Created(c *gin.Context, data any) { c.JSON(http.StatusCreated, envelope{Data: data}) }

func NoContent(c *gin.Context) { c.Status(http.StatusNoContent) }

func Page(c *gin.Context, data any, meta Meta) {
	if meta.PageSize > 0 {
		meta.TotalPages = (meta.TotalItems + meta.PageSize - 1) / meta.PageSize
	}
	c.JSON(http.StatusOK, envelope{Data: data, Meta: &meta})
}

// Error แปลง error ใด ๆ ให้เป็น response ที่เหมาะสม โดยไม่รั่วรายละเอียดภายใน
//
// ใช้ AbortWithStatusJSON เพื่อให้เรียกจาก middleware ได้ด้วย — middleware ที่ตอบ error
// ต้องหยุด chain ไม่ให้ handler ทำงานต่อ ส่วนใน handler การ abort ไม่มีผลเสียเพราะ return อยู่แล้ว
func Error(c *gin.Context, err error) {
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		apiErr = ErrInternal
	}
	if apiErr.Status >= 500 {
		slog.Error("request failed",
			"method", c.Request.Method, "path", c.Request.URL.Path, "error", err.Error())
		c.AbortWithStatusJSON(apiErr.Status, gin.H{"error": ErrInternal})
		return
	}
	c.AbortWithStatusJSON(apiErr.Status, gin.H{"error": apiErr})
}

// ---------------------------------------------------------------- request parsing

const maxJSONBody = 1 << 20 // 1MB

// DecodeJSON อ่าน body เป็น JSON แบบเข้มงวด
//
// ไม่ใช้ c.ShouldBindJSON ของ gin เพราะต้องการ DisallowUnknownFields (กัน client
// ส่งฟิลด์ที่พิมพ์ผิดมาแล้วเงียบ ๆ ไม่มีผล) และต้องการจำกัดขนาด body ตั้งแต่ระดับ reader
func DecodeJSON(c *gin.Context, dst any) error {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxJSONBody)
	dec := json.NewDecoder(c.Request.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return BadRequest("รูปแบบข้อมูล JSON ไม่ถูกต้อง").Wrap(err)
	}
	if dec.More() {
		return BadRequest("body ต้องมี JSON object เพียงชุดเดียว")
	}
	return nil
}

func ParseUUID(raw string) (uuid.UUID, error) {
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, BadRequest("รหัสอ้างอิงไม่ถูกต้อง").Wrap(err)
	}
	return id, nil
}

// Pagination อ่าน ?page= &page_size= พร้อมค่าเริ่มต้นและเพดานที่ปลอดภัย
func Pagination(c *gin.Context) (page, pageSize, offset int) {
	page = 1
	pageSize = 20
	if v := c.Query("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			page = n
		}
	}
	if v := c.Query("page_size"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			pageSize = min(n, 100)
		}
	}
	return page, pageSize, (page - 1) * pageSize
}
