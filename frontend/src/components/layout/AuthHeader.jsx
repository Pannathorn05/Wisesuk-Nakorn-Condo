import { Link } from "react-router-dom";
import { IconBuilding } from "../icons";
import "./AuthHeader.css";

// header ของหน้า auth ฝั่งผู้ดูแล — prototype หน้า 39 ภาพ 51 โชว์แค่โลโก้อย่างเดียว
// ไม่มีเมนูนำทางและไม่มีปุ่มเข้าสู่ระบบ/สมัครสมาชิกแบบ Header ของ Guest
// (แยกเป็นคนละ component ตั้งใจ ไม่เอาไปยัดเป็นโหมดหนึ่งของ Header เดิมที่มีเมนู/เมนูมือถือของมันเอง)
export function AuthHeader() {
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
      </div>
    </header>
  );
}
