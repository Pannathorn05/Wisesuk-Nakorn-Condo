import { useState } from "react";
import { FilterDropdown } from "./FilterDropdown";
import { buildMonthGrid, getMonthLabel, THAI_WEEKDAYS, toDateKey } from "../../utils/calendar";
import { IconCalendar, IconChevronLeft, IconChevronRight } from "../icons";
import "./DateFilterDropdown.css";

// ปฏิทินเลือกวันเดียว (ไม่ใช่ date range) ตรงกับ prototype ภาพ 17 — ดู docs/task/frontend/03-room-search.md
export function DateFilterDropdown({ value, onChange }) {
  const [viewDate, setViewDate] = useState(() => (value ? new Date(value) : new Date()));

  const label = value
    ? new Date(value).toLocaleDateString("th-TH", { day: "2-digit", month: "2-digit", year: "2-digit" })
    : "วว/ดด/ปป";

  const days = buildMonthGrid(viewDate);
  const todayKey = toDateKey(new Date());

  return (
    <FilterDropdown icon={<IconCalendar width={18} height={18} />} label={label}>
      {({ close }) => (
        <div className="date-picker">
          <div className="date-picker__header">
            <button
              type="button"
              onClick={() => setViewDate((d) => new Date(d.getFullYear(), d.getMonth() - 1, 1))}
              aria-label="เดือนก่อนหน้า"
            >
              <IconChevronLeft width={18} height={18} />
            </button>
            <strong>
              {getMonthLabel(viewDate)} {viewDate.getFullYear() + 543}
            </strong>
            <button
              type="button"
              onClick={() => setViewDate((d) => new Date(d.getFullYear(), d.getMonth() + 1, 1))}
              aria-label="เดือนถัดไป"
            >
              <IconChevronRight width={18} height={18} />
            </button>
          </div>

          <div className="date-picker__weekdays">
            {THAI_WEEKDAYS.map((w) => (
              <span key={w}>{w.slice(0, 3)}</span>
            ))}
          </div>

          <div className="date-picker__grid">
            {days.map((day) => (
              <button
                type="button"
                key={day.key}
                className={`date-picker__day ${!day.inCurrentMonth ? "is-outside" : ""} ${
                  day.key === value ? "is-selected" : ""
                } ${day.key === todayKey ? "is-today" : ""}`}
                onClick={() => {
                  onChange(day.key);
                  close();
                }}
              >
                {day.date.getDate()}
              </button>
            ))}
          </div>

          {value && (
            <button
              type="button"
              className="date-picker__clear"
              onClick={() => {
                onChange("");
                close();
              }}
            >
              ล้างวันที่
            </button>
          )}
        </div>
      )}
    </FilterDropdown>
  );
}
