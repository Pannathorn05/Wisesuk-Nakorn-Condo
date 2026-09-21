import { useRef, useState } from "react";
import { useClickOutside } from "../../../hooks/useClickOutside";
import { deleteStaff } from "../../../api/superadminApi";
import { IconClose } from "../../../components/icons";
import "./StaffModal.css";

// ยืนยันก่อนลบผู้ดูแล (docs/task/frontend/08-superadmin-staff.md, FE-38)
// prototype ไม่มีภาพ confirm แต่การลบย้อนกลับไม่ได้ จึงเพิ่มไว้ (ตกลงไว้ใน Gap)
//
// ⚠️ docs/openapi.yaml เขียนว่า "soft delete" แต่โค้ดจริงของ backend คือ
// `DELETE FROM users WHERE id = $1 AND role = 'admin'` (account/repository.go) = ลบแถวถาวร
// ข้อความยืนยันจึงต้องบอกตามความจริงว่าลบถาวร ไม่ใช่แค่ปิดการใช้งาน
export function ConfirmDeleteModal({ open, staff, onClose, onDeleted }) {
  const boxRef = useRef(null);
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState(false);

  useClickOutside(boxRef, onClose, open);

  if (!open || !staff) return null;

  async function handleDelete() {
    setError("");
    setSubmitting(true);
    try {
      await deleteStaff(staff.id);
      onDeleted();
    } catch (err) {
      // 400 = ลบบัญชีตัวเองไม่ได้ (ถึง UI จะซ่อนปุ่มลบของตัวเองไว้แล้ว ก็ต้องรองรับให้ครบ)
      setError(err.message);
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className="staff-modal__backdrop" role="dialog" aria-modal="true" aria-label="ยืนยันการลบผู้ดูแล">
      <div className="staff-modal__box staff-modal__box--sm" ref={boxRef}>
        <button type="button" className="staff-modal__close" onClick={onClose} aria-label="ปิด">
          <IconClose width={20} height={20} />
        </button>

        <h2>ลบผู้ดูแล</h2>
        <p className="staff-modal__note">
          ต้องการลบบัญชีของ <strong>{staff.first_name} {staff.last_name}</strong> ({staff.email}) ใช่หรือไม่?
          บัญชีจะถูกลบออกจากระบบถาวรและกู้คืนไม่ได้ สาขาที่ดูแลอยู่จะว่างลงทันที
        </p>

        {error && (
          <p className="staff-modal__banner" role="alert">
            {error}
          </p>
        )}

        <div className="staff-modal__actions">
          <button type="button" className="btn btn-outline" onClick={onClose} disabled={submitting}>
            ยกเลิก
          </button>
          <button type="button" className="btn btn-primary" onClick={handleDelete} disabled={submitting}>
            {submitting ? "กำลังลบ..." : "ลบผู้ดูแล"}
          </button>
        </div>
      </div>
    </div>
  );
}
