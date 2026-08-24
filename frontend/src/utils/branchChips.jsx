// chip ไฮไลต์บนหน้ารวมสาขา — เฉพาะที่มี field จริงรองรับเท่านั้น (ดู Gap ใน docs/task/frontend/02-branches.md)
// ห้ามเดา/คำนวณตัวเลขขนาดห้องหรือจำนวนลิฟท์ที่ไม่มีข้อมูลจริง

// เรียงตามลำดับความสำคัญที่น่าจะดึงดูดคนดูมากที่สุดก่อน — เลือกแสดงแค่ 2 ตัวแรกที่สาขานั้นมีจริง
const AMENITY_CHIP_LABEL = {
  "convenience-store": "ใกล้ร้านสะดวกซื้อ",
  parking: "มีที่จอดรถ",
  elevator: "มีลิฟต์",
  furniture: "มีเฟอร์นิเจอร์ครบครัน",
  laundry: "เครื่องซักผ้าหยอดเหรียญ",
  wifi: "มีไวไฟ",
  aircon: "แอร์ทุกห้อง",
};

export function getBranchHighlightChips(branch) {
  const chips = [];

  if (branch.building_count && branch.floor_count) {
    chips.push(`${branch.building_count} อาคาร ${branch.floor_count} ชั้น`);
  }

  const amenityCodes = new Set((branch.amenities || []).map((a) => a.code));
  for (const code of Object.keys(AMENITY_CHIP_LABEL)) {
    if (chips.length >= 3) break;
    if (amenityCodes.has(code)) chips.push(AMENITY_CHIP_LABEL[code]);
  }

  return chips;
}
