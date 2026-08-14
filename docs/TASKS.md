# TASKS — ลำดับงานสร้างระบบจองห้องพักวิเศษสุขนครคอนโด

อ้างอิง `AGENTS.md`, `docs/SPEC.md` (AC-1…AC-22), `docs/PLAN.md`
ทำเรียงตามลำดับ ทีละ task · task เขียนเทสมาก่อน task เขียนโค้ดของ slice เดียวกันเสมอ

**คำสั่งที่ใช้ใน DoD**
- `CHECK` = `cd backend && gofmt -l . && go build ./... && go vet ./...` ต้องไม่มี output จาก gofmt และไม่มี error
- `TEST` = `cd backend && go test ./... -race` ต้องผ่านทั้งหมด
- **DoD ของ task เขียนเทส คือเทสรันแล้วแดง** ไม่ใช่เขียว — ถ้าเขียวแปลว่าเทสไม่ได้ทดสอบอะไร

---

## Phase 0 — Walking skeleton

**T-01 · สร้าง `docs/openapi.yaml`**
- ทำ: เขียน contract ทุก endpoint ตามหัวข้อ API Contract ใน SPEC ให้ครบ ไม่เพิ่ม ไม่เปลี่ยนชื่อ ไม่ตัด · รวม schema ของ error envelope และรายการ error code ทั้ง 12 ตัว
- ขึ้นกับ: —
- DoD: `npx @redocly/cli lint docs/openapi.yaml` ผ่าน · ไล่เทียบกับ SPEC ทีละบรรทัดแล้วจำนวน endpoint ตรงกันพอดี · ยังไม่มีโค้ด Go แม้แต่บรรทัดเดียว

**T-02 · โครงเทส: 1 database ต่อ 1 test**
- ทำ: helper สร้าง/ลบ database ต่อเทส, ตั้ง `UPLOAD_DIR` เป็น `t.TempDir()`, ออก token ตามบทบาท, ใส่ข้อมูลตั้งต้นขั้นต่ำ
- ขึ้นกับ: T-01
- DoD: เทสตัวอย่าง 3 ตัวรันขนานด้วย `-race` แล้วไม่ชน state กัน · database ถูกลบทิ้งครบหลังจบ

**T-03 · เทส error envelope และกฎ 400/422 (AC-21)**
- ทำ: เทสทุกบรรทัดของ AC-21 — 400 vs 422, `fields` โผล่เฉพาะ `validation_failed`, 401 vs 403, 404/405/500 ใช้รูปแบบเดียวกัน
- ขึ้นกับ: T-02
- DoD: `go test ./internal/httpx/...` รันแล้ว **แดงทุกเคส** · checklist เทียบกับ AC-21 ครบทุกบรรทัด

**T-04 · โค้ด typed error + middleware แปลง error**
- ทำ: typed error ครบ 12 code, middleware ตัวเดียวแปลงเป็น status + body, ข้อความไทยรวมที่เดียว
- ขึ้นกับ: T-03
- DoD: `TEST` ผ่าน · `CHECK` สะอาด · grep แล้วไม่มีการประกอบ error JSON นอก middleware

**T-05 · เทสเวลาและวันที่ (AC-22)**
- ทำ: เทสทุกบรรทัดของ AC-22 รวมเคสนาฬิกาเซิร์ฟเวอร์เป็น UTC แล้วยิงตอน 23:30 เวลาไทย
- ขึ้นกับ: T-02
- DoD: รันแล้ว **แดง** · ครอบทั้ง check_in, move_in_date, transferred_at, appointment_at

**T-06 · โค้ดตัวช่วยเวลา Asia/Bangkok**
- ทำ: จุดเดียวที่ตัดสิน "วันนี้/อดีต/อนาคต" · เก็บ UTC · ตอบกลับเป็น ISO 8601 พร้อม offset
- ขึ้นกับ: T-05
- DoD: `TEST` ผ่าน · `CHECK` สะอาด

