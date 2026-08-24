import { IconChevronLeft, IconChevronRight } from "../icons";
import "./Pagination.css";

// ใช้กับทุก endpoint ที่คืน meta.page/total_pages แบบเดียวกัน (docs/openapi.yaml PageMeta)
export function Pagination({ page, totalPages, onChange }) {
  const total = totalPages || 1;

  return (
    <nav className="pagination" aria-label="เปลี่ยนหน้า">
      <button
        type="button"
        className="pagination__btn"
        onClick={() => onChange(page - 1)}
        disabled={page <= 1}
        aria-label="หน้าก่อนหน้า"
      >
        <IconChevronLeft width={18} height={18} />
      </button>

      <span className="pagination__label">
        หน้า {page} จาก {total}
      </span>

      <button
        type="button"
        className="pagination__btn"
        onClick={() => onChange(page + 1)}
        disabled={page >= total}
        aria-label="หน้าถัดไป"
      >
        <IconChevronRight width={18} height={18} />
      </button>
    </nav>
  );
}
