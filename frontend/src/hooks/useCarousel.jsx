import { useEffect, useMemo, useState } from "react";

// carousel วนลูปไม่มีที่สิ้นสุดแบบเลื่อนภาพ (ไม่ใช่สลับรูปเดี่ยว ๆ) — ใช้ร่วมกันทั้ง HeroSection
// (หน้าแรก) และ BranchGallery (หน้ารายละเอียดสาขา) ที่มาของเทคนิค/บั๊กที่เจอระหว่างทำ ดูคอมเมนต์เดิม
// ที่ HeroSection.jsx (ก่อนแยกออกมาเป็น hook นี้)
//
// เทคนิค: ต่อโคลนรูปสุดท้ายไว้หน้าแถบ และโคลนรูปแรกไว้ท้ายแถบ [last, ...slides, first] แล้วเลื่อน
// trackIndex ปกติ (1..slideCount คือรูปจริง) พอเลื่อนไปชนโคลนหัว/ท้าย (0 หรือ slideCount+1) ค่อยสลับ
// transition ปิดชั่วคราวแล้ว "สแนป" กลับไปตำแหน่งจริงแบบมองไม่เห็นรอยต่อ
//
// autoplayMs: undefined/0 = ไม่เล่นอัตโนมัติ
export function useCarousel(slides, { autoplayMs } = {}) {
  const slideCount = slides.length;

  const extendedSlides = useMemo(() => {
    if (slideCount <= 1) return slides;
    return [slides[slideCount - 1], ...slides, slides[0]];
  }, [slides, slideCount]);

  // มีแค่ 0-1 รูป ไม่มีโคลนหัว/ท้ายให้ชน (extendedSlides = slides ตรง ๆ) trackIndex เลยต้องเริ่มที่ 0
  // (ชี้รูปจริงตัวเดียวใน array ที่ไม่ได้ถูกยืด) ไม่ใช่ 1 แบบกรณีปกติ ไม่งั้น translateX(-100%) จะเลื่อน
  // ไปช่องที่ไม่มีรูปอยู่ กลายเป็นจอว่าง
  const [trackIndex, setTrackIndex] = useState(() => (slideCount <= 1 ? 0 : 1));
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
    if (!autoplayMs || slideCount <= 1) return undefined;
    const timer = setTimeout(next, autoplayMs);
    return () => clearTimeout(timer);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [realIndex, slideCount, autoplayMs]);

  return { slideCount, extendedSlides, trackIndex, realIndex, smooth, prev, next, goTo, handleTrackTransitionEnd };
}