**T-07 · เทส skeleton `GET /health`**
- ทำ: ยิงผ่าน router จริงตั้งแต่ middleware ลงถึง database
- ขึ้นกับ: T-02
- DoD: รันแล้ว **แดง**

**T-08 · โค้ด skeleton + migration ฐาน**
- ทำ: entrypoint, router, handler→service→repo ของ `/health`, migration แรกที่มีตารางหลักและ constraint ตาม PLAN §4
- ขึ้นกับ: T-07
- DoD: `TEST` ผ่าน · `CHECK` สะอาด · `GET /health` ตอบ 200 ตาม contract ใน T-01

---

## Phase 1 — บัญชีและ token

**T-09 · เทสสมัครสมาชิก (AC-1)** — ขึ้นกับ T-04, T-08 · DoD: แดง · ครบทุกบรรทัด AC-1 รวมอีเมลซ้ำแบบไม่สนตัวพิมพ์, รหัส 73 ตัว, ฟิลด์ `role` แถม, หลายช่องผิดพร้อมกัน
**T-10 · โค้ดสมัครสมาชิก** — ขึ้นกับ T-09 · `POST /api/v1/auth/register` · DoD: `TEST` ผ่าน · `CHECK` สะอาด

**T-11 · เทสเข้าสู่ระบบ (AC-2)** — ขึ้นกับ T-10 · DoD: แดง · รวมเทสวัด median 50 ครั้งต่อกรณี และเคส `account_disabled`
**T-12 · โค้ดเข้าสู่ระบบ** — ขึ้นกับ T-11 · `POST /api/v1/auth/login` · DoD: `TEST` ผ่าน · `CHECK` สะอาด

**T-13 · เทส token และเปลี่ยนรหัสผ่าน (AC-3)** — ขึ้นกับ T-12 · DoD: แดง · รวมเคสยิง refresh พร้อมกัน 2 request ด้วย goroutine จริง และเคยเช็คว่าตารางเก็บแค่ hash
**T-14 · โค้ด token และเปลี่ยนรหัสผ่าน** — ขึ้นกับ T-13 · `POST /auth/refresh`, `/auth/logout`, `/me/password` · DoD: `TEST` ผ่าน · `CHECK` สะอาด

**T-15 · เทสโปรไฟล์และแจ้งเตือน** — ขึ้นกับ T-14 · DoD: แดง · ครอบบรรทัด notifications ของ AC-5 (อ่าน id ที่ไม่ใช่ของตัวเอง → 204 ไม่เปลี่ยนอะไร)
**T-16 · โค้ดโปรไฟล์และแจ้งเตือน** — ขึ้นกับ T-15 · `GET/PUT /me`, `GET /me/notifications`, `POST /me/notifications/read` · DoD: `TEST` ผ่าน · `CHECK` สะอาด

---

## Phase 2 — สิทธิ์และขอบเขตข้อมูล

**T-17 · เทสล็อกสาขาของ admin (AC-4)** — ขึ้นกับ T-16 · DoD: แดง · ครบทุกบรรทัดรวมเคสย้ายสาขาแล้วใช้ token ใบเดิมต่อ และเคส 403 vs 404
**T-18 · โค้ด middleware สิทธิ์รายสาขา** — ขึ้นกับ T-17 · อ่าน `branch_id` จากฐานข้อมูลทุกคำขอ JWT เก็บแค่ `user_id` + `role` · DoD: `TEST` ผ่าน · `CHECK` สะอาด

**T-19 · เทสขอบเขตข้อมูลของ member (AC-5)** — ขึ้นกับ T-18 · DoD: แดง · ครบทุกบรรทัด
**T-20 · โค้ดขอบเขตข้อมูลของ member** — ขึ้นกับ T-19 · DoD: `TEST` ผ่าน · `CHECK` สะอาด

---

## Phase 3 — สาขาและห้อง

