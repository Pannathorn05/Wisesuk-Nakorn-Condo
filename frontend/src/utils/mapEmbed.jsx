// เดารูปแบบ map_url อย่างปลอดภัย — ไม่มีข้อมูลตัวอย่างจริงตอนนี้ (ว่างเปล่าทุกสาขา) จึงรองรับทั้ง 2 กรณี
// ห้ามฝัง API key แผนที่ใหม่เอง (docs/c.md ข้อ 22) — ใช้ URL ที่ backend ให้มาตรง ๆ เท่านั้น
export function isEmbeddableMapUrl(url) {
  if (!url) return false;
  return /\/maps\/embed|output=embed/i.test(url);
}

// สาขาทุกสาขายัง map_url ว่างอยู่ทั้งหมด (backend ยังไม่มี field นี้ให้มา) ระหว่างรอ สร้าง URL แผนที่
// ฝัง (embed) จากที่อยู่จริงของสาขา (branch.address จาก backend) แทน — รูปแบบ "q=...&output=embed"
// เป็นวิธีฝังแผนที่ของ Google ที่ไม่ต้องขอ/ฝัง API key เลย (ต่างจาก Maps Embed API ทางการที่ต้องมี key)
// ยังคงตรงตามกติกา "ห้ามฝัง API key ลง source code" (docs/c.md ข้อ 22) เพราะไม่มี key เข้ามาเกี่ยวข้อง
//
// ค่านี้เข้าเงื่อนไข isEmbeddableMapUrl ด้านบนพอดี (มี output=embed) — BranchMapPanel.jsx จึงเลือก
// เรนเดอร์เป็น <iframe> ให้เองตามโค้ดเดิมที่มีอยู่แล้ว ไม่ต้องเพิ่ม component/แก้ layout ใหม่ —
// พอวันไหน backend ส่ง map_url จริงมา ค่านั้นจะถูกใช้ก่อนเสมอ (ดู BranchMapPanel.jsx)
export function buildAddressMapEmbedUrl(address) {
  if (!address) return "";
  return `https://www.google.com/maps?q=${encodeURIComponent(address)}&output=embed`;
}

