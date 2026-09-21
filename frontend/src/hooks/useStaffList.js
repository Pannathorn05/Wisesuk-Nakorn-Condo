import { useAsync } from "./useAsync";
import { listStaff, listAllBranches } from "../api/superadminApi";

// รายชื่อผู้ดูแล + รายการสาขา (docs/task/frontend/08-superadmin-staff.md, FE-35)
// ดึงสาขามาด้วยเพราะต้องใช้ 2 ที่: แปลง branch_id -> ชื่อสาขาในตาราง และเติมตัวเลือกใน modal
// ยิงคู่กันด้วย Promise.all ใน fetcher เดียว จะได้มี loading/error/refetch ชุดเดียวคุมทั้งหน้า
export function useStaffList() {
  const { data, loading, error, refetch } = useAsync(
    (signal) =>
      Promise.all([listStaff(signal), listAllBranches(signal)]).then(([staffRes, branchRes]) => ({
        staff: staffRes.data ?? [],
        branches: branchRes.data ?? [],
      })),
    []
  );

  const branches = data?.branches || [];
  const staff = data?.staff || [];

  // สาขาที่ยังไม่มีผู้ดูแล — ใช้กันไม่ให้ผู้ใช้เลือกสาขาที่จะโดน 422 กลับมาอยู่ดี
  // (backend: หนึ่งสาขามีผู้ดูแลได้คนเดียว) ส่ง excludeUserId เพื่อไม่ให้สาขาของคนที่กำลังแก้อยู่หายไปจากตัวเลือก
  const getAvailableBranches = (excludeUserId) => {
    const taken = new Set(
      staff.filter((u) => u.role === "admin" && u.branch_id && u.id !== excludeUserId).map((u) => u.branch_id)
    );
    return branches.map((b) => ({ ...b, takenBy: taken.has(b.id) ? staff.find((u) => u.branch_id === b.id) : null }));
  };

  return { staff, branches, getAvailableBranches, loading, error, refetch };
}
