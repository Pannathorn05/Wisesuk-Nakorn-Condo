import { IconKeycard, IconGuard, IconParking, IconBathroom } from "../icons";
import "./FacilitiesStrip.css";

// static ตาม prototype ตรง ๆ — ไม่ใช่ข้อมูลชุดเดียวกับ GET /api/v1/amenities (ดู docs/task/frontend/01-homepage.md)
const FACILITIES = [
  { icon: IconKeycard, label: "คีย์การ์ด" },
  { icon: IconGuard, label: "รปภ. 24 ชั่วโมง" },
  { icon: IconParking, label: "ที่จอดรถ" },
  { icon: IconBathroom, label: "ห้องน้ำในตัว" },
];

export function FacilitiesStrip() {
  return (
    <section className="facilities">
      <div className="container facilities__grid">
        {FACILITIES.map(({ icon: Icon, label }) => (
          <div className="facilities__item" key={label}>
            <Icon className="facilities__icon" />
            <span>{label}</span>
          </div>
        ))}
      </div>
    </section>
  );
}
