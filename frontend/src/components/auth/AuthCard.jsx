import "./AuthCard.css";

// การ์ดฟอร์มลอยกลางพื้นหลังครีม ใช้ร่วมกันทั้งหน้า Login/Register (docs/task/frontend/05-login-register.md)
export function AuthCard({ title, subtitle, error, children, footer }) {
  return (
    <section className="section auth-page">
      <div className="container auth-page__container">
        <div className="auth-card">
          <div className="auth-card__header">
            <h1>{title}</h1>
            <p>{subtitle}</p>
          </div>

          {error && (
            <p className="auth-card__banner" role="alert">
              {error}
            </p>
          )}

          {children}

          {footer && <div className="auth-card__footer">{footer}</div>}
        </div>
      </div>
    </section>
  );
}
