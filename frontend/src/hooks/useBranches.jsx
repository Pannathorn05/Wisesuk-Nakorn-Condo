import { useAsync } from "./useAsync";
import { listBranches } from "../api/branchApi";

// ใช้ร่วมกันระหว่าง section "สาขาของเรา" (FE-04) และ "ประเภทห้องพัก" (FE-05)
// เพื่อไม่ต้องยิง GET /api/v1/branches ซ้ำสองครั้งในหน้าเดียว
export function useBranches() {
  const { data, loading, error, refetch } = useAsync(
    (signal) => listBranches(signal).then((res) => res.data ?? []),
    []
  );

  const activeBranches = (data || []).filter((b) => b.is_active);

  return { branches: data, activeBranches, loading, error, refetch };
}
