import { useState } from "react";
import { NavLink, useNavigate } from "react-router-dom";
import { IconBuilding, IconMenu, IconClose } from "../icons";
import { useAuthState } from "../../hooks/useAuthState";
import { logout as logoutRequest } from "../../api/authApi";
import { clearTokens, getRefreshToken } from "../../utils/authStorage";
import "./Header.css";

const NAV_LINKS = [
  { to: "/", label: "หน้าหลัก", end: true },
  { to: "/branches", label: "สาขาของเรา" },
  { to: "/rooms", label: "ค้นหาห้องพัก" },
  { to: "/contact", label: "ติดต่อเรา" },
];

export function Header() {
  const [menuOpen, setMenuOpen] = useState(false);
  const loggedIn = useAuthState();
  const navigate = useNavigate();

  // best-effort — backend ตอบ 204 เสมอไม่ว่า token จะใช้ได้หรือไม่ (docs/openapi.yaml) เลยไม่ต้องรอ/
  // เช็คผลลัพธ์ก่อนเคลียร์ token ฝั่ง client จะได้ไม่ทำให้ผู้ใช้ค้างรอถ้าเน็ตหลุดตอนกด logout
  function handleLogout() {
    const refreshToken = getRefreshToken();
    if (refreshToken) logoutRequest(refreshToken).catch(() => {});
    clearTokens();
    setMenuOpen(false);
    navigate("/");
  }

  return (
    <header className="site-header">
      <div className="container site-header__inner">
        <NavLink to="/" className="site-header__brand" onClick={() => setMenuOpen(false)}>
          <IconBuilding className="site-header__brand-icon" width={32} height={32} />
          <span className="site-header__brand-text">
            <strong>วิเศษสุขนครคอนโด</strong>
            <small>และหอพักในเครือ</small>
          </span>
        </NavLink>

        <nav className={`site-header__nav ${menuOpen ? "is-open" : ""}`} aria-label="เมนูหลัก">
          {NAV_LINKS.map((link) => (
            <NavLink
              key={link.to}
              to={link.to}
              end={link.end}
              className={({ isActive }) => `site-header__nav-link ${isActive ? "is-active" : ""}`}
              onClick={() => setMenuOpen(false)}
            >
              {link.label}
            </NavLink>
          ))}

          <div className="site-header__nav-auth">
            {loggedIn ? (
              <button type="button" className="btn btn-outline" onClick={handleLogout}>
                ออกจากระบบ
              </button>
            ) : (
              <>
                <NavLink to="/login" className="btn btn-outline" onClick={() => setMenuOpen(false)}>
                  เข้าสู่ระบบ
                </NavLink>
                <NavLink to="/register" className="btn btn-primary" onClick={() => setMenuOpen(false)}>
                  สมัครสมาชิก
                </NavLink>
              </>
            )}
          </div>
        </nav>

        <div className="site-header__auth-desktop">
          {loggedIn ? (
            <button type="button" className="btn btn-outline" onClick={handleLogout}>
              ออกจากระบบ
            </button>
          ) : (
            <>
              <NavLink to="/login" className="btn btn-outline">
                เข้าสู่ระบบ
              </NavLink>
              <NavLink to="/register" className="btn btn-primary">
                สมัครสมาชิก
              </NavLink>
            </>
          )}
        </div>

        <button
          type="button"
          className="site-header__toggle"
          aria-label={menuOpen ? "ปิดเมนู" : "เปิดเมนู"}
          aria-expanded={menuOpen}
          onClick={() => setMenuOpen((open) => !open)}
        >
          {menuOpen ? <IconClose /> : <IconMenu />}
        </button>
      </div>
    </header>
  );
}
