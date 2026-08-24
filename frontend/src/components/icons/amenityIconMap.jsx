import {
  IconTv,
  IconWind,
  IconWifi,
  IconFridge,
  IconShower,
  IconSofa,
  IconShield,
  IconCamera,
  IconElevator,
  IconCar,
  IconWashingMachine,
  IconStore,
  IconAmenityFallback,
} from "./index.jsx";

// map ตรงกับ field `icon` ที่ GET /api/v1/amenities ส่งจริง (ตรวจครบ 12 ตัวแล้ว — ดู docs/task/frontend/02-branches.md FE-08)
// ถ้า backend เพิ่ม amenity ใหม่ที่มี icon string ไม่อยู่ในนี้ ให้ตกไปที่ fallback แทนที่จะพัง
const ICON_BY_KEY = {
  tv: IconTv,
  wind: IconWind,
  wifi: IconWifi,
  refrigerator: IconFridge,
  shower: IconShower,
  sofa: IconSofa,
  shield: IconShield,
  camera: IconCamera,
  elevator: IconElevator,
  car: IconCar,
  "washing-machine": IconWashingMachine,
  store: IconStore,
};

export function getAmenityIcon(iconKey) {
  return ICON_BY_KEY[iconKey] || IconAmenityFallback;
}
