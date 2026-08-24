import { useEffect, useMemo, useState } from "react";
import { Link } from "react-router-dom";
import { IconBuilding, IconChevronLeft, IconChevronRight } from "../icons";
import { getAllBranchPhotos } from "../../assets/branchPhotos";
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
  const slideCount = slides.length;

  // เทคนิค infinite loop carousel: ต่อโคลนรูปสุดท้ายไว้หน้าแถบ และโคลนรูปแรกไว้ท้ายแถบ
  // [last, ...slides, first] แล้วเลื่อน trackIndex ปกติ (1..slideCount คือรูปจริง) พอเลื่อนไปชนโคลน
  // หัว/ท้าย (0 หรือ slideCount+1) ค่อยสลับ transition ปิดชั่วคราวแล้ว "สแนป" กลับไปตำแหน่งจริงแบบ
  // มองไม่เห็นรอยต่อ — ทำให้ก่อนหน้า/ถัดไปเลื่อนสั้น ๆ ทิศทางถูกทุกครั้ง ไม่ใช่กวาดยาวข้ามทุกรูปตอนวน
  const extendedSlides = useMemo(() => {
    if (slideCount <= 1) return slides;
    return [slides[slideCount - 1], ...slides, slides[0]];
  }, [slides, slideCount]);

  const [trackIndex, setTrackIndex] = useState(1);
  const [smooth, setSmooth] = useState(true);
  const realIndex = slideCount > 0 ? (((trackIndex - 1) % slideCount) + slideCount) % slideCount : 0;

  // บั๊กที่เจอ: ถ้ากด "ถัดไป" รัว ๆ (หรือ autoplay ดันไปชนจังหวะเดียวกับที่เพิ่งกดเอง) trackIndex
  // จะบวกเพิ่มไปเรื่อย ๆ ได้โดยไม่รอให้ transitionend (handleTrackTransitionEnd ด้านล่าง) สแนปกลับมา
  // ในช่วง [0, slideCount+1] ก่อน — เกินขอบ extendedSlides ไปแล้ว index จะไม่มีรูปจริงอยู่ (undefined)
  // กลายเป็นรูปหายไปเลยตามที่เจอ กัน 2 ชั้น: (1) ระหว่างกำลังสแนปอยู่ (smooth=false) เมิน
  // next/prev เพิ่มไปก่อน — ช่วงนั้นสั้นมาก (2 เฟรม) กดไม่ทันอยู่แล้วในทางปฏิบัติ (2) เผื่อไว้อีกชั้น
  // ให้ wrap ด้วย modulo เองเลยไม่ให้ค่าหลุดขอบ [0, slideCount+1] ไม่ว่าจะเกิดอะไรขึ้น
  const prev = () => {
    if (!smooth) return;
    setTrackIndex((i) => (i - 1 < 0 ? slideCount : i - 1));
  };
  const next = () => {
    if (!smooth) return;
    setTrackIndex((i) => (i + 1 > slideCount + 1 ? 1 : i + 1));
  };
  const goTo = (i) => setTrackIndex(i + 1);

  // ปิด transition ชั่วคราวหลังสแนป แล้วเปิดกลับหลังจากเบราว์เซอร์วาดเฟรมที่ "สแนปแล้ว" ไปแสดงจริง
  // ก่อน (double requestAnimationFrame) ไม่งั้นถ้าเปิด transition คืนเร็วเกินไป การขยับกลับจะเห็นเป็น
  // แอนิเมชั่นเลื่อนย้อนไปด้วย
  useEffect(() => {
    if (smooth) return undefined;
    const raf1 = requestAnimationFrame(() => {
      const raf2 = requestAnimationFrame(() => setSmooth(true));
      return () => cancelAnimationFrame(raf2);
    });
    return () => cancelAnimationFrame(raf1);
  }, [smooth]);

  const handleTrackTransitionEnd = (e) => {
    if (e.target !== e.currentTarget || e.propertyName !== "transform") return;
    if (trackIndex === 0) {
      setSmooth(false);
      setTrackIndex(slideCount);
    } else if (trackIndex === slideCount + 1) {
      setSmooth(false);
      setTrackIndex(1);
    }
  };

  // เปลี่ยนรูปเองอัตโนมัติ — ใช้ setTimeout ที่ตั้งใหม่ทุกครั้งที่ realIndex เปลี่ยน (ไม่ว่าจะมาจาก
  // autoplay เองหรือคนกดปุ่ม/จุดเอง) ทำให้กดเองแล้วนับเวลาถอยหลังเริ่มใหม่ ไม่ใช่เปลี่ยนซ้อนทันที
  useEffect(() => {
    if (slideCount <= 1) return undefined;
    const timer = setTimeout(next, AUTOPLAY_MS);
    return () => clearTimeout(timer);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [realIndex, slideCount]);

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
            {slideCount === 0 ? (
              <div className="hero__track">
                <div className="hero__slide" role="img" aria-label="รูปตัวอย่างห้องพัก">
                  <IconBuilding width={64} height={64} />
                </div>
              </div>
            ) : (
              // เลื่อนทั้งแถบรูปด้วย transform แทนการสลับรูปเดี่ยว ๆ — ได้แอนิเมชั่นเลื่อนตอนเปลี่ยนรูป
              // รูปที่กำลังโชว์ (is-active) มีเอฟเฟกต์ซูมช้า ๆ เพิ่มความน่าสนใจระหว่างที่ค้างอยู่
              <div
                className={`hero__track ${smooth ? "" : "hero__track--no-transition"}`}
                style={{ transform: `translateX(-${trackIndex * 100}%)` }}
                onTransitionEnd={handleTrackTransitionEnd}
              >
                {extendedSlides.map((src, i) => {
                  const isActive = i === trackIndex;
                  return (
                    <div
                      className="hero__slide"
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
                <button type="button" className="hero__nav hero__nav--prev" onClick={prev} aria-label="รูปก่อนหน้า">
                  <IconChevronLeft />
                </button>
                <button type="button" className="hero__nav hero__nav--next" onClick={next} aria-label="รูปถัดไป">
                  <IconChevronRight />
                </button>
                <div className="hero__dots">
                  {slides.map((src, i) => (
                    <button
                      key={src}
                      type="button"
                      className={`hero__dot ${i === realIndex ? "is-active" : ""}`}
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
    </section>
  );
}
