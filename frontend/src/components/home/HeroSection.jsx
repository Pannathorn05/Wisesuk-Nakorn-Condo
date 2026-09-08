import { useState } from "react";
import { Link } from "react-router-dom";
import { IconBuilding, IconChevronLeft, IconChevronRight } from "../icons";
import { getAllBranchPhotos } from "../../assets/branchPhotos";
import { useCarousel } from "../../hooks/useCarousel";
import "../common/Carousel.css";
import "./HeroSection.css";

const SLIDE_TARGET_COUNT = 15; // จำนวนรูปที่อยากสุ่มมาโชว์ (ถ้ามีรูปในคลังน้อยกว่านี้ ใช้เท่าที่มี)
const AUTOPLAY_MS = 5000;

// สุ่มหยิบ n รูปแบบไม่ซ้ำจาก pool — Fisher–Yates แบบสั้น พอสำหรับ list ขนาดนี้
function pickRandomPhotos(pool, n) {
  const shuffled = [...pool];
  for (let i = shuffled.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1));
    [shuffled[i], shuffled[j]] = [shuffled[j], shuffled[i]];
  }
  return shuffled.slice(0, n);
}

export function HeroSection() {
  // สุ่มครั้งเดียวตอน mount (lazy initializer) — ไม่สุ่มใหม่ทุก re-render แต่จะสุ่มชุดใหม่ทุกครั้งที่
  // เข้าหน้านี้ใหม่ ดึงจากรูปสาขาจริงที่มีอยู่แล้ว (src/assets/branchPhotos.jsx) ไม่ใช่ mock/hardcode ใหม่
  const [slides] = useState(() => pickRandomPhotos(getAllBranchPhotos(), SLIDE_TARGET_COUNT));
  const { slideCount, extendedSlides, trackIndex, realIndex, smooth, prev, next, goTo, handleTrackTransitionEnd } =
    useCarousel(slides, { autoplayMs: AUTOPLAY_MS });

  return (
    <section className="hero">
      <div className="container hero__inner">
        <div className="hero__panel">
          <div className="hero__card">
            <div className="hero__text">
              <h1>หอพักวิเศษสุขนคร คอนโด</h1>
              <div className="hero__intro">
                <span className="hero__intro-arrow" aria-hidden="true">
                  ⟶
                </span>
                <p>
                  ปลอดภัย ราคาคุ้มค่า
                  <br />
                  พร้อมสิ่งอำนวยความสะดวกครบครัน
                  <br />
                  ทำเลดี ใกล้ห้างสรรพสินค้า และระบบขนส่งสาธารณะ
                </p>
              </div>
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
              {slideCount === 0 ? (
                <div className="carousel__track">
                  <div className="carousel__slide" role="img" aria-label="รูปตัวอย่างห้องพัก">
                    <IconBuilding width={64} height={64} />
                  </div>
                </div>
              ) : (
                // เลื่อนทั้งแถบรูปด้วย transform แทนการสลับรูปเดี่ยว ๆ — ได้แอนิเมชั่นเลื่อนตอนเปลี่ยนรูป
                // รูปที่กำลังโชว์ (is-active) มีเอฟเฟกต์ซูมช้า ๆ เพิ่มความน่าสนใจระหว่างที่ค้างอยู่
                <div
                  className={`carousel__track ${smooth ? "" : "carousel__track--no-transition"}`}
                  style={{ transform: `translateX(-${trackIndex * 100}%)` }}
                  onTransitionEnd={handleTrackTransitionEnd}
                >
                  {extendedSlides.map((src, i) => {
                    const isActive = i === trackIndex;
                    return (
                      <div
                        className="carousel__slide"
                        key={`${src}-${i}`}
                        role="img"
                        aria-label={`รูปตัวอย่างห้องพัก ${realIndex + 1} จาก ${slideCount}`}
                        aria-hidden={!isActive}
                      >
                        <img src={src} alt="" className={isActive ? "is-active" : ""} />
                      </div>
                    );
                  })}
                </div>
              )}
              {slideCount > 1 && (
                <>
                  <button type="button" className="carousel__nav carousel__nav--prev" onClick={prev} aria-label="รูปก่อนหน้า">
                    <IconChevronLeft />
                  </button>
                  <button type="button" className="carousel__nav carousel__nav--next" onClick={next} aria-label="รูปถัดไป">
                    <IconChevronRight />
                  </button>
                  <div className="carousel__dots">
                    {slides.map((src, i) => (
                      <button
                        key={src}
                        type="button"
                        className={`carousel__dot ${i === realIndex ? "is-active" : ""}`}
                        aria-label={`ไปรูปที่ ${i + 1}`}
                        onClick={() => goTo(i)}
                      />
                    ))}
                  </div>
                </>
              )}
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
