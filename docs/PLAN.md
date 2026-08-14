# PLAN — ระบบจองห้องพักวิเศษสุขนครคอนโด

แผนการสร้างตาม `docs/SPEC.md` (AC-1…AC-22) และกติกาใน `AGENTS.md`

## 1. ขอบเขตและข้อสมมติ

- สร้าง REST API ให้ครบทุก AC แล้วต่อด้วย React frontend
- `docs/openapi.yaml` **ยังไม่มีในโปรเจกต์** ต้องสร้างเป็นงานแรก เพราะเป็นทั้งแหล่งความจริงของ contract และต้นทางของ typed client ฝั่ง frontend
- SPEC ระบุ Frontend เป็น out of scope (v1) — แผนนี้ขยายขอบเขตออกไป ต้องแก้บรรทัดนั้นใน SPEC ก่อนเริ่ม M8

**ฐานข้อมูล**: ตาม requirement — runtime ใช้ **SQLite** ส่วน test ใช้ **PostgreSQL** (1 database ต่อ 1 test)
ผลที่ตามมาซึ่งต้องยอมรับ: พฤติกรรมที่ต่างกันระหว่างสอง engine จะไม่ถูกทดสอบ โดยเฉพาะ 4 จุดที่ SPEC พึ่งพาโดยตรง

| เรื่อง | SQLite | PostgreSQL |
|---|---|---|
| ล็อกแถวห้อง (AC-8) | ไม่มี `SELECT FOR UPDATE` ต้องใช้ `BEGIN IMMEDIATE` | `SELECT FOR UPDATE` |
| รหัสจองรันเลข (AC-8) | ไม่มี sequence ต้องใช้ตาราง counter ใน transaction เดียวกัน | `SEQUENCE` |
| ยอดเงิน (AC-6) | ไม่มี `numeric` — เก็บเป็นจำนวนเต็มหน่วยสตางค์ | `numeric` |
| partial unique index (AC-19) | รองรับ | รองรับ |

→ SQL ในชั้น repo ต้องเขียนแบบพอร์ตได้ทั้งสอง engine และแยกเฉพาะ 3 จุดข้างบนออกเป็นจุดต่อที่สลับตาม engine ได้
**ข้อเสนอ**: เลือก engine เดียวทั้ง runtime และ test จะได้ไม่ต้องแบกภาระนี้ — รอการตัดสินใจ

## 2. สถาปัตยกรรม Backend

```
HTTP → handler → service → repository → database
```

| ชั้น | รู้อะไร | ห้ามรู้ |
|---|---|---|
| handler | Gin, HTTP status, พาร์ส/ตรวจรูปแบบ input, แปลง typed error เป็น response | กฎธุรกิจ, SQL |
| service | กฎธุรกิจทั้งหมดใน SPEC, ขอบเขต transaction, สิทธิ์รายสาขา, ปล่อย typed error | Gin, HTTP status, SQL |
| repository | SQL ล้วน, map แถวเป็น struct | กฎธุรกิจ, HTTP |

- service คุยกับ repository ผ่าน interface ที่ service เป็นเจ้าของ
- transaction เปิด/ปิดที่ service ผ่าน TxManager ที่มีอยู่แล้ว ห้าม repo เปิดเอง
- สิทธิ์รายสาขา (AC-4) บังคับที่ service ทุกครั้ง โดยอ่าน `branch_id` **จากฐานข้อมูลตาม user ในทุกคำขอ** — JWT เก็บแค่ `user_id` + `role`
- โครงไฟล์คงตาม `AGENTS.md` (module ละ handler / service / repository / model / router)

## 3. Error handling

- service ปล่อย **typed error** ที่พกรหัสตาม SPEC: `validation_failed` (+ fields), `invalid_state`, `room_unavailable`, `not_found`, `forbidden`, `unauthorized`, `conflict`, `invalid_credentials`, `account_disabled`
- middleware ตัวเดียวแปลง typed error → HTTP status + body ตามรูปแบบใน SPEC ที่เดียวในระบบ ห้าม handler ประกอบ error เอง
- กฎ status ตาม AC-21: 400 = พาร์ส/ชนิด/query param ผิด · 422 = ค่าถูกชนิดแต่ผิดกติกาธุรกิจ · `fields` มีเฉพาะ `validation_failed`
- error ที่ไม่รู้จัก → 500 `internal_error` พร้อม log ฝั่งเซิร์ฟเวอร์ ไม่หลุด stack trace/SQL ออก response (AC-15)
- ข้อความผู้ใช้เป็นภาษาไทยทั้งหมด เก็บรวมที่เดียว รวมถึง error ระดับ router (404/405/body พัง)

## 4. Schema ที่ SPEC บังคับ

ตามกติกา AGENTS.md ข้อ 4 — กฎธุรกิจต้องมี constraint รองรับ ไม่ใช่เช็คแค่ในโค้ด

