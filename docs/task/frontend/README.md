# Frontend Guest — แผนงานทีละหน้า

แตก task จาก `frontend/prototype/ตัวอย่าง Website v2.pdf` ทีละหน้าจอ ตามขอบเขต Guest ที่กำหนดใน `docs/c.md` และ `.claude/CLAUDE.md`
อ้างอิง API จริงจาก `docs/openapi.yaml` เท่านั้น — ไม่เดา ไม่สร้าง endpoint ใหม่ ไม่ mock แทนของจริง

**หมายเหตุเรื่องเลข task**: ไฟล์ในโฟลเดอร์นี้ใช้ prefix **FE-** (ไม่ใช่ `T-`) เพราะ `docs/TASKS.md` ใช้ `T-01…T-70` ไปแล้วทั้งฝั่ง backend และฝั่ง frontend รวม role (Phase 8: T-59…T-67) เพื่อไม่ให้เลขชนกันและอ้างอิงกำกวมเวลาพูดถึง "T-17" ว่าเป็นของ Phase ไหน

**DoD ของทุก task ที่ดึงข้อมูลจาก API ต้องมี E2E checklist ครบ 4 สถานะ**: `loading` / `empty` / `error` (พร้อมปุ่มลองใหม่) / `success` — ส่วนไหนเป็น static content ล้วน (ไม่มี API) ไม่ต้องมี 4 สถานะนี้ แต่ต้องระบุไว้ชัดว่าเป็น static

## สถานะรายหน้า

| # | หน้า (ตาม prototype) | อ้างอิงในไฟล์ PDF | ไฟล์ task | สถานะ |
|---|---|---|---|---|
| 1 | Homepage | หน้า 1, รูปภาพที่ 1–2 | [01-homepage.md](01-homepage.md) | **implement แล้ว** (FE-00…FE-07) |
| 2 | หน้ารวมสาขา (Branches List) | หน้า 2, รูปภาพที่ 3 | [02-branches.md](02-branches.md) | **implement แล้ว** (FE-08…FE-13) |
| 3 | หน้ารายละเอียดสาขา + แผนที่ (BranchDetail + Map) | หน้า 3–8, รูปภาพที่ 4–12 | [02-branches.md](02-branches.md) | **implement แล้ว** (FE-08…FE-13) |
| 4 | หน้าค้นหาห้องพัก + Filter + ผลลัพธ์ + modal ข้อจำกัด Guest | หน้า 9–12, รูปภาพที่ 13–17 | [03-room-search.md](03-room-search.md) | **implement แล้ว** (FE-14…FE-18) |
| 5 | หน้า Contact | หน้า 13, รูปภาพที่ 18 | [04-contact.md](04-contact.md) | **implement แล้ว** (FE-19…FE-20) |
| 6 | หน้า Login / Register | หน้า 14, รูปภาพที่ 19–20 | [05-login-register.md](05-login-register.md) | **implement แล้ว** (FE-21…FE-25) |

อัปเดตตารางนี้ทุกครั้งที่เปิดดู prototype หน้าใหม่แล้วแตก task เพิ่ม

## สถานะการ implement Homepage (2026-08-15)

โค้ดจริงอยู่ที่ `frontend/` (Create React App เดิม — ไม่ได้เปลี่ยน build tool) ครบทั้ง FE-00–FE-07:

- `src/api/`, `src/hooks/useBranches.js` — เชื่อม `GET /api/v1/branches` จริง (curl ทดสอบกับ backend ที่รันอยู่แล้วก่อนเขียนโค้ด)
- `src/components/layout/` (Header + Footer), `src/components/home/` (Hero, Facilities, Branches, StayType, CTA)
- `src/routes/AppRoutes.jsx` — route Guest ครบ 7 เส้นทาง หน้าที่ยัง "รอ" ใช้ `ComingSoonPage` กันพัง ไม่ error
- ตรวจแล้ว: `npm test` ผ่าน (2/2), `npm run build` compile สำเร็จไม่มี warning, responsive ตรวจจริงที่ 390/820/1440px ผ่านครบ (ใช้ Chrome DevTools Protocol ตั้ง mobile viewport จริง ไม่ใช่แค่ resize หน้าต่าง), ตรวจ git diff แล้วไม่มีไฟล์ backend ถูกแก้
- Gap ที่ต้องตัดสินใจต่อ (ไม่บล็อกการใช้งาน แต่ควรรู้): `Branch.cover_image_url` และ `RoomType.image_url` ยังว่างเปล่าทุกรายการในข้อมูลจริง (ยังไม่มีการอัปโหลดรูป) หน้าเว็บ fallback เป็น placeholder ไอคอนอาคารไปก่อน

