import { Link } from "react-router-dom";
import { ImageWithFallback } from "../common/ImageWithFallback";
import { getBranchCover } from "../../assets/branchPhotos";
import { getBranchHighlightChips } from "../../utils/branchChips";
import { getBranchStayAvailabilityText } from "../../utils/branchStay";
import { IconPin, IconPhone, IconLine, IconBed } from "../icons";
import "./BranchListRow.css";

export function BranchListRow({ branch }) {
  const chips = getBranchHighlightChips(branch);

  return (
    <article className="branch-row">
      <div className="branch-row__image">
        <ImageWithFallback src={getBranchCover(branch)} alt={branch.name} />
      </div>

      <div className="branch-row__body">
        <h3>{branch.name}</h3>

        <div className="branch-row__meta">
          <p>
            <IconPin width={18} height={18} /> {branch.address}
          </p>
          {branch.phones?.length > 0 && (
            <p>
              <IconPhone width={18} height={18} /> {branch.phones.join(", ")}
            </p>
          )}
          {branch.line_id && (
            <p>
              <IconLine width={18} height={18} /> {branch.line_id}
            </p>
          )}
        </div>

        {chips.length > 0 && (
          <div className="branch-row__chips">
            {chips.map((chip) => (
              <span className="branch-row__chip" key={chip}>
                {chip}
              </span>
            ))}
          </div>
        )}

        {branch.tagline && (
          <p className="branch-row__tagline">
            <IconPin width={16} height={16} /> {branch.tagline}
          </p>
        )}

        <hr className="branch-row__divider" />

        <div className="branch-row__footer">
          <span className="branch-row__availability">
            <IconBed width={18} height={18} /> {getBranchStayAvailabilityText(branch)}
          </span>
          <Link to={`/branches/${branch.id}`} className="btn btn-primary">
            ดูรายละเอียด →
          </Link>
        </div>
      </div>
    </article>
  );
}
