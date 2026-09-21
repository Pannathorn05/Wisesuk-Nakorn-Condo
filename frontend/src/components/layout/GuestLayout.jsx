import { Outlet } from "react-router-dom";
import { Header } from "./Header";
import { Footer } from "./Footer";

// โครงหน้าฝั่ง Guest (header เมนูเต็ม + footer) — ย้ายมาจาก App.js เพื่อให้โซนผู้ดูแลใช้โครงคนละแบบได้
// (docs/task/frontend/06-admin-login.md, FE-26) ทุกหน้าเดิมของ Guest ยังได้ layout เดิมเป๊ะ
export function GuestLayout() {
  return (
    <>
      <Header />
      <main>
        <Outlet />
      </main>
      <Footer />
    </>
  );
}
