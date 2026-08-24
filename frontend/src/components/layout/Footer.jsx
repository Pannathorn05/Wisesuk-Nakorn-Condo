import { Link } from "react-router-dom";
import { IconBuilding, IconLine, IconPhone } from "../icons";
import { useBranches } from "../../hooks/useBranches";
import { Skeleton } from "../common/Skeleton";
import { ErrorState } from "../common/ErrorState";
import "./Footer.css";

const MENU_LINKS = [
  { to: "/", label: "หน้าหลัก" },
  { to: "/branches", label: "สาขาของเรา" },
  { to: "/rooms", label: "ค้นหาห้องพัก" },
  { to: "/contact", label: "ติดต่อเรา" },
];

export function Footer() {
  const { activeBranches, loading, error, refetch } = useBranches();

  return (
    <footer className="site-footer">
      <div className="container site-footer__grid">
        <div className="site-footer__brand">
          <div className="site-footer__brand-title">
            <IconBuilding width={28} height={28} />
            <strong>วิเศษสุขนครคอนโด</strong>
          </div>
          <p>และหอพักในเครือ</p>
        </div>

        <div className="site-footer__col">
          <h4>เมนู</h4>
          <ul>
            {MENU_LINKS.map((link) => (
              <li key={link.to}>
                <Link to={link.to}>{link.label}</Link>
              </li>
            ))}
          </ul>
        </div>

        <div className="site-footer__col">
          <h4>สาขา</h4>
          {loading && (
            <ul className="site-footer__skeleton-list">
              {[1, 2, 3].map((i) => (
                <li key={i}>
                  <Skeleton width="80%" height="0.9rem" />
                </li>
              ))}
            </ul>
          )}
          {error && <ErrorState message="โหลดรายชื่อสาขาไม่สำเร็จ" onRetry={refetch} />}
          {!loading && !error && activeBranches.length === 0 && <p>ยังไม่มีสาขาที่เปิดให้บริการ</p>}
          {!loading && !error && activeBranches.length > 0 && (
            <ul>
              {activeBranches.map((b) => (
                <li key={b.id}>
                  <Link to={`/branches/${b.id}`}>{b.name}</Link>
                </li>
              ))}
            </ul>
          )}
        </div>

        <div className="site-footer__col">
          <h4>ติดต่อเรา</h4>
          {loading && <Skeleton width="70%" height="0.9rem" />}
          {!loading && !error && (
            <ul className="site-footer__contact">
              {activeBranches[0]?.line_id && (
                <li>
                  <IconLine width={18} height={18} /> {activeBranches[0].line_id}
                </li>
              )}
              {activeBranches.map((b) =>
                b.phones?.length ? (
                  <li key={b.id}>
                    <IconPhone width={18} height={18} />
                    <span>
                      {b.phones.join(", ")}
                      <small>{b.name}</small>
                    </span>
                  </li>
                ) : null
              )}
            </ul>
          )}
        </div>
      </div>

      <div className="site-footer__bottom">
        <div className="container">© {new Date().getFullYear()} วิเศษสุขนคร คอนโด และหอพักในเครือ สงวนลิขสิทธิ์</div>
      </div>
    </footer>
  );
}