**T-21 · เทสข้อมูลสาขาฝั่ง guest** — ขึ้นกับ T-20 · DoD: แดง · รวมเคส branchID ไม่ใช่ UUID → 400 และไม่พบ → 404
**T-22 · โค้ดข้อมูลสาขาฝั่ง guest** — ขึ้นกับ T-21 · `GET /branches`, `/branches/:branchID`, `/amenities`, `/room-types` · DoD: `TEST` ผ่าน · `CHECK` สะอาด

**T-23 · เทสแก้ข้อมูลสาขา** — ขึ้นกับ T-22 · DoD: แดง · รวมเคสข้ามสาขา → 403
**T-24 · โค้ดแก้ข้อมูลสาขา** — ขึ้นกับ T-23 · `PUT /admin/branch`, `/branch/amenities`, `/branch/nearby` · DoD: `TEST` ผ่าน · `CHECK` สะอาด

**T-25 · เทสรูปสาขา (AC-14)** — ขึ้นกับ T-24 · DoD: แดง · ครบทุกบรรทัดรวม `javascript:`, `data:`, URL ไม่มี host, และ `GET /uploads/*` → 404
**T-26 · โค้ดรูปสาขา** — ขึ้นกับ T-25 · `POST/DELETE /admin/branch/images` · DoD: `TEST` ผ่าน · `CHECK` สะอาด

**T-27 · เทสจัดการห้องฝั่ง admin (AC-19)** — ขึ้นกับ T-26 · DoD: แดง · ครบทุกบรรทัดรวมเลขห้องซ้ำกับห้องที่ลบแล้ว, ปล่อยห้องจาก occupied, ลบห้องที่มีใบ active → 409
**T-28 · โค้ดจัดการห้องฝั่ง admin** — ขึ้นกับ T-27 · `GET/POST /admin/rooms`, `PUT/PATCH/DELETE /admin/rooms/:roomID` + partial unique index · DoD: `TEST` ผ่าน · `CHECK` สะอาด

**T-29 · เทสค้นหาห้องว่าง (AC-12)** — ขึ้นกับ T-28 · DoD: แดง · ครบทุกบรรทัดรวม page เกิน total_pages, page_size=500, เคสขอบวันที่ที่ต้องอยู่ในผลลัพธ์
**T-30 · โค้ดค้นหาห้องว่าง** — ขึ้นกับ T-29 · `GET /rooms/search`, `GET /rooms/:roomID` · DoD: `TEST` ผ่าน · `CHECK` สะอาด

---

## Phase 4 — การจอง

**T-31 · เทสจองรายวัน (AC-6)** — ขึ้นกับ T-30 · DoD: แดง · ครบทุกบรรทัดรวม snapshot ราคา, จอง 400 คืนไม่ล้น, ห้องถูกลบ → 404
**T-32 · โค้ดจองรายวัน** — ขึ้นกับ T-31 · `POST /api/v1/bookings` (stay_type daily) · DoD: `TEST` ผ่าน · `CHECK` สะอาด

**T-33 · เทสจองรายเดือน (AC-7)** — ขึ้นกับ T-32 · DoD: แดง · ครบทุกบรรทัดรวม snapshot ค่าทำสัญญา
**T-34 · โค้ดจองรายเดือน** — ขึ้นกับ T-33 · `POST /api/v1/bookings` (stay_type monthly) · DoD: `TEST` ผ่าน · `CHECK` สะอาด

**T-35 · เทสกันจองซ้อน (AC-8)** — ขึ้นกับ T-34 · DoD: แดง · ครบทุกบรรทัด · ต้องมี goroutine จริง 2 เคส: จองห้องเดียวกันพร้อมกัน และจอง 100 ห้องพร้อมกันแล้วรหัสไม่ซ้ำ
**T-36 · โค้ดกันจองซ้อน** — ขึ้นกับ T-35 · ล็อกแถวห้องตลอด transaction + unique constraint + รหัสจองรันเลข · DoD: `TEST` ผ่าน · `CHECK` สะอาด

