import { useMemo, useState } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { useBranches } from "../../../hooks/useBranches";
import { useRoomSearch } from "../../../hooks/useRoomSearch";
import { isLoggedIn } from "../../../utils/authStorage";
import { buildBranchesMap } from "../../../utils/roomChips";
import { Skeleton } from "../../../components/common/Skeleton";
import { ErrorState } from "../../../components/common/ErrorState";
import { EmptyState } from "../../../components/common/EmptyState";
import { Pagination } from "../../../components/common/Pagination";
import { FilterBar } from "../../../components/room-search/FilterBar";
import { RoomResultCard } from "../../../components/room-search/RoomResultCard";
import { GuestBookingModal } from "../../../components/room-search/GuestBookingModal";
import "./RoomsPage.css";

function getErrorMessage(error) {
  if (error?.code === "bad_request") return "เงื่อนไขค้นหาไม่ถูกต้อง";
  return error?.message || "ค้นหาห้องพักไม่สำเร็จ";
}

export function RoomsPage() {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();

  // อ่านค่าเริ่มต้นจาก query string (ปุ่ม "ดูรายละเอียด" ของ StayTypeSection หน้า Home ลิงก์มาด้วย ?stay_type=)
  const [filters, setFilters] = useState({
    branchId: searchParams.get("branch_id") || "",
    stayType: searchParams.get("stay_type") || "",
    date: searchParams.get("date") || "",
  });
  const [page, setPage] = useState(Number(searchParams.get("page")) || 1);
  const [modalOpen, setModalOpen] = useState(false);

  const { activeBranches, loading: branchesLoading, error: branchesError, refetch: refetchBranches } = useBranches();
  const { rooms, meta, loading, error, refetch } = useRoomSearch(filters, page);

  const branchesById = useMemo(() => buildBranchesMap(activeBranches), [activeBranches]);

  function syncQuery(nextFilters, nextPage) {
    const params = {};
    if (nextFilters.branchId) params.branch_id = nextFilters.branchId;
    if (nextFilters.stayType) params.stay_type = nextFilters.stayType;
    if (nextFilters.date) params.date = nextFilters.date;
    if (nextPage > 1) params.page = String(nextPage);
    setSearchParams(params, { replace: true });
  }

  function handleFilterChange(patch) {
    const next = { ...filters, ...patch };
    setFilters(next);
    setPage(1);
    syncQuery(next, 1);
  }

  function handlePageChange(nextPage) {
    setPage(nextPage);
    syncQuery(filters, nextPage);
    window.scrollTo({ top: 0, behavior: "smooth" });
  }

  function handleBook(room) {
    if (isLoggedIn()) {
      navigate(`/rooms/${room.id}`);
    } else {
      setModalOpen(true);
    }
  }

  return (
    <section className="section rooms-page">
      <div className="container">
        <div className="section-heading rooms-page__heading">
          <h1>ค้นหาห้องพัก</h1>
          <p>เลือกสาขา อาคาร และประเภทห้องที่ต้องการ</p>
        </div>

        <FilterBar
          filters={filters}
          onFilterChange={handleFilterChange}
          branches={activeBranches}
          branchesLoading={branchesLoading}
          branchesError={branchesError}
          onRetryBranches={refetchBranches}
        />

        {loading && (
          <div className="room-results-grid">
            {[1, 2, 3, 4, 5, 6].map((i) => (
              <div className="room-card" key={i}>
                <Skeleton height="200px" radius="0" />
                <div className="room-card__body">
                  <Skeleton width="60%" height="1.1rem" />
                  <Skeleton width="80%" height="0.9rem" style={{ marginTop: 8 }} />
                </div>
              </div>
            ))}
          </div>
        )}

        {!loading && error && <ErrorState message={getErrorMessage(error)} onRetry={refetch} />}

        {!loading && !error && rooms.length === 0 && (
          <EmptyState message="ไม่พบห้องพักที่ตรงกับเงื่อนไขที่เลือก ลองปรับตัวกรองแล้วค้นหาใหม่อีกครั้ง" />
        )}

        {!loading && !error && rooms.length > 0 && (
          <>
            <div className="room-results-grid">
              {rooms.map((room) => (
                <RoomResultCard room={room} branchesById={branchesById} onBook={handleBook} key={room.id} />
              ))}
            </div>
            <Pagination page={meta?.page || 1} totalPages={meta?.total_pages || 1} onChange={handlePageChange} />
          </>
        )}
      </div>

      <GuestBookingModal open={modalOpen} onClose={() => setModalOpen(false)} />
    </section>
  );
}
