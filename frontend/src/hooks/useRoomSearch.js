import { useAsync } from "./useAsync";
import { searchRooms } from "../api/roomApi";

// mapping วันที่ตามกฎใน Gap ของ docs/task/frontend/03-room-search.md:
// UI มีช่องวันที่ช่องเดียว (ไม่มี check_in/check_out แยก) — ส่งเป็น check_in ตอน daily,
// move_in_date ตอน monthly, ไม่ส่งเลยตอนไม่ระบุ stay_type (ไม่รู้จะแปลเป็น field ไหน)
function buildParams(filters, page) {
  const params = {
    branch_id: filters.branchId || undefined,
    stay_type: filters.stayType || undefined,
    page,
  };

  if (filters.date && filters.stayType === "daily") params.check_in = filters.date;
  if (filters.date && filters.stayType === "monthly") params.move_in_date = filters.date;

  return params;
}

export function useRoomSearch(filters, page) {
  const { data, loading, error, refetch } = useAsync(
    (signal) => searchRooms(buildParams(filters, page), signal),
    [filters.branchId, filters.stayType, filters.date, page]
  );

  return {
    rooms: data?.data || [],
    meta: data?.meta || null,
    loading,
    error,
    refetch,
  };
}
