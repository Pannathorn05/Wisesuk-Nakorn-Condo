// ที่เก็บ token ฝั่ง client จุดเดียว — ทุกที่ที่ต้องอ่าน/เขียน token ต้องผ่านไฟล์นี้เท่านั้น
// field name (access_token / refresh_token) ตรงกับ docs/openapi.yaml AuthResponse

const ACCESS_TOKEN_KEY = "wisetsuk_access_token";
const REFRESH_TOKEN_KEY = "wisetsuk_refresh_token";

export function getAccessToken() {
  return localStorage.getItem(ACCESS_TOKEN_KEY);
}

export function getRefreshToken() {
  return localStorage.getItem(REFRESH_TOKEN_KEY);
}

export function setTokens({ access_token, refresh_token }) {
  if (access_token) localStorage.setItem(ACCESS_TOKEN_KEY, access_token);
  if (refresh_token) localStorage.setItem(REFRESH_TOKEN_KEY, refresh_token);
}

export function clearTokens() {
  localStorage.removeItem(ACCESS_TOKEN_KEY);
  localStorage.removeItem(REFRESH_TOKEN_KEY);
}

export function isLoggedIn() {
  return Boolean(getAccessToken());
}
