import { useState } from "react";
import { IconBuilding } from "../icons";
import "./ImageWithFallback.css";

// backend มี field รูปอยู่แล้ว (เช่น Branch.cover_image_url) แต่ยังไม่มีการอัปโหลดรูปจริง (ค่าว่าง)
// นี่คือ empty-state ของรูป ไม่ใช่การ hardcode รูปแทนของจริง — ดู docs/task/frontend/01-homepage.md
export function ImageWithFallback({ src, alt, className = "" }) {
  const [broken, setBroken] = useState(false);

  if (!src || broken) {
    return (
      <div className={`img-fallback ${className}`} role="img" aria-label={alt}>
        <IconBuilding width={48} height={48} />
      </div>
    );
  }

  return <img src={src} alt={alt} className={className} onError={() => setBroken(true)} />;
}
