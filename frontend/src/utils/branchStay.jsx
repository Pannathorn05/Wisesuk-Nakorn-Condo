// เช็คว่า "มี key อยู่" ไม่ใช่ "ค่ามากกว่า 0" — daily_price_from/monthly_price_min หายไปทั้ง key
// เมื่อสาขานั้นไม่มีห้องประเภทนั้นขาย (ไม่ใช่ 0) ดู Gap ใน docs/task/frontend/02-branches.md
export function branchHasDaily(branch) {
  return Object.prototype.hasOwnProperty.call(branch, "daily_price_from");
}

export function branchHasMonthly(branch) {
  return Object.prototype.hasOwnProperty.call(branch, "monthly_price_min");
}

export function getBranchStayAvailabilityText(branch) {
  const hasDaily = branchHasDaily(branch);
  const hasMonthly = branchHasMonthly(branch);

  if (hasDaily && hasMonthly) return "มีห้องพักทั้งรายวัน และรายเดือน";
  if (hasMonthly) return "มีห้องพักรายเดือนพร้อมเข้าพัก";
  if (hasDaily) return "มีห้องพักรายวันพร้อมเข้าพัก";
  return "ติดต่อสอบถามห้องว่าง";
}
