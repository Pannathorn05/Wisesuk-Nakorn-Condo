import { Link } from "react-router-dom";
import "./ComingSoonPage.css";

// หน้ากันพัง (placeholder) สำหรับ route ที่ยังไม่ได้ implement ในรอบนี้
// ตาม docs/task/frontend/README.md — จะแทนที่ทีละหน้าเมื่อแตก task จาก prototype หน้านั้น ๆ แล้ว
export function ComingSoonPage({ title }) {
  return (
    <section className="section coming-soon">
      <div className="container coming-soon__box">
        <h1>{title}</h1>
        <p>หน้านี้กำลังอยู่ระหว่างพัฒนา</p>
        <Link to="/" className="btn btn-primary">
          กลับหน้าหลัก
        </Link>
      </div>
    </section>
  );
}