// พิกัด (lat, lng) จริงของแต่ละสาขา — backend มี field latitude/longitude ใน schema อยู่แล้ว
// (backend/internal/database/migrations/0001_init.sql) แต่ยังว่างทุกสาขาในข้อมูลจริงตอนนี้ ระหว่างรอ
// ใช้พิกัดจริงที่มีอยู่แล้วปักหมุดแทน — แม่นยำกว่าการเดาพิกัดจากข้อความที่อยู่ (buildAddressMapEmbedUrl
// ด้านบน อาจปักหมุดคลาดเคลื่อนถ้า Google หาที่อยู่ไม่เจอเป๊ะ) คีย์ตรงกับ branch.slug จริงจาก
// backend/cmd/seed/main.go — ถ้า backend ส่ง latitude/longitude มาจริงในอนาคต ควรสลับไปใช้ค่านั้นแทนที่นี่
//
// placeUrl คือลิงก์ Google Maps ตัวเต็มที่ maps.app.goo.gl แต่ละลิงก์ (คอมเมนต์กำกับไว้) พาไปจริง ๆ
// (resolve แล้วครั้งเดียว เก็บผลลัพธ์ไว้ตรงนี้) ต่างจาก q=lat,lng ตรงที่ผูกกับ Place ID จริงของ Google
// Business Profile นั้นเลย เปิดแล้วเจอหน้าข้อมูลสถานที่ครบ (ชื่อ/รีวิว/รูป) แบบเดียวกับกดหาชื่อเจอเป๊ะ
// ส่วน q=lat,lng (ใช้ทำ embed iframe ด้านล่าง) แม่นยำเรื่องตำแหน่งหมุดแต่เป็นแค่พิกัดดิบ ไม่ผูกกับ
// Business Profile ไหน — กดขยายจาก iframe (ไอคอนมุมขวาล่างของ Google เอง คุมไม่ได้) เลยไปเจอแค่หน้า
// "13°xx'xx.x"N ..." เฉย ๆ ไม่มีข้อมูลสถานที่ ต้องมีลิงก์ placeUrl แยกไว้กดเปิดแทนสำหรับกรณีนี้
const BRANCH_COORDS = {
  "prachauthit-45": {
    lat: 13.648103084625639,
    lng: 100.49987572042214,
    // JFXX+6W กรุงเทพมหานคร — resolve จาก https://maps.app.goo.gl/HEdef8W6H8T8GzMX7
    placeUrl:
      "https://www.google.com/maps/place/%E0%B8%AB%E0%B8%AD%E0%B8%9E%E0%B8%B1%E0%B8%81+%E0%B8%A7%E0%B8%B4%E0%B9%80%E0%B8%A8%E0%B8%A9%E0%B8%AA%E0%B8%B8%E0%B8%82%E0%B8%99%E0%B8%84%E0%B8%A3+%E0%B8%9B%E0%B8%A3%E0%B8%B0%E0%B8%8A%E0%B8%B2%E0%B8%AD%E0%B8%B8%E0%B8%97%E0%B8%B4%E0%B8%A8+45/@13.6480581,100.4998146,20z/data=!4m6!3m5!1s0x30e2a24c43a89e65:0x94faa8b07a56c2d8!8m2!3d13.6480354!4d100.4998724!16s%2Fg%2F1hm2sj1nt",
  },
  bangkhae: {
    lat: 13.70132274131765,
    lng: 100.40788426442103,
    // PC25+F5 กรุงเทพมหานคร — resolve จาก https://maps.app.goo.gl/165dk9mDxENdxdvr5
    placeUrl:
      "https://www.google.com/maps/place/%E0%B8%A7%E0%B8%B4%E0%B9%80%E0%B8%A8%E0%B8%A9%E0%B8%AA%E0%B8%B8%E0%B8%82%E0%B8%99%E0%B8%84%E0%B8%A3%E0%B8%84%E0%B8%AD%E0%B8%99%E0%B9%82%E0%B8%94/@13.7011716,100.4078789,17z/data=!4m6!3m5!1s0x30e297d9cb98c2a9:0xab2bdb6c12bbccf2!8m2!3d13.7011716!4d100.4078789!16s%2Fg%2F1hc2_8kjm",
  },
  "charoenkrung-place": {
    lat: 13.699253763025506,
    lng: 100.49522691040536,
    // MFXW+63 กรุงเทพมหานคร — resolve จาก https://maps.app.goo.gl/iRGbgcwGqUzU8faw5
    placeUrl:
      "https://www.google.com/maps/place/%E0%B9%80%E0%B8%88%E0%B8%A3%E0%B8%B4%E0%B8%8D%E0%B8%81%E0%B8%A3%E0%B8%B8%E0%B8%87+%E0%B9%80%E0%B8%9E%E0%B8%A5%E0%B8%AA/@13.7104593,100.4572956,13z/data=!4m6!3m5!1s0x30e29884cf45046f:0xdf4f42beb638aa5b!8m2!3d13.6980161!4d100.4952011!16s%2Fg%2F1pztq3z17",
  },
};

function buildCoordsMapEmbedUrl(lat, lng) {
  return `https://www.google.com/maps?q=${lat},${lng}&output=embed`;
}

// จุดเดียวที่ตัดสินใจว่าจะฝังแผนที่ (iframe) ของสาขาด้วยอะไร — ลำดับความสำคัญ:
// 1) branch.map_url จาก backend (source of truth ถ้ามีจริง)
// 2) พิกัด lat/lng จริงที่รู้อยู่แล้ว (BRANCH_COORDS ด้านบน)
// 3) เดาจากข้อความที่อยู่ (ตกมาถึงตรงนี้เฉพาะสาขาที่ยังไม่มีพิกัดใน BRANCH_COORDS)
export function getBranchMapEmbedUrl(branch) {
  if (!branch) return "";
  if (branch.map_url) return branch.map_url;
  const coords = BRANCH_COORDS[branch.slug];
  if (coords) return buildCoordsMapEmbedUrl(coords.lat, coords.lng);
  return buildAddressMapEmbedUrl(branch.address);
}

// URL สำหรับปุ่ม "เปิดใน Google Maps" (คลิกแล้วออกไปเปิดเว็บ Google Maps จริง คนละอันกับ iframe ฝัง
// ด้านบน) ใช้ placeUrl ที่ผูกกับ Business Profile จริงก่อนเสมอถ้ามี ไม่งั้นตกไปที่ลิงก์แบบเดียวกับ
// embed (พิกัดดิบ/เดาจากที่อยู่) ซึ่งเปิดแล้วจะเจอแค่หน้าพิกัดเฉย ๆ ไม่มีข้อมูลสถานที่ (เคสที่ backend
// ยังไม่มี map_url และไม่มี placeUrl ที่รู้จักในนี้)
export function getBranchOpenMapUrl(branch) {
  if (!branch) return "";
  if (branch.map_url) return branch.map_url;
  const coords = BRANCH_COORDS[branch.slug];
  if (coords?.placeUrl) return coords.placeUrl;
  return getBranchMapEmbedUrl(branch);
}
