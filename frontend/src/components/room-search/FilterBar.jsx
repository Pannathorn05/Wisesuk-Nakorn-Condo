import { BranchFilterDropdown } from "./BranchFilterDropdown";
import { SelectFilterDropdown } from "./SelectFilterDropdown";
import { DateFilterDropdown } from "./DateFilterDropdown";
import { IconBed } from "../icons";
import "./FilterBar.css";

// ตรงกับภาพ 16 ใน prototype — นี่คือ stay_type ไม่ใช่ room_type (ดู Gap ใน docs/task/frontend/03-room-search.md)
const STAY_TYPE_OPTIONS = [
  { value: "", label: "ทุกประเภท" },
  { value: "daily", label: "รายวัน" },
  { value: "monthly", label: "รายเดือน" },
];

export function FilterBar({ filters, onFilterChange, branches, branchesLoading, branchesError, onRetryBranches }) {
  return (
    <div className="filter-bar">
      <BranchFilterDropdown
        branches={branches}
        loading={branchesLoading}
        error={branchesError}
        onRetry={onRetryBranches}
        value={filters.branchId}
        onChange={(branchId) => onFilterChange({ branchId })}
      />
      <SelectFilterDropdown
        icon={<IconBed width={18} height={18} />}
        options={STAY_TYPE_OPTIONS}
        value={filters.stayType}
        defaultLabel="ทุกประเภท"
        onChange={(stayType) => onFilterChange({ stayType, date: "" })}
      />
      <DateFilterDropdown value={filters.date} onChange={(date) => onFilterChange({ date })} />
    </div>
  );
}
