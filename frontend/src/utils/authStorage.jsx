// ที่เก็บ token ฝั่ง client จุดเดียว — ทุกที่ที่ต้องอ่าน/เขียน token ต้องผ่านไฟล์นี้เท่านั้น
// field name (access_token / refresh_token) ตรงกับ docs/openapi.yaml AuthResponse

const ACCESS_TOKEN_KEY = "wisetsuk_access_token";
const REFRESH_TOKEN_KEY = "wisetsuk_refresh_token";
// เก็บ user ที่ login อยู่ (มากับ AuthResult ของ POST /auth/login|/auth/register) ไว้ด้วย เพราะ
// role ต้องใช้ตัดสินใจ redirect/กัน route ของโซนผู้ดูแล และต้องรอดจากการ reload หน้า
// (docs/openapi.yaml UserRole เขียนกำกับเองว่า frontend ใช้ค่านี้เลือกปลายทางหลังล็อกอิน)
const USER_KEY = "wisetsuk_user";

// event ภายในแท็บเดียวกัน — component อื่น (เช่น Header, FE-24) ที่ต้องรู้ทันทีว่า login/logout
// เปลี่ยนสถานะ ต้อง subscribe ผ่าน onAuthChange แทนที่จะ poll localStorage เอง (window "storage"
// event ของเบราว์เซอร์ยิงเฉพาะข้ามแท็บ ไม่ยิงในแท็บที่เป็นคนเปลี่ยนค่าเอง จึงต้องมี event ของเราเองด้วย)
const AUTH_EVENT = "wisetsuk_auth_changed";

export function getAccessToken() {
  return localStorage.getItem(ACCESS_TOKEN_KEY);
}

export function getRefreshToken() {
  return localStorage.getItem(REFRESH_TOKEN_KEY);
}

// รับ AuthResult ทั้งก้อนได้เลย (มี user มาด้วยก็เก็บให้ ไม่มีก็ข้าม)
export function setTokens({ access_token, refresh_token, user }) {
  if (access_token) localStorage.setItem(ACCESS_TOKEN_KEY, access_token);
  if (refresh_token) localStorage.setItem(REFRESH_TOKEN_KEY, refresh_token);
  if (user) localStorage.setItem(USER_KEY, JSON.stringify(user));
  window.dispatchEvent(new Event(AUTH_EVENT));
}

export function clearTokens() {
  localStorage.removeItem(ACCESS_TOKEN_KEY);
  localStorage.removeItem(REFRESH_TOKEN_KEY);
  localStorage.removeItem(USER_KEY);
  window.dispatchEvent(new Event(AUTH_EVENT));
}

export function getCurrentUser() {
  const raw = localStorage.getItem(USER_KEY);
  if (!raw) return null;
  try {
    return JSON.parse(raw);
  } catch {
    // ข้อมูลเสีย (เช่นถูกแก้มือใน devtools) — ถือว่าไม่มี user ดีกว่าให้ทั้งแอปพัง
    return null;
  }
}

export function getCurrentRole() {
  return getCurrentUser()?.role || null;
}

export function isLoggedIn() {
  return Boolean(getAccessToken());
}

// คืนฟังก์ชันยกเลิก subscribe (ใช้ใน useEffect cleanup) — ฟัง "storage" ด้วยเพื่อ sync ข้ามแท็บ
export function onAuthChange(callback) {
  window.addEventListener(AUTH_EVENT, callback);
  window.addEventListener("storage", callback);
  return () => {
    window.removeEventListener(AUTH_EVENT, callback);
    window.removeEventListener("storage", callback);
  };
}
