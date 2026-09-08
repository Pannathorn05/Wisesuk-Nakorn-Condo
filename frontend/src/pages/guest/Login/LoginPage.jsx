import { useState } from "react";
import { Link, Navigate, useNavigate } from "react-router-dom";
import { login } from "../../../api/authApi";
import { setTokens, isLoggedIn } from "../../../utils/authStorage";
import { validateLogin } from "../../../utils/authValidation";
import { AuthCard } from "../../../components/auth/AuthCard";
import { AuthInput } from "../../../components/auth/AuthInput";
import { IconMail, IconLockClosed } from "../../../components/icons";
import "../../../components/auth/AuthForm.css";

// หน้าเข้าสู่ระบบ (docs/task/frontend/05-login-register.md, FE-22) — ต่อ POST /auth/login จริง
export function LoginPage() {
  const navigate = useNavigate();
  const [form, setForm] = useState({ email: "", password: "", remember: false });
  const [fieldErrors, setFieldErrors] = useState({});
  const [formError, setFormError] = useState("");
  const [submitting, setSubmitting] = useState(false);

  // login อยู่แล้วไม่ควรเห็นฟอร์มนี้ซ้ำ (ดู Gap ใน 05-login-register.md)
  if (isLoggedIn()) return <Navigate to="/" replace />;

  function updateField(key, value) {
    setForm((f) => ({ ...f, [key]: value }));
    setFieldErrors((e) => ({ ...e, [key]: undefined }));
  }

  async function handleSubmit(e) {
    e.preventDefault();
    const errors = validateLogin(form);
    setFieldErrors(errors);
    if (Object.keys(errors).length > 0) return;

    setFormError("");
    setSubmitting(true);
    try {
      const res = await login({ email: form.email.trim(), password: form.password });
      setTokens(res.data);
      navigate("/");
    } catch (err) {
      setFormError(err.message);
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <AuthCard
      title="เข้าสู่ระบบ"
      subtitle="เข้าสู่ระบบเพื่อจองห้องพักออนไลน์"
      error={formError}
      footer={
        <>
          ยังไม่มีบัญชี? <Link to="/register">สมัครสมาชิก</Link>
        </>
      }
    >
      <form className="auth-form" onSubmit={handleSubmit} noValidate>
        <AuthInput
          label="อีเมล"
          type="email"
          icon={<IconMail width={18} height={18} />}
          placeholder="your@gmail.com"
          autoComplete="email"
          value={form.email}
          onChange={(e) => updateField("email", e.target.value)}
          error={fieldErrors.email}
        />

        <AuthInput
          label="รหัสผ่าน"
          type="password"
          icon={<IconLockClosed width={18} height={18} />}
          placeholder="••••••••"
          autoComplete="current-password"
          value={form.password}
          onChange={(e) => updateField("password", e.target.value)}
          error={fieldErrors.password}
        />

        <div className="auth-form__meta">
          <label className="auth-form__remember">
            <input
              type="checkbox"
              checked={form.remember}
              onChange={(e) => updateField("remember", e.target.checked)}
            />
            จดจำฉัน
          </label>
          <Link to="/contact" className="auth-form__forgot">
            ลืมรหัสผ่าน?
          </Link>
        </div>

        <button type="submit" className="btn btn-primary" disabled={submitting}>
          {submitting ? "กำลังเข้าสู่ระบบ..." : "เข้าสู่ระบบ"}
        </button>
      </form>
    </AuthCard>
  );
}
