import { IconBuilding, IconChevronLeft, IconChevronRight } from "../icons";
import { useCarousel } from "../../hooks/useCarousel";
import "../common/Carousel.css";
import "./BranchGallery.css";

const AUTOPLAY_MS = 5000; // เท่ากับ carousel รูปตัวอย่างห้องพักหน้า homepage (HeroSection)

// carousel รูปภาพสาขา — ลักษณะ/พฤติกรรมเดียวกับ carousel หน้า homepage (เลื่อนภาพ + จุดบอกตำแหน่ง +
// เปลี่ยนรูปอัตโนมัติ) ใช้ hook กลาง useCarousel ร่วมกัน ถ้าไม่มีรูปจริงเลย (ยังไม่ได้ถ่าย/ยังไม่
// อัปโหลด) โชว์ empty state แทนช่องว่าง
export function BranchGallery({ photos }) {
  const slides = photos || [];
  const { slideCount, extendedSlides, trackIndex, realIndex, smooth, prev, next, goTo, handleTrackTransitionEnd } =
    useCarousel(slides, { autoplayMs: AUTOPLAY_MS });

  if (slideCount === 0) {
    return (
      <div className="branch-media-box branch-gallery__empty">
        <IconBuilding width={48} height={48} />
        <p>ยังไม่มีรูปภาพสำหรับสาขานี้</p>
      </div>
    );
  }

  return (
    <div className="branch-media-box branch-gallery">
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
              aria-label={`รูปบรรยากาศ ${realIndex + 1} จาก ${slideCount}`}
              aria-hidden={!isActive}
            >
              <img src={src} alt="" className={isActive ? "is-active" : ""} />
            </div>
          );
        })}
      </div>
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
  );
}
