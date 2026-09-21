import { apiClient } from "./client";

// GET /api/v1/admin/dashboard -> { data: { branches: BranchSummary[], recent_activities: ActivityLog[] } }
// endpoint เดียวได้ครบทั้งหน้า และ backend เป็นคนตัดสินเองว่าเห็นกี่สาขา
// (spec: "Admin ได้ BranchSummary ของสาขาตัวเองรายการเดียว ส่วน Super Admin ได้ครบทุกสาขา")
export function getAdminDashboard(signal) {
  return apiClient.get("/admin/dashboard", { signal });
}
