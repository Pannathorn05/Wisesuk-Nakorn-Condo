import { IconPin, IconPhone, IconLine } from "../icons";
import { BranchMapPanel } from "./BranchMapPanel";
import "./BranchContactCard.css";

// การ์ด "ติดต่อสาขานี้" ในหน้ารายละเอียดสาขา — Line/เบอร์โทรศัพท์ ย้ายมาจาก BranchInfoCard เดิม
// (field เดียวกัน ไม่มีข้อมูลใหม่) reuse BranchMapPanel ตรง ๆ สำหรับกล่องแผนที่ย่อ เหมือนที่
// ContactBranchCard (หน้า "ติดต่อเรา") ทำไว้แล้ว ไม่เขียน logic embed/empty-state ใหม่ซ้ำ
export function BranchContactCard({ branch }) {
  return (
    <div className="branch-detail-card branch-contact-card">
      <h3>
        <IconPin width={18} height={18} /> ติดต่อสาขานี้
      </h3>
      <ul className="branch-contact-card__info">
        {branch.line_id && (
          <li>
            <IconLine width={18} height={18} />
            <span>Line Official</span>
            <strong>{branch.line_id}</strong>
          </li>
        )}
        {branch.phones?.length > 0 && (
          <li>
            <IconPhone width={18} height={18} />
            <span>เบอร์โทรศัพท์</span>
            <strong>{branch.phones.join(", ")}</strong>
          </li>
        )}
        {!branch.line_id && !branch.phones?.length && (
          <li className="branch-contact-card__empty">ยังไม่มีช่องทางติดต่อสำหรับสาขานี้</li>
        )}
      </ul>

      <div className="branch-contact-card__map">
        <BranchMapPanel branch={branch} />
      </div>
    </div>
  );
}
