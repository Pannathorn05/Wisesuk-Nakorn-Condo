import { Link } from "react-router-dom";
import { isLoggedIn } from "../../utils/authStorage";
import "./CtaSection.css";

export function CtaSection() {
  const loggedIn = isLoggedIn();

  return (
    <section className="section cta-section">
      <div className="container">
        <div className="cta-box">
          {loggedIn ? (
            <>
              <h2>พร้อมจองห้องพักแล้วหรือยัง?</h2>
              <p>ค้นหาห้องว่างและจองออนไลน์ได้ทันที</p>
              <div className="cta-box__actions">
                <Link to="/rooms" className="btn btn-primary">
                  ค้นหาห้องพัก
                </Link>
              </div>
            </>
          ) : (
            <>
              <h2>พร้อมจองห้องพักแล้วหรือยัง?</h2>
              <p>สมัครสมาชิกวันนี้ เพื่อจองห้องพักออนไลน์ได้ทันที</p>
              <div className="cta-box__actions">
                <Link to="/login" className="btn btn-primary">
                  เข้าสู่ระบบ
                </Link>
                <Link to="/register" className="btn btn-outline-light">
                  สมัครสมาชิก
                </Link>
              </div>
            </>
          )}
        </div>
      </div>
    </section>
  );
}
