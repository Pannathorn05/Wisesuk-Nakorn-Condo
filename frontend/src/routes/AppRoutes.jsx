import { Routes, Route } from "react-router-dom";
import { HomePage } from "../pages/guest/Home/HomePage";
import { BranchesPage } from "../pages/guest/Branches/BranchesPage";
import { BranchDetailPage } from "../pages/guest/BranchDetail/BranchDetailPage";
import { RoomsPage } from "../pages/guest/Rooms/RoomsPage";
import { RoomDetailPage } from "../pages/guest/RoomDetail/RoomDetailPage";
import { ContactPage } from "../pages/guest/Contact/ContactPage";
import { LoginPage } from "../pages/guest/Login/LoginPage";
import { RegisterPage } from "../pages/guest/Register/RegisterPage";
import { NotFoundPage } from "../pages/NotFoundPage";

// เส้นทางฝั่ง Guest ตาม docs/c.md ข้อ 9 — หน้าที่ยังไม่ implement ใช้ ComingSoonPage กันพัง
// (แตก task ทีละหน้าใน docs/task/frontend/ ตาม prototype)
export function AppRoutes() {
  return (
    <Routes>
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
    </Routes>
  );
}
