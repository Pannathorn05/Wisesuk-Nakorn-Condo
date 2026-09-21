import { NavLink, Outlet, useNavigate } from "react-router-dom";
import { getCurrentUser, clearTokens, getRefreshToken } from "../../utils/authStorage";
import { getRoleLabel } from "../../utils/activityLog";
import { logout as logoutRequest } from "../../api/authApi";
import { IconGrid, IconUsers, IconPin, IconLogs, IconLogout } from "../icons";
import "./AdminShell.css";

// โครงหน้าโซนผู้ดูแลแบบมี sidebar (prototype หน้า 39 ภาพ 52) — layout ที่ 3 ของโปรเจกต์
// ต่อจาก GuestLayout (header เมนู + footer) และ AdminAuthLayout (header โลโก้อย่างเดียว)
//
// เมนูที่ยังไม่มีหน้าจริงชี้ไป placeholder ไว้ก่อน (ยังไม่อนุมัติให้ทำ ดู docs/c.md ข้อ 5)
// แต่ต้องกดได้จริงทุกปุ่ม ไม่ปล่อยเป็นปุ่มตาย
const MENU = [
  { to: "/superadmin", label: "แดชบอร์ด", Icon: IconGrid, end: true },
  { to: "/superadmin/staff", label: "จัดการผู้ดูแลระบบ", Icon: IconUsers },
  { to: "/superadmin/branches", label: "จัดการสาขา", Icon: IconPin },
  { to: "/superadmin/activity-logs", label: "activity log", Icon: IconLogs },
];

export function AdminShell() {
  const navigate = useNavigate();
  const user = getCurrentUser();

  // เหมือนปุ่ม logout ใน Header (FE-24): ยิงบอก backend แบบ best-effort แล้วเคลียร์ฝั่ง client เลย
  // ไม่รอผล เพราะ backend ตอบ 204 เสมอ และผู้ใช้ไม่ควรค้างถ้าเน็ตมีปัญหา
  function handleLogout() {
    const refreshToken = getRefreshToken();
    if (refreshToken) logoutRequest(refreshToken).catch(() => {});
    clearTokens();
    navigate("/admin/login", { replace: true });
  }

  return (
    <div className="admin-shell">
      <aside className="admin-shell__sidebar">
        <div className="admin-shell__user">
          <strong>{[user?.first_name, user?.last_name].filter(Boolean).join(" ") || "ผู้ดูแลระบบ"}</strong>
          <span>{getRoleLabel(user?.role)}</span>
        </div>

        <nav className="admin-shell__nav" aria-label="เมนูผู้ดูแลระบบ">
          {MENU.map(({ to, label, Icon, end }) => (
            <NavLink
              key={to}
              to={to}
              end={end}
              className={({ isActive }) => `admin-shell__nav-link ${isActive ? "is-active" : ""}`}
            >
              <Icon width={20} height={20} />
              {label}
            </NavLink>
          ))}
        </nav>

        <button type="button" className="admin-shell__logout" onClick={handleLogout}>
          <IconLogout width={20} height={20} />
          Logout
        </button>
      </aside>

      <main className="admin-shell__content">
        <Outlet />
      </main>
    </div>
  );
}
