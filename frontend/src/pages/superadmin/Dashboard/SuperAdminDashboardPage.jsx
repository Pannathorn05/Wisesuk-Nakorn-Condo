import { useAdminDashboard } from "../../../hooks/useAdminDashboard";
import { Skeleton } from "../../../components/common/Skeleton";
import { ErrorState } from "../../../components/common/ErrorState";
import { EmptyState } from "../../../components/common/EmptyState";
import { formatActivityText, formatActivityTime } from "../../../utils/activityLog";
import { formatPrice } from "../../../utils/formatPrice";
import { IconDoor, IconClipboardClock } from "../../../components/icons";
import "./SuperAdminDashboardPage.css";

// วันที่ที่เปิดดูหน้านี้ — prototype เขียน "สรุปข้อมูลประจำวันที่ 22 มิถุนายน 2569" แต่ไม่มี field นี้
// ใน API (ดู Gap ใน docs/task/frontend/07-superadmin-dashboard.md) จึงใช้วันที่ปัจจุบันฝั่ง client
// เป็นไทย พ.ศ. ห้าม hardcode วันที่ที่เห็นในภาพ
function todayThai() {
  return new Date().toLocaleDateString("th-TH", { day: "numeric", month: "long", year: "numeric" });
}

// การ์ดตัวเลข 1 ช่อง — ตัวเศษ/ตัวส่วน map ตาม description ที่ spec กำกับไว้รายฟิลด์
function StatTile({ label, Icon, value, total, unit }) {
  return (
    <div className="stat-tile">
      <div className="stat-tile__head">
        <span>{label}</span>
        <Icon width={20} height={20} />
      </div>
      <div className="stat-tile__body">
        <strong>{formatPrice(value)}</strong>
        <span className="stat-tile__total">
          / {formatPrice(total)} {unit}
        </span>
      </div>
    </div>
  );
}

// แดชบอร์ดหัวหน้าผู้ดูแลระบบ (docs/task/frontend/07-superadmin-dashboard.md, FE-31/FE-32)
// prototype หน้า 39 ภาพ 52 — ข้อมูลทั้งหน้ามาจาก GET /api/v1/admin/dashboard ครั้งเดียว
export function SuperAdminDashboardPage() {
  const { branches, recentActivities, loading, error, refetch } = useAdminDashboard();

  return (
    <div className="sa-dashboard">
      <div className="sa-dashboard__heading">
        <h1>แดชบอร์ด</h1>
        <p>สรุปข้อมูลประจำวันที่ {todayThai()}</p>
      </div>

      {loading && (
        <div className="sa-dashboard__branches">
          {[1, 2, 3].map((i) => (
            <Skeleton key={i} height="180px" radius="20px" />
          ))}
        </div>
      )}

      {!loading && error && <ErrorState message="โหลดข้อมูลแดชบอร์ดไม่สำเร็จ" onRetry={refetch} />}

      {!loading && !error && branches.length === 0 && (
        <EmptyState message="ยังไม่มีข้อมูลสาขาให้สรุป" />
      )}

      {!loading && !error && branches.length > 0 && (
        <div className="sa-dashboard__branches">
          {branches.map((b) => (
            <section className="branch-summary-card" key={b.branch_id}>
              <h2>{b.branch_name}</h2>
              <div className="branch-summary-card__tiles">
                <StatTile
                  label="ห้องว่างรายวันวันนี้"
                  Icon={IconDoor}
                  value={b.daily_rooms_free}
                  total={b.daily_rooms_total}
                  unit="ห้อง"
                />
                <StatTile
                  label="ห้องว่างรายเดือนวันนี้"
                  Icon={IconDoor}
                  value={b.monthly_rooms_free}
                  total={b.monthly_rooms_total}
                  unit="ห้อง"
                />
                <StatTile
                  label="รอตรวจสอบการจอง"
                  Icon={IconClipboardClock}
                  value={b.pending_review}
                  total={b.bookings_total}
                  unit="รายการ"
                />
              </div>
            </section>
          ))}
        </div>
      )}

      {!loading && !error && (
        <section className="activity-box">
          <h2>กิจกรรมล่าสุด</h2>
          {recentActivities.length === 0 ? (
            <p className="activity-box__empty">ยังไม่มีกิจกรรมในระบบ</p>
          ) : (
            <ul className="activity-box__list">
              {recentActivities.map((a) => (
                <li key={a.id}>
                  <span className="activity-box__text">{formatActivityText(a)}</span>
                  <time className="activity-box__time" dateTime={a.created_at}>
                    {formatActivityTime(a.created_at)}
                  </time>
                </li>
              ))}
            </ul>
          )}
        </section>
      )}
    </div>
  );
}
