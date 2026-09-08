import { apiClient } from "./client";

// POST /api/v1/auth/register -> AuthResult { access_token, refresh_token, user }
// body ต้องมีแค่ email/password/first_name/last_name/phone (additionalProperties: false — ห้ามส่ง role)
export function register(body) {
  return apiClient.post("/auth/register", body);
}

// POST /api/v1/auth/login -> AuthResult เหมือนกัน
export function login(body) {
  return apiClient.post("/auth/login", body);
}

// POST /api/v1/auth/logout -> 204 เสมอ (docs/openapi.yaml: "ตอบ 204 เสมอ ไม่ว่า token จะใช้ได้หรือไม่")
export function logout(refreshToken) {
  return apiClient.post("/auth/logout", { refresh_token: refreshToken });
}
