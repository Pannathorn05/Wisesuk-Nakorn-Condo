import { Link } from "react-router-dom";
import { ImageWithFallback } from "../common/ImageWithFallback";
import { getBranchCover } from "../../assets/branchPhotos";
import { IconArrowRight } from "../icons";
import "./OtherBranchesSidebar.css";

export function OtherBranchesSidebar({ branches }) {
  if (!branches || branches.length === 0) return null;

  return (
    <aside className="other-branches">
      <h3>สาขาอื่น</h3>
      {branches.map((branch) => (
        <Link to={`/branches/${branch.id}`} className="other-branches__card" key={branch.id}>
          <div className="other-branches__image">
            <ImageWithFallback src={getBranchCover(branch)} alt={branch.name} />
          </div>
          <div className="other-branches__body">
            <strong>{branch.name}</strong>
            <p>{branch.address}</p>
            <span>
              ดูรายละเอียด <IconArrowRight width={14} height={14} />
            </span>
          </div>
        </Link>
      ))}
    </aside>
  );
}
