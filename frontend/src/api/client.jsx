import { getAccessToken } from "../utils/authStorage";

// จุดเดียวที่รู้จัก base URL ของ backend — ห้ามมีการ fetch/axios ตรงในหน้า component
// (docs/c.md ข้อ 8, .claude/CLAUDE.md ส่วน Architecture)
const BASE_URL = (process.env.REACT_APP_API_BASE_URL || "http://localhost:8080/api/v1").replace(/\/+$/, "");

// backend เสิร์ฟไฟล์ static (GET /uploads/*filepath, GET /files/:assetID) ที่ root ของ server
// ไม่ได้อยู่ใต้ /api/v1 — ตัด /api/v1 ออกจาก BASE_URL เอาไว้ให้ที่อื่นประกอบ URL รูป static ได้
// โดยไม่ต้องรู้จัก host ของ backend เอง (ดู src/assets/branchPhotos.jsx)
export const API_ORIGIN = BASE_URL.replace(/\/api\/v1\/?$/, "");

// ตารางแปล error.code (docs/openapi.yaml ErrorCode) เป็นข้อความไทยสำรอง
// ปกติ backend ส่ง error.message เป็นภาษาไทยมาให้อยู่แล้ว ตารางนี้ใช้เฉพาะตอน message ไม่มา
const ERROR_CODE_FALLBACK = {
  unauthorized: "กรุณาเข้าสู่ระบบก่อนใช้งานส่วนนี้",
  forbidden: "ไม่มีสิทธิ์เข้าถึงข้อมูลนี้",
  not_found: "ไม่พบข้อมูลที่ต้องการ",
  conflict: "ข้อมูลถูกใช้งานไปแล้ว",
  bad_request: "คำขอไม่ถูกต้อง",
  validation_failed: "ข้อมูลไม่ถูกต้อง กรุณาตรวจสอบอีกครั้ง",
  method_not_allowed: "ไม่รองรับคำขอนี้",
  internal_error: "เกิดข้อผิดพลาดที่เซิร์ฟเวอร์ กรุณาลองใหม่",
  invalid_credentials: "อีเมลหรือรหัสผ่านไม่ถูกต้อง",
  account_disabled: "บัญชีนี้ถูกระงับการใช้งาน",
  room_unavailable: "ห้องนี้ไม่ว่างแล้ว",
  invalid_state: "สถานะปัจจุบันทำรายการนี้ไม่ได้",
};

export class ApiError extends Error {
  constructor({ code, message, fields, status }) {
    super(message || ERROR_CODE_FALLBACK[code] || "เกิดข้อผิดพลาด กรุณาลองใหม่");
    this.name = "ApiError";
    this.code = code;
    this.fields = fields || null;
    this.status = status;
  }
}

// error ที่ยิงไม่ถึง backend เลย (เน็ตหลุด, backend ปิดอยู่, CORS ฯลฯ)
export class NetworkError extends Error {
  constructor(cause) {
    super("เชื่อมต่อเซิร์ฟเวอร์ไม่ได้ กรุณาตรวจสอบอินเทอร์เน็ตแล้วลองใหม่");
    this.name = "NetworkError";
    this.cause = cause;
  }
}

function buildUrl(path, params) {
  const url = new URL(BASE_URL + path);
  if (params) {
    Object.entries(params).forEach(([key, value]) => {
      if (value === undefined || value === null || value === "") return;
      url.searchParams.set(key, value);
    });
  }
  return url.toString();
}

async function request(path, { method = "GET", params, body, signal } = {}) {
  const headers = { Accept: "application/json" };
  if (body !== undefined) headers["Content-Type"] = "application/json";

  const token = getAccessToken();
  if (token) headers.Authorization = `Bearer ${token}`;

  let res;
  try {
    res = await fetch(buildUrl(path, params), {
      method,
      headers,
      body: body !== undefined ? JSON.stringify(body) : undefined,
      signal,
    });
  } catch (cause) {
    if (cause?.name === "AbortError") throw cause;
    throw new NetworkError(cause);
  }

  if (res.status === 204) return null;

  const text = await res.text();
  const json = text ? JSON.parse(text) : null;

  if (!res.ok) {
    const err = json?.error || {};
    throw new ApiError({ code: err.code, message: err.message, fields: err.fields, status: res.status });
  }

  return json;
}

export const apiClient = {
  get: (path, opts) => request(path, { ...opts, method: "GET" }),
  post: (path, body, opts) => request(path, { ...opts, method: "POST", body }),
  put: (path, body, opts) => request(path, { ...opts, method: "PUT", body }),
  patch: (path, body, opts) => request(path, { ...opts, method: "PATCH", body }),
  delete: (path, opts) => request(path, { ...opts, method: "DELETE" }),
};
