import { apiClient } from "./client";

// GET /api/v1/branches -> { data: Branch[] }
export function listBranches(signal) {
  return apiClient.get("/branches", { signal });
}

// GET /api/v1/branches/:branchID -> { data: BranchDetail }
export function getBranch(branchId, signal) {
  return apiClient.get(`/branches/${branchId}`, { signal });
}
