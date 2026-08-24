import { useRef } from "react";
import { Link } from "react-router-dom";
import { useClickOutside } from "../../hooks/useClickOutside";
import { IconClose } from "../icons";
import "./GuestBookingModal.css";

// docs/c.md ข้อ 19: Guest ที่ยังไม่ login กด "จอง" ต้องเห็น flow นี้ ห้ามเข้า booking flow ของ Member ตรง ๆ
export function GuestBookingModal({ open, onClose }) {
  const boxRef = useRef(null);
  useClickOutside(boxRef, onClose, open);

  if (!open) return null;

  return (
    <div className="guest-booking-modal__backdrop" role="dialog" aria-modal="true">
      <div className="guest-booking-modal__box" ref={boxRef}>
        <button type="button" className="guest-booking-modal__close" onClick={onClose} aria-label="ปิด">
          <IconClose width={18} height={18} />
        </button>
        <h3>โปรดลงทะเบียน / เข้าสู่ระบบ</h3>
        <p>เพื่อทำการจอง</p>
        <div className="guest-booking-modal__actions">
          <Link to="/login" className="btn btn-outline" onClick={onClose}>
            เข้าสู่ระบบ
          </Link>
          <Link to="/register" className="btn btn-primary" onClick={onClose}>
            สมัครสมาชิก
          </Link>
        </div>
      </div>
    </div>
  );
}
