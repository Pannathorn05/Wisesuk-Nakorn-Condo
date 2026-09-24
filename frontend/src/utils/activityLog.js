// แปลค่า ActivityAction (enum อังกฤษจาก backend) เป็นข้อความไทยสำหรับกล่อง "กิจกรรมล่าสุด"
// ต้องครบทุกค่าใน enum ของ docs/openapi.yaml — ถ้า backend เพิ่มค่าใหม่ทีหลังจะตกไปที่ fallback
// (คืนค่า enum ดิบ) ดีกว่าโชว์ช่องว่าง แต่ควรกลับมาเติมตารางนี้ให้ครบเมื่อรู้ว่ามีค่าใหม่
const ACTION_LABELS = {
  "auth.register": "สมัครสมาชิก",
  "auth.login": "เข้าสู่ระบบ",
  "auth.oauth_register": "สมัครสมาชิกผ่านบัญชีภายนอก",
  "auth.oauth_login": "เข้าสู่ระบบผ่านบัญชีภายนอก",
  "auth.oauth_link": "เชื่อมบัญชีภายนอก",
  "user.update_profile": "แก้ไขข้อมูลส่วนตัว",
  "user.change_password": "เปลี่ยนรหัสผ่าน",
  "booking.create": "สร้างการจอง",
  "booking.submit_payment": "แจ้งชำระเงิน",
  "booking.cancel": "ยกเลิกการจอง",
  "booking.approve": "อนุมัติการจอง",
  "booking.reject": "ปฏิเสธการจอง",
  "booking.set_appointment": "นัดหมายทำสัญญา",
  "room.create": "เพิ่มห้องพัก",
  "room.update": "แก้ไขข้อมูลห้องพัก",
  "room.update_status": "เปลี่ยนสถานะห้องพัก",
  "room.update_image": "เปลี่ยนรูปห้องพัก",
  "room.delete": "ลบห้องพัก",
  "branch.update": "แก้ไขข้อมูลสาขา",
  "branch.set_amenities": "ตั้งค่าสิ่งอำนวยความสะดวกของสาขา",
  "branch.update_nearby": "แก้ไขสถานที่ใกล้เคียงของสาขา",
  "branch.update_cover": "เปลี่ยนรูปปกสาขา",
  "branch.add_image": "เพิ่มรูปสาขา",
  "branch.delete_image": "ลบรูปสาขา",
  "admin.create": "เพิ่มบัญชีผู้ดูแล",
  "admin.update": "แก้ไขบัญชีผู้ดูแล",
  "admin.delete": "ลบบัญชีผู้ดูแล",
};

// ActorRole ใช้ enum เดียวกับ UserRole (member/admin/superadmin)
const ROLE_LABELS = {
  member: "สมาชิก",
  admin: "Admin",
  superadmin: "Superadmin",
};

export function getActionLabel(action) {
  return ACTION_LABELS[action] || action;
}

export function getRoleLabel(role) {
  return ROLE_LABELS[role] || role;
}

// ประกอบข้อความหนึ่งบรรทัดตามรูปแบบใน prototype ภาพ 52: "Admin ประชาอุทิศ 45 อนุมัติการจอง"
// branch_name อาจไม่มี (เช่น superadmin ที่ไม่ได้ผูกกับสาขาไหน) — ข้ามไปเฉย ๆ ไม่ต้องเว้นช่องว่างค้าง
//
// หมายเหตุ: prototype มีคำว่า "รายวัน" ต่อท้ายด้วย แต่ ActivityLog ไม่มี field ประเภทการเข้าพักเลย
// (ดู Gap ใน docs/task/frontend/07-superadmin-dashboard.md) จึงไม่ใส่ ห้ามเดาหรือไปยิง API หาเพิ่มรายใบ
export function formatActivityText(activity) {
  return [getRoleLabel(activity.actor_role), activity.branch_name, getActionLabel(activity.action)]
    .filter(Boolean)
    .join(" ");
}

// "2026-03-12 13:34" ตามรูปแบบในภาพ 52 (ไม่ใช่ toLocaleString เพราะภาพใช้ ค.ศ. + 24 ชม. แบบนี้ตรง ๆ)
export function formatActivityTime(isoString) {
  const d = new Date(isoString);
  if (Number.isNaN(d.getTime())) return "";

  const pad = (n) => String(n).padStart(2, "0");
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`;
}
