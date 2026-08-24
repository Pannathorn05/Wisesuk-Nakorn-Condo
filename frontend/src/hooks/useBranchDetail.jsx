import { useAsync } from "./useAsync";
import { getBranch } from "../api/branchApi";

export function useBranchDetail(branchId) {
  const { data, loading, error, refetch } = useAsync(
    (signal) => getBranch(branchId, signal).then((res) => res.data),
    [branchId]
  );

  return { branch: data, loading, error, refetch };
}
