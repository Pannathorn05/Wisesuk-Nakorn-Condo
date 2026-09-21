import { Routes, Route } from "react-router-dom";
import { GuestLayout } from "../components/layout/GuestLayout";
import { AdminAuthLayout } from "../components/layout/AdminAuthLayout";
import { AdminShell } from "../components/layout/AdminShell";
import { RequireRole } from "../components/auth/RequireRole";
import { HomePage } from "../pages/guest/Home/HomePage";
import { BranchesPage } from "../pages/guest/Branches/BranchesPage";
import { BranchDetailPage } from "../pages/guest/BranchDetail/BranchDetailPage";
import { RoomsPage } from "../pages/guest/Rooms/RoomsPage";
import { RoomDetailPage } from "../pages/guest/RoomDetail/RoomDetailPage";
import { ContactPage } from "../pages/guest/Contact/ContactPage";
import { LoginPage } from "../pages/guest/Login/LoginPage";
import { RegisterPage } from "../pages/guest/Register/RegisterPage";
import { AdminLoginPage } from "../pages/admin/AdminLogin/AdminLoginPage";
import { SuperAdminDashboardPage } from "../pages/superadmin/Dashboard/SuperAdminDashboardPage";
import { StaffPage } from "../pages/superadmin/Staff/StaffPage";
import { ComingSoonPage } from "../pages/guest/ComingSoonPage";
import { NotFoundPage } from "../pages/NotFoundPage";

// เส้นทางฝั่ง Guest ตาม docs/c.md ข้อ 9 + โซนผู้ดูแล (docs/task/frontend/06-admin-login.md)
// หน้าที่ยังไม่ implement ใช้ ComingSoonPage กันพัง (แตก task ทีละหน้าใน docs/task/frontend/)
export function AppRoutes() {
  return (
    <Routes>
      {/* โซน Guest — header เมนูเต็ม + footer */}
      <Route element={<GuestLayout />}>
        <Route path="/" element={<HomePage />} />
        <Route path="/branches" element={<BranchesPage />} />
        <Route path="/branches/:branchId" element={<BranchDetailPage />} />
        <Route path="/branches/:branchId/map" element={<BranchDetailPage />} />
        <Route path="/rooms" element={<RoomsPage />} />
        <Route path="/rooms/:roomId" element={<RoomDetailPage />} />
        <Route path="/contact" element={<ContactPage />} />
        <Route path="/login" element={<LoginPage />} />
        <Route path="/register" element={<RegisterPage />} />
        <Route path="*" element={<NotFoundPage />} />
      </Route>

      {/* โซนผู้ดูแล: หน้า login — header โลโก้อย่างเดียว ไม่มี footer */}
      <Route element={<AdminAuthLayout />}>
        <Route path="/admin/login" element={<AdminLoginPage />} />
      </Route>

      {/* ปลายทางหลัง login — ยังเป็น placeholder เพราะหน้าแดชบอร์ด/หน้าจัดการยังไม่อยู่ในขอบเขต
          (docs/c.md ข้อ 5 ยกเว้นเฉพาะหน้า login เท่านั้น) แต่ guard ทำงานจริงแล้ว */}
      <Route element={<RequireRole roles={["admin"]} />}>
        <Route element={<AdminAuthLayout />}>
          <Route path="/admin" element={<ComingSoonPage title="ระบบผู้ดูแลระบบ" />} />
        </Route>
      </Route>

      {/* โซนหัวหน้าผู้ดูแล — layout sidebar (ภาพ 52) แดชบอร์ดทำจริงแล้ว
          ส่วนอีก 3 เมนูยังเป็น placeholder เพราะยังไม่อนุมัติให้ทำ (docs/c.md ข้อ 5) */}
      <Route element={<RequireRole roles={["superadmin"]} />}>
        <Route element={<AdminShell />}>
          <Route path="/superadmin" element={<SuperAdminDashboardPage />} />
          <Route path="/superadmin/staff" element={<StaffPage />} />
          <Route path="/superadmin/branches" element={<ComingSoonPage title="จัดการสาขา" />} />
          <Route path="/superadmin/activity-logs" element={<ComingSoonPage title="Activity log" />} />
        </Route>
      </Route>
    </Routes>
  );
}
