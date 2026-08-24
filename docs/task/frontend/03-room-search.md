# Phase FE-3 — ค้นหาห้องพัก + ผลลัพธ์ (Guest)

อ้างอิง: `frontend/prototype/ตัวอย่าง Website v2.pdf` หน้า 9–12 (รูปภาพที่ 13–17)

| รูปภาพ | หน้า PDF | เนื้อหา |
|---|---|---|
| 13 | 9 | หน้าค้นหาห้องพัก: filter bar (สาขา/ประเภท/วันที่) + ผลลัพธ์เป็นการ์ด 9 ใบ |
| 14 | 10 | กด "จองเลย" ตอนยังไม่ login → modal "โปรดลงทะเบียน / เข้าสู่ระบบ เพื่อทำการจอง" |
| 15 | 11 | dropdown "ทุกสาขา" เปิดอยู่ — list เลือกได้ทีละสาขา (single-select, มี ✓) |
| 16 | 11 | dropdown "ทุกประเภท" เปิดอยู่ — list มีแค่ 3 ตัวเลือก: ทุกประเภท / รายวัน / รายเดือน (นี่คือ `stay_type` **ไม่ใช่** room_type) |
| 17 | 12 | dropdown วันที่เปิดอยู่ — ปฏิทินเลือกได้ **วันเดียว** (ไม่ใช่ date range ไม่มีช่อง check_in/check_out แยกกัน) |

ตรวจแล้ว: ภาพ 15–17 เป็นหน้าเดียวกับภาพ 13 แค่เปิด dropdown คนละตัว ไม่ใช่หน้าแยก — จึงแตกเป็น **1 หน้า** (filter bar + grid ผลลัพธ์) ไม่ใช่ 5 หน้า

DoD ทุก task ที่ดึงข้อมูลจาก API ต้องผ่าน **E2E checklist 4 สถานะ: loading / empty / error (มีปุ่มลองใหม่) / success**
ขึ้นกับ `docs/c.md`, `.claude/CLAUDE.md`: ห้ามแก้ backend, ห้าม hardcode ข้อมูลที่ API มีจริง, ห้าม mock แทน API จริง, **ห้ามให้ Guest ที่ยังไม่ login เข้า booking flow ของ Member โดยตรง** (docs/c.md ข้อ 19)

---

## Section map จาก prototype

**Header**: title "ค้นหาห้องพัก" + subtext static "เลือกสาขา อาคาร และประเภทห้องที่ต้องการ"

**Filter bar** (แถบขาวมนเดียว มี 3 dropdown):
1. 📍 สาขา — ตัวเลือก: ทุกสาขา (default) / รายชื่อสาขาจริงจาก `GET /api/v1/branches` (single-select)
2. 🛏 ประเภท — ตัวเลือกตายตัวแค่ 3 อัน: ทุกประเภท (default) / รายวัน / รายเดือน → ผูกกับ `stay_type` ของ API
3. 📅 วันที่ — ปฏิทินเลือกวันเดียว placeholder "วว/ดด/ปป" (ไม่มีช่องวันที่ 2 ช่องสำหรับ check_in/check_out แยกกัน)

**ผลลัพธ์**: grid การ์ดห้อง (การ์ดละ: รูป, หัวข้อ "ห้องพักรายวัน"/"ห้องพักรายเดือน" ตาม `stay_type`, ชื่อสาขา, แถบ chip, เส้นคั่น, ราคา + หน่วย, บรรทัดค่าน้ำ/ไฟเล็ก ๆ, ปุ่ม "จองเลย" ทึบมุมขวา)

**Modal ข้อจำกัดของ Guest** (ภาพ 14): กด "จองเลย" ตอนยังไม่ login → popup กลาง "โปรดลงทะเบียน / เข้าสู่ระบบ เพื่อทำการจอง" + ปุ่ม "เข้าสู่ระบบ" (ไป `/login`) / "สมัครสมาชิก" (ไป `/register`) + ปุ่มปิด (×)

---

## Gap ที่เจอ — เช็คกับ backend ที่รันจริงแล้ว (curl `localhost:8080`)

