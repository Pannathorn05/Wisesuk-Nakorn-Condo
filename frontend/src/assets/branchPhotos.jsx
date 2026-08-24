import { API_ORIGIN } from "../api/client";

// รูปห้อง/สาขา — backend field รูป (Branch.cover_image_url, Room.image_url) ยังว่างเปล่าทุกรายการ
// ในข้อมูลจริงตอนนี้ (ยังไม่มีการอัปโหลดผ่าน admin upload API)
//
// ระหว่างรอ ใช้รูปจริงของแต่ละสาขาที่มีอยู่แล้วใน backend/uploads/<branch-slug>/ เป็น fallback
// แทน — backend เสิร์ฟโฟลเดอร์นี้แบบ read-only อยู่แล้วที่ GET /uploads/*filepath
// (backend/internal/routes/routes.go) ไม่ต้องแก้ backend ใด ๆ ทั้งสิ้น แค่ประกอบ URL ไปเรียก
// คีย์ของ map ตรงกับ branch.slug จริงจาก backend/cmd/seed/main.go
//
// ลำดับความสำคัญยังเหมือนเดิม: API (cover_image_url / image_url) มาก่อนเสมอถ้ามีจริง (backend
// เป็น source of truth) ตกมาที่ชุดนี้เป็นลำดับสอง ถ้า backend ไม่มีรูปและไม่มีไฟล์ fallback นี้ด้วย
// ค่อยปล่อยให้ ImageWithFallback (src/components/common/ImageWithFallback.jsx) แสดง placeholder ไอคอนแทน

// ประกอบ URL ไปยังไฟล์ใน backend/uploads/<slug>/<relPath> — เข้ารหัสทีละ segment ไม่ใช่ทั้งพาธ
// เพราะบางไฟล์มีทั้งช่องว่างและอยู่ในโฟลเดอร์ย่อย (เช่น "views/v-45-1 (1).jpg")
function uploadUrl(slug, relPath) {
  const encodedPath = relPath.split("/").map(encodeURIComponent).join("/");
  return `${API_ORIGIN}/uploads/${slug}/${encodedPath}`;
}

// bangkhae และ charoenkrung-place เก็บไฟล์แบบแบน ตั้งชื่อเรียงเลข 1..n ต่อเนื่อง
// firstNumber (ถ้าระบุ) ดันไฟล์เลขนั้นขึ้นเป็นรูปแรกแทน (ใช้เป็นรูปปก) ที่เหลือเรียงเลข 1..n ตามเดิม
function numberedPhotos(slug, prefix, count, firstNumber) {
  const numbers = Array.from({ length: count }, (_, i) => i + 1);
  if (firstNumber) {
    const rest = numbers.filter((n) => n !== firstNumber);
    numbers.splice(0, numbers.length, firstNumber, ...rest);
  }
  return numbers.map((n) => uploadUrl(slug, `${prefix}-${n}.jpg`));
}

const BRANCH_PHOTOS = {
  bangkhae: numberedPhotos("bangkhae", "bk", 45, 23),
  "charoenkrung-place": numberedPhotos("charoenkrung-place", "ckp", 44, 41),
  // prachauthit-45 เก็บแยกโฟลเดอร์ย่อยตามหมวด (views = ภายนอกอาคาร, days/month = ห้องพักรายวัน/รายเดือน)
  // ชื่อไฟล์ไม่เรียงเลขต่อเนื่อง (มี "(1)" ต่อท้ายจากการดาวน์โหลดซ้ำ) จึงต้องระบุทีละไฟล์
  "prachauthit-45": [
    "views/v-45-1 (1).jpg",
    "views/v-45-3.jpg",
    "views/v-45-2 (1).jpg",
    "days/d-45-1.jpg",
    "days/d-45-2 (3).jpg",
    "month/m-45-1 (1).jpg",
    "month/m-45-2 (1).jpg",
    "month/m-45-3 (1).jpg",
    "month/m-45-4 (1).jpg",
    "month/m-45-5 (1).jpg",
    "month/m-45-6 (1).jpg",
    "month/m-45-7 (1).jpg",
    "month/m-45-8 (1).jpg",
    "month/m-45-9.jpg",
    "month/m-45-10 (1).jpg",
  ].map((relPath) => uploadUrl("prachauthit-45", relPath)),
};

export function getBranchGalleryPhotos(slug) {
  return BRANCH_PHOTOS[slug] || [];
}

// รวมรูปของทุกสาขาเป็น pool เดียว — ใช้กับจุดที่อยากสุ่มรูปโชว์แบบไม่ผูกกับสาขาใดสาขาหนึ่ง
// (เช่น hero carousel หน้าแรก) ไม่ใช่การ hardcode ข้อมูลใหม่ แค่รวม array ที่มีอยู่แล้วด้านบน
export function getAllBranchPhotos() {
  return Object.values(BRANCH_PHOTOS).flat();
}

export function getBranchCoverPhoto(slug) {
  return BRANCH_PHOTOS[slug]?.[0] || null;
}

// ใช้ทุกที่ที่ต้องโชว์รูปปกของสาขา — API มาก่อนเสมอถ้ามีจริง (backend เป็น source of truth)
// ตกมาที่ asset ในโปรเจกต์เป็นลำดับสอง ไม่ใช่การ hardcode ข้อมูลที่ backend มีอยู่แล้ว
export function getBranchCover(branch) {
  return branch?.cover_image_url || getBranchCoverPhoto(branch?.slug) || null;
}

// ใช้กับการ์ดผลการค้นหาห้อง (ห้องพักรายวัน/รายเดือน) — Room.image_url ยังว่างทุกห้องในข้อมูลจริง
// เหมือนกัน แต่ส่วนนี้ "ยังไม่ต้องดึงรูป" ตามที่แจ้งไว้ (ต่างจากรูปสาขาที่ดึงจาก backend/uploads
// แล้ว) เลยไม่ตกไปใช้รูปปกของสาขาเป็น fallback แบบ getBranchCover — ปล่อยให้ ImageWithFallback
// แสดง placeholder ไอคอนไปก่อนจนกว่าจะมีรูปห้องจริง (branchesById ยังรับไว้เผื่อกลับมาใช้ทีหลัง)
export function getRoomImage(room, branchesById) {
  return room.image_url || null;
}
