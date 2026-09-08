// ที่เก็บ token ฝั่ง client จุดเดียว — ทุกที่ที่ต้องอ่าน/เขียน token ต้องผ่านไฟล์นี้เท่านั้น
// field name (access_token / refresh_token) ตรงกับ docs/openapi.yaml AuthResponse

const ACCESS_TOKEN_KEY = "wisetsuk_access_token";
const REFRESH_TOKEN_KEY = "wisetsuk_refresh_token";

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

export function setTokens({ access_token, refresh_token }) {
  if (access_token) localStorage.setItem(ACCESS_TOKEN_KEY, access_token);
  if (refresh_token) localStorage.setItem(REFRESH_TOKEN_KEY, refresh_token);
  window.dispatchEvent(new Event(AUTH_EVENT));
}

export function clearTokens() {
  localStorage.removeItem(ACCESS_TOKEN_KEY);
  localStorage.removeItem(REFRESH_TOKEN_KEY);
  window.dispatchEvent(new Event(AUTH_EVENT));
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
