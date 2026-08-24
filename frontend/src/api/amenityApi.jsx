import { apiClient } from "./client";

// GET /api/v1/amenities -> { data: Amenity[] }
// หมายเหตุ: นี่คือ catalog สิ่งอำนวยความสะดวกระดับห้อง/สาขา (ใช้ในหน้ารายละเอียดสาขา/filter ห้อง)
// ไม่ใช่ข้อมูลชุดเดียวกับแถบ facilities บน Homepage — ดู docs/task/frontend/01-homepage.md
export function listAmenities(signal) {
  return apiClient.get("/amenities", { signal });
}
