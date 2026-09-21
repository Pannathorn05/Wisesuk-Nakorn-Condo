// ลิงก์ "เพิ่มเพื่อน" ของ LINE Official Account — รูปแบบทางการคือ https://line.me/R/ti/p/<LINE ID>
// โดย ID ของ OA ขึ้นต้นด้วย @ เสมอ และต้อง encode เป็น %40 ไม่งั้นบางเบราว์เซอร์/แอปจะตัดทิ้ง
// เปิดบนมือถือที่มีแอป LINE จะเด้งเข้าแอปหน้าเพิ่มเพื่อนให้เลย บนเดสก์ท็อปจะเปิดหน้าเว็บของ OA แทน
//
// ค่า line_id มาจาก backend (GET /api/v1/branches -> Branch.line_id) ไม่ได้ hardcode ไว้ในโค้ด
export function getLineAddFriendUrl(lineId) {
  const id = lineId?.trim();
  if (!id) return null;

  const withPrefix = id.startsWith("@") ? id : `@${id}`;
  return `https://line.me/R/ti/p/${encodeURIComponent(withPrefix)}`;
}
