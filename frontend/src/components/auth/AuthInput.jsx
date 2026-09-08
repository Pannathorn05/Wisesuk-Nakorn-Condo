import { useId, useState } from "react";
import { IconEye, IconEyeOff } from "../icons";
import "./AuthInput.css";

// ช่อง input ใช้ร่วมกันทั้งหน้า Login/Register (docs/task/frontend/05-login-register.md, FE-21)
// type="password" มีปุ่มไอคอนตา toggle แสดง/ซ่อนในตัว (ตรงกับภาพ 19/20 ในprototype — 2 การ์ดในภาพ
// คือ state เดียวกันของ toggle นี้เอง ไม่ใช่ 2 หน้าคนละแบบ) type อื่นเป็นช่อง input ธรรมดา + ไอคอนซ้าย
export function AuthInput({ label, icon, type = "text", error, id, ...inputProps }) {
  const autoId = useId();
  const inputId = id || autoId;
  const [showPassword, setShowPassword] = useState(false);
  const isPassword = type === "password";

  return (
    <div className="auth-input">
      <label htmlFor={inputId}>{label}</label>
      <div className={`auth-input__field ${error ? "has-error" : ""}`}>
        <span className="auth-input__icon">{icon}</span>
        <input id={inputId} type={isPassword && showPassword ? "text" : type} {...inputProps} />
        {isPassword && (
          <button
            type="button"
            className="auth-input__toggle"
            onClick={() => setShowPassword((v) => !v)}
            aria-label={showPassword ? "ซ่อนรหัสผ่าน" : "แสดงรหัสผ่าน"}
          >
            {showPassword ? <IconEyeOff width={18} height={18} /> : <IconEye width={18} height={18} />}
          </button>
        )}
      </div>
      {error && <p className="auth-input__error">{error}</p>}
    </div>
  );
}