**T-37 · เทสดูการจองของตัวเอง** — ขึ้นกับ T-36 · DoD: แดง · รวม pagination และเคสเปิดใบคนอื่น → 404
**T-38 · โค้ดดูการจองของตัวเอง** — ขึ้นกับ T-37 · `GET /api/v1/bookings`, `GET /api/v1/bookings/:bookingID` · DoD: `TEST` ผ่าน · `CHECK` สะอาด

---

## Phase 5 — ชำระเงินและสลิป

**T-39 · เทสแจ้งชำระเงิน (AC-9)** — ขึ้นกับ T-38 · DoD: แดง · ครบทุกบรรทัดรวมไฟล์ .txt เปลี่ยนนามสกุล, ไฟล์เกินขนาดต้องไม่เขียนลงดิสก์, เก็บประวัติหลายแถว
**T-40 · โค้ดแจ้งชำระเงิน** — ขึ้นกับ T-39 · `POST /api/v1/bookings/:bookingID/payment` · DoD: `TEST` ผ่าน · `CHECK` สะอาด

**T-41 · เทสเสิร์ฟไฟล์สลิป (AC-17)** — ขึ้นกับ T-40 · DoD: แดง · ครบทุกบรรทัดรวม member คนอื่น/admin คนละสาขา → 404 และเคส path traversal
**T-42 · โค้ดเสิร์ฟไฟล์สลิป** — ขึ้นกับ T-41 · `GET /api/v1/bookings/:bookingID/slip` · DoD: `TEST` ผ่าน · `CHECK` สะอาด

---

## Phase 6 — วงจรอนุมัติ

**T-43 · เทสอนุมัติ/ปฏิเสธ (AC-10)** — ขึ้นกับ T-42 · DoD: แดง · ครบทุกบรรทัดรวมปฏิเสธแล้วกลับเป็น pending_payment + expires_at +24 ชม. และ approve พร้อมกัน 2 request
**T-44 · โค้ดอนุมัติ/ปฏิเสธ** — ขึ้นกับ T-43 · `POST /admin/bookings/:bookingID/approve`, `/reject` · DoD: `TEST` ผ่าน · `CHECK` สะอาด

**T-45 · เทสยกเลิก (AC-11)** — ขึ้นกับ T-44 · DoD: แดง · ครบทุกบรรทัดรวมห้อง maintenance ต้องไม่ถูกทับ, คืนห้องไม่สำเร็จต้องไม่ยกเลิกใบ, cancel พร้อมกัน 2 request
**T-46 · โค้ดยกเลิก** — ขึ้นกับ T-45 · `POST /api/v1/bookings/:bookingID/cancel` · DoD: `TEST` ผ่าน · `CHECK` สะอาด

**T-47 · เทสหมดอายุอัตโนมัติ (AC-16)** — ขึ้นกับ T-46 · DoD: แดง · ครบทุกบรรทัดรวม 10 นาที, 24 ชม.หลังถูกปฏิเสธ, หยุดนับเมื่อเข้า awaiting_review
**T-48 · โค้ดหมดอายุอัตโนมัติ** — ขึ้นกับ T-47 · งานเบื้องหลัง + คอลัมน์ `expires_at` รายใบ · DoD: `TEST` ผ่าน · `CHECK` สะอาด

**T-49 · เทสนัดหมายทำสัญญา (AC-20)** — ขึ้นกับ T-48 · DoD: แดง · ครบทุกบรรทัด
**T-50 · โค้ดนัดหมายทำสัญญา** — ขึ้นกับ T-49 · `PUT /admin/bookings/:bookingID/appointment` · DoD: `TEST` ผ่าน · `CHECK` สะอาด

---

## Phase 7 — บันทึกและผู้ดูแล

