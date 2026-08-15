package httpx

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

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

// รหัส error ทั้ง 12 ตัวที่ SPEC อนุญาต ห้ามมีรหัสอื่นหลุดออกไปหา client
const (
	CodeUnauthorized       = "unauthorized"
	CodeForbidden          = "forbidden"
	CodeNotFound           = "not_found"
	CodeConflict           = "conflict"
	CodeBadRequest         = "bad_request"
	CodeValidationFailed   = "validation_failed"
	CodeMethodNotAllowed   = "method_not_allowed"
	CodeInternalError      = "internal_error"
	CodeInvalidCredentials = "invalid_credentials"
	CodeAccountDisabled    = "account_disabled"
	CodeRoomUnavailable    = "room_unavailable"
	CodeInvalidState       = "invalid_state"
)

// ข้อความภาษาไทยที่ผู้ใช้เห็นของ error กลางทุกตัวอยู่ที่นี่ที่เดียว
// ข้อความเฉพาะเรื่อง (เช่น เหตุผลที่เปลี่ยนสถานะการจองไม่ได้) ประกาศไว้ต้น service ของ module นั้น
const (
	msgUnauthorized       = "กรุณาเข้าสู่ระบบ"
	msgForbidden          = "คุณไม่มีสิทธิ์เข้าถึงส่วนนี้"
	msgNotFound           = "ไม่พบข้อมูลที่ต้องการ"
	msgConflict           = "ข้อมูลนี้มีอยู่ในระบบแล้ว"
	msgInternal           = "เกิดข้อผิดพลาดภายในระบบ"
	msgMethodNotAllowed   = "ไม่รองรับ HTTP method นี้"
	msgInvalidCredentials = "อีเมลหรือรหัสผ่านไม่ถูกต้อง"
	msgAccountDisabled    = "บัญชีนี้ถูกระงับการใช้งาน กรุณาติดต่อผู้ดูแลระบบ"
	msgRoomUnavailable    = "ขออภัย ห้องนี้ถูกจองไปแล้วในช่วงวันที่ที่เลือก กรุณาเลือกห้องหรือวันที่อื่น"
	msgValidationFailed   = "ข้อมูลที่กรอกไม่ถูกต้อง"
	msgInvalidJSON        = "รูปแบบข้อมูล JSON ไม่ถูกต้อง"
	msgSingleJSONObject   = "body ต้องมี JSON object เพียงชุดเดียว"
	msgFieldNotAllowed    = "ไม่อนุญาตให้ส่งฟิลด์นี้"
	msgInvalidID          = "รหัสอ้างอิงไม่ถูกต้อง"
)

var (
	ErrUnauthorized = NewError(http.StatusUnauthorized, CodeUnauthorized, msgUnauthorized)
	ErrForbidden    = NewError(http.StatusForbidden, CodeForbidden, msgForbidden)
	ErrNotFound     = NewError(http.StatusNotFound, CodeNotFound, msgNotFound)
	ErrConflict     = NewError(http.StatusConflict, CodeConflict, msgConflict)
	ErrInternal     = NewError(http.StatusInternalServerError, CodeInternalError, msgInternal)

	ErrMethodNotAllowed   = NewError(http.StatusMethodNotAllowed, CodeMethodNotAllowed, msgMethodNotAllowed)
	ErrInvalidCredentials = NewError(http.StatusUnauthorized, CodeInvalidCredentials, msgInvalidCredentials)
	ErrAccountDisabled    = NewError(http.StatusForbidden, CodeAccountDisabled, msgAccountDisabled)
	ErrRoomUnavailable    = NewError(http.StatusConflict, CodeRoomUnavailable, msgRoomUnavailable)
)

func BadRequest(message string) *APIError {
	return NewError(http.StatusBadRequest, CodeBadRequest, message)
}

// InvalidState ใช้เมื่อสถานะปัจจุบันของข้อมูลทำรายการที่ขอไม่ได้
// รับข้อความมาเพราะเหตุผลต่างกันไปตามเรื่อง ผู้ใช้ต้องรู้ว่าติดตรงไหน
func InvalidState(message string) *APIError {
	return NewError(http.StatusConflict, CodeInvalidState, message)
}

func ValidationFailed(fields map[string]string) *APIError {
	return &APIError{
		Status:  http.StatusUnprocessableEntity,
		Code:    CodeValidationFailed,
		Message: msgValidationFailed,
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
		// ฟิลด์ที่ไม่รู้จักคือ "ค่าถูกชนิดแต่ผิดกติกา" ไม่ใช่ body พัง จึงเป็น 422 ตาม AC-21
		// และต้องบอกชื่อฟิลด์กลับไปใน fields เพื่อให้ client รู้ว่าตัวไหนต้องห้าม (AC-1)
		if name, ok := unknownFieldName(err); ok {
			return ValidationFailed(map[string]string{name: msgFieldNotAllowed})
		}
		return BadRequest(msgInvalidJSON).Wrap(err)
	}
	if dec.More() {
		return BadRequest(msgSingleJSONObject)
	}
	return nil
}

// unknownFieldName ดึงชื่อฟิลด์ออกจาก error ของ encoding/json
//
// encoding/json ไม่มี error type สำหรับกรณีนี้ มีแต่ข้อความ จึงต้องอ่านจากข้อความตรง ๆ
// รูปแบบคือ: json: unknown field "role"
func unknownFieldName(err error) (string, bool) {
	const prefix = `json: unknown field `
	msg := err.Error()
	if !strings.HasPrefix(msg, prefix) {
		return "", false
	}
	name := strings.Trim(strings.TrimPrefix(msg, prefix), `"`)
	if name == "" {
		return "", false
	}
	return name, true
}

func ParseUUID(raw string) (uuid.UUID, error) {
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, BadRequest(msgInvalidID).Wrap(err)
	}
	return id, nil
}

const (
	defaultPage     = 1
	defaultPageSize = 20
	maxPageSize     = 100
)

// Pagination อ่าน ?page= &page_size= พร้อมค่าเริ่มต้นและเพดานที่ปลอดภัย
//
// ค่าที่พาร์สไม่ได้หรือไม่เป็นจำนวนเต็มบวกคือ 400 ไม่ใช่การเงียบ ๆ ใช้ค่าเริ่มต้น
// เพราะการกลืน input ที่ผิดทำให้ client เข้าใจผิดว่าคำขอถูกต้อง (AC-21)
func Pagination(c *gin.Context) (page, pageSize, offset int, err error) {
	if page, err = positiveInt(c, "page", defaultPage); err != nil {
		return 0, 0, 0, err
	}
	if pageSize, err = positiveInt(c, "page_size", defaultPageSize); err != nil {
		return 0, 0, 0, err
	}
	pageSize = min(pageSize, maxPageSize)
	return page, pageSize, (page - 1) * pageSize, nil
}

func positiveInt(c *gin.Context, key string, fallback int) (int, error) {
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" {
		return fallback, nil
	}
	n, convErr := strconv.Atoi(raw)
	if convErr != nil || n < 1 {
		return 0, BadRequest("พารามิเตอร์ " + key + " ต้องเป็นจำนวนเต็มบวก")
	}
	return n, nil
}
