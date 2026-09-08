// ไอคอนเส้น (stroke, currentColor) เซ็ตเดียวใช้ทั้งเว็บ แทนที่จะติดตั้ง icon library เพิ่ม
// เพราะจำนวนไอคอนที่ต้องใช้รอบนี้มีไม่มาก (docs/c.md ข้อ 6: อย่าติดตั้ง package ที่ไม่จำเป็นจริง ๆ)
import React from "react";

const base = {
  width: 24,
  height: 24,
  viewBox: "0 0 24 24",
  fill: "none",
  stroke: "currentColor",
  strokeWidth: 1.6,
  strokeLinecap: "round",
  strokeLinejoin: "round",
};

export function IconKeycard(props) {
  return (
    <svg {...base} {...props}>
      <rect x="3" y="6" width="18" height="12" rx="2" />
      <circle cx="8" cy="12" r="2" />
      <line x1="13" y1="10" x2="18" y2="10" />
      <line x1="13" y1="14" x2="18" y2="14" />
    </svg>
  );
}

export function IconGuard(props) {
  return (
    <svg {...base} {...props}>
      <path d="M12 3l7 3v5c0 4.5-3 8-7 10-4-2-7-5.5-7-10V6z" />
      <path d="M9 12l2 2 4-4" />
    </svg>
  );
}

export function IconParking(props) {
  return (
    <svg {...base} {...props}>
      <rect x="3" y="4" width="18" height="16" rx="2" />
      <path d="M9 16V8h3.5a2.5 2.5 0 010 5H9" />
    </svg>
  );
}

export function IconBathroom(props) {
  return (
    <svg {...base} {...props}>
      <path d="M4 12h16v2a5 5 0 01-5 5H9a5 5 0 01-5-5v-2z" />
      <path d="M7 12V6a2 2 0 012-2 2 2 0 012 2" />
      <line x1="4" y1="19" x2="4" y2="21" />
      <line x1="20" y1="19" x2="20" y2="21" />
    </svg>
  );
}

export function IconChevronLeft(props) {
  return (
    <svg {...base} {...props}>
      <polyline points="15 18 9 12 15 6" />
    </svg>
  );
}

export function IconChevronRight(props) {
  return (
    <svg {...base} {...props}>
      <polyline points="9 18 15 12 9 6" />
    </svg>
  );
}

export function IconArrowRight(props) {
  return (
    <svg {...base} {...props}>
      <line x1="4" y1="12" x2="19" y2="12" />
      <polyline points="13 6 19 12 13 18" />
    </svg>
  );
}

export function IconMenu(props) {
  return (
    <svg {...base} {...props}>
      <line x1="3" y1="6" x2="21" y2="6" />
      <line x1="3" y1="12" x2="21" y2="12" />
      <line x1="3" y1="18" x2="21" y2="18" />
    </svg>
  );
}

export function IconClose(props) {
  return (
    <svg {...base} {...props}>
      <line x1="6" y1="6" x2="18" y2="18" />
      <line x1="18" y1="6" x2="6" y2="18" />
    </svg>
  );
}

export function IconPhone(props) {
  return (
    <svg {...base} {...props}>
      <path d="M5 4h3l2 5-2.5 1.5a11 11 0 005 5L14 13l5 2v3a2 2 0 01-2 2C9.5 20 4 14.5 4 7a2 2 0 011-2z" />
    </svg>
  );
}

export function IconLine(props) {
  return (
    <svg {...base} {...props}>
      <rect x="3" y="5" width="18" height="14" rx="4" />
      <path d="M7 9v6M11 9v6M11 9l3 6V9M17 9v6h2" />
    </svg>
  );
}

export function IconBed(props) {
  return (
    <svg {...base} {...props}>
      <path d="M3 19v-8a2 2 0 012-2h14a2 2 0 012 2v8" />
      <path d="M3 19v2M21 19v2" />
      <path d="M3 13h18" />
      <path d="M6 13V9a1 1 0 011-1h4a1 1 0 011 1v4" />
    </svg>
  );
}

export function IconChevronDown(props) {
  return (
    <svg {...base} {...props}>
      <polyline points="6 9 12 15 18 9" />
    </svg>
  );
}

export function IconCheck(props) {
  return (
    <svg {...base} {...props}>
      <polyline points="5 12 10 17 19 7" />
    </svg>
  );
}