**T-51 · เทส activity log (AC-13)** — ขึ้นกับ T-50 · DoD: แดง · ครบทั้ง 23 action, เคส actor เป็น system, เคสเขียน log ล้มเหลวแล้วธุรกรรมหลักไม่ rollback
**T-52 · โค้ด activity log** — ขึ้นกับ T-51 · `GET /admin/activity-logs` + `actor_id` nullable · DoD: `TEST` ผ่าน · `CHECK` สะอาด

**T-53 · เทสจัดการบัญชีผู้ดูแล (AC-18)** — ขึ้นกับ T-52 · DoD: แดง · ครบทุกบรรทัดรวมลบ/ระงับแล้วเพิกถอน token ทันที และอีเมลห้ามใช้ซ้ำแม้ลบแล้ว
**T-54 · โค้ดจัดการบัญชีผู้ดูแล** — ขึ้นกับ T-53 · `GET/POST /superadmin/staff`, `PUT/DELETE /superadmin/staff/:userID`, `GET /superadmin/branches` · DoD: `TEST` ผ่าน · `CHECK` สะอาด

**T-55 · เทสแดชบอร์ดและรายชื่อสมาชิก** — ขึ้นกับ T-54 · DoD: แดง · รวมเคสขอบเขตสาขา
**T-56 · โค้ดแดชบอร์ดและรายชื่อสมาชิก** — ขึ้นกับ T-55 · `GET /admin/dashboard`, `/members`, `/members/:memberID/bookings`, `/bookings` · DoD: `TEST` ผ่าน · `CHECK` สะอาด

**T-57 · เทสความทนทาน (AC-15)** — ขึ้นกับ T-56 · DoD: แดง · ครบทุกบรรทัดรวม panic ใน handler, 5xx ต้องไม่มีชื่อตาราง/SQL, error ระดับ router เป็นภาษาไทย
**T-58 · โค้ดความทนทาน** — ขึ้นกับ T-57 · recovery middleware + handler 404/405 + ตรวจข้อความไทยครบ · DoD: `TEST` ผ่าน · `CHECK` สะอาด

---

## Phase 8 — Frontend

ทุก task ฝั่ง UI ตรวจด้วย E2E checklist 4 สถานะ: **loading · error · empty · success**

**T-59 · generate typed client + wrapper ชั้นเดียว**
- ทำ: generate client จาก `docs/openapi.yaml`, wrapper ที่ดูแล base URL + แนบ access token + refresh อัตโนมัติเมื่อ 401, ตารางแปลง `error.code` เป็นข้อความไทย
- ขึ้นกับ: T-01, T-58
- DoD: สคริปต์ generate รันซ้ำได้ผลเหมือนเดิม · `grep -r "fetch(" src/` เจอเฉพาะในไฟล์ wrapper · เปลี่ยน contract แล้ว generate ใหม่ทำให้ type พัง (พิสูจน์ว่า client ผูกกับ contract จริง)

**T-60 · หน้าเข้าสู่ระบบและสมัครสมาชิก** — ขึ้นกับ T-59 · E2E: loading ปุ่มถูก disable · error แสดงข้อความไทยจาก `invalid_credentials` และ `account_disabled` แยกกัน · empty ไม่กรอกอะไรแล้วกด submit เห็น fields ครบทุกช่องในครั้งเดียว · success เข้าระบบแล้วเด้งไปหน้าหลักพร้อม token

**T-61 · หน้าค้นหาห้องว่าง** — ขึ้นกับ T-60 · E2E: loading โครงร่างระหว่างค้นหา · error ใส่ stay_type ผิดแล้วเห็นข้อความไทย · empty ไม่พบห้องต้องเห็นข้อความ ไม่ใช่หน้าขาว · success เห็นผลลัพธ์ + pagination ตรงกับ meta

