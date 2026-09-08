import { Link } from "react-router-dom";
import { ImageWithFallback } from "../common/ImageWithFallback";
import { getBranchCover } from "../../assets/branchPhotos";
import { IconArrowRight, IconPin } from "../icons";
import "./OtherBranchesSection.css";

// เดิมเป็น sidebar ข้าง gallery หลัก (OtherBranchesSidebar) — ย้ายมาเป็น section เต็มความกว้างท้าย
// หน้าแทน (ข้อมูล/route เดิมทุกอย่าง แค่เปลี่ยนตำแหน่ง/รูปแบบการ์ดจากแนวตั้งเป็นแนวนอน)
export function OtherBranchesSection({ branches }) {
  if (!branches || branches.length === 0) return null;

  return (
    <div className="other-branches-section">
      <h3>สาขาอื่นๆ ของเรา</h3>
      <div className="other-branches-section__grid">
        {branches.map((branch) => (
          <Link to={`/branches/${branch.id}`} className="other-branches-section__card" key={branch.id}>
            <div className="other-branches-section__image">
              <ImageWithFallback src={getBranchCover(branch)} alt={branch.name} />
            </div>
            <div className="other-branches-section__body">
              <strong>{branch.name}</strong>
              <p>
                <IconPin width={14} height={14} /> {branch.address}
              </p>
              <span>
                ดูรายละเอียด <IconArrowRight width={14} height={14} />
              </span>
            </div>
          </Link>
        ))}
      </div>
    </div>
  );
}
