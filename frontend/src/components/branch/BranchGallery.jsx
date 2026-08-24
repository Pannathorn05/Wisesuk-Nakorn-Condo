import { useState } from "react";
import { IconBuilding, IconChevronLeft, IconChevronRight } from "../icons";
import "./BranchGallery.css";

// carousel รูปภาพสาขา — ถ้าไม่มีรูปจริงเลย (ยังไม่ได้ถ่าย/ยังไม่อัปโหลด) โชว์ empty state แทนช่องว่าง
export function BranchGallery({ photos }) {
  const [index, setIndex] = useState(0);

  if (!photos || photos.length === 0) {
    return (
      <div className="branch-media-box branch-gallery__empty">
        <IconBuilding width={48} height={48} />
        <p>ยังไม่มีรูปภาพสำหรับสาขานี้</p>
      </div>
    );
  }

  const prev = () => setIndex((i) => (i - 1 + photos.length) % photos.length);
  const next = () => setIndex((i) => (i + 1) % photos.length);

  return (
    <div className="branch-media-box branch-gallery">
      <img src={photos[index]} alt={`รูปบรรยากาศ ${index + 1} จาก ${photos.length}`} />
      {photos.length > 1 && (
        <>
          <button type="button" className="branch-gallery__nav branch-gallery__nav--prev" onClick={prev} aria-label="รูปก่อนหน้า">
            <IconChevronLeft />
          </button>
          <button type="button" className="branch-gallery__nav branch-gallery__nav--next" onClick={next} aria-label="รูปถัดไป">
            <IconChevronRight />
          </button>
          <div className="branch-gallery__dots">
            {photos.map((photo, i) => (
              <button
                key={photo}
                type="button"
                className={`branch-gallery__dot ${i === index ? "is-active" : ""}`}
                aria-label={`ไปรูปที่ ${i + 1}`}
                onClick={() => setIndex(i)}
              />
            ))}
          </div>
        </>
      )}
    </div>
  );
}
