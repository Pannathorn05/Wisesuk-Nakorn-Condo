import { IconPin, IconPhone, IconLine } from "../icons";
import { BranchMapPanel } from "../branch/BranchMapPanel";
import "./ContactBranchCard.css";

// การ์ดข้อมูลติดต่อ 1 สาขา ใช้ในหน้า "ติดต่อเรา" (docs/task/frontend/04-contact.md, FE-19)
// reuse BranchMapPanel เดิมจาก FE-12 ตรง ๆ สำหรับกล่องแผนที่ — ไม่เขียน logic embed/empty-state ใหม่
export function ContactBranchCard({ branch }) {
  return (
    <div className="contact-card">
      <h3>{branch.name}</h3>
      <ul className="contact-card__info">
        <li>
          <IconPin width={18} height={18} />
          <span>{branch.address}</span>
        </li>
        {branch.phones?.length > 0 && (
          <li>
            <IconPhone width={18} height={18} />
            <span>{branch.phones.join(", ")}</span>
          </li>
        )}
        {branch.line_id && (
          <li>
            <IconLine width={18} height={18} />
            <span>{branch.line_id}</span>
          </li>
        )}
        {branch.map_url && (
          <li>
            <IconPin width={18} height={18} />
            <a href={branch.map_url} target="_blank" rel="noopener noreferrer">
              {branch.map_url}
            </a>
          </li>
        )}
      </ul>

      <BranchMapPanel mapUrl={branch.map_url} branchName={branch.name} />
    </div>
  );
}