- **Dropdown "ประเภท" คือ `stay_type` ไม่ใช่ room_type** — ยืนยันจากภาพ 16 ที่ dropdown มีแค่ "ทุกประเภท/รายวัน/รายเดือน" 3 ตัวเลือกตายตัว ไม่ใช่รายชื่อ room_type ("ห้องแอร์"/"ห้องพัดลม"/"ห้องเปล่า") ที่มี 9 แบบต่างกันในแต่ละสาขา — **ห้ามผูกกับ `GET /api/v1/room-types` โดยเข้าใจผิดว่าเป็น dropdown นี้**
- **chip ใต้หัวข้อการ์ดในผลลัพธ์ไม่มี field เดียวที่ตรงกับ prototype เป๊ะ ๆ** — `Room` object จาก `GET /api/v1/rooms/search` มีแค่ `room_type_name, size_sqm, description` ต่อห้อง **ไม่มี** amenity list ต่อห้อง (เช่น "เตียง", "ตู้เย็น", "โต๊ะ" ที่เห็นใน prototype เป็นข้อมูลสมมติระดับห้อง ไม่มี field จริงรองรับ) → ใช้เท่าที่มีจริง: (ก) `room_type_name` เป็น chip แรกถ้ามี (ข) `{size_sqm} ตร.ม.` ถ้ามี (ค) สิ่งอำนวยความสะดวกจริงของ**สาขา**นั้น (`branch.amenities` — ต้อง join เอง จาก `room.branch_id` เทียบกับผลของ `GET /api/v1/branches` เพราะ endpoint ค้นหาห้องไม่แนบ amenities ของสาขามาให้) **ห้ามยัด "เตียง/ตู้เย็น/โต๊ะ" เป็น field ตายตัวเพราะไม่มีข้อมูลจริงรองรับระดับห้อง**
- **ปฏิทินในหน้า prototype เลือกได้วันเดียว ไม่มีช่อง check_in/check_out แยก** แต่ API ต้องการ `check_in`+`check_out` คู่กันตอน `stay_type=daily` (ตาม AC-12) และต้องการแค่ `move_in_date` ตอน `stay_type=monthly` — **ตัดสินใจ**: วันที่ที่เลือกจาก dropdown เดียวนี้ ส่งเป็น `move_in_date` เสมอเมื่อ `stay_type=monthly`, และส่งเป็น `check_in` อย่างเดียวเมื่อ `stay_type=daily` (**ไม่ส่ง `check_out`** เพราะ UI ไม่มีช่องให้เลือกวันที่สอง — backend อนุญาตให้ omit ได้เพราะ schema ไม่ได้ require ทั้งคู่) ถ้าเลือก "ทุกประเภท" (ไม่ระบุ stay_type) ให้ไม่ส่งวันที่เลย เพราะไม่รู้ว่าจะแปลเป็น field ไหน
- **ข้อมูลจริงตอนนี้มีห้องแค่ 6 ห้อง** (`meta.total_items = 6` ตอนไม่ filter) ไม่ใช่ 9 ใบแบบ prototype — เป็นเรื่องปกติของข้อมูล seed ไม่ต้องพยายามทำให้ตรง 9 ใบ, `total_pages = 1` เสมอตอนนี้เพราะ `page_size` default 20 > 6 แต่ยังต้องเขียน pagination UI รองรับกรณีข้อมูลเยอะขึ้นในอนาคต (ตาม AC-12: `page_size` สูงสุดจริงคือ 100 แม้ขอ 500)
- **`stay_type` ผิดค่า/ไม่รู้จัก → 400** (`"stay_type ต้องเป็น daily หรือ monthly"`), **`branch_id` ที่ไม่ใช่ UUID → 400** เช็คแล้วตรงกับ error envelope เดิมที่ `api/client.js` แปลงให้อยู่แล้ว ไม่ต้องเขียน error handling เพิ่ม
- **ปุ่ม "จองเลย" ตอน login แล้ว**: prototype ไม่ได้โชว์ (ภาพ 14 โชว์แค่เคสยังไม่ login) และ booking flow ของ Member ยังไม่อยู่ใน scope รอบนี้ (docs/c.md ข้อ 5, 19) — **ตัดสินใจ**: ถ้า login อยู่แล้วให้พาไปหน้า `/rooms/:roomID` (รายละเอียดห้อง) แทนที่จะเข้า booking flow ตรง ๆ — หน้ารายละเอียดห้องยังไม่ได้แตก task ในรอบนี้ ให้ทำเป็น route placeholder (`ComingSoonPage`) ไปก่อนเหมือน route อื่นที่ยังไม่ implement

---

**FE-14 · Filter bar (สาขา/ประเภท/วันที่) — static UI ก่อน ยังไม่ต้องยิง API**
- ทำ: 3 dropdown ตาม section map — สาขาใช้ `GET /api/v1/branches` จริง (loading/error ของ dropdown นี้เอง), ประเภทเป็น 3 ตัวเลือกตายตัว (static, ไม่มี API), วันที่เป็นปฏิทินเลือกวันเดียว (component ใหม่หรือ `<input type="date">` ของเบราว์เซอร์ก็ได้ถ้าลดความซับซ้อนได้จริงและยังใช้งานได้ตามหน้าที่เดิม — ถ้าเลือกทำปฏิทิน custom ให้ยึด UX ปีเดือนวันแบบ prototype ไม่ต้องทำ range) · อ่าน query string เริ่มต้นจาก URL (`?stay_type=`, `?branch_id=`) มาตั้งค่า filter เริ่มต้น เพราะปุ่มจากหน้า Home (`StayTypeSection`) ลิงก์มาด้วย query นี้อยู่แล้ว
- ขึ้นกับ: FE-00 (จาก `01-homepage.md`)
- DoD: dropdown สาขา — E2E checklist ครบ (loading/error มีปุ่มลองใหม่/empty ไม่มีสาขา active/success รายชื่อตรงกับ API) · เลือก filter แล้ว state เปลี่ยนจริง (ยังไม่ต้องยิง search) · เข้าเว็บด้วย `/rooms?stay_type=daily` แล้ว dropdown ประเภทต้องตั้งต้นเป็น "รายวัน" ให้อัตโนมัติ

