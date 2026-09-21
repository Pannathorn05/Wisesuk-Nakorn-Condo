import { useAsync } from "./useAsync";
import { getAdminDashboard } from "../api/adminApi";

// แดชบอร์ดผู้ดูแล/หัวหน้าผู้ดูแล (docs/task/frontend/07-superadmin-dashboard.md, FE-31)
// ยิงครั้งเดียวได้ทั้งการ์ดสรุปรายสาขาและกล่องกิจกรรมล่าสุด ไม่ต้องยิง /admin/activity-logs ซ้ำในหน้านี้
export function useAdminDashboard() {
  const { data, loading, error, refetch } = useAsync((signal) => getAdminDashboard(signal).then((res) => res.data), []);

  return {
    branches: data?.branches || [],
    recentActivities: data?.recent_activities || [],
    loading,
    error,
    refetch,
  };
}
