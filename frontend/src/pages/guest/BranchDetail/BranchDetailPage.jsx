import { useParams, useLocation, useNavigate, Link } from "react-router-dom";
import { useBranchDetail } from "../../../hooks/useBranchDetail";
import { useBranches } from "../../../hooks/useBranches";
import { Skeleton } from "../../../components/common/Skeleton";
import { ErrorState } from "../../../components/common/ErrorState";
import { BranchGallery } from "../../../components/branch/BranchGallery";
import { BranchMapPanel } from "../../../components/branch/BranchMapPanel";
import { BranchStatCards } from "../../../components/branch/BranchStatCards";
import { BranchInfoCard } from "../../../components/branch/BranchInfoCard";
import { BranchContactCard } from "../../../components/branch/BranchContactCard";
import { BranchAmenitiesCard } from "../../../components/branch/BranchAmenitiesCard";
import { OtherBranchesSection } from "../../../components/branch/OtherBranchesSection";
import { NearbyPlacesSection } from "../../../components/branch/NearbyPlacesSection";
import { getBranchGalleryPhotos } from "../../../assets/branchPhotos";
import { IconPin } from "../../../components/icons";
import "./BranchDetailPage.css";

// error.code จาก ApiError (docs/openapi.yaml ErrorCode) — แปลเป็นข้อความเฉพาะบริบทหน้านี้
function getErrorMessage(error) {
  if (error?.code === "bad_request") return "รหัสสาขาไม่ถูกต้อง";
  if (error?.code === "not_found") return "ไม่พบสาขาที่ต้องการ";
  return error?.message || "โหลดข้อมูลสาขาไม่สำเร็จ";
}

export function BranchDetailPage() {
  const { branchId } = useParams();
  const location = useLocation();
  const navigate = useNavigate();
  const isMapTab = location.pathname.endsWith("/map");

  const { branch, loading, error, refetch } = useBranchDetail(branchId);
  const { activeBranches } = useBranches();

  const otherBranches = activeBranches.filter((b) => b.id !== branchId).slice(0, 2);

  const goToTab = (tab) => {
    navigate(tab === "map" ? `/branches/${branchId}/map` : `/branches/${branchId}`);
  };

  if (loading) {
    return (
      <section className="section branch-detail-page">
        <div className="container">
          <div className="branch-detail-page__top">
            <Skeleton height="360px" radius="20px" />
            <div>
              <Skeleton width="40%" height="2rem" />
              <Skeleton width="60%" height="1rem" style={{ marginTop: 12 }} />
              <Skeleton height="140px" radius="12px" style={{ marginTop: 24 }} />
            </div>
          </div>
          <div className="branch-detail-section">
            <Skeleton height="240px" />
            <Skeleton height="240px" />
          </div>
        </div>
      </section>
    );
  }

  if (error) {
    return (
      <section className="section branch-detail-page">
        <div className="container">
          <ErrorState message={getErrorMessage(error)} onRetry={error?.code === "bad_request" ? undefined : refetch} />
        </div>
      </section>
    );
  }

  return (
    <section className="section branch-detail-page">
      <div className="container">
        <div className="branch-detail-page__top">
          <div className="branch-detail-page__gallery-col">
            <div className="branch-tabs" role="tablist">
              <button
                type="button"
                role="tab"
                aria-selected={!isMapTab}
                className={`branch-tabs__btn ${!isMapTab ? "is-active" : ""}`}
                onClick={() => goToTab("photos")}
              >
                รูปภาพ
              </button>
              <button
                type="button"
                role="tab"
                aria-selected={isMapTab}
                className={`branch-tabs__btn ${isMapTab ? "is-active" : ""}`}
                onClick={() => goToTab("map")}
              >
                แผนที่
              </button>
            </div>

            {isMapTab ? (
              <BranchMapPanel branch={branch} />
            ) : (
              <BranchGallery photos={getBranchGalleryPhotos(branch.slug)} />
            )}
          </div>

          <div className="branch-detail-page__header-col">
            <span className="branch-detail-page__badge">
              <IconPin width={14} height={14} /> สาขา
            </span>
            <h1>{branch.name}</h1>
            <p className="branch-detail-page__address">
              <IconPin width={16} height={16} /> {branch.address}
            </p>

            <BranchStatCards branch={branch} />
            <BranchInfoCard branch={branch} />

            <div className="branch-detail-page__actions">
              <Link to={`/rooms?branch_id=${branch.id}`} className="btn btn-primary">
                จองห้องพัก
              </Link>
              {branch.phones?.length > 0 ? (
                <a href={`tel:${branch.phones[0]}`} className="btn btn-outline">
                  ติดต่อสอบถาม
                </a>
              ) : (
                <Link to="/contact" className="btn btn-outline">
                  ติดต่อสอบถาม
                </Link>
              )}
            </div>
          </div>
        </div>

        <div className="branch-detail-section">
          <BranchAmenitiesCard amenities={branch.amenities} />
          <BranchContactCard branch={branch} />
        </div>

        <NearbyPlacesSection nearbyPlaces={branch.nearby_places} />

        <OtherBranchesSection branches={otherBranches} />
      </div>
    </section>
  );
}
