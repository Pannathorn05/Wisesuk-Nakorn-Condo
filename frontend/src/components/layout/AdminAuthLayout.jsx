import { Outlet } from "react-router-dom";
import { AuthHeader } from "./AuthHeader";
import "./AdminAuthLayout.css";

// โครงหน้า auth ของโซนผู้ดูแล — header โลโก้อย่างเดียว และ "ไม่มี footer"
// ตาม prototype หน้า 39 ภาพ 51 (ในภาพมีแค่ header บาง ๆ กับการ์ดฟอร์มกลางพื้นหลังครีม)
export function AdminAuthLayout() {
  return (
    <>
      <AuthHeader />
      <main className="admin-auth-layout">
        <Outlet />
      </main>
    </>
  );
}
