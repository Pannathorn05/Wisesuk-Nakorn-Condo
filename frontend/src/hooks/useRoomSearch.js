import { useAsync } from "./useAsync";
import { searchRooms } from "../api/roomApi";

// mapping วันที่ตรงกับ GET /rooms/search จริง (docs/openapi.yaml): stay_type = "daily" ส่งเป็นช่วง
// check_in/check_out คู่กัน (backend รองรับทั้งคู่อยู่แล้ว), stay_type = "monthly" ส่ง move_in_date
// วันเดียว, ไม่ระบุ stay_type ไม่ส่งวันที่เลย (ไม่รู้จะแปลเป็น field ไหน — DateFilterDropdown ปิดการ
// เลือกวันที่ไปเลยในเคสนี้ ไม่ปล่อยให้เลือกได้ทั้งที่ไม่มีผลกรองจริงแบบเดิม)
function buildParams(filters, page) {
  const params = {
    branch_id: filters.branchId || undefined,
    stay_type: filters.stayType || undefined,
    page,
  };

  if (filters.stayType === "daily") {
    if (filters.dateFrom) params.check_in = filters.dateFrom;
    if (filters.dateTo) params.check_out = filters.dateTo;
  }
  if (filters.stayType === "monthly" && filters.dateFrom) params.move_in_date = filters.dateFrom;

  return params;
}

export function useRoomSearch(filters, page) {
  const { data, loading, error, refetch } = useAsync(
    (signal) => searchRooms(buildParams(filters, page), signal),
    [filters.branchId, filters.stayType, filters.dateFrom, filters.dateTo, page]
  );

  return {
    rooms: data?.data || [],
    meta: data?.meta || null,
    loading,
    error,
    refetch,
  };
}
