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
| 5 | หน้า Contact | หน้า 13, รูปภาพที่ 18 | [04-contact.md](04-contact.md) | **แตก task แล้ว** (FE-19…FE-20) รอ implement |
| 6 | หน้า Login / Register | ยังไม่ได้เปิดดู | — | รอ (มี placeholder route กันพัง) |

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
