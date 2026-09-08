import { useRef, useState } from "react";
import { useClickOutside } from "../../hooks/useClickOutside";
import { IconChevronDown } from "../icons";
import "./FilterDropdown.css";

// trigger + panel กลางของ dropdown ทั้ง 3 ตัวใน filter bar (สาขา/ประเภท/วันที่)
// เนื้อหาข้างในต่างกัน (list เลือก / ปฏิทิน) แต่พฤติกรรมเปิด-ปิดเหมือนกันหมด
export function FilterDropdown({ icon, label, disabled, children }) {
  const [open, setOpen] = useState(false);
  const ref = useRef(null);

  useClickOutside(ref, () => setOpen(false), open);

  return (
    <div className="filter-dropdown" ref={ref}>
      <button
        type="button"
        className={`filter-dropdown__trigger ${open ? "is-open" : ""}`}
        onClick={() => setOpen((o) => !o)}
        aria-expanded={open}
        disabled={disabled}
      >
        {icon}
        <span>{label}</span>
        <IconChevronDown width={16} height={16} />
      </button>
      {open && !disabled && (
        <div className="filter-dropdown__panel" role="listbox">
          {typeof children === "function" ? children({ close: () => setOpen(false) }) : children}
        </div>
      )}
    </div>
  );
}
