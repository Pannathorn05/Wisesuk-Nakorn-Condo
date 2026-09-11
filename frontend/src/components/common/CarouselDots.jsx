import "./Carousel.css";

// จำนวนจุดสูงสุดที่โชว์พร้อมกัน — สาขาที่มีรูปเยอะ (บางแค 45 รูป, เจริญกรุงเพลส 44 รูป) ถ้าโชว์
// ครบทุกจุดแถบจุดจะยาวล้นออกนอกกรอบรูป เลยใช้ "หน้าต่าง" เลื่อนตามรูปที่กำลังดูอยู่แทน
// (จำนวนรูปจริงไม่ได้ลดลง แค่จุดที่แสดงพร้อมกันเท่านั้นที่จำกัด)
const MAX_VISIBLE_DOTS = 7;

export function CarouselDots({ count, activeIndex, onSelect }) {
  if (count <= 1) return null;

  const visible = Math.min(MAX_VISIBLE_DOTS, count);
  const half = Math.floor(MAX_VISIBLE_DOTS / 2);
  // ให้จุดที่ active อยู่กลางหน้าต่างเสมอ ยกเว้นตอนอยู่ต้น/ท้ายชุดที่ต้องหนีบไว้ไม่ให้หลุดขอบ
  const start = count <= MAX_VISIBLE_DOTS ? 0 : Math.min(Math.max(activeIndex - half, 0), count - visible);

  return (
    <div className="carousel__dots">
      {Array.from({ length: visible }, (_, offset) => {
        const index = start + offset;
        // จุดริมสุดที่ยัง "มีรูปต่อไปอีก" ย่อขนาดลงเพื่อบอกใบ้ว่ายังไม่หมดแค่นี้
        const isEdge =
          (offset === 0 && start > 0) || (offset === visible - 1 && start + visible < count);

        return (
          <button
            key={index}
            type="button"
            className={`carousel__dot ${index === activeIndex ? "is-active" : ""} ${isEdge ? "is-edge" : ""}`}
            aria-label={`ไปรูปที่ ${index + 1} จาก ${count}`}
            onClick={() => onSelect(index)}
          />
        );
      })}
    </div>
  );
}
