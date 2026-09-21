import { apiClient } from "./client";

// GET /api/v1/superadmin/staff -> { data: User[] } (รวมตัว superadmin เองด้วย)
export function listStaff(signal) {
  return apiClient.get("/superadmin/staff", { signal });
}

// GET /api/v1/superadmin/branches -> { data: Branch[] } รวมสาขาที่ปิดใช้งาน
// ใช้ทั้งแปลง branch_id -> ชื่อสาขาในตาราง และเติมตัวเลือกสาขาใน modal
export function listAllBranches(signal) {
  return apiClient.get("/superadmin/branches", { signal });
}

// POST /api/v1/superadmin/staff -> 201 User
// body: CreateStaffRequest {email, first_name, last_name, phone, branch_id} — ทั้ง 5 ช่อง required
// ไม่มีรหัสผ่าน: backend ตั้งรหัสตั้งต้นจาก env ให้เอง + must_change_password = true
export function createStaff(body) {
  return apiClient.post("/superadmin/staff", body);
}

// PUT /api/v1/superadmin/staff/{userID} -> 200 User
// body: UpdateStaffRequest ทุก field optional — branch_id ห้ามส่งเมื่อเป้าหมายเป็น superadmin
export function updateStaff(userId, body) {
  return apiClient.put(`/superadmin/staff/${userId}`, body);
}

// DELETE /api/v1/superadmin/staff/{userID} -> 204 (soft delete) · 400 ถ้าลบบัญชีตัวเอง
export function deleteStaff(userId) {
  return apiClient.delete(`/superadmin/staff/${userId}`);
}
