// ปลายทางหลัง login แยกตาม role — ไม่ได้คิดเอง แต่ตามที่ docs/openapi.yaml เขียนกำกับ UserRole ไว้ว่า
// "frontend ใช้ค่านี้ตัดสินใจว่าจะพาผู้ใช้ไปหน้าไหนหลังล็อกอิน
//  (member → หน้าสมาชิก, admin → หน้าผู้ดูแล, superadmin → หน้าหัวหน้าผู้ดูแล)"
//
// หน้าสมาชิก/แดชบอร์ดจริงยังไม่อยู่ในขอบเขต (docs/c.md ข้อ 5) ตอนนี้ member จึงกลับหน้าแรก
// ส่วน admin/superadmin ไปหน้า placeholder กันพังไปก่อน
const HOME_BY_ROLE = {
  member: "/",
  admin: "/admin",
  superadmin: "/superadmin",
};

export const ADMIN_ROLES = ["admin", "superadmin"];

export function getHomePathForRole(role) {
  return HOME_BY_ROLE[role] || "/";
}

export function isAdminRole(role) {
  return ADMIN_ROLES.includes(role);
}
