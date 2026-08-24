import { useEffect } from "react";

// ปิด dropdown/modal เมื่อคลิกนอกกล่อง หรือกด Escape — ใช้ร่วมกันทุก dropdown ใน filter bar และ modal
export function useClickOutside(ref, onOutside, active) {
  useEffect(() => {
    if (!active) return undefined;

    function handlePointer(e) {
      if (ref.current && !ref.current.contains(e.target)) onOutside();
    }
    function handleKey(e) {
      if (e.key === "Escape") onOutside();
    }

    document.addEventListener("mousedown", handlePointer);
    document.addEventListener("keydown", handleKey);
    return () => {
      document.removeEventListener("mousedown", handlePointer);
      document.removeEventListener("keydown", handleKey);
    };
  }, [ref, onOutside, active]);
}
