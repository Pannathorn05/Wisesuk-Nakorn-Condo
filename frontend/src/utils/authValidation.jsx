// กฎ validate ฝั่ง client mirror ตรงกับ backend (docs/openapi.yaml RegisterRequest/LoginRequest)
// เช็คก่อนยิง request เพื่อลดจำนวน round-trip ที่ต้อง fail จริง — backend ยังเป็นคนตัดสินสุดท้ายเสมอ
// (422 validation_failed พร้อม error.fields ต้องแสดงผลได้ด้วยไม่ว่าจะเช็คตรงนี้พลาดจุดไหนไป)
const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
const PHONE_RE = /^[0-9]{9,10}$/;

export function validateLogin({ email, password }) {
  const errors = {};
  if (!email.trim()) errors.email = "กรุณากรอกอีเมล";
  else if (!EMAIL_RE.test(email)) errors.email = "รูปแบบอีเมลไม่ถูกต้อง";
  if (!password) errors.password = "กรุณากรอกรหัสผ่าน";
  return errors;
}

// คีย์ของ errors ที่คืนเป็น snake_case ตรงกับชื่อ field ของ backend เป๊ะ (first_name/last_name/phone/
// email/password) ยกเว้น confirm_password ที่เป็นกฎฝั่ง client ล้วน (ไม่มีใน RegisterRequest) —
// ตั้งใจให้ตรงกันเพื่อ merge กับ error.fields จาก response 422 ได้ตรง ๆ โดยไม่ต้องแปลงชื่อคีย์ไปมา
export function validateRegister({ first_name, last_name, phone, email, password, confirm_password }) {
  const errors = {};
  if (!first_name.trim()) errors.first_name = "กรุณากรอกชื่อ";
  if (!last_name.trim()) errors.last_name = "กรุณากรอกนามสกุล";
  if (!PHONE_RE.test(phone)) errors.phone = "เบอร์โทรศัพท์ต้องเป็นตัวเลข 9-10 หลัก";
  if (!email.trim()) errors.email = "กรุณากรอกอีเมล";
  else if (!EMAIL_RE.test(email)) errors.email = "รูปแบบอีเมลไม่ถูกต้อง";
  if (password.length < 8 || password.length > 72) errors.password = "รหัสผ่านต้องมีความยาว 8-72 ตัวอักษร";
  if (confirm_password !== password) errors.confirm_password = "รหัสผ่านไม่ตรงกัน";
  return errors;
}