export function IconCalendar(props) {
  return (
    <svg {...base} {...props}>
      <rect x="3" y="5" width="18" height="16" rx="2" />
      <line x1="3" y1="10" x2="21" y2="10" />
      <line x1="8" y1="3" x2="8" y2="7" />
      <line x1="16" y1="3" x2="16" y2="7" />
    </svg>
  );
}

export function IconPin(props) {
  return (
    <svg {...base} {...props}>
      <path d="M12 21s7-6.6 7-11.5A7 7 0 105 9.5C5 14.4 12 21 12 21z" />
      <circle cx="12" cy="9.5" r="2.3" />
    </svg>
  );
}

export function IconStar(props) {
  return (
    <svg {...base} {...props}>
      <path d="M12 3l2.6 5.6 6.1.7-4.5 4.2 1.2 6-5.4-3-5.4 3 1.2-6-4.5-4.2 6.1-.7z" />
    </svg>
  );
}

export function IconWallet(props) {
  return (
    <svg {...base} {...props}>
      <rect x="3" y="6" width="18" height="13" rx="2" />
      <path d="M3 10h18" />
      <circle cx="16" cy="14.5" r="1.4" />
    </svg>
  );
}

export function IconDroplet(props) {
  return (
    <svg {...base} {...props}>
      <path d="M12 3s6 6.8 6 11a6 6 0 11-12 0c0-4.2 6-11 6-11z" />
    </svg>
  );
}

export function IconBolt(props) {
  return (
    <svg {...base} {...props}>
      <path d="M13 2 5 14h6l-1 8 9-12h-6z" />
    </svg>
  );
}

/* ===== ไอคอน amenity — ต้องครบตาม `icon` field ที่ GET /api/v1/amenities ส่งจริง 12 ตัว ===== */

export function IconTv(props) {
  return (
    <svg {...base} {...props}>
      <rect x="3" y="5" width="18" height="12" rx="2" />
      <line x1="8" y1="21" x2="16" y2="21" />
      <line x1="12" y1="17" x2="12" y2="21" />
    </svg>
  );
}

export function IconWind(props) {
  return (
    <svg {...base} {...props}>
      <path d="M3 8h11a2.5 2.5 0 10-2.5-2.5" />
      <path d="M3 12h15a2.5 2.5 0 11-2.5 2.5" />
      <path d="M3 16h9a2.5 2.5 0 102.5 2.5" />
    </svg>
  );
}

export function IconWifi(props) {
  return (
    <svg {...base} {...props}>
      <path d="M2 8.5a15.5 15.5 0 0120 0" />
      <path d="M5.5 12.2a10.8 10.8 0 0113 0" />
      <path d="M9 15.8a5.8 5.8 0 016 0" />
      <line x1="12" y1="19" x2="12" y2="19.01" />
    </svg>
  );
}

export function IconFridge(props) {
  return (
    <svg {...base} {...props}>
      <rect x="5" y="2" width="14" height="20" rx="2" />
      <line x1="5" y1="10" x2="19" y2="10" />
      <line x1="8" y1="5" x2="8" y2="7" />
      <line x1="8" y1="13" x2="8" y2="15" />
    </svg>
  );
}

export function IconShower(props) {
  return (
    <svg {...base} {...props}>
      <path d="M6 9a6 6 0 0112 0" />
      <line x1="4" y1="9" x2="20" y2="9" />
      <line x1="7" y1="13" x2="7" y2="13.01" />
      <line x1="12" y1="13" x2="12" y2="13.01" />
      <line x1="17" y1="13" x2="17" y2="13.01" />
      <line x1="7" y1="17" x2="7" y2="17.01" />
      <line x1="12" y1="17" x2="12" y2="17.01" />
      <line x1="17" y1="17" x2="17" y2="17.01" />
    </svg>
  );
}

export function IconSofa(props) {
  return (
    <svg {...base} {...props}>
      <path d="M4 12V9a2 2 0 012-2h12a2 2 0 012 2v3" />
      <path d="M3 12h18v4a1 1 0 01-1 1H4a1 1 0 01-1-1v-4z" />
      <line x1="4" y1="17" x2="4" y2="20" />
      <line x1="20" y1="17" x2="20" y2="20" />
    </svg>
  );
}

export function IconShield(props) {
  return (
    <svg {...base} {...props}>
      <path d="M12 3l7 3v5c0 4.5-3 8-7 10-4-2-7-5.5-7-10V6z" />
      <path d="M9 12l2 2 4-4" />
    </svg>
  );
}

