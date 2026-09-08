import { Link } from "react-router-dom";
import { Skeleton } from "../common/Skeleton";
import { ErrorState } from "../common/ErrorState";
import { EmptyState } from "../common/EmptyState";
import { ImageWithFallback } from "../common/ImageWithFallback";
import { getBranchCover } from "../../assets/branchPhotos";
import { IconArrowRight, IconPin } from "../icons";
import "./BranchesSection.css";

export function BranchesSection({ activeBranches, loading, error, refetch }) {
  return (
    <section className="section branches-section">
      <div className="container">
        <div className="section-heading">
          <h2>สาขาของเรา</h2>
          <p>เลือกสาขาที่สะดวกใกล้คุณ</p>
        </div>

        {loading && (
          <div className="branches-section__grid">
            {[1, 2, 3].map((i) => (
              <div className="branch-card" key={i}>
                <Skeleton height="200px" radius="20px" />
                <div className="branch-card__body">
                  <Skeleton width="60%" height="1.1rem" />
                  <Skeleton width="85%" height="0.9rem" style={{ marginTop: 8 }} />
                </div>
              </div>
            ))}
          </div>
        )}

        {!loading && error && <ErrorState message="โหลดรายชื่อสาขาไม่สำเร็จ" onRetry={refetch} />}

        {!loading && !error && activeBranches.length === 0 && (
          <EmptyState message="ตอนนี้ยังไม่มีสาขาที่เปิดให้บริการ" />
        )}

        {!loading && !error && activeBranches.length > 0 && (
          <div className="branches-section__grid">
            {activeBranches.map((branch) => (
              <Link to={`/branches/${branch.id}`} className="branch-card" key={branch.id}>
                <div className="branch-card__image">
                  <ImageWithFallback src={getBranchCover(branch)} alt={branch.name} />
                </div>

                <div className="branch-card__body">
                  <h3>{branch.name}</h3>
                  <p>
                    <IconPin width={18} height={18} /> {branch.address}
                  </p>

                  <span className="branch-card__cta">
                    ดูรายละเอียด
                    <IconArrowRight width={14} height={14} />
                  </span>
                </div>
              </Link>
            ))}
          </div>
        )}
      </div>
    </section>
  );
}
