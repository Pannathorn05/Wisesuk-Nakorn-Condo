import { useState } from "react";
import { NavLink } from "react-router-dom";
import { IconBuilding, IconMenu, IconClose } from "../icons";
import "./Header.css";

const NAV_LINKS = [
  { to: "/", label: "หน้าหลัก", end: true },
  { to: "/branches", label: "สาขาของเรา" },
  { to: "/rooms", label: "ค้นหาห้องพัก" },
  { to: "/contact", label: "ติดต่อเรา" },
];

export function Header() {
  const [menuOpen, setMenuOpen] = useState(false);

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
            <NavLink to="/login" className="btn btn-outline" onClick={() => setMenuOpen(false)}>
              เข้าสู่ระบบ
            </NavLink>
            <NavLink to="/register" className="btn btn-primary" onClick={() => setMenuOpen(false)}>
              สมัครสมาชิก
            </NavLink>
          </div>
        </nav>

        <div className="site-header__auth-desktop">
          <NavLink to="/login" className="btn btn-outline">
            เข้าสู่ระบบ
          </NavLink>
          <NavLink to="/register" className="btn btn-primary">
            สมัครสมาชิก
          </NavLink>
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
