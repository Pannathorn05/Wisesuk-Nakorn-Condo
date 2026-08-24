import { FilterDropdown } from "./FilterDropdown";
import { Skeleton } from "../common/Skeleton";
import { ErrorState } from "../common/ErrorState";
import { IconCheck, IconPin } from "../icons";

// filter สาขา — มี loading/error/empty ของตัวเองเพราะดึงจาก GET /api/v1/branches จริง
// (ดู FE-14 ใน docs/task/frontend/03-room-search.md)
export function BranchFilterDropdown({ branches, loading, error, onRetry, value, onChange }) {
  const current = branches.find((b) => b.id === value);
  const label = current ? current.name : "ทุกสาขา";

  return (
    <FilterDropdown icon={<IconPin width={18} height={18} />} label={label}>
      {({ close }) => (
        <>
          {loading && (
            <div className="filter-dropdown__loading">
              <Skeleton width="100%" height="1rem" />
              <Skeleton width="80%" height="1rem" />
              <Skeleton width="60%" height="1rem" />
            </div>
          )}

          {!loading && error && <ErrorState message="โหลดรายชื่อสาขาไม่สำเร็จ" onRetry={onRetry} />}

          {!loading && !error && (
            <>
              <button
                type="button"
                className={`filter-dropdown__option ${!value ? "is-selected" : ""}`}
                onClick={() => {
                  onChange("");
                  close();
                }}
              >
                <IconCheck width={16} height={16} />
                ทุกสาขา
              </button>
              {branches.length === 0 && <p className="filter-dropdown__empty">ยังไม่มีสาขาที่เปิดให้บริการ</p>}
              {branches.map((branch) => (
                <button
                  type="button"
                  key={branch.id}
                  className={`filter-dropdown__option ${branch.id === value ? "is-selected" : ""}`}
                  onClick={() => {
                    onChange(branch.id);
                    close();
                  }}
                >
                  <IconCheck width={16} height={16} />
                  {branch.name}
                </button>
              ))}
            </>
          )}
        </>
      )}
    </FilterDropdown>
  );
}
