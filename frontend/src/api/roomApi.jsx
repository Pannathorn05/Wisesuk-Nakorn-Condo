import { apiClient } from "./client";

// GET /api/v1/rooms/search -> { data: Room[], meta: PageMeta }
// เตรียมไว้ให้หน้าค้นหาห้องพัก (ยังไม่ implement หน้านั้นในรอบนี้)
export function searchRooms(params, signal) {
  return apiClient.get("/rooms/search", { params, signal });
}

// GET /api/v1/rooms/:roomID -> { data: Room }
export function getRoom(roomId, signal) {
  return apiClient.get(`/rooms/${roomId}`, { signal });
}
