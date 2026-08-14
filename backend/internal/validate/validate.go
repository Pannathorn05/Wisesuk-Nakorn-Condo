package validate

import (
	"net/url"
	"regexp"
	"strings"
	"unicode"

	"backend/internal/httpx"
)

var (
	emailRe = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	phoneRe = regexp.MustCompile(`^[0-9\-\s()+]{8,20}$`)
)

// Validator เก็บ error รายฟิลด์ แล้วแปลงเป็น 422 response ทีเดียว
type Validator struct {
	fields map[string]string
}

func New() *Validator { return &Validator{fields: map[string]string{}} }

func (v *Validator) add(field, msg string) {
	if _, exists := v.fields[field]; !exists {
		v.fields[field] = msg
	}
}

func (v *Validator) Check(ok bool, field, msg string) {
	if !ok {
		v.add(field, msg)
	}
}

func (v *Validator) Required(field, value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		v.add(field, "กรุณากรอกข้อมูลนี้")
	}
	return value
}

func (v *Validator) MaxLen(field, value string, max int) {
	if len([]rune(value)) > max {
		v.add(field, "ความยาวต้องไม่เกิน "+itoa(max)+" ตัวอักษร")
	}
}

func (v *Validator) Email(field, value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		v.add(field, "กรุณากรอกอีเมล")
	} else if !emailRe.MatchString(value) {
		v.add(field, "รูปแบบอีเมลไม่ถูกต้อง")
	}
	return value
}

func (v *Validator) Phone(field, value string, required bool) string {
	value = strings.TrimSpace(value)
	if value == "" {
		if required {
			v.add(field, "กรุณากรอกเบอร์โทรศัพท์")
		}
		return value
	}
	if !phoneRe.MatchString(value) {
		v.add(field, "รูปแบบเบอร์โทรศัพท์ไม่ถูกต้อง")
	}
	return value
}

// ImageURL ตรวจ URL รูปภาพที่ client ส่งมา (รูปถูกโฮสต์ไว้ที่บริการภายนอก)
//
// รับเฉพาะ http/https ที่มี host จริง เพื่อกัน javascript:, data: และ path เปล่า ๆ
// ที่จะกลายเป็น XSS เมื่อฝั่งเว็บเอาไปใส่ใน <img src> หรือ <a href>
func (v *Validator) ImageURL(field, value string, required bool) string {
	value = strings.TrimSpace(value)
	if value == "" {
		if required {
			v.add(field, "กรุณาระบุลิงก์รูปภาพ")
		}
		return value
	}
	if len(value) > 2048 {
		v.add(field, "ลิงก์ยาวเกินไป")
		return value
	}
	u, err := url.Parse(value)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		v.add(field, "ลิงก์รูปภาพต้องขึ้นต้นด้วย http:// หรือ https://")
	}
	return value
}

// Password บังคับความยาว 8+ และต้องมีทั้งตัวอักษรและตัวเลข
func (v *Validator) Password(field, value string) {
	if len(value) < 8 {
		v.add(field, "รหัสผ่านต้องมีอย่างน้อย 8 ตัวอักษร")
		return
	}
	if len(value) > 72 { // เพดานของ bcrypt
		v.add(field, "รหัสผ่านต้องยาวไม่เกิน 72 ตัวอักษร")
		return
	}
	var hasLetter, hasDigit bool
	for _, r := range value {
		switch {
		case unicode.IsLetter(r):
			hasLetter = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
	}
	if !hasLetter || !hasDigit {
		v.add(field, "รหัสผ่านต้องมีทั้งตัวอักษรและตัวเลข")
	}
}

func (v *Validator) Positive(field string, value float64) {
	if value <= 0 {
		v.add(field, "ต้องเป็นตัวเลขมากกว่า 0")
	}
}

func (v *Validator) OK() bool { return len(v.fields) == 0 }

// Err คืน *httpx.APIError เมื่อมี error มิฉะนั้นคืน nil
func (v *Validator) Err() error {
	if v.OK() {
		return nil
	}
	return httpx.ValidationFailed(v.fields)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