export function IconCamera(props) {
  return (
    <svg {...base} {...props}>
      <path d="M4 8h3l1.5-2h7L17 8h3a1 1 0 011 1v9a1 1 0 01-1 1H4a1 1 0 01-1-1V9a1 1 0 011-1z" />
      <circle cx="12" cy="13" r="3.2" />
    </svg>
  );
}

export function IconElevator(props) {
  return (
    <svg {...base} {...props}>
      <rect x="5" y="2" width="14" height="20" rx="1.5" />
      <polyline points="10 9 12 7 14 9" />
      <polyline points="10 15 12 17 14 15" />
    </svg>
  );
}

export function IconCar(props) {
  return (
    <svg {...base} {...props}>
      <path d="M4 16V11l2-5h12l2 5v5" />
      <path d="M2 16h20v2a1 1 0 01-1 1h-2a1 1 0 01-1-1v-1H6v1a1 1 0 01-1 1H3a1 1 0 01-1-1v-2z" />
      <circle cx="7" cy="16" r="1.3" />
      <circle cx="17" cy="16" r="1.3" />
    </svg>
  );
}

export function IconWashingMachine(props) {
  return (
    <svg {...base} {...props}>
      <rect x="4" y="3" width="16" height="18" rx="2" />
      <line x1="7" y1="6" x2="9" y2="6" />
      <circle cx="12" cy="14" r="5" />
      <circle cx="12" cy="14" r="2" />
    </svg>
  );
}

export function IconStore(props) {
  return (
    <svg {...base} {...props}>
      <path d="M4 9l1-5h14l1 5" />
      <path d="M4 9a2.2 2.2 0 004.4 0 2.2 2.2 0 004.4 0 2.2 2.2 0 004.4 0 2.2 2.2 0 004.4 0" />
      <path d="M5 9v10a1 1 0 001 1h12a1 1 0 001-1V9" />
      <line x1="9" y1="20" x2="9" y2="14" />
      <line x1="15" y1="20" x2="15" y2="14" />
    </svg>
  );
}

export function IconAmenityFallback(props) {
  return (
    <svg {...base} {...props}>
      <circle cx="12" cy="12" r="9" />
      <path d="M9 12l2 2 4-4" />
    </svg>
  );
}

export function IconBuilding(props) {
  return (
    <svg {...base} {...props}>
      <rect x="4" y="3" width="10" height="18" rx="1" />
      <rect x="14" y="9" width="6" height="12" rx="1" />
      <line x1="7" y1="7" x2="7" y2="7.01" />
      <line x1="10" y1="7" x2="10" y2="7.01" />
      <line x1="7" y1="11" x2="7" y2="11.01" />
      <line x1="10" y1="11" x2="10" y2="11.01" />
      <line x1="7" y1="15" x2="7" y2="15.01" />
      <line x1="10" y1="15" x2="10" y2="15.01" />
    </svg>
  );
}

/* ===== ไอคอนสำหรับหน้า Login/Register (docs/task/frontend/05-login-register.md, FE-21) ===== */

export function IconMail(props) {
  return (
    <svg {...base} {...props}>
      <rect x="3" y="5" width="18" height="14" rx="2" />
      <path d="M3 6.5l9 6 9-6" />
    </svg>
  );
}

export function IconLockClosed(props) {
  return (
    <svg {...base} {...props}>
      <rect x="4" y="10" width="16" height="10" rx="2" />
      <path d="M7.5 10V7a4.5 4.5 0 019 0v3" />
      <circle cx="12" cy="15" r="1.4" />
    </svg>
  );
}

export function IconEye(props) {
  return (
    <svg {...base} {...props}>
      <path d="M2 12s3.5-7 10-7 10 7 10 7-3.5 7-10 7-10-7-10-7z" />
      <circle cx="12" cy="12" r="3" />
    </svg>
  );
}

export function IconEyeOff(props) {
  return (
    <svg {...base} {...props}>
      <path d="M3 3l18 18" />
      <path d="M10.6 5.2A10.6 10.6 0 0112 5c6.5 0 10 7 10 7a15.6 15.6 0 01-3.4 4.4M6.6 6.6C3.7 8.5 2 12 2 12s3.5 7 10 7a9.9 9.9 0 004.4-1" />
      <path d="M9.9 10a3 3 0 004.2 4.2" />
    </svg>
  );
}

export function IconUser(props) {
  return (
    <svg {...base} {...props}>
      <circle cx="12" cy="8" r="4" />
      <path d="M4 21c0-4.4 3.6-8 8-8s8 3.6 8 8" />
    </svg>
  );
}
