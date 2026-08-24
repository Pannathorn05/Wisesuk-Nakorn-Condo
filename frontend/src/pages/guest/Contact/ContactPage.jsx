import { useBranches } from "../../../hooks/useBranches";
import { Skeleton } from "../../../components/common/Skeleton";
import { ErrorState } from "../../../components/common/ErrorState";
import { EmptyState } from "../../../components/common/EmptyState";
import { ContactBranchCard } from "../../../components/contact/ContactBranchCard";
import "./ContactPage.css";

// หน้าติดต่อเรา (docs/task/frontend/04-contact.md, FE-19) — แทนที่ ComingSoonPage placeholder
// "พร้อมให้บริการทุกวัน 8:30 - 17:30 น." เป็น business copy คงที่ ไม่มี field API รองรับ (ดู Gap ในไฟล์ task)
export function ContactPage() {
  const { activeBranches, loading, error, refetch } = useBranches();

  return (
    <section className="section contact-page">
      <div className="container">
        <div className="section-heading">
          <h1>ติดต่อเรา</h1>
          <p>พร้อมให้บริการทุกวัน 8:30 - 17:30 น.</p>
        </div>

        {loading && (
          <div className="contact-page__grid">
            {[1, 2, 3].map((i) => (
              <div className="contact-card-skeleton" key={i}>
                <Skeleton width="60%" height="1.2rem" />
                <Skeleton width="90%" height="0.9rem" style={{ marginTop: 12 }} />
                <Skeleton width="70%" height="0.9rem" style={{ marginTop: 8 }} />
                <Skeleton width="50%" height="0.9rem" style={{ marginTop: 8 }} />
                <Skeleton height="260px" style={{ marginTop: 16 }} />
              </div>
            ))}
          </div>
        )}

        {!loading && error && <ErrorState message="โหลดข้อมูลสาขาไม่สำเร็จ" onRetry={refetch} />}

        {!loading && !error && activeBranches.length === 0 && (
          <EmptyState message="ตอนนี้ยังไม่มีสาขาที่เปิดให้บริการ" />
        )}

        {!loading && !error && activeBranches.length > 0 && (
          <div className="contact-page__grid">
            {activeBranches.map((branch) => (
              <ContactBranchCard branch={branch} key={branch.id} />
            ))}
          </div>
        )}
      </div>
    </section>
  );
}
