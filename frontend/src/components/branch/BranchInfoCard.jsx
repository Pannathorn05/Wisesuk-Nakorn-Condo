import { formatPrice } from "../../utils/formatPrice";
import { branchHasDaily, branchHasMonthly } from "../../utils/branchStay";
import "./BranchDetailCards.css";

function getDepositText(branch) {
  if (branch.deposit > 0) return `${formatPrice(branch.deposit)} บาท`;
  if (branch.advance_payment > 0) return `จ่ายล่วงหน้า ${formatPrice(branch.advance_payment)} บาท`;
  // ยังไม่มี branch ไหนตั้งค่าตัวเลขจริง (เป็น 0 ทั้งหมด) — ใช้ข้อความสูตรคำนวณตาม business rule ทั่วไปของระบบไปก่อน
  // ดู Gap ใน docs/task/frontend/02-branches.md
  return "จ่ายล่วงหน้า 3 เดือนของราคาห้องพัก";
}

function getParkingText(branch) {
  const hasParking = (branch.amenities || []).some((a) => a.code === "parking");
  // ไม่มี field ราคาที่จอดรถจริงใน backend ตอนนี้ — บอกได้แค่มี/ไม่มี ห้ามเดาราคา (ดู Gap)
  return hasParking ? "มีที่จอดรถ" : "ไม่มีที่จอดรถ";
}

export function BranchInfoCard({ branch }) {
  const rows = [
    ["รายวัน", branchHasDaily(branch) ? `เริ่มต้น ${formatPrice(branch.daily_price_from)} บาท/วัน` : "-"],
    [
      "รายเดือน",
      branchHasMonthly(branch)
        ? `เริ่มต้น ${formatPrice(branch.monthly_price_min)} - ${formatPrice(branch.monthly_price_max)} บาท/เดือน`
        : "-",
    ],
    ["เงินประกัน", getDepositText(branch)],
    ["ค่าน้ำ", `${formatPrice(branch.water_rate)} บาท/ยูนิต`],
    ["ค่าไฟ", `${formatPrice(branch.electric_rate)} บาท/ยูนิต`],
    ["ที่จอดรถ", getParkingText(branch)],
    ["จำนวนตึก", `${branch.building_count} ตึก ${branch.floor_count} ชั้น/ตึก`],
    ["Line", branch.line_id || "-"],
    ["เบอร์โทรศัพท์", branch.phones?.length ? branch.phones.join(", ") : "-"],
    ["ที่อยู่", branch.address],
  ];

  return (
    <div className="branch-detail-card">
      <h3>รายละเอียด</h3>
      <dl className="branch-detail-card__list">
        {rows.map(([label, value]) => (
          <div className="branch-detail-card__row" key={label}>
            <dt>{label}:</dt>
            <dd>{value}</dd>
          </div>
        ))}
      </dl>
    </div>
  );
}
