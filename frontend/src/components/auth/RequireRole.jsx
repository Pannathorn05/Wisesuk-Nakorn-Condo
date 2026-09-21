import { Navigate, Outlet, useLocation } from "react-router-dom";
import { isLoggedIn, getCurrentRole } from "../../utils/authStorage";
import { getHomePathForRole } from "../../utils/roleRoutes";

// กัน route ของโซนผู้ดูแล (docs/task/frontend/06-admin-login.md, FE-28)
// - ยังไม่ login       -> เด้งไปหน้า login ของผู้ดูแล (จำหน้าที่ตั้งใจจะเข้าไว้ใน state.from)
// - login แล้วแต่ role ไม่ถึง -> ส่งกลับหน้าแรกของ role ตัวเอง ไม่ปล่อยให้เห็นหน้าในโซนนี้
//
// อ่าน role จาก authStorage (localStorage) ไม่ใช่ state ใน memory — reload หน้าแล้วต้องยังกันได้อยู่
// หมายเหตุด้านความปลอดภัย: guard นี้เป็นแค่ UX ฝั่ง client เท่านั้น ตัวจริงที่กันข้อมูลคือ backend
// ที่เช็ค token/role ของทุก endpoint อยู่แล้ว (ห้ามพึ่ง guard นี้แทนการตรวจสิทธิ์ฝั่ง server)
export function RequireRole({ roles, redirectTo = "/admin/login" }) {
  const location = useLocation();

  if (!isLoggedIn()) {
    return <Navigate to={redirectTo} replace state={{ from: location.pathname }} />;
  }

  const role = getCurrentRole();
  if (roles && !roles.includes(role)) {
    return <Navigate to={getHomePathForRole(role)} replace />;
  }

  return <Outlet />;
}
