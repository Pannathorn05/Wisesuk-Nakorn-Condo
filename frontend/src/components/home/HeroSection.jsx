import { useState } from "react";
import { Link } from "react-router-dom";
import { IconBuilding, IconChevronLeft, IconChevronRight } from "../icons";
import "./HeroSection.css";

// ยังไม่มี asset รูปห้องจริงในโปรเจกต์ (ต้องรอรูปจาก backend/ทีมงาน — cover_image_url ของห้อง/สาขายังว่างอยู่ทั้งหมด)
// ใช้ placeholder แบบมีป้ายกำกับชัดเจนแทนการหยิบภาพจาก prototype PDF มาใช้ตรง ๆ
const SLIDE_COUNT = 4;

export function HeroSection() {
  const [index, setIndex] = useState(0);

  const prev = () => setIndex((i) => (i - 1 + SLIDE_COUNT) % SLIDE_COUNT);
  const next = () => setIndex((i) => (i + 1) % SLIDE_COUNT);

  return (
    <section className="hero">
      <div className="container hero__inner">
        <div className="hero__card">
          <div className="hero__text">
            <h1>หอพักวิเศษสุขนคร คอนโด</h1>
            <p>
              ปลอดภัย ราคาคุ้มค่า
              <br />
              พร้อมสิ่งอำนวยความสะดวกครบครัน
              <br />
              ทำเลดี ใกล้ห้างสรรพสินค้า และระบบขนส่งสาธารณะ
            </p>
            <div className="hero__actions">
              <Link to="/rooms" className="btn btn-primary">
                ค้นหาห้องพัก
              </Link>
              <Link to="/branches" className="btn btn-outline">
                ดูสาขาทั้งหมด
              </Link>
            </div>
          </div>

          <div className="hero__carousel">
            <div className="hero__slide" role="img" aria-label={`รูปตัวอย่างห้องพัก ${index + 1} จาก ${SLIDE_COUNT}`}>
              <IconBuilding width={64} height={64} />
            </div>
            <button type="button" className="hero__nav hero__nav--prev" onClick={prev} aria-label="รูปก่อนหน้า">
              <IconChevronLeft />
            </button>
            <button type="button" className="hero__nav hero__nav--next" onClick={next} aria-label="รูปถัดไป">
              <IconChevronRight />
            </button>
            <div className="hero__dots">
              {Array.from({ length: SLIDE_COUNT }).map((_, i) => (
                <button
                  key={i}
                  type="button"
                  className={`hero__dot ${i === index ? "is-active" : ""}`}
                  aria-label={`ไปรูปที่ ${i + 1}`}
                  onClick={() => setIndex(i)}
                />
              ))}
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
