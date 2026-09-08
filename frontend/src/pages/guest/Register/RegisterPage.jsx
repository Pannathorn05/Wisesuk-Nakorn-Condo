import { useState } from "react";
import { Link, Navigate, useNavigate } from "react-router-dom";
import { register } from "../../../api/authApi";
import { setTokens, isLoggedIn } from "../../../utils/authStorage";
import { validateRegister } from "../../../utils/authValidation";
import { AuthCard } from "../../../components/auth/AuthCard";
import { AuthInput } from "../../../components/auth/AuthInput";
import { IconUser, IconPhone, IconMail, IconLockClosed } from "../../../components/icons";
import "../../../components/auth/AuthForm.css";

const INITIAL_FORM = {
  first_name: "",
  last_name: "",
  phone: "",
  email: "",
  password: "",
  confirm_password: "",
};

// หน้าสมัครสมาชิก (docs/task/frontend/05-login-register.md, FE-23) — ต่อ POST /auth/register จริง
// สมัครสำเร็จ backend คืน token คู่มาให้เลย (login ให้อัตโนมัติ) ไม่ต้องยิง login ซ้ำ
export function RegisterPage() {
  const navigate = useNavigate();
  const [form, setForm] = useState(INITIAL_FORM);
  const [fieldErrors, setFieldErrors] = useState({});
  const [formError, setFormError] = useState("");
  const [submitting, setSubmitting] = useState(false);

  if (isLoggedIn()) return <Navigate to="/" replace />;

  function updateField(key, value) {
    setForm((f) => ({ ...f, [key]: value }));
    setFieldErrors((e) => ({ ...e, [key]: undefined }));
  }

  async function handleSubmit(e) {
    e.preventDefault();
    const errors = validateRegister(form);
    setFieldErrors(errors);
    if (Object.keys(errors).length > 0) return;

    setFormError("");
    setSubmitting(true);
    try {
      // ห้ามส่ง field อื่นนอกจาก 5 ตัวนี้เด็ดขาด (RegisterRequest additionalProperties: false —
      // ส่ง confirm_password หรือ role แถมไปจะได้ 422 ทันที)
      const res = await register({
        first_name: form.first_name.trim(),
        last_name: form.last_name.trim(),
        phone: form.phone.trim(),
        email: form.email.trim(),
        password: form.password,
      });
      setTokens(res.data);
      navigate("/");
    } catch (err) {
      if (err.fields) setFieldErrors((prev) => ({ ...prev, ...err.fields }));
      else setFormError(err.message);
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <AuthCard
      title="สมัครสมาชิก"
      subtitle="สร้างบัญชีเพื่อจองห้องพักออนไลน์"
      error={formError}
      footer={
        <>
          มีบัญชีแล้ว? <Link to="/login">เข้าสู่ระบบ</Link>
        </>
      }
    >
      <form className="auth-form" onSubmit={handleSubmit} noValidate>
        <div className="auth-form__row">
          <AuthInput
            label="ชื่อ"
            icon={<IconUser width={18} height={18} />}
            placeholder="ชื่อจริง"
            autoComplete="given-name"
            value={form.first_name}
            onChange={(e) => updateField("first_name", e.target.value)}
            error={fieldErrors.first_name}
          />
          <AuthInput
            label="นามสกุล"
            icon={<IconUser width={18} height={18} />}
            placeholder="นามสกุล"
            autoComplete="family-name"
            value={form.last_name}
            onChange={(e) => updateField("last_name", e.target.value)}
            error={fieldErrors.last_name}
          />
        </div>

        <AuthInput
          label="เบอร์โทรศัพท์"
          type="tel"
          icon={<IconPhone width={18} height={18} />}
          placeholder="08xxxxxxxx"
          autoComplete="tel"
          value={form.phone}
          onChange={(e) => updateField("phone", e.target.value)}
          error={fieldErrors.phone}
        />

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
          autoComplete="new-password"
          value={form.password}
          onChange={(e) => updateField("password", e.target.value)}
          error={fieldErrors.password}
        />

        <AuthInput
          label="ยืนยันรหัสผ่าน"
          type="password"
          icon={<IconLockClosed width={18} height={18} />}
          placeholder="••••••••"
          autoComplete="new-password"
          value={form.confirm_password}
          onChange={(e) => updateField("confirm_password", e.target.value)}
          error={fieldErrors.confirm_password}
        />

        <button type="submit" className="btn btn-primary" disabled={submitting}>
          {submitting ? "กำลังสมัครสมาชิก..." : "สมัครสมาชิก"}
        </button>
      </form>
    </AuthCard>
  );
}
