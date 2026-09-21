import { useRef, useState } from "react";
import { useClickOutside } from "../../../hooks/useClickOutside";
import { createStaff, updateStaff } from "../../../api/superadminApi";
import { IconClose } from "../../../components/icons";
import "./StaffModal.css";

const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
const PHONE_RE = /^[0-9]{9,10}$/;

function initialForm(staff) {
  return {
    first_name: staff?.first_name || "",
    last_name: staff?.last_name || "",
    email: staff?.email || "",
    phone: staff?.phone || "",
    branch_id: staff?.branch_id || "",
  };
}

// modal เพิ่ม/แก้ไขผู้ดูแล (prototype ภาพ 54/55 — โครงเดียวกัน ต่างแค่หัวข้อกับค่าที่เติมไว้)
// FE-36 (เพิ่ม) + FE-37 (แก้ไข) ใช้ component เดียวกัน
//
// เรื่องสิทธิ์สาขา (ทีมเคาะแล้ว ดู docs/c.md ข้อ 5):
//   - superadmin = ทุกสาขาเสมอ → ซ่อนกล่องเลือกสาขา และ "ห้ามส่ง branch_id" ตามที่ spec กำหนด
//   - admin = เลือกได้สาขาเดียว (backend รับ branch_id เป็น string เดี่ยว และหนึ่งสาขามีผู้ดูแลได้คนเดียว)
//     ต่างจาก prototype ที่วาดเป็น checkbox หลายช่อง — adapt ตาม backend ตาม hard rule ข้อ 4
export function StaffFormModal({ open, staff, branchOptions, onClose, onSaved }) {
  const boxRef = useRef(null);
  const [form, setForm] = useState(() => initialForm(staff));
  const [fieldErrors, setFieldErrors] = useState({});
  const [formError, setFormError] = useState("");
  const [submitting, setSubmitting] = useState(false);

  useClickOutside(boxRef, onClose, open);

  if (!open) return null;

  const isEdit = Boolean(staff);
  const isSuperadminTarget = staff?.role === "superadmin";
  const freeBranches = branchOptions.filter((b) => !b.takenBy);
  const noBranchAvailable = !isSuperadminTarget && freeBranches.length === 0 && !form.branch_id;

  function updateField(key, value) {
    setForm((f) => ({ ...f, [key]: value }));
    setFieldErrors((e) => ({ ...e, [key]: undefined }));
  }

  // mirror กฎของ backend ไว้ฝั่ง client เพื่อลด round-trip ที่ต้อง fail จริง
  function validate() {
    const errors = {};
    if (!form.first_name.trim()) errors.first_name = "กรุณากรอกชื่อ";
    if (!form.last_name.trim()) errors.last_name = "กรุณากรอกนามสกุล";
    if (!PHONE_RE.test(form.phone.trim())) errors.phone = "เบอร์โทรศัพท์ต้องเป็นตัวเลข 9-10 หลัก";
    if (!isEdit) {
      if (!form.email.trim()) errors.email = "กรุณากรอกอีเมล";
      else if (!EMAIL_RE.test(form.email.trim())) errors.email = "รูปแบบอีเมลไม่ถูกต้อง";
    }
    if (!isSuperadminTarget && !form.branch_id) errors.branch_id = "กรุณาเลือกสาขาที่รับผิดชอบ";
    return errors;
  }

  async function handleSubmit(e) {
    e.preventDefault();
    const errors = validate();
    setFieldErrors(errors);
    if (Object.keys(errors).length > 0) return;

    setFormError("");
    setSubmitting(true);
    try {
      if (isEdit) {
        const body = {
          first_name: form.first_name.trim(),
          last_name: form.last_name.trim(),
          phone: form.phone.trim(),
        };
        // ห้ามส่ง branch_id เมื่อเป้าหมายเป็น superadmin (spec กำหนดไว้ตรง ๆ)
        if (!isSuperadminTarget) body.branch_id = form.branch_id;
        await updateStaff(staff.id, body);
      } else {
        await createStaff({
          email: form.email.trim(),
          first_name: form.first_name.trim(),
          last_name: form.last_name.trim(),
          phone: form.phone.trim(),
          branch_id: form.branch_id,
        });
      }
      onSaved();
    } catch (err) {
      // 422 มาพร้อม error.fields (เช่น อีเมลซ้ำ / สาขามีผู้ดูแลแล้ว) โชว์ใต้ช่องนั้นตรง ๆ
      if (err.fields) setFieldErrors((prev) => ({ ...prev, ...err.fields }));
      else setFormError(err.message);
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className="staff-modal__backdrop" role="dialog" aria-modal="true" aria-label={isEdit ? "แก้ไขผู้ดูแล" : "เพิ่มผู้ดูแลใหม่"}>
      <div className="staff-modal__box" ref={boxRef}>
        <button type="button" className="staff-modal__close" onClick={onClose} aria-label="ปิด">
          <IconClose width={20} height={20} />
        </button>

        <h2>{isEdit ? "แก้ไขผู้ดูแล" : "เพิ่มผู้ดูแลใหม่"}</h2>

        {formError && (
          <p className="staff-modal__banner" role="alert">
            {formError}
          </p>
        )}

        <form onSubmit={handleSubmit} noValidate>
          <div className="staff-modal__row">
            <label className="staff-field">
              ชื่อ
              <input value={form.first_name} onChange={(e) => updateField("first_name", e.target.value)} />
              {fieldErrors.first_name && <span className="staff-field__error">{fieldErrors.first_name}</span>}
            </label>
            <label className="staff-field">
              นามสกุล
              <input value={form.last_name} onChange={(e) => updateField("last_name", e.target.value)} />
              {fieldErrors.last_name && <span className="staff-field__error">{fieldErrors.last_name}</span>}
            </label>
          </div>

          <div className="staff-modal__row">
            <label className="staff-field">
              อีเมล
              <input
                type="email"
                value={form.email}
                onChange={(e) => updateField("email", e.target.value)}
                disabled={isEdit}
                title={isEdit ? "แก้อีเมลของผู้ดูแลไม่ได้ (backend ไม่รับ field นี้ตอนแก้ไข)" : undefined}
              />
              {fieldErrors.email && <span className="staff-field__error">{fieldErrors.email}</span>}
            </label>
            <label className="staff-field">
              เบอร์โทรศัพท์
              <input type="tel" value={form.phone} onChange={(e) => updateField("phone", e.target.value)} />
              {fieldErrors.phone && <span className="staff-field__error">{fieldErrors.phone}</span>}
            </label>
          </div>

          {isSuperadminTarget ? (
            <p className="staff-modal__note">หัวหน้าผู้ดูแลระบบดูแลทุกสาขาอยู่แล้ว จึงไม่ต้องกำหนดสิทธิ์รายสาขา</p>
          ) : (
            <div className="staff-field">
              <span className="staff-field__label">สิทธิ์ดูแลสาขา</span>
              <div className="staff-branches">
                {branchOptions.length === 0 && <p className="staff-modal__note">ยังไม่มีสาขาในระบบ</p>}
                {branchOptions.map((b) => {
                  const isCurrent = b.id === form.branch_id;
                  const disabled = Boolean(b.takenBy) && !isCurrent;
                  return (
                    <label key={b.id} className={`staff-branches__item ${disabled ? "is-disabled" : ""}`}>
                      <input
                        type="radio"
                        name="branch_id"
                        value={b.id}
                        checked={isCurrent}
                        disabled={disabled}
                        onChange={() => updateField("branch_id", b.id)}
                      />
                      <span>{b.name}</span>
                      {disabled && (
                        <small>
                          — มีผู้ดูแลแล้ว ({b.takenBy.first_name} {b.takenBy.last_name})
                        </small>
                      )}
                    </label>
                  );
                })}
              </div>
              {/* prototype เขียนว่า "เลือกได้หลายสาขา" แต่ backend รับได้สาขาเดียวและหนึ่งสาขามีผู้ดูแล
                  ได้คนเดียว จึงต้องแก้ข้อความให้ตรงความจริง (ดู Gap ใน 08-superadmin-staff.md) */}
              <small className="staff-field__hint">
                เลือกได้สาขาเดียว — หนึ่งสาขามีผู้ดูแลได้คนเดียว และ Admin จะเข้าถึงเฉพาะสาขาที่ได้รับสิทธิ์
              </small>
              {fieldErrors.branch_id && <span className="staff-field__error">{fieldErrors.branch_id}</span>}
              {noBranchAvailable && (
                <span className="staff-field__error">
                  ทุกสาขามีผู้ดูแลครบแล้ว ต้องย้ายหรือลบผู้ดูแลเดิมออกก่อนจึงจะเพิ่มคนใหม่ได้
                </span>
              )}
            </div>
          )}

          <div className="staff-modal__actions">
            <button type="button" className="btn btn-outline" onClick={onClose} disabled={submitting}>
              ยกเลิก
            </button>
            <button type="submit" className="btn btn-primary" disabled={submitting || noBranchAvailable}>
              {submitting ? "กำลังบันทึก..." : "บันทึก"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
