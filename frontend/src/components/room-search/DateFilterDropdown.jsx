import { useState } from "react";
import { FilterDropdown } from "./FilterDropdown";
import { buildMonthGrid, getMonthLabel, THAI_WEEKDAYS, toDateKey } from "../../utils/calendar";
import { IconCalendar, IconChevronLeft, IconChevronRight } from "../icons";
import "./DateFilterDropdown.css";

// key จาก buildMonthGrid เป็น "YYYY-MM-DD" (ค.ศ.) เรียงตามตัวอักษรได้ตรงกับลำดับวันที่จริงอยู่แล้ว
// (ใช้เทียบ key < key ตอนเลือกช่วงด้านล่างได้เลย ไม่ต้องแปลงเป็น Date) แปลงเป็น วว/ดด/ปป (พ.ศ. 2 หลัก)
// สำหรับโชว์ในป้ายกำกับเท่านั้น
function formatShort(key) {
  const [y, m, d] = key.split("-");
  const beYear = String(Number(y) + 543).slice(-2);
  return `${d}/${m}/${beYear}`;
}

// ปฏิทินเลือกวันที่ของฟิลเตอร์ค้นหาห้องพัก — โหมดขึ้นกับ stay_type ที่เลือกไว้ (backend
// GET /rooms/search รับ check_in/check_out คู่กัน หรือ move_in_date เดี่ยว ๆ ดู docs/openapi.yaml):
//   - "daily" (รายวัน): เลือกเป็นช่วงเข้าพัก-ออก 2 คลิก (check_in, check_out)
//   - "monthly" (รายเดือน): เลือกวันเดียว (move_in_date)
//   - ยังไม่เลือกประเภท: ปิดใช้งานฟิลเตอร์นี้ไปเลย เพราะไม่รู้จะ map ค่าที่เลือกไปเป็น field ไหนของ
//     backend — ของเดิมปล่อยให้เลือกได้แต่ไม่ถูกส่งไปกรองผลจริง (ดู Gap เดิมใน
//     docs/task/frontend/03-room-search.md) ผู้ใช้เข้าใจผิดว่า "เลือกวันที่ไม่มีผล" ได้ ปิดไปเลยชัดเจนกว่า
export function DateFilterDropdown({ stayType, from, to, onChange }) {
  const isRange = stayType === "daily";
  const disabled = !stayType;

  const [viewDate, setViewDate] = useState(() => new Date(from || Date.now()));

  const label = disabled
    ? "เลือกประเภทก่อน"
    : isRange
    ? from && to
      ? `${formatShort(from)} - ${formatShort(to)}`
      : from
      ? `${formatShort(from)} - วว/ดด/ปป`
      : "วว/ดด/ปป - วว/ดด/ปป"
    : from
    ? formatShort(from)
    : "วว/ดด/ปป";

  const days = buildMonthGrid(viewDate);
  const todayKey = toDateKey(new Date());

  // สถานะการเลือกช่วงแบบ 2 คลิก: ยังไม่มีจุดเริ่ม (from ว่าง) → คลิกแรกตั้ง from แล้วรอคลิกที่สอง
  // (from มีแต่ to ว่าง) → คลิกที่สองตั้ง to แล้วปิด เว้นแต่คลิกวันที่ก่อน from เอง ถือว่าเริ่มช่วงใหม่
  // (from ใหม่ = วันนั้น) ครบคู่แล้ว (from และ to มีทั้งคู่) คลิกใด ๆ ถือเป็นการเริ่มช่วงใหม่เช่นกัน
  function handlePick(key, close) {
    if (!isRange) {
      onChange(key, "");
      close();
      return;
    }
    if (from && !to) {
      if (key >= from) {
        onChange(from, key);
        close();
      } else {
        onChange(key, "");
      }
      return;
    }
    onChange(key, "");
  }

  return (
    <FilterDropdown icon={<IconCalendar width={18} height={18} />} label={label} disabled={disabled}>
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

          {isRange && (
            <p className="date-picker__hint">{from && !to ? "เลือกวันที่ออก" : "เลือกวันที่เข้าพัก"}</p>
          )}

          <div className="date-picker__weekdays">
            {THAI_WEEKDAYS.map((w) => (
              <span key={w}>{w.slice(0, 3)}</span>
            ))}
          </div>

          <div className="date-picker__grid">
            {days.map((day) => {
              const isStart = day.key === from;
              const isEnd = isRange && day.key === to;
              const isInRange = isRange && from && to && day.key > from && day.key < to;
              return (
                <div
                  className={`date-picker__cell ${isInRange ? "is-in-range" : ""} ${
                    isStart ? "is-range-start" : ""
                  } ${isEnd ? "is-range-end" : ""}`}
                  key={day.key}
                >
                  <button
                    type="button"
                    className={`date-picker__day ${!day.inCurrentMonth ? "is-outside" : ""} ${
                      isStart || isEnd ? "is-selected" : ""
                    } ${day.key === todayKey ? "is-today" : ""}`}
                    onClick={() => handlePick(day.key, close)}
                  >
                    {day.date.getDate()}
                  </button>
                </div>
              );
            })}
          </div>

          {(from || to) && (
            <button
              type="button"
              className="date-picker__clear"
              onClick={() => {
                onChange("", "");
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
