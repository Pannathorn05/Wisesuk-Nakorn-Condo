import { useState } from "react";
import { useStaffList } from "../../../hooks/useStaffList";
import { getCurrentUser } from "../../../utils/authStorage";
import { Skeleton } from "../../../components/common/Skeleton";
import { ErrorState } from "../../../components/common/ErrorState";
import { EmptyState } from "../../../components/common/EmptyState";
import { IconTrash } from "../../../components/icons";
import { StaffFormModal } from "./StaffFormModal";
import { ConfirmDeleteModal } from "./ConfirmDeleteModal";
import "./StaffPage.css";

// หน้าจัดการผู้ดูแลระบบ (prototype หน้า 40 ภาพ 53) — docs/task/frontend/08-superadmin-staff.md
export function StaffPage() {
  const { staff, branches, getAvailableBranches, loading, error, refetch } = useStaffList();
  const [formTarget, setFormTarget] = useState(null); // { staff } = แก้ไข, {} = เพิ่มใหม่, null = ปิด
  const [deleteTarget, setDeleteTarget] = useState(null);
  const currentUser = getCurrentUser();

  const branchName = (branchId) => branches.find((b) => b.id === branchId)?.name || "—";

  function handleSaved() {
    setFormTarget(null);
    refetch();
  }

  function handleDeleted() {
    setDeleteTarget(null);
    refetch();
  }

  return (
    <div className="staff-page">
      <div className="staff-page__head">
        <div>
          <h1>จัดการผู้ดูแลระบบ</h1>
          <p>เพิ่ม/แก้ไข บัญชี Admin สาขา และกำหนดสิทธิ์ดูแลสาขา</p>
        </div>
        <button type="button" className="btn btn-primary" onClick={() => setFormTarget({})}>
          เพิ่มผู้ดูแล
        </button>
      </div>

      {loading && (
        <div className="staff-page__skeleton">
          {[1, 2, 3].map((i) => (
            <Skeleton key={i} height="56px" radius="12px" />
          ))}
        </div>
      )}

      {!loading && error && <ErrorState message="โหลดรายชื่อผู้ดูแลไม่สำเร็จ" onRetry={refetch} />}

      {!loading && !error && staff.length === 0 && <EmptyState message="ยังไม่มีบัญชีผู้ดูแลในระบบ" />}

      {!loading && !error && staff.length > 0 && (
        <div className="staff-table__wrap">
          <table className="staff-table">
            <thead>
              <tr>
                <th>ผู้ดูแล</th>
                <th>บทบาท</th>
                <th>สิทธิ์ดูแลสาขา</th>
                <th>จัดการ</th>
              </tr>
            </thead>
            <tbody>
              {staff.map((u) => {
                const isSuperadmin = u.role === "superadmin";
                // ลบบัญชีตัวเองไม่ได้ (backend ตอบ 400) และ prototype ก็ไม่มีไอคอนลบในแถว superadmin
                const canDelete = !isSuperadmin && u.id !== currentUser?.id;
                return (
                  <tr key={u.id}>
                    <td>
                      <strong>
                        {u.first_name} {u.last_name}
                      </strong>
                      <small>{u.email}</small>
                    </td>
                    <td>
                      {isSuperadmin ? <span className="staff-badge">Superadmin</span> : "Admin"}
                    </td>
                    <td>{isSuperadmin ? "ทุกสาขา" : branchName(u.branch_id)}</td>
                    <td>
                      <div className="staff-table__actions">
                        <button type="button" className="staff-table__edit" onClick={() => setFormTarget({ staff: u })}>
                          แก้ไข
                        </button>
                        {canDelete && (
                          <button
                            type="button"
                            className="staff-table__delete"
                            onClick={() => setDeleteTarget(u)}
                            aria-label={`ลบ ${u.first_name} ${u.last_name}`}
                          >
                            <IconTrash width={18} height={18} />
                          </button>
                        )}
                      </div>
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}

      <StaffFormModal
        open={Boolean(formTarget)}
        key={formTarget?.staff?.id || (formTarget ? "new" : "closed")}
        staff={formTarget?.staff}
        branchOptions={getAvailableBranches(formTarget?.staff?.id)}
        onClose={() => setFormTarget(null)}
        onSaved={handleSaved}
      />

      <ConfirmDeleteModal
        open={Boolean(deleteTarget)}
        staff={deleteTarget}
        onClose={() => setDeleteTarget(null)}
        onDeleted={handleDeleted}
      />
    </div>
  );
}
