import { useBranches } from "../../../hooks/useBranches";
import { Skeleton } from "../../../components/common/Skeleton";
import { ErrorState } from "../../../components/common/ErrorState";
import { EmptyState } from "../../../components/common/EmptyState";
import { BranchListRow } from "../../../components/branch/BranchListRow";
import "./BranchesPage.css";

export function BranchesPage() {
  const { activeBranches, loading, error, refetch } = useBranches();

  return (
    <section className="section branches-page">
      <div className="container">
        <div className="section-heading">
          <h1>สาขาในเครือทั้งหมด</h1>
          <p>
            เครือข่ายหอพักวิเศษสุขนคร ครอบคลุมทั่วกรุงเทพฯ
            <br />
            พร้อมให้บริการคุณภาพเดียวกันทุกสาขา
          </p>
        </div>

        {loading && (
          <div>
            {[1, 2, 3].map((i) => (
              <div className="branch-row-skeleton" key={i}>
                <Skeleton height="260px" radius="0" />
                <div className="branch-row-skeleton__body">
                  <Skeleton width="40%" height="1.4rem" />
                  <Skeleton width="70%" height="1rem" style={{ marginTop: 12 }} />
                  <Skeleton width="55%" height="1rem" style={{ marginTop: 8 }} />
                </div>
              </div>
            ))}
          </div>
        )}

        {!loading && error && <ErrorState message="โหลดรายชื่อสาขาไม่สำเร็จ" onRetry={refetch} />}

        {!loading && !error && activeBranches.length === 0 && (
          <EmptyState message="ตอนนี้ยังไม่มีสาขาที่เปิดให้บริการ" />
        )}

        {!loading &&
          !error &&
          activeBranches.length > 0 &&
          activeBranches.map((branch) => <BranchListRow branch={branch} key={branch.id} />)}
      </div>
    </section>
  );
}
