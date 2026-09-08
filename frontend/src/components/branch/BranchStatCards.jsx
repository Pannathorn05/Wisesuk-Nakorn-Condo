import { IconCalendar } from "../icons";
import { branchHasDaily, branchHasMonthly } from "../../utils/branchStay";
import { formatPrice } from "../../utils/formatPrice";
import "./BranchStatCards.css";

// การ์ดราคาแบบย่อ (รายวัน/รายเดือน) โชว์เด่นบนหัวหน้ารายละเอียดสาขา — ใช้ field เดียวกับที่
// BranchInfoCard เคยรวมไว้ในลิสต์รายละเอียดทั้งหมด (daily_price_from / monthly_price_min/max)
// แค่แยกออกมาเป็นการ์ดเด่นแทนแถวข้อความธรรมดา ไม่มี field ใหม่/ไม่ hardcode ราคาเอง
export function BranchStatCards({ branch }) {
  const hasDaily = branchHasDaily(branch);
  const hasMonthly = branchHasMonthly(branch);

  if (!hasDaily && !hasMonthly) return null;

  return (
    <div className="branch-stat-cards">
      {hasDaily && (
        <div className="branch-stat-card">
          <IconCalendar width={28} height={28} />
          <span>รายวัน</span>
          <strong>เริ่มต้น {formatPrice(branch.daily_price_from)} บาท/วัน</strong>
        </div>
      )}
      {hasMonthly && (
        <div className="branch-stat-card">
          <IconCalendar width={28} height={28} />
          <span>รายเดือน</span>
          <strong>
            เริ่มต้น {formatPrice(branch.monthly_price_min)} - {formatPrice(branch.monthly_price_max)} บาท/เดือน
          </strong>
        </div>
      )}
    </div>
  );
}