- `bookings.expires_at` เป็นคอลัมน์รายใบ (AC-16) ไม่คำนวณจาก `created_at`
- สถานะการจองมี 4 ค่าเท่านั้น บังคับด้วย CHECK — **ไม่มี `rejected`**
- unique index กันจองซ้อน: ห้ามมีใบสถานะ `pending_payment`/`awaiting_review`/`approved` ซ้ำต่อห้อง สำหรับรายเดือน (AC-8)
- `rooms`: partial unique `(branch_id, room_number) WHERE deleted_at IS NULL` (AC-19)
- `users.email` unique ทั้งตาราง ไม่สนใจ `deleted_at` (AC-18)
- `payment_attempts` เก็บได้หลายแถวต่อ 1 การจอง ไม่เขียนทับ (AC-9)
- `activity_logs.actor_id` เป็น nullable รองรับ `booking.auto_expire` ที่ผู้ทำเป็น system (AC-13)
- เวลาเก็บเป็น UTC ตัดสินวันที่ด้วย Asia/Bangkok ที่ชั้น service (AC-22)

## 5. Frontend

```
pages → components → generated client
```

- generate typed client จาก `docs/openapi.yaml` ทุกครั้งที่ contract เปลี่ยน โฟลเดอร์ที่ generate ห้ามแก้ด้วยมือ
- **component ห้ามเรียก `fetch` เอง** — มี wrapper ชั้นเดียวที่ดูแล base URL, แนบ access token, ต่อ refresh token อัตโนมัติเมื่อ 401 (AC-3)
- pages รู้จัก routing/สิทธิ์/สถานะหน้าจอ · components รับ props แล้ว render อย่างเดียว
- แปลง `error.code` เป็นข้อความผู้ใช้ที่ตารางเดียวกลาง ไม่กระจายตามหน้า
- หน้าจอตาม `Prototype/ตัวอย่าง Website.pdf`

## 6. Testing

- `go test` + `net/http/httptest` ยิงผ่าน router จริงตั้งแต่ middleware ลงไป
- **1 database ต่อ 1 test** สร้างใหม่และล้างทิ้งเมื่อจบ ไม่แชร์ state ข้ามเทส รันขนานได้
- `t.TempDir()` เป็น `UPLOAD_DIR` ของแต่ละเทส — ไฟล์สลิปไม่ปนกันและถูกลบอัตโนมัติ (AC-9, AC-17)
- helper กลางตัวเดียวสำหรับ: ตั้ง app + ใส่ข้อมูลตั้งต้นขั้นต่ำ + ออก token ตามบทบาท
- **เทสที่ห้ามขาด** — AC ที่มีคำว่า "พร้อมกัน" ต้องรันด้วย goroutine จริง: AC-3 (refresh ซ้ำ), AC-8 (จองห้องเดียวกัน + 100 รหัสไม่ซ้ำ), AC-10 (approve ซ้อน), AC-11 (cancel ซ้อน)
- AC-2 (timing) และ AC-15 (parameterized query) แยกออกจาก suite ปกติ — ตัวแรกวัด median 50 ครั้ง ตัวหลังตรวจด้วย code review
- ทุก AC ต้องมีเทสอ้างถึงชัดเจน ตั้งชื่อเทสให้บอกได้ว่าปิด AC ข้อไหน

## 7. ลำดับงาน

ทำทีละ milestone ตาม AGENTS.md ข้อ 2 — จบแต่ละขั้นต้อง build ผ่าน เทสเขียว และ AC ที่ระบุปิดครบ

| # | งาน | ปิด AC |
|---|---|---|
| M0 | `docs/openapi.yaml` + โครง error กลาง + helper เทส + migration ตามข้อ 4 | 21 |
| M1 | สมัคร / เข้าสู่ระบบ / token / เปลี่ยนรหัสผ่าน | 1, 2, 3, 22 |
| M2 | middleware สิทธิ์ + ล็อกสาขา + ขอบเขตข้อมูลของ member | 4, 5 |
| M3 | สาขา, ห้อง, ค้นหาห้องว่าง, รูปภาพเป็น URL | 12, 14, 19 |
| M4 | สร้างการจองรายวัน/รายเดือน + กันจองซ้อน + รหัสจอง | 6, 7, 8 |
| M5 | แจ้งชำระเงิน + เสิร์ฟสลิปแบบตรวจสิทธิ์ | 9, 17 |
| M6 | อนุมัติ/ปฏิเสธ/ยกเลิก/หมดอายุอัตโนมัติ/นัดหมาย + แจ้งเตือน | 10, 11, 16, 20 |
| M7 | activity log 23 action + แดชบอร์ด + จัดการบัญชีผู้ดูแล | 13, 18 |
| M8 | ความทนทาน: panic recovery, error ระดับ router, ข้อความไทยครบ | 15 |
| M9 | Frontend: generate client + wrapper + pages ตาม prototype | — |

## 8. เสร็จเมื่อไร

- `gofmt -l . && go build ./... && go vet ./...` ผ่านสะอาด
- ทุก AC ใน SPEC มีเทสที่ผ่าน และชี้ได้ว่าเทสไหนปิด AC ข้อไหน
- ทุก endpoint ตรงกับ `docs/openapi.yaml` และตารางใน `backend/README.md`
- ไม่มี SQL อยู่นอกชั้น repository และไม่มีการประกอบ error นอก middleware กลาง
- frontend ไม่มี `fetch` นอก wrapper ชั้นเดียว
