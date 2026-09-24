import { Link, useNavigate } from "react-router-dom";
import { IconBuilding } from "../icons";
import { useAuthState } from "../../hooks/useAuthState";
import { logout as logoutRequest } from "../../api/authApi";
import { clearTokens, getRefreshToken } from "../../utils/authStorage";
import "./AuthHeader.css";

// header ของหน้า auth ฝั่งผู้ดูแล — prototype หน้า 39 ภาพ 51 โชว์แค่โลโก้อย่างเดียว
// ไม่มีเมนูนำทางและไม่มีปุ่มเข้าสู่ระบบ/สมัครสมาชิกแบบ Header ของ Guest
// (แยกเป็นคนละ component ตั้งใจ ไม่เอาไปยัดเป็นโหมดหนึ่งของ Header เดิมที่มีเมนู/เมนูมือถือของมันเอง)
//
// layout นี้ครอบหน้า /admin (placeholder ของ role admin) ด้วย ซึ่งไม่มี sidebar แบบ AdminShell
// จึงต้องมีปุ่มออกจากระบบตรงนี้ ไม่งั้น admin ออกจากระบบไม่ได้เลย (BUG-18) — โชว์เฉพาะตอน login อยู่
// หน้า /admin/login จึงยังเป็นโลโก้อย่างเดียวตามภาพ 51
export function AuthHeader() {
  const loggedIn = useAuthState();
  const navigate = useNavigate();

  // เหมือนปุ่ม logout ใน Header/AdminShell: ยิงบอก backend แบบ best-effort แล้วเคลียร์ฝั่ง client เลย
  function handleLogout() {
    const refreshToken = getRefreshToken();
    if (refreshToken) logoutRequest(refreshToken).catch(() => {});
    clearTokens();
    navigate("/admin/login", { replace: true });
  }

  return (
    <header className="auth-header">
      <div className="container auth-header__inner">
        <Link to="/" className="auth-header__brand">
          <IconBuilding width={32} height={32} />
          <span className="auth-header__brand-text">
            <strong>วิเศษสุขนครคอนโด</strong>
            <small>และหอพักในเครือ</small>
          </span>
        </Link>

        {loggedIn && (
          <button type="button" className="btn btn-outline" onClick={handleLogout}>
            ออกจากระบบ
          </button>
        )}
      </div>
    </header>
  );
}
