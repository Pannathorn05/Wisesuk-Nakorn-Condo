import { useParams, useLocation, useNavigate } from "react-router-dom";
import { useBranchDetail } from "../../../hooks/useBranchDetail";
import { useBranches } from "../../../hooks/useBranches";
import { Skeleton } from "../../../components/common/Skeleton";
import { ErrorState } from "../../../components/common/ErrorState";
import { BranchGallery } from "../../../components/branch/BranchGallery";
import { BranchMapPanel } from "../../../components/branch/BranchMapPanel";
import { OtherBranchesSidebar } from "../../../components/branch/OtherBranchesSidebar";
import { BranchInfoCard } from "../../../components/branch/BranchInfoCard";
import { BranchAmenitiesCard } from "../../../components/branch/BranchAmenitiesCard";
import { NearbyPlacesSection } from "../../../components/branch/NearbyPlacesSection";
import { getBranchGalleryPhotos } from "../../../assets/branchPhotos";
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
          <Skeleton width="40%" height="2rem" />
          <Skeleton width="60%" height="1rem" style={{ marginTop: 12 }} />
          <div className="branch-detail-page__media-row" style={{ marginTop: 24 }}>
            <Skeleton height="360px" />
            <Skeleton height="360px" />
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
        <h1>{branch.name}</h1>
        <p className="branch-detail-page__address">{branch.address}</p>

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

        <div className="branch-detail-page__media-row">
          {isMapTab ? (
            <BranchMapPanel mapUrl={branch.map_url} branchName={branch.name} />
          ) : (
            <BranchGallery photos={getBranchGalleryPhotos(branch.slug)} />
          )}
          <OtherBranchesSidebar branches={otherBranches} />
        </div>

        <div className="branch-detail-section">
          <BranchInfoCard branch={branch} />
          <BranchAmenitiesCard amenities={branch.amenities} />
        </div>

        <NearbyPlacesSection nearbyPlaces={branch.nearby_places} />
      </div>
    </section>
  );
}
