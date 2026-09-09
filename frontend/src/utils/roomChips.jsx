// chip ใต้หัวข้อการ์ดผลลัพธ์ — ใช้เฉพาะ field ที่มีจริง (ดู Gap ใน docs/task/frontend/03-room-search.md)
// Room จาก GET /api/v1/rooms/search ไม่มี amenity list ต่อห้อง ต้อง join กับ amenities ของสาขาเอาเอง
// branchesById: Map<branchId, Branch> — สร้างจาก GET /api/v1/branches (ดู buildBranchesMap)
export function getRoomChips(room, branchesById) {
  const chips = [];

  if (room.room_type_name) chips.push(room.room_type_name);
  if (room.size_sqm) chips.push(`${room.size_sqm} ตร.ม.`);

  const branch = branchesById.get(room.branch_id);
  for (const amenity of branch?.amenities || []) {
    chips.push(amenity.name);
  }

  return chips;
}

export function buildBranchesMap(branches) {
  return new Map(branches.map((branch) => [branch.id, branch]));
}
