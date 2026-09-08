import { formatPrice } from "../../utils/formatPrice";
import { IconWallet, IconDroplet, IconBolt, IconParking, IconBuilding } from "../icons";
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

// รายวัน/รายเดือน ย้ายไปโชว์เป็นการ์ดเด่นด้านบน (BranchStatCards) และ Line/เบอร์โทร/ที่อยู่ ย้ายไป
// การ์ด "ติดต่อสาขานี้" (BranchContactCard) แล้ว เหลือแค่รายละเอียดอื่นที่ไม่มีการ์ดเฉพาะของตัวเอง
export function BranchInfoCard({ branch }) {
  const rows = [
    [IconWallet, "เงินประกัน", getDepositText(branch)],
    [IconDroplet, "ค่าน้ำ", `${formatPrice(branch.water_rate)} บาท/ยูนิต`],
    [IconBolt, "ค่าไฟ", `${formatPrice(branch.electric_rate)} บาท/ยูนิต`],
    [IconParking, "ที่จอดรถ", getParkingText(branch)],
    [IconBuilding, "จำนวนตึก", `${branch.building_count} ตึก ${branch.floor_count} ชั้น/ตึก`],
  ];

  return (
    <div className="branch-detail-card branch-detail-card--info">
      <h3>รายละเอียดเพิ่มเติม</h3>
      <dl className="branch-detail-card__list">
        {rows.map(([Icon, label, value]) => (
          <div className="branch-detail-card__row" key={label}>
            <Icon width={18} height={18} />
            <dt>{label}:</dt>
            <dd>{value}</dd>
          </div>
        ))}
      </dl>
    </div>
  );
}