## สถานะการ implement Branches List + Detail + Map (2026-08-15)

ครบทั้ง FE-08–FE-13 ตามที่ระบุใน [02-branches.md](02-branches.md):

- `src/components/branch/` (BranchListRow, BranchGallery, BranchMapPanel, OtherBranchesSidebar, BranchInfoCard, BranchAmenitiesCard, NearbyPlacesSection), `src/components/icons/amenityIconMap.js` (ครบ 12 icon จริงจาก `GET /api/v1/amenities` + fallback)
- `src/utils/branchChips.js`, `branchStay.js`, `nearbyPlaces.js`, `mapEmbed.js` — logic ตามกฎใน Gap ของ `02-branches.md` (เช็ค key หายไม่ใช่ค่า 0, ห้ามเดาตัวเลขที่ไม่มี field รองรับ)
- เชื่อมจริง: `GET /api/v1/branches` (list + sidebar "สาขาอื่น"), `GET /api/v1/branches/:branchID` (detail, แยกข้อความ 400 "รหัสสาขาไม่ถูกต้อง" กับ 404 "ไม่พบสาขาที่ต้องการ")
- Deep-link `/branches/:id/map` ทำงานถูกต้อง (tab แผนที่ active ทันทีไม่ต้องกด) — ตรวจผ่าน screenshot จริง
- ตรวจแล้ว: `npm test` ผ่าน, `npm run build` compile สำเร็จไม่มี warning, responsive จริงที่ 390/820/1440px ผ่านครบ (CDP mobile viewport), backend ไม่ถูกแก้

**เรื่องรูปภาพที่ต้องรู้**: มีรูปจริง (ไม่มีลายน้ำ) อยู่ใน `src/assets/45/` และ `src/assets/bk/` ใช้แสดงเป็น gallery/cover ของ 2 สาขานั้นแล้ว — ส่วน `src/assets/ckp/` (เจริญกรุงเพลส) **ทั้ง 4 ไฟล์มีลายน้ำ "RentHub.in.th Verifiled Photo" ติดอยู่** (รูปจากเว็บ listing เจ้าอื่น ไม่ใช่รูปของธุรกิจนี้เอง) โค้ดจึง**ตั้งใจไม่ใช้ไฟล์ชุดนี้** ปล่อยให้ตกไปที่ placeholder ไอคอนแทน — ไม่ได้ลบไฟล์ทิ้งเผื่อต้องตรวจสอบ/เปลี่ยนเอง แต่ห้ามเอาขึ้นเว็บจริงตามที่เป็นอยู่ (มีความเสี่ยงเรื่องลิขสิทธิ์/แบรนด์คู่แข่งปนอยู่ในเว็บของเราเอง) ต้องหารูปจริงของสาขานี้มาแทนก่อนขึ้นโปรดักชัน

## สถานะการ implement Room Search (2026-08-15)

ครบทั้ง FE-14–FE-18 ตามที่ระบุใน [03-room-search.md](03-room-search.md):

- `src/components/room-search/` (FilterDropdown, BranchFilterDropdown, SelectFilterDropdown, DateFilterDropdown, FilterBar, RoomResultCard, GuestBookingModal), `src/components/common/Pagination.jsx`
- `src/hooks/useRoomSearch.js`, `src/utils/roomChips.js`, `src/utils/calendar.js` — mapping วันที่/chip ตามกฎใน Gap ของ `03-room-search.md` (ปฏิทินวันเดียว → ส่งเป็น check_in ตอน daily / move_in_date ตอน monthly เท่านั้น, chip ใช้แค่ field จริงที่มี + join amenities จากสาขา)
- เชื่อมจริง: `GET /api/v1/rooms/search` (filter branch/stay_type/วันที่ + pagination), `GET /api/v1/branches` (join amenities + สร้าง filter สาขา)
- Modal ข้อจำกัด Guest (`GuestBookingModal`) ทำงานตาม docs/c.md ข้อ 19 — ยังไม่ login กด "จองเลย" เห็น modal, login แล้วพาไป `/rooms/:roomID` (placeholder ใหม่ ยังไม่แตก task)
- ตรวจแล้ว: `npm test` ผ่าน, `npm run build` compile สำเร็จไม่มี warning, ตรวจ dropdown/modal ทั้ง 3 ตัวด้วย screenshot จริง (คลิกจริงผ่าน CDP ไม่ใช่แค่ดู CSS) ทั้ง desktop และ mobile (390px), backend ไม่ถูกแก้