**T-62 · หน้าจองห้อง** — ขึ้นกับ T-61 · E2E: loading ระหว่างส่งคำขอ · error ห้องถูกจองตัดหน้าแล้วเห็นข้อความจาก `room_unavailable` · empty ไม่เลือกวันที่แล้วปุ่มยังกดไม่ได้ · success ได้ booking_code และเห็นเวลาหมดอายุ 10 นาที

**T-63 · หน้าแจ้งชำระเงิน** — ขึ้นกับ T-62 · E2E: loading ระหว่างอัปโหลด · error ไฟล์ผิดชนิด/ใหญ่เกินเห็นข้อความไทย · empty ยังไม่แนบไฟล์แล้วส่งไม่ได้ · success สถานะเปลี่ยนเป็นรอตรวจสอบทันที

**T-64 · หน้าการจองของฉัน** — ขึ้นกับ T-63 · E2E: loading · error · empty ยังไม่เคยจองต้องเห็นข้อความชวนไปค้นหาห้อง · success เห็นรายการ กรองตามสถานะได้ ยกเลิกใบ pending_payment ได้และใบอื่นปุ่มยกเลิกไม่ขึ้น

**T-65 · หน้าอนุมัติ/ปฏิเสธของ admin** — ขึ้นกับ T-64 · E2E: loading · error ปฏิเสธโดยไม่ใส่เหตุผลเห็น validation ไทย · empty ไม่มีใบรอตรวจต้องเห็นข้อความ · success อนุมัติแล้วแถวหายจากคิวและสลิปเปิดดูได้

**T-66 · หน้าจัดการห้องของ admin** — ขึ้นกับ T-65 · E2E: loading · error เลขห้องซ้ำเห็นข้อความไทย · empty สาขายังไม่มีห้อง · success เพิ่ม/แก้/เปลี่ยนสถานะ/ลบห้องแล้วรายการอัปเดต

**T-67 · หน้าจัดการผู้ดูแลของ super admin** — ขึ้นกับ T-66 · E2E: loading · error ระงับหรือลบตัวเองแล้วเห็นข้อความไทย · empty · success เพิ่ม/แก้/ลบ staff แล้วรายการอัปเดต

---

## Phase 9 — Hardening

**T-68 · quality gates บนเครื่อง**
- ทำ: สคริปต์เดียวรวม `gofmt -l .`, `go build ./...`, `go vet ./...`, `go test ./... -race`, lint `docs/openapi.yaml`, ตรวจว่าไม่มี `fetch(` นอก wrapper · ต่อเข้า pre-commit hook
- ขึ้นกับ: T-67
- DoD: รันสคริปต์เดียวจบทุกอย่างและ exit code เป็น 0 · แกล้งใส่โค้ดที่ format ผิดแล้วสคริปต์ต้องแดง

**T-69 · CI ด้วย GitHub Actions**
- ทำ: workflow รัน quality gates ชุดเดียวกับ T-68 ทุก push และ pull request · ใช้ service container สำหรับฐานข้อมูลของเทส
- ขึ้นกับ: T-68
- DoD: workflow เขียวบน branch หลัก · เปิด PR ที่เทสพังแล้ว CI ต้องแดงและบล็อกการ merge

**T-70 · cross-agent review ด้วย session ใหม่**
- ทำ: เปิด session ใหม่ที่ไม่มีบริบทเดิม ให้ไล่ตรวจ 3 อย่าง — (ก) ทุก AC ใน SPEC มีเทสที่ชี้ตัวได้จริง (ข) ทุก endpoint ตรงกับ `docs/openapi.yaml` ไม่มีเส้นทางส่วนเกิน (ค) ไม่มี SQL นอกชั้น repository และไม่มีการประกอบ error นอก middleware
- ขึ้นกับ: T-69
- DoD: รายงานผลเป็นข้อ ๆ · ทุกข้อที่พบต้องถูกแก้หรือบันทึกเหตุผลที่ไม่แก้ · ไม่มีข้อค้างที่ไม่มีคำตอบ
