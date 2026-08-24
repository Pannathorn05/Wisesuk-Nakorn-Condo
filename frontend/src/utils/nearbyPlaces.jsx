const CATEGORY_LABEL = {
  education: "สถานศึกษา",
  hospital: "โรงพยาบาล",
  shopping: "ห้างสรรพสินค้าและแหล่งช้อปปิ้ง",
  park: "สวนสาธารณะ",
};

// จัดกลุ่ม nearby_places ตาม category แล้วเรียงภายในกลุ่มด้วย sort_order
// ต้องรับมือกรณี nearbyPlaces เป็น undefined ได้ (2 ใน 3 สาขาไม่มี key นี้เลยจาก backend — ดู Gap)
export function groupNearbyPlaces(nearbyPlaces) {
  const groups = new Map();

  for (const place of nearbyPlaces || []) {
    const key = place.category || "other";
    if (!groups.has(key)) groups.set(key, []);
    groups.get(key).push(place);
  }

  for (const list of groups.values()) {
    list.sort((a, b) => (a.sort_order ?? 0) - (b.sort_order ?? 0));
  }

  return Array.from(groups.entries()).map(([category, places]) => ({
    category,
    label: CATEGORY_LABEL[category] || category,
    places,
  }));
}