## สถานะการ implement Login / Register (2026-09-08)

ครบทั้ง FE-21–FE-25 ตามที่ระบุใน [05-login-register.md](05-login-register.md):

- `src/api/authApi.jsx` (register/login/logout), `src/utils/authValidation.js` (mirror กฎ validate ของ backend), `src/utils/authStorage.jsx` เพิ่ม event `onAuthChange` ให้ component อื่นรู้ทันทีว่า login/logout เปลี่ยนสถานะ (ไม่ต้อง reload), `src/hooks/useAuthState.js`
- `src/components/auth/` ใหม่ (`AuthCard`, `AuthInput` มี toggle แสดง/ซ่อนรหัสผ่านในตัว, `AuthForm.css`) ใช้ร่วมกันทั้ง Login/Register, ไอคอนใหม่ 4 ตัวใน `src/components/icons/index.jsx` (`IconMail`, `IconLockClosed`, `IconEye`/`IconEyeOff`, `IconUser`)
- เชื่อมจริง: `POST /api/v1/auth/register`, `POST /api/v1/auth/login`, `POST /api/v1/auth/logout` — **เจอ gap สำคัญตอน curl จริง**: response จริงห่อด้วย `{ data: {...} }` ต่างจากตัวอย่าง schema เปล่า ๆ ใน `docs/openapi.yaml` (ตรงกับ pattern ที่ `branchApi`/`useBranches` เดิมก็ต้อง unwrap `.data` เหมือนกันอยู่แล้ว) แก้โค้ดให้ unwrap ถูกจุดแล้ว
- `Header.jsx` สลับปุ่ม login/register เป็นปุ่ม "ออกจากระบบ" ปุ่มเดียวเมื่อ login อยู่ (ตาม Gap ที่ตกลงไว้ในไฟล์ task — ไม่มี dropdown เมนูสมาชิกเพราะเข้าข่าย Member area ที่ห้ามทำ)
- ตรวจแล้วด้วย browser จริง (Playwright) ต่อ backend ที่รันอยู่จริง ไม่ mock: สมัครสมาชิกจริงสำเร็จ → token เก็บถูกคีย์ → header เปลี่ยนเป็น "ออกจากระบบ" ทันที → logout แล้ว token ถูกลบ → login ซ้ำด้วยบัญชีเดิมสำเร็จ → เข้า `/login` ซ้ำตอน login อยู่แล้ว redirect กลับหน้าแรกอัตโนมัติ → รหัสผ่านผิดโชว์ข้อความจาก backend จริง ("อีเมลหรือรหัสผ่านไม่ถูกต้อง") → สมัครอีเมลซ้ำโชว์ field error ใต้ช่องอีเมลจริง ("อีเมลนี้ถูกใช้งานแล้ว") → ยืนยันรหัสผ่านไม่ตรงกันถูกกันไว้ฝั่ง client ก่อนยิง request จริง
- `npm run build` compile สำเร็จไม่มี warning, `eslint` ทั้ง `src/` ผ่านสะอาด, responsive จริงที่ 390/1280px ไม่มี horizontal scroll ทั้ง Login/Register, backend ไม่ถูกแก้
- การตัดสินใจตาม Gap ในไฟล์ task: ลิงก์ "ลืมรหัสผ่าน?" พาไปหน้า `/contact` (ไม่มี endpoint จริงให้เชื่อม), checkbox "จดจำฉัน" ยังไม่ผูก logic (token เก็บ `localStorage` เสมอเหมือนเดิม), หลัง login/register สำเร็จ redirect กลับหน้าแรกเสมอ (ยังไม่ทำ "จำหน้าที่มาจาก")
