import { useState } from "react";
import { Navigate, useLocation, useNavigate } from "react-router-dom";
import { login } from "../../../api/authApi";
import { setTokens, clearTokens, isLoggedIn, getCurrentRole } from "../../../utils/authStorage";
import { validateLogin } from "../../../utils/authValidation";
import { getHomePathForRole, isAdminRole } from "../../../utils/roleRoutes";
import { AuthCard } from "../../../components/auth/AuthCard";
import { AuthInput } from "../../../components/auth/AuthInput";
import { IconMail, IconLockClosed } from "../../../components/icons";
import "../../../components/auth/AuthForm.css";

// หน้าเข้าสู่ระบบของผู้ดูแล/หัวหน้าผู้ดูแล (docs/task/frontend/06-admin-login.md, FE-27)
// prototype หน้า 39 ภาพ 51 — ฟอร์มหน้าตาเหมือนหน้า Login ฝั่ง Guest ทุกจุด ยกเว้นไม่มีลิงก์สมัครสมาชิก
// จึง reuse AuthCard/AuthInput/AuthForm.css ชุดเดิมทั้งหมด ไม่สร้าง component ใหม่
//
// backend ใช้ POST /auth/login ตัวเดียวกับ Guest (spec เขียนว่า "ใช้ endpoint เดียวทุกบทบาท")
// การแยกว่าใครเข้าโซนไหนได้จึงเป็นหน้าที่ frontend ล้วน ๆ (ดู handleSubmit ด้านล่าง)
export function AdminLoginPage() {
  const navigate = useNavigate();
  const location = useLocation();
  const [form, setForm] = useState({ email: "", password: "", remember: false });
  const [fieldErrors, setFieldErrors] = useState({});
  const [formError, setFormError] = useState("");
  const [submitting, setSubmitting] = useState(false);

  // login ค้างอยู่แล้วไม่ต้องเห็นฟอร์มนี้ซ้ำ — ส่งไปหน้าแรกของ role ที่ login อยู่
  if (isLoggedIn()) return <Navigate to={getHomePathForRole(getCurrentRole())} replace />;

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
      const result = res.data;

      // backend ตอบ 200 ให้ทุก role ที่รหัสผ่านถูก รวมถึง member ที่ไม่มีสิทธิ์ในโซนนี้ —
      // ต้องปฏิเสธเองที่ frontend และล้าง token ที่เพิ่งได้มาทิ้ง ไม่ปล่อยให้ค้างสถานะ login ไว้
      if (!isAdminRole(result?.user?.role)) {
        clearTokens();
        setFormError("บัญชีนี้ไม่มีสิทธิ์เข้าใช้ระบบผู้ดูแล กรุณาเข้าสู่ระบบที่หน้าสำหรับสมาชิก");
        return;
      }

      setTokens(result);
      // ถ้าโดน guard เด้งมาจากหน้าไหน กลับไปหน้านั้น ไม่งั้นไปหน้าแรกตาม role
      const from = location.state?.from;
      navigate(from || getHomePathForRole(result.user.role), { replace: true });
    } catch (err) {
      setFormError(err.message);
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <AuthCard
      title="เข้าสู่ระบบ"
      subtitle="สำหรับผู้ดูแลระบบและหัวหน้าผู้ดูแลระบบ"
      error={formError}
      footer={<span className="auth-form__hint">ลืมรหัสผ่าน? กรุณาติดต่อหัวหน้าผู้ดูแลระบบเพื่อรีเซ็ตรหัสผ่าน</span>}
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
        </div>

        <button type="submit" className="btn btn-primary" disabled={submitting}>
          {submitting ? "กำลังเข้าสู่ระบบ..." : "เข้าสู่ระบบ"}
        </button>
      </form>
    </AuthCard>
  );
}
