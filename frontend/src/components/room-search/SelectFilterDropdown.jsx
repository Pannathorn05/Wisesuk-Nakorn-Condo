import { FilterDropdown } from "./FilterDropdown";
import { IconCheck } from "../icons";

// dropdown แบบ single-select ใช้ร่วมกันทั้ง filter สาขา และ filter ประเภท (stay_type)
export function SelectFilterDropdown({ icon, options, value, onChange, defaultLabel }) {
  const current = options.find((o) => o.value === value);
  const label = current ? current.label : defaultLabel;

  return (
    <FilterDropdown icon={icon} label={label}>
      {({ close }) =>
        options.map((option) => (
          <button
            type="button"
            key={option.value || "__all__"}
            className={`filter-dropdown__option ${option.value === value ? "is-selected" : ""}`}
            onClick={() => {
              onChange(option.value);
              close();
            }}
            role="option"
            aria-selected={option.value === value}
          >
            <IconCheck width={16} height={16} />
            {option.label}
          </button>
        ))
      }
    </FilterDropdown>
  );
}
