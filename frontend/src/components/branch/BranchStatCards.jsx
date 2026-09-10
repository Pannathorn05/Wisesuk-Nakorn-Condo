import { IconCalendar } from "../icons";
import { branchHasDaily, branchHasMonthly } from "../../utils/branchStay";
import { formatPrice } from "../../utils/formatPrice";
import "./BranchStatCards.css";

// การ์ดราคาแบบย่อ (รายวัน/รายเดือน) โชว์เด่นบนหัวหน้ารายละเอียดสาขา — ใช้ field เดียวกับที่
// BranchInfoCard เคยรวมไว้ในลิสต์รายละเอียดทั้งหมด (daily_price_from / monthly_price_min/max)
// แค่แยกออกมาเป็นการ์ดเด่นแทนแถวข้อความธรรมดา ไม่มี field ใหม่/ไม่ hardcode ราคาเอง
//
// โชว์ทั้ง 2 การ์ดเสมอเพื่อให้ทุกสาขาหน้าตาเหมือนกัน — สาขาที่ไม่มีห้องประเภทนั้นขาย (field ราคา
// หายไปทั้ง key ดู branchStay.jsx) เขียนบอกตรง ๆ ว่า "ไม่มีห้องพักรายวัน/รายเดือน" ห้ามโชว์เป็นขีด
// หรือเลข 0 เพราะจะทำให้เข้าใจผิดว่ามีห้องขายแต่ไม่ระบุราคา
export function BranchStatCards({ branch }) {
  const hasDaily = branchHasDaily(branch);
  const hasMonthly = branchHasMonthly(branch);

  if (!hasDaily && !hasMonthly) return null;

  return (
    <div className="branch-stat-cards">
      <div className={`branch-stat-card ${hasDaily ? "" : "is-unavailable"}`}>
        <IconCalendar width={28} height={28} />
        <span>รายวัน</span>
        <strong>
          {hasDaily ? `เริ่มต้น ${formatPrice(branch.daily_price_from)} บาท/วัน` : "ไม่มีห้องพักรายวัน"}
        </strong>
      </div>
      <div className={`branch-stat-card ${hasMonthly ? "" : "is-unavailable"}`}>
        <IconCalendar width={28} height={28} />
        <span>รายเดือน</span>
        <strong>
          {hasMonthly
            ? `เริ่มต้น ${formatPrice(branch.monthly_price_min)} - ${formatPrice(branch.monthly_price_max)} บาท/เดือน`
            : "ไม่มีห้องพักรายเดือน"}
        </strong>
      </div>
    </div>
  );
}
