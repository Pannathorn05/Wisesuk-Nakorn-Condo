import { Link } from "react-router-dom";
import { Skeleton } from "../common/Skeleton";
import { ErrorState } from "../common/ErrorState";
import { EmptyState } from "../common/EmptyState";
import { IconBed, IconArrowRight } from "../icons";
import { formatPrice } from "../../utils/formatPrice";
import "./StayTypeSection.css";

// ราคาคำนวณจาก field daily_price_from / monthly_price_min / monthly_price_max
// ของ GET /api/v1/branches (aggregate ข้ามสาขา is_active) — ไม่มี endpoint แยกตามประเภทห้อง
// ดูเหตุผลใน docs/task/frontend/01-homepage.md (FE-05)
function aggregatePriceRanges(branches) {
  const withDaily = branches.filter((b) => typeof b.daily_price_from === "number" && b.daily_price_from > 0);
  const withMonthly = branches.filter(
    (b) => typeof b.monthly_price_min === "number" && b.monthly_price_min > 0
  );

  return {
    daily: withDaily.length
      ? { min: Math.min(...withDaily.map((b) => b.daily_price_from)) }
      : null,
    monthly: withMonthly.length
      ? {
          min: Math.min(...withMonthly.map((b) => b.monthly_price_min)),
          max: Math.max(...withMonthly.map((b) => b.monthly_price_max ?? b.monthly_price_min)),
        }
      : null,
  };
}

export function StayTypeSection({ activeBranches, loading, error, refetch }) {
  const ranges = activeBranches ? aggregatePriceRanges(activeBranches) : null;
  const hasAnyPrice = ranges && (ranges.daily || ranges.monthly);

  return (
    <section className="section room-type-section">
      <div className="container">
        <div className="section-heading">
          <h2>ประเภทห้องพัก</h2>
          <p>เลือกประเภทที่คุณต้องการ</p>
        </div>

        {loading && (
          <div className="room-type-section__grid">
            {[1, 2].map((i) => (
              <div className="room-type-card" key={i}>
                <Skeleton height="320px" radius="0" />
                <div className="room-type-card__body">
                  <Skeleton width="50%" height="1.1rem" />
                  <Skeleton width="70%" height="0.9rem" style={{ marginTop: 8 }} />
                </div>
              </div>
            ))}
          </div>
        )}

        {!loading && error && <ErrorState message="โหลดข้อมูลราคาห้องพักไม่สำเร็จ" onRetry={refetch} />}

        {!loading && !error && !hasAnyPrice && <EmptyState message="ยังไม่มีข้อมูลราคาห้องพัก" />}

        {!loading && !error && hasAnyPrice && (
          <div className="room-type-section__grid">
            {ranges.daily && (
              <Link to="/rooms?stay_type=daily" className="room-type-card">
                <div className="room-type-card__image">
                  <IconBed width={64} height={64} />
                </div>
                <div className="room-type-card__body">
                  <h3>ห้องพักรายวัน</h3>
                  <p>ราคาเริ่มต้น {formatPrice(ranges.daily.min)} บาท/คืน</p>
                  <span className="room-type-card__link">
                    ดูรายละเอียด <IconArrowRight width={16} height={16} />
                  </span>
                </div>
              </Link>
            )}
            {ranges.monthly && (
              <Link to="/rooms?stay_type=monthly" className="room-type-card">
                <div className="room-type-card__image">
                  <IconBed width={64} height={64} />
                </div>
                <div className="room-type-card__body">
                  <h3>ห้องพักรายเดือน</h3>
                  <p>
                    ราคา {formatPrice(ranges.monthly.min)} - {formatPrice(ranges.monthly.max)} บาท/เดือน
                  </p>
                  <span className="room-type-card__link">
                    ดูรายละเอียด <IconArrowRight width={16} height={16} />
                  </span>
                </div>
              </Link>
            )}
          </div>
        )}
      </div>
    </section>
  );
}