**FE-15 · เชื่อมผลการค้นหาจริง + การ์ดผลลัพธ์**
- ทำ: ยิง `GET /api/v1/rooms/search` ทุกครั้งที่ filter เปลี่ยน (branch_id, stay_type, check_in หรือ move_in_date ตามกฎใน Gap) · การ์ดผลลัพธ์: รูป (ใช้ asset จริงถ้ามี — reuse `getBranchGalleryPhotos`/placeholder ตาม branch, `room.image_url` ยังว่างทุกห้องเหมือนกับ field รูปอื่น ๆ ที่เจอมาก่อนหน้านี้), หัวข้อ "ห้องพักรายวัน"/"ห้องพักรายเดือน" ตาม `stay_type`, ชื่อสาขา (`branch_name`), chip ตามกฎใน Gap (room_type_name, size_sqm, amenities จริงของสาขา — ต้องดึง `GET /api/v1/branches` มา join เอง), ราคา + หน่วย (บาท/วัน หรือ บาท/เดือน ตาม stay_type), บรรทัดค่าน้ำ-ไฟ
- ขึ้นกับ: FE-14
- DoD: E2E checklist ครบ — loading = skeleton grid ระหว่างเปลี่ยน filter, empty = ไม่พบห้องตรงเงื่อนไข ต้องเห็นข้อความชัดเจนไม่ใช่หน้าขาว, error = `stay_type`/`branch_id` ผิดรูปแบบ (400) ต้องเห็นข้อความไทย + ปุ่มลองใหม่, success = จำนวนการ์ด/ราคา/สาขาตรงกับ API เป๊ะ (เทียบกับ `meta.total_items` จริงตอนนั้น ไม่ใช่ 9 ใบตาม prototype) · เปลี่ยน filter แล้วผลลัพธ์อัปเดตจริงไม่ค้างของเก่า

**FE-16 · Modal ข้อจำกัด Guest ตอนกด "จองเลย"**
- ทำ: ปุ่ม "จองเลย" ทุกใบเช็ค `isLoggedIn()` (util เดิมจาก `01-homepage.md`) — ยังไม่ login → เปิด modal ตามภาพ 14 เป๊ะ ๆ (หัวข้อ, ปุ่ม เข้าสู่ระบบ/สมัครสมาชิก, ปุ่มปิด, ปิดได้ด้วยคลิก backdrop/กด Esc ด้วย) — login อยู่แล้ว → พาไป `/rooms/:roomID` (route placeholder ใหม่ ยังไม่ implement หน้าจริง)
- ขึ้นกับ: FE-15
- DoD: checklist ปุ่ม "เข้าสู่ระบบ"/"สมัครสมาชิก" ในโมดัลนำทางไป `/login`/`/register` ถูกต้อง (คงค้าง query กลับมาที่หน้านี้ได้ก็ดีแต่ไม่บังคับในรอบนี้) · ปิดโมดัลได้ 3 ทาง (ปุ่ม ×, คลิกนอกกล่อง, Esc) แล้วกลับมาอยู่หน้าเดิมไม่มีอะไรเปลี่ยน · จำลอง login แล้วกด "จองเลย" ไม่เห็น modal เปลี่ยนไปเป็นนำทางแทน

**FE-17 · Pagination**
- ทำ: ใช้ `meta.page`/`meta.total_pages`/`meta.total_items` จาก response ควบคุมปุ่มก่อนหน้า/ถัดไป (หรือเลขหน้า) — ต้องใช้งานได้จริงแม้ข้อมูลตอนนี้มีหน้าเดียว (ปุ่มต้อง disable ถูกต้องตอน `total_pages<=1`)
- ขึ้นกับ: FE-15
- DoD: E2E checklist ผูกกับ call เดียวกับ FE-15 · `page` เกิน `total_pages` ต้องได้ `data` ว่างไม่ error (ตาม AC-12) และ UI ต้องไม่พังกับเคสนี้ · เปลี่ยนหน้าแล้ว query string อัปเดตตาม (deep-link กลับมาหน้าเดิมได้)

**FE-18 · Responsive + Visual QA**
- ทำ: ไล่เทียบกับ prototype ทีละจุดตาม checklist เดียวกับ `01-homepage.md`/`02-branches.md` — เน้น filter bar ที่ต้องยุบเรียงเป็นแนวตั้งบน mobile โดยไม่ overflow แนวนอน (ใช้ `minmax(0, 1fr)` ตามบทเรียนเดิม) และ modal ต้องอยู่กลางจอพอดีทั้ง desktop/mobile ไม่ล้นขอบ
- ขึ้นกับ: FE-14, FE-15, FE-16, FE-17
- DoD: ตรวจ responsive จริงด้วย viewport แคบจริง (390px ผ่าน device metrics override ไม่ใช่แค่ย่อหน้าต่าง) ตามบทเรียนจาก `01-homepage.md` · `npm run build` ผ่านไม่มี error/warning สำคัญ · ไม่มีไฟล์ backend ถูกแก้ (เช็ค git diff)
