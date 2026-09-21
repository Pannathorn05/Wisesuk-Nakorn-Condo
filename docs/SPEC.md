# SPEC ระบบจองห้องพักวิเศษสุขนครคอนโด (Backend)

> เขียนใหม่ทั้งไฟล์โดยอ้างอิงโค้ดจริงใน `backend/` (ไม่ใช่เอกสารเก่า) — ทุกข้อคือพฤติกรรมที่ทดสอบผ่าน API ได้จริง
> ยกเว้นข้อที่ทำเครื่องหมาย **[ยังไม่ implement]** ซึ่งเป็นของที่ตกลงจะทำต่อแต่โค้ดปัจจุบันยังไม่มี

## Users
- Guest: ดูสาขา, ดูห้องพัก, ค้นหาห้องว่างตามวันที่/ราคา/ประเภท — ไม่ต้องล็อกอิน
- Member: จองห้อง (รายวัน/รายเดือน), แจ้งชำระเงินพร้อมสลิป, ดู/ยกเลิกการจองของตัวเอง, รับแจ้งเตือน, ล็อกอินด้วยอีเมล/รหัสผ่าน **หรือ** Google/Facebook OAuth (ดู AC-24)
- Admin: จัดการห้องและรายละเอียดสาขา, อนุมัติ/ปฏิเสธการจอง, นัดหมายทำสัญญา, ดูแดชบอร์ดและ activity log — **เฉพาะสาขาที่ตัวเองผูกอยู่ 1 สาขาเท่านั้น**
- Super Admin: ทุกอย่างที่ Admin ทำได้ในทุกสาขา + เพิ่ม/แก้/ลบบัญชีผู้ดูแล + ตั้งรหัสผ่านใหม่ให้ผู้ดูแล
- Admin ผูกกับ 1 สาขาเสมอ — บังคับด้วย DB constraint `admin_requires_branch` (admin ต้องมี branch_id, บทบาทอื่นต้องไม่มี)
- **สาขาของ admin ถูกฝังมากับ JWT access token ตอนออก ไม่ใช่การ query DB ทุกคำขอ** (ดู AC-4 — นี่คือจุดที่ต่างจากความเข้าใจทั่วไป)

## Out of scope (v1 ปัจจุบัน)
- Frontend (มีแค่ REST API)
- สร้าง/ลบสาขา (มีแค่ `PUT /admin/branch` แก้รายละเอียดสาขาที่มีอยู่แล้วจาก `cmd/seed`)
- จัดการประเภทห้อง (`room_types`) และรายการสิ่งอำนวยความสะดวกกลาง (`amenities`) — ใส่ผ่าน `cmd/seed` เท่านั้น มีแค่ `PUT /admin/branch/amenities` ให้เลือกจากรายการที่มีอยู่
- รูปห้องหลายรูปต่อห้อง และสิ่งอำนวยความสะดวกรายห้อง — ตาราง `room_images`/`room_amenities` มีอยู่ในฐานข้อมูลแต่ไม่มี API ใช้งาน (ห้องมีได้แค่ `image_url` เดียว)
- ชำระเงินออนไลน์ / ตัดบัตร (ใช้โอนเงินแล้วแนบสลิปให้แอดมินตรวจ)
- คืนเงิน, ค่าปรับ, ใบเสร็จ/ใบกำกับภาษี
- ยืนยันอีเมล, เปลี่ยนอีเมล
- ระงับบัญชีสมาชิก (member) — ระงับได้เฉพาะบัญชี admin ผ่าน `PUT /superadmin/staff/:userID`
- ส่ง SMS ทุกกรณี และส่งอีเมลทุกกรณี — **ไม่มี SMTP/mailer ในระบบเลยสักจุด** (แจ้งเตือนทั้งหมดเก็บในระบบให้ดึงไปแสดงผ่าน `/me/notifications` เท่านั้น)
- Admin/Super Admin รีเซ็ตรหัสผ่านให้ **member** — ทำได้เฉพาะรีเซ็ตรหัสผ่านบัญชี **staff/admin** ผ่าน `PUT /superadmin/staff/:userID` (ดู AC-18)
- Rate limit / account lockout ที่ `/auth/login` — ยอมรับความเสี่ยง brute force ใน v1
- endpoint จบสัญญา/เช็คเอาต์ — ปล่อยห้องด้วย `PATCH /admin/rooms/:roomID/status` แทน
- การจำกัดขนาดหรือชนิดของ `nearby_places.category` — เป็น free text ไม่ใช่ enum (แค่ตกลงกันเป็นธรรมเนียมในโค้ด)

## Acceptance Criteria

AC-1: สมัครสมาชิกได้ และได้สิทธิ์ member เสมอ
  - POST `/auth/register` ด้วยอีเมลใหม่ + รหัสถูกกติกา → 201 พร้อม access_token, refresh_token, user.role = "member"
  - Field ที่รับจริงมีแค่ `email, password, first_name, last_name, phone` — ส่งฟิลด์ที่ไม่รู้จัก (เช่น `role`) มาด้วย → **422 validation_failed** (JSON decoder ปฏิเสธฟิลด์แปลกหน้าทุกช่องทั้งระบบ ไม่ใช่การเช็ค `role` โดยเฉพาะ ดู AC-21)
  - อีเมลซ้ำ (ไม่สนตัวพิมพ์ `A@b.com` = `a@b.com`, เทียบแบบ lowercase) → 422 validation_failed + fields.email
  - อีเมลซ้ำกับบัญชี staff ที่ถูกระงับแล้ว (`is_active=false`) → 422 เช่นกัน (คอลัมน์ `email` unique เสมอ ไม่สนสถานะ active — ระบบไม่มี soft delete แบบ `deleted_at` ที่ไหนเลย มีแต่ `is_active`)
  - กรอกผิดหลายช่อง → 422 และ `fields` มีครบทุกช่องในครั้งเดียว ไม่ใช่ทีละช่อง
  - รหัสผ่านต้องยาว **8–72 ตัว** และมีทั้งตัวอักษรและตัวเลขอย่างน้อยอย่างละ 1 ตัว ไม่ผ่านข้อใด → 422
  - รหัสยาว 73 ตัว → 422 (ห้ามปล่อยให้ bcrypt ตัดทิ้งเงียบ ๆ — cap ตรงกับ bcrypt เป๊ะ)
  - `phone` ต้องตรง regex `^[0-9\-\s()+]{8,20}$` (ตัวเลข/`-`/เว้นวรรค/วงเล็บ/`+` รวม 8–20 ตัวอักษร ไม่ใช่ "เลข 9-10 หลัก" ตามที่เข้าใจกันมาก่อน) ไม่ผ่าน → 422
  - `first_name`/`last_name` ต้องไม่ว่างและยาวไม่เกิน 100 ตัว ไม่ผ่าน → 422
  - body พาร์สไม่ได้หรือไม่ใช่ JSON → 400 (ดู AC-21)

AC-2: เข้าสู่ระบบด้วย endpoint เดียวทุกบทบาท (รวมถึงบัญชีที่สมัครผ่าน OAuth)
  - อีเมล+รหัสถูก → 200 พร้อม token คู่
  - อีเมลไม่มีในระบบ **หรือ** รหัสผิด → 401 code "invalid_credentials" ข้อความเหมือนกันทั้งสองกรณี
  - บัญชีที่สมัครผ่าน OAuth และยังไม่เคยตั้งรหัสผ่าน (`password_hash = ''`) → ล็อกอินด้วยรหัสผ่านไม่ได้ → 401 "invalid_credentials" เหมือนกรณีรหัสผิด (ไม่บอกใบ้ว่าเป็นบัญชี OAuth)
  - จับเวลา 2 กรณี "อีเมลไม่มี" กับ "รหัสผิด" กรณีละ 50 ครั้ง เทียบค่า median → ต่างกันไม่เกิน 20% (ระบบมี dummy bcrypt hash สุ่มตอน boot ไว้เผาเวลาเมื่อไม่พบอีเมลหรือบัญชีไม่มีรหัสผ่านจริง)
  - บัญชี `is_active = false` → 403 code "account_disabled"
    **ยอมรับความเสี่ยง:** โค้ดนี้เปิดเผยว่าอีเมลนั้นมีบัญชีอยู่จริง แลกกับข้อความที่ผู้ใช้เข้าใจได้

AC-3: อายุและการหมุนเวียนของ token
  - access token อายุ **30 นาที** (`ACCESS_TOKEN_TTL_MIN`, default 30) — ใช้หลังหมดอายุ → 401
  - refresh token อายุ **30 วัน** (`REFRESH_TOKEN_TTL_DAY`, default 30)
  - POST `/auth/refresh` ด้วยใบที่ยังไม่หมดอายุ → 200 พร้อม token คู่ใหม่ (rotation: ใบเดิมถูกเพิกถอนทันทีที่ใช้)
  - ใช้ใบเดิมซ้ำอีกครั้ง → 401
  - ใบที่หมดอายุแล้ว / ใบที่ logout ไปแล้ว / สตริงที่ไม่ใช่ token → 401 ข้อความเดียวกันทุกกรณี
  - ตาราง `refresh_tokens` เก็บเฉพาะ sha256 hash ไม่มี token ดิบ
  - เปลี่ยนรหัสผ่านสำเร็จ (`POST /me/password`) → refresh token ทุกใบของบัญชีนั้นใช้ไม่ได้ทันที
    **ยอมรับความเสี่ยง:** access token ใบเดิมยังใช้ได้จนหมดอายุ (สูงสุด 30 นาที)
  - POST `/auth/logout` ด้วย token ที่ผิด/หมดอายุ/ไม่มีจริง → 204 เหมือนกันหมด (ไม่บอกใบ้ว่าใบไหนมีจริง)

AC-4: Admin ถูกล็อกไว้ที่สาขาเดียว ผ่านค่าใน JWT (กฎเหล็ก)
  - `branch_id` ของ admin **ถูกฝังลงใน access token ตอนออก token** ไม่ใช่ query DB ทุกคำขอ (`internal/auth/security.go`) — แลกความเร็วกับความสดของข้อมูล
  - admin สาขา A เรียก endpoint ใต้ `/admin` พร้อม `?branch_id=<สาขา B>` → **403**
  - admin สาขา A เปิดทรัพยากรรายชิ้นของสาขา B (เช่น `GET /bookings/:bookingID`) → **404** (ผ่าน `hideCrossBranch` — แปลง 403 เป็น 404 เฉพาะ endpoint รายชิ้น เพื่อไม่ยืนยันว่ารหัสนั้นมีอยู่จริง)
  - admin สาขา A ไม่ระบุ branch_id → ระบบบังคับใช้สาขา A ไม่ใช่ "ทุกสาขา"
  - GET `/admin/bookings` ของ admin สาขา A → ทุกแถวมี branch_id เป็นสาขา A
  - super admin ไม่ระบุ branch_id → เห็นทุกสาขา (`access.Branch` คืน `nil` = ไม่กรอง)
  - **super admin ย้าย admin จากสาขา A ไป B แล้ว admin คนนั้นใช้ token ใบเดิมต่อ → ยังเห็นสาขา A เหมือนเดิมจนกว่าจะออก access token ใบใหม่ (refresh หรือ login ใหม่)** — ตรงข้ามกับสมมติฐาน "เห็นทันที" เพราะ branch มาจาก JWT ไม่ใช่ query DB สด

AC-5: Member เห็นเฉพาะของตัวเอง
  - member A เปิดการจองของ member B (`GET /bookings/:bookingID`) → 404 (ไม่ใช่ 403 เพื่อไม่ให้รู้ว่ารหัสจองนั้นมีจริง)
  - member เรียก endpoint ใต้ `/admin` หรือ `/superadmin` → 403
  - `POST /me/notifications/read` ส่ง `{"id": "ntf-xxx"}` ที่ไม่ใช่ของตัวเองหรือไม่มีจริง → 204 โดยไม่เปลี่ยนอะไร (ไม่บอกใบ้)
  - `POST /me/notifications/read` **ไม่ส่ง `id` มาเลย** → mark ทุกรายการที่ยังไม่อ่านของ user นั้นเป็นอ่านแล้วทั้งหมด (endpoint เดียวรองรับทั้ง "อ่าน 1 รายการ" และ "อ่านทั้งหมด" ผ่าน `id` ที่เป็น optional)
  - `GET /me/notifications` คืนสูงสุด **30 รายการล่าสุด** พร้อม `unread_count`

AC-6: จองรายวันคิดเงินตามจำนวนคืน
  - **จำนวนคืน = `int((check_out - check_in).Hours() / 24)`** เช่น 5–8 มี.ค. = 3 คืน
  - จอง 3 คืน ห้อง 1,200/คืน → 201 และ `total_amount = 3600`
  - `total_amount` คำนวณและเขียนลงแถวใบจองตอนสร้างครั้งเดียว — แก้ราคาห้องเป็น 2,000 หลังจองแล้ว ไม่กระทบใบเดิม (ไม่มีการคำนวณซ้ำภายหลัง)
  - `check_out_date` ต้องอยู่หลัง `check_in_date` เท่านั้น (ไม่ใช่แค่ ≥) — เท่ากันหรือก่อนหน้า → 422
  - `check_in_date` เป็นวันที่ผ่านมาแล้ว (เทียบเวลาไทย ดู AC-22) → 422
  - จองแบบ `stay_type=daily` บนห้องที่ `stay_type=monthly` (หรือกลับกัน) → 422
  - `room_id` ไม่มีในระบบ หรือ `is_active=false` → 404
  - `room_id` ไม่อยู่ในรูปแบบ `rm-001` → 400
  - ไม่มีข้อจำกัดจำนวนคืนสูงสุด และไม่มีฟิลด์จำนวนผู้เข้าพัก (guest count) ในระบบเลย
  - ฟิลด์ที่ต้องส่งเพิ่ม: `guest_first_name`, `guest_last_name`, `guest_phone` (required), `emergency_phone`/`emergency_relation` (optional)

AC-7: จองรายเดือนเก็บแค่ค่าทำสัญญา
  - ห้อง 5,000/เดือน สาขาเก็บค่าทำสัญญา (`branches.contract_fee`) 500 → 201 และ `total_amount = 500` ไม่ใช่ 5000 (ค่าเช่ารายเดือนจริงไม่ได้ถูกบันทึกไว้ที่ใบจองเลย)
  - `contract_fee` ถูกคัดลอกลงใบจองตอนสร้าง — แก้ค่าทำสัญญาของสาขาเป็น 800 หลังจองแล้ว → `total_amount` ของใบเดิมยังเป็น 500
  - `move_in_date` เป็นวันที่ผ่านมาแล้ว → 422
  - `contract_date` เป็น optional ไม่บังคับส่ง
  - จองแบบ `stay_type=monthly` บนห้องที่เปิดขายแบบ daily → 422

AC-8: ห้องเดียวจองซ้อนไม่ได้ (กฎเหล็ก)
  - **ห้องถูกล็อกเมื่อมีใบสถานะ `pending_payment` / `awaiting_review` / `approved`** — `cancelled` ไม่ล็อก (สถานะที่มีในระบบมีแค่ 4 ค่านี้ ไม่มี `rejected`/`completed` แล้ว — ถูกตัดออกจาก enum ในภายหลัง)
  - **การจองรายเดือน (`stay_type=monthly`) ที่อยู่ในสถานะล็อกจะล็อกทั้งห้องแบบไม่สนวันที่เลย** — มีใบรายเดือนค้างอยู่ใบเดียวพอที่จะกันการจองใหม่ทุกประเภทในห้องนั้น
  - **สองใบรายวันทับกันเมื่อ `ใหม่.check_in < เก่า.check_out AND ใหม่.check_out > เก่า.check_in`** — จองรายวันทับช่วงของใบที่ล็อกอยู่ → 409 code "room_unavailable"
  - จองรายวันที่ `check_in` ตรงกับ `check_out` ของใบเดิมพอดี → **201 สำเร็จ** (ตามสูตรข้างบนถือว่าไม่ทับ)
  - 2 request จองห้องเดียวกันพร้อมกัน → สำเร็จ 1 ใบเท่านั้น (ล็อกแถวห้องด้วย row lock ตลอด transaction ผ่าน `LockForBooking`)
  - จองห้องสถานะ `maintenance` หรือ `occupied` → 409
  - **`booking_code` รูปแบบ `PT-001` (เลขจาก sequence รันต่อเนื่อง, `LPAD` อย่างน้อย 3 หลัก — ไม่ใช่ 6 หลักแบบ `PT-000001`)** เกิน 999 แล้วจะยาวขึ้นเป็น 4+ หลักตามจริง ไม่ตัดทิ้ง
  - **URL ทุกเส้นใช้ entity id ที่มี prefix `bkg-001` ไม่ใช่ booking_code** — `GET /api/v1/bookings/PT-001` → 400 (booking_code ใช้แสดงผลและค้นหาเท่านั้น)

AC-9: แจ้งชำระเงินพร้อมสลิป
  - การจองสถานะ `pending_payment` + แนบสลิป → 200 และสถานะเป็น `awaiting_review`
  - การจองสถานะอื่น (`awaiting_review`/`approved`/`cancelled`) แจ้งซ้ำ → 409 code "invalid_state"
  - `amount` ต้องเป็นค่าบวก (`> 0`) ไม่ผ่าน → 422 — **ไม่มีการเทียบกับ `total_amount` เลยในโค้ด** ยอดที่แจ้งไม่ตรงกับใบจอง **รับเข้าระบบตามปกติเสมอ ไม่ถูก flag พิเศษใด ๆ** — แอดมินต้องเทียบเองจาก response ที่มีทั้ง `booking.total_amount` และ `payment.amount`
  - `transferred_at` ห้ามเป็นอนาคตเกิน **5 นาที** → เกิน → 422 — **ไม่มีขอบเขตย้อนหลังเลย** (แจ้งวันโอนเมื่อ 2 ปีก่อนก็ผ่านการตรวจ ไม่มี validation กันไว้ในปัจจุบัน)
  - `slip` (ไฟล์) ว่าง → 422 — ไม่ส่งไฟล์มาเลย
  - ไฟล์ใหญ่เกิน `MAX_UPLOAD_MB` (default 5) → 400 ก่อนโดน `ParseMultipartForm` เขียนไฟล์ (จำกัดด้วย `http.MaxBytesReader` และ `LimitReader` สองชั้น)
  - **การแจ้งชำระเงินเก็บเป็นประวัติหลายแถวต่อ 1 การจอง** (`payments.booking_id` ไม่มี unique constraint) ห้ามเขียนทับแถวเดิม
  - ไฟล์ที่บันทึกตั้งชื่อด้วย UUID ไม่ใช้ชื่อจาก client

AC-10: อนุมัติ/ปฏิเสธการจอง
  - อนุมัติใบสถานะ `awaiting_review` เท่านั้น → 200 สถานะเป็น `approved`
  - อนุมัติใบ **รายเดือน** → ห้องเปลี่ยนเป็น `occupied` ในทรานแซกชันเดียวกัน
  - **ปฏิเสธใบสถานะ `awaiting_review` → 200 และใบกลับไปสถานะ `pending_payment`** โดยห้องยังถูกล็อกอยู่ (สถานะยังอยู่ในกลุ่มล็อก ดู AC-8)
    - `reason` เก็บที่แถว **payment** (`payments.reject_reason`) ไม่ใช่ที่ใบจอง — คอลัมน์ `bookings.reject_reason` เดิมถูกลบไปแล้วในภายหลัง
  - ปฏิเสธโดยไม่ระบุ `reason` → 422
  - อนุมัติหรือปฏิเสธใบที่ไม่ได้อยู่สถานะ `awaiting_review` → 409 code "invalid_state"
  - admin สาขา A อนุมัติใบของสาขา B → 404 (เป็นทรัพยากรรายชิ้น ดู AC-4)
  - อนุมัติหรือปฏิเสธสำเร็จ → สมาชิกได้แจ้งเตือน 1 รายการ

AC-11: ยกเลิกแล้วห้องต้องกลับมาขายได้ (กฎเหล็ก)
  - member ยกเลิกใบสถานะ `pending_payment` ของตัวเอง → 200 สถานะเป็น `cancelled`
  - **member ยกเลิกใบสถานะ `awaiting_review` หรือ `approved` เอง → 409** (ต้องให้แอดมินจัดการแทน — เช็คเฉพาะฝั่ง member เท่านั้น)
  - admin/super admin ในสาขาของตัวเอง ยกเลิกใบสถานะใดก็ได้ที่ยังไม่ cancelled → 200
  - ยกเลิกใบที่ยกเลิกไปแล้ว → 409 code "invalid_state"
  - ยกเลิกใบรายเดือนที่อนุมัติแล้ว → ห้องกลับเป็น `available` (ยังไม่มีเทสยืนยันกรณีห้องถูกเปลี่ยนเป็น `maintenance` ไปแล้วระหว่างทาง — ถือเป็นช่องว่างที่ต้องเพิ่มเทสก่อนเชื่อพฤติกรรมนี้ 100%)
  - ห้องที่ถูกปล่อยแล้ว → จองใหม่ได้ทันที

AC-12: ค้นหาห้องว่าง
  - ระบุ `check_in`/`check_out` → ไม่มีห้องที่มีการจองรายวันช่วงคาบเกี่ยวกันในผลลัพธ์ (สูตรทับกันเดียวกับ AC-8) และไม่มีห้องที่มีใบรายเดือนล็อกอยู่
  - ห้องที่ `check_out` ของใบเดิมตรงกับ `check_in` ที่ค้นหา → **ต้องอยู่ในผลลัพธ์**
  - ห้องสถานะ `maintenance`/`occupied`, ห้องที่ `is_active=false`, ห้องของสาขาที่ปิดใช้งาน → ไม่อยู่ในผลลัพธ์
  - `page_size=500` → คืนไม่เกิน **100** รายการ (`maxPageSize=100`)
  - ไม่ระบุ `page` → หน้า 1 ขนาด 20 พร้อม `meta.total_items`/`meta.total_pages`
  - `page`/`page_size` เป็น 0 / ติดลบ / ไม่ใช่ตัวเลข → **400** (ไม่ถูก clamp เงียบ ๆ)
  - `stay_type` เป็นค่าที่ไม่รู้จัก → 400
  - `branch_id`/`room_id` ที่ไม่อยู่ในรูปแบบ `brn-001`/`rm-001` → 400

AC-13: บันทึกทุกการกระทำที่เปลี่ยนข้อมูล
  - action string ที่มีจริงในโค้ดตอนนี้ (25 รายการ):
    `auth.register`, `auth.login`, `auth.oauth_login`, `auth.oauth_link`, `auth.oauth_register`,
    `user.update_profile`, `user.change_password`,
    `admin.create`, `admin.update`, `admin.delete`,
    `booking.create`, `booking.submit_payment`, `booking.cancel`, `booking.approve`, `booking.reject`, `booking.set_appointment`,
    `branch.update`, `branch.update_nearby`, `branch.update_cover`, `branch.delete_image`,
    `room.create`, `room.update`, `room.update_image`, `room.update_status`, `room.delete`
  - **ช่องว่างที่รู้แล้ว:** `PUT /admin/branch/amenities` (`SetAmenities`) และ `POST /admin/branch/images` (`AddImage`) **ไม่ได้เขียน activity log** — ควรเพิ่ม `branch.set_amenities`/`branch.add_image` ทีหลัง
  - ทำ action ใด ๆ ในลิสต์ข้างบน → `activity_logs` มีแถวใหม่: ผู้ทำ (id+ชื่อ+role), สาขา, action, entity, ip, เวลา, `detail` (jsonb)
  - เขียน log ล้มเหลว → ธุรกรรมหลักที่สำเร็จแล้วต้องไม่ rollback ตาม
  - admin สาขา A เรียก `GET /admin/activity-logs` → เห็นเฉพาะ log สาขา A (super admin ไม่กรอง branch_id → เห็นทุกสาขา)
  - filter รองรับ: `action`, `search`, `actor_id`, `actor_role` (ต้องเป็น role ที่รู้จัก ไม่งั้น 400), `branch_id`, `page`/`page_size`

AC-14: รูปภาพสาขา/ห้อง — รองรับสองทาง (อัปโหลดเป็นไฟล์ และรับ URL ภายนอก)
  - **ทางที่ 1 — อัปโหลดไฟล์จริง**: `POST /admin/branch/images/upload`, `POST /admin/branch/cover`, `POST /admin/rooms/:roomID/image` (multipart `image`) → เก็บเนื้อไฟล์เป็น blob ในตาราง `assets` (dedup ด้วย sha256 ผ่าน `UNIQUE(checksum)`) แล้วคืน `image_url`/`cover_image_url` เป็น `{PUBLIC_BASE_URL}/files/{assetID}`
  - `GET /files/{assetID}` เปิดสาธารณะ ไม่ต้องมี token ตอบ `ETag`+`Cache-Control: immutable`
  - ชนิดไฟล์ตรวจจากเนื้อไฟล์จริง ไม่เชื่อนามสกุลหรือ `Content-Type` ที่ client ส่งมา
  - ไฟล์ใหญ่เกิน `MAX_UPLOAD_MB` → 400 (จำกัดเดียวกับสลิป AC-9)
  - **ทางที่ 2 — ใส่ URL ภายนอก**: `POST /admin/branch/images` (JSON `{image_url, caption, sort_order}`) → validate ว่าเป็น `http`/`https` ที่มี host จริง, ยาวไม่เกิน 2048 ตัว, ปฏิเสธ `javascript:`/`data:`/path สัมพัทธ์ที่ไม่มี host → ไม่ผ่าน → 422
  - `DELETE /admin/branch/images/:imageID` → 204 | 403 คนละสาขา | 404
  - ห้องมีได้แค่ `image_url` เดียว (ไม่มี gallery หลายรูปต่อห้องผ่าน API — ดู Out of scope)

AC-15: ความปลอดภัยและความทนทานพื้นฐาน
  - ทุก query ใช้ parameterized ไม่มีการต่อ string ค่าเข้า SQL — **ตรวจด้วย code review ไม่ใช่เทสผ่าน API**
  - error 5xx → body ต้องไม่มี stack trace, ชื่อตาราง หรือข้อความ SQL (log รายละเอียดจริงไว้ฝั่ง server เท่านั้น)
  - panic ใน handler → server ไม่ล้ม และ client ได้ 500 เป็น JSON
  - ข้อความ error ทุกตัวที่ผู้ใช้เห็นเป็นภาษาไทย รวมถึง error ระดับ router (404 เส้นทางไม่มีจริง, 405 method ผิด, body พาร์สไม่ได้)
  - `JWT_SECRET` สั้นกว่า 32 ตัว หรือ `DATABASE_URL` ว่าง → server ไม่ start เลย (fail fast ตอน boot)

AC-16: การจองที่ไม่จ่ายเงินหมดอายุเอง **[ยังไม่ implement]**
  - **สถานะปัจจุบันของโค้ด:** ไม่มีคอลัมน์ `expires_at` ในตาราง `bookings`, ไม่มี cron/worker/goroutine ใด ๆ คอยเช็คหมดอายุ — ใบจองสถานะ `pending_payment` ที่ไม่มีใครแจ้งชำระเงินจะค้างอยู่แบบนั้น**ตลอดไป** จนกว่าจะมีคนกด cancel เอง หรือ admin ยกเลิกให้
  - **พฤติกรรมที่ตกลงจะทำต่อ (ยังไม่มีในโค้ด):**
    - สร้างใบจอง → ตั้ง `expires_at` = เวลาปัจจุบัน + 10 นาที
    - ถูกปฏิเสธแล้วกลับเป็น `pending_payment` → ตั้ง `expires_at` ใหม่ = เวลาที่ปฏิเสธ + 24 ชั่วโมง
    - ใบที่เลย `expires_at` และยังเป็น `pending_payment` → ระบบ (ไม่ใช่คน) เปลี่ยนเป็น `cancelled` และห้องถูกปล่อยตาม AC-11
    - action ใน activity log ควรเป็น `booking.auto_expire` โดย `actor_role = "system"`, `actor_id` เป็น NULL (ต้องแก้ schema ให้ `actor_id` เป็น nullable ก่อน — ตอนนี้ยังไม่ได้เช็คว่าเป็น nullable อยู่แล้วหรือไม่)
  - ห้ามเขียนเทสที่ยืนยันว่ามี auto-expire จนกว่าจะ implement จริง

AC-17: ไฟล์สลิปโอนเงิน — สถานะปัจจุบัน **ยังไม่มีการตรวจสิทธิ์ก่อนเสิร์ฟ**
  - **สถานะปัจจุบันของโค้ด:** สลิปเก็บเป็นไฟล์บนดิสก์ที่ `UPLOAD_DIR` เสิร์ฟผ่าน `GET /uploads/*filepath` ซึ่งเป็น route **นอก** `/api/v1` และ **ไม่มี auth middleware คลุมอยู่เลย** — ใครก็ตามที่รู้ URL (ชื่อไฟล์เป็น UUID เดาไม่ได้ แต่ไม่มีการเช็คว่าเป็นเจ้าของจริงหรือไม่) เปิดดูสลิปได้
  - ไฟล์เก็บนอก directory ที่เว็บเสิร์ฟตรง ๆ ได้ (path ใน DB เท่านั้น ห้าม client ส่ง path มาต่อเอง — กัน path traversal ได้ แต่ไม่กันการเข้าถึงโดยไม่มีสิทธิ์)
  - **TODO ที่ตกลงจะทำต่อ:** ย้ายมาเสิร์ฟผ่าน endpoint ที่ตรวจสิทธิ์ เช่น `GET /api/v1/bookings/:bookingID/slip` (เจ้าของใบจอง / admin สาขาเดียวกัน / super admin เท่านั้น → 200, คนอื่น/ไม่มีจริง → 404 เหมือนกันหมด, ไม่มี token → 401) — **endpoint นี้ยังไม่มีอยู่ในโค้ดปัจจุบัน**

AC-18: Super Admin จัดการบัญชีผู้ดูแล
  - `POST /superadmin/staff` รับแค่ `email, first_name, last_name, phone, branch_id` — **ไม่รับ password** → 201 และ role เป็น admin
  - รหัสผ่านตั้งอัตโนมัติจาก `STAFF_DEFAULT_PASSWORD` (ถ้าไม่ตั้ง env นี้ fallback ไปที่ `SEED_DEFAULT_PASSWORD`, ค่า default สุดท้ายคือ `"Wisetsuk!2026"`) และตั้ง `must_change_password = true` ทันที — ค่านี้ติดมากับ `GET /me` ให้ frontend บังคับเปลี่ยนได้
  - อีเมลซ้ำ (รวมถึงซ้ำกับบัญชีที่ระงับแล้ว) → 422
  - `branch_id` ไม่พบ → 422
  - `PUT /superadmin/staff/:userID` ระงับตัวเอง (`is_active=false`) → 400
  - `DELETE /superadmin/staff/:userID` ลบตัวเอง → 400
  - เป้าหมายเป็น member หรือไม่มีจริง → 404
  - `PUT /superadmin/staff/:userID` ส่ง `password` มาด้วย (optional field) → ตั้งรหัสผ่านใหม่ให้ทันที **และ** ตั้ง `must_change_password = true` กลับมาอีกครั้ง — นี่คือช่องทางเดียวที่ Super Admin รีเซ็ตรหัสผ่านให้คนอื่นได้ (เฉพาะบัญชี staff เท่านั้น)
  - ลบหรือระงับ staff สำเร็จ → refresh token ทุกใบของบัญชีนั้นถูกเพิกถอนทันที
  - **การลบ staff เป็นการตั้ง `is_active=false` (ไม่มี `deleted_at` ในระบบเลย)** — `GET /superadmin/staff` ไม่แสดงคนที่ถูกลบ แต่ `activity_logs` เดิมยังแสดงชื่อผู้ทำได้ครบ
  - ย้ายสาขา staff → **ไม่**เพิกถอน token และผู้ใช้เห็นสาขาใหม่ก็ต่อเมื่อออก access token ใบใหม่แล้ว (ดู AC-4 — ไม่ใช่ "เห็นทันที")

AC-19: จัดการห้องและการปล่อยห้อง
  - field ของ `POST`/`PUT /admin/rooms`: `branch_id, room_type_id, room_number, building, floor, stay_type, price, water_rate, electric_rate, size_sqm, description, image_url, status`
  - `room_number` ซ้ำในสาขาเดียวกัน (`UNIQUE(branch_id, room_number)` แบบไม่มีเงื่อนไข) → 409/422 แล้วแต่ error mapping
  - **ลบห้อง = ตั้ง `is_active = false` แบบไม่มีเงื่อนไข ไม่เช็คว่ามีใบจองค้างอยู่หรือไม่เลย** — ลบห้องที่มีใบสถานะ `pending_payment`/`awaiting_review`/`approved` ค้างอยู่ก็สำเร็จ 204 ปกติ (ช่องว่างที่ควรพิจารณาเพิ่ม guard ทีหลัง)
  - **`room_number` ของห้องที่ถูกลบไปแล้ว (`is_active=false`) นำมาใช้ซ้ำไม่ได้** เพราะ unique constraint ไม่มี `WHERE is_active` — สร้างห้องเลขเดิมซ้ำ → ชนกับห้องเก่าที่ถูกปิดใช้งานไปแล้ว (คนละพฤติกรรมจากระบบที่ใช้ `deleted_at` แบบ partial unique index)
  - `price` ต้องเป็นค่าบวก (`>0`), `floor>0`, `water_rate>=0`, `electric_rate>=0` ไม่ผ่าน → 422
  - `PATCH /admin/rooms/:roomID/status` **ไม่มีการจำกัดว่าเปลี่ยนจากสถานะไหนไปไหนได้บ้าง** — ส่งค่าที่ valid ใน 3 ค่า (`available`/`occupied`/`maintenance`) ผ่านได้เสมอไม่ว่าสถานะปัจจุบันจะเป็นอะไร และ **ไม่เช็คว่ามีใบจอง active ค้างอยู่หรือไม่เลย** (ต่างจากที่เคยเข้าใจว่าจะมี 409 ตอนเปลี่ยนเป็น maintenance ขณะมีใบค้าง — ปัจจุบันไม่มี guard นี้)
  - `status` ที่ไม่อยู่ใน 3 ค่าที่กำหนด → 422

AC-20: นัดหมายทำสัญญา
  - `PUT /admin/bookings/:bookingID/appointment` บนใบรายวัน → 400
  - `appointment_at` เป็นเวลาในอดีต → 422
  - นัดหมายสำเร็จ → สมาชิกได้แจ้งเตือน 1 รายการที่อ้าง booking_code
  - (มีเทส state machine คุมเงื่อนไข "ต้องเป็นใบรายเดือนที่ approved แล้ว" อยู่ที่ `booking/statemachine_test.go` — ผิดสถานะ → 409 code "invalid_state")

AC-21: รหัสสถานะและรูปแบบ error สอดคล้องกันทั้งระบบ
  - format error: `{"error": {"code": "...", "message": "...", "fields": {...}}}` (`fields` ปรากฏเฉพาะตอน code = "validation_failed" เท่านั้น)
  - format สำเร็จ: `{"data": ...}` และรายการที่แบ่งหน้าจะมี `meta: {page, page_size, total_items, total_pages}` เพิ่ม
  - **มี error code ทั้งหมด 12 ตัวพอดีในโค้ด (คอมเมนต์ในซอร์สยืนยันเลขนี้ตรงกัน):**
    `unauthorized`, `forbidden`, `not_found`, `conflict`, `bad_request`, `validation_failed`, `method_not_allowed`, `internal_error`, `invalid_credentials`, `account_disabled`, `room_unavailable`, `invalid_state`
  - **400 bad_request** ใช้เมื่อ: พาร์ส body ไม่ได้, ชนิดข้อมูลผิด, query param ผิดรูป (branch_id ไม่อยู่ในรูปแบบ `brn-001`), ไฟล์อัปโหลดไม่ถูกต้อง/ใหญ่เกิน
  - **422 validation_failed** ใช้เมื่อ: ค่าถูกชนิดแต่ผิดกติกาธุรกิจ **หรือ** body มีฟิลด์ที่ไม่รู้จัก (JSON decoder เข้มงวด — unknown field → 422 ไม่ใช่ 400)
  - ไม่มี token หรือ token ใช้ไม่ได้ → 401 · มี token แต่บทบาทไม่พอ → 403
  - error ทุกตัวรวมถึง 404/405/500 ใช้ format เดียวกันหมด ไม่มี response ที่หลุดออกนอกรูปแบบนี้

AC-22: เวลาและวันที่
  - เก็บทุก timestamp เป็น **UTC** ในฐานข้อมูล (`TIMESTAMPTZ` ทุกคอลัมน์)
  - ตัดสิน "วันนี้ / อดีต / อนาคต" ด้วยโซนเวลาไทย (`internal/timex`, มีเทสคุมที่ `timex_test.go`) สำหรับ `check_in_date`, `move_in_date`, `transferred_at`, `appointment_at`
  - วันที่ใน response เป็น ISO 8601 พร้อม offset

AC-23: ลืมรหัสผ่าน — ขอลิงก์รีเซ็ตทางอีเมล **[ยังไม่ implement]**
  - **สถานะปัจจุบันของโค้ด:** ไม่มี route `/auth/forgot-password` หรือ `/auth/reset-password` เลย, ไม่มี SMTP/mailer package ในระบบ, ไม่มีตาราง `password_resets`, ไม่มี rate limiter ใด ๆ — member ที่ลืมรหัสผ่านตอนนี้ **ไม่มีทางกู้คืนเองได้เลย** ต้องให้ Super Admin ช่วย แต่ Super Admin ก็รีเซ็ตได้แค่บัญชี staff เท่านั้น (ดู AC-18, Out of scope)
  - **พฤติกรรมที่ตกลงจะทำต่อ (ยังไม่มีในโค้ด) — คงไว้ตามสเปกเดิมเพื่อเป็นแนวทางตอน implement จริง:**
    - `POST /auth/forgot-password {email}` → ตอบ **204 เหมือนกันทุกกรณี** ไม่ว่าอีเมลจะมีบัญชีหรือไม่ ไม่บอกใบ้ว่าอีเมลใดมีบัญชี, เผาเวลาเท่ากันด้วย dummy hash เหมือน AC-2
    - บัญชี `is_active=false` → 204 เช่นกัน แต่ไม่ส่งอีเมลจริง
    - token เก็บเฉพาะ sha256 hash (เหมือน refresh token), อายุ 15 นาที, ใช้ได้ครั้งเดียว, ขอใบใหม่แล้วใบเก่าใช้ไม่ได้ทันที
    - จำกัด 5 ครั้ง/ชม./อีเมล นับที่ DB ไม่ใช่ในหน่วยความจำ
    - ลิงก์ชี้ไปหน้าเว็บ frontend ด้วย env ตัวใหม่ (ระบบยังไม่มี `PUBLIC_APP_URL` — มีแค่ `PUBLIC_BASE_URL` ที่เป็นโดเมน API สำหรับไฟล์รูป ต้องแยกตัวแปรกันชัดเจนตอน implement)
    - `POST /auth/reset-password {token,new_password,confirm_password}` สำเร็จ → 204, เพิกถอน refresh token ทุกใบ, บันทึก activity log `user.reset_password`
    - `APP_ENV=production` โดยไม่ตั้งค่า SMTP ที่จำเป็น → server ไม่ start (เหมือนที่ `JWT_SECRET`/`DATABASE_URL` เช็คตอน boot อยู่แล้ว — ดู AC-15)
  - ห้ามเขียนเทสที่ยืนยันว่ามี flow นี้จนกว่าจะ implement จริง

AC-24: เข้าสู่ระบบด้วย Google/Facebook OAuth
  - ผู้ให้บริการที่รองรับ: `google`, `facebook` เท่านั้น (บังคับด้วย DB check `user_identities.provider IN ('google','facebook')`) — provider ที่ไม่ได้ตั้ง `GOOGLE_CLIENT_ID`/`FACEBOOK_CLIENT_ID` ฯลฯ จะถูกปิดใช้งานเงียบ ๆ (ไม่ fatal ตอน boot)
  - `GET /auth/oauth` → คืนรายชื่อ provider ที่เปิดใช้งานอยู่จริงเท่านั้น
  - `GET /auth/oauth/:provider` → provider ปิดอยู่ → 404, provider เปิดอยู่ → ตั้ง cookie `HttpOnly`/`SameSite=Lax` อายุ 10 นาทีสำหรับ PKCE state+verifier แล้ว 302 ไปหน้า consent ของ provider
  - `GET /auth/oauth/:provider/callback` → ตรวจ `state` ตรงกับ cookie, แลก `code` เป็น token ฝั่ง server (PKCE), จากนั้น **302 กลับไป frontend เสมอ** (`FRONTEND_OAUTH_CALLBACK_URL`) พร้อม `?code=...` (exchange code ใช้ครั้งเดียว) หรือ `?error=...` — endpoint นี้ไม่เคยตอบ JSON ตรง ๆ
  - `POST /auth/oauth/exchange {code}` → แลก exchange code (อายุ **2 นาที**, เก็บ hash ใน `oauth_exchange_codes`, ใช้ได้ครั้งเดียว) เป็น access+refresh token จริง
  - ล็อกอินครั้งแรกด้วย OAuth: มี `user_identities` ของ provider+subject นี้อยู่แล้ว → login เข้าบัญชีเดิม (`auth.oauth_login`); ไม่มีแต่มีบัญชีอีเมลเดียวกันอยู่แล้ว → ผูก identity เข้ากับบัญชีเดิม (`auth.oauth_link`); ไม่มีทั้งคู่ → สร้างบัญชีใหม่ role `member` โดย `password_hash=''` (`auth.oauth_register`)
  - บัญชีที่ยังไม่เคยตั้งรหัสผ่าน (`password_hash=''`) เรียก `POST /me/password` → ระบบข้ามการเช็ค current_password ให้ (ถือเป็นการตั้งรหัสผ่านครั้งแรก ไม่ใช่เปลี่ยนรหัส)

AC-25: แดชบอร์ดและรายงานสำหรับ Admin
  - `GET /admin/dashboard` คืน `{branches: [...], recent_activities: [...]}` — `branches` เป็น array ของสถิติรายสาขา (`branch_id, branch_name, daily_rooms_free, daily_rooms_total, monthly_rooms_free, monthly_rooms_total, pending_review, bookings_total`) ไม่มีตัวเลขสรุปรวมทุกสาขาแยกต่างหาก
  - `recent_activities` คืน **ล่าสุด 10 รายการพอดี** (`const dashboardActivities = 10`) จาก activity log ที่ admin เห็นได้ (สาขาตัวเองหรือทุกสาขาถ้าเป็น super admin)
  - `GET /admin/activity-logs` filter: `action`, `search`, `actor_id`, `actor_role`, `branch_id`, `page`/`page_size` — ทุก endpoint ใน AC นี้ยังไม่มีเทสอัตโนมัติคุม (`reporting` module ไม่มีไฟล์ `_test.go`) ถือเป็น spec ที่ทดสอบได้แต่ยังไม่ถูกทดสอบจริง

## Error format (ทั้งระบบ)
```
{"error": {"code": "machine_readable_code", "message": "ข้อความไทย", "fields": {"ชื่อฟิลด์": "ข้อความไทย"}}}
```
`fields` มีเฉพาะตอน code = "validation_failed"
code ที่ใช้ (12 ตัวเท่านั้น): `unauthorized` · `forbidden` · `not_found` · `conflict` · `bad_request` · `validation_failed` ·
`method_not_allowed` · `internal_error` · `invalid_credentials` · `account_disabled` · `room_unavailable` · `invalid_state`

## API Contract

```
GET  /health                                  → 200 {"status":"healthy","time"}
GET  /files/:assetID                          → 200 ไฟล์ (public, ETag+immutable) | 404
GET  /uploads/*filepath                       → 200 ไฟล์ (public, ไม่มี auth — ดู AC-17) | 404

--- Guest ---
POST /api/v1/auth/register {email,password,first_name,last_name,phone}
                                              → 201 | 400 | 422
POST /api/v1/auth/login {email,password}      → 200 | 401 invalid_credentials | 403 account_disabled
POST /api/v1/auth/refresh {refresh_token}     → 200 | 401
POST /api/v1/auth/logout {refresh_token}      → 204 เสมอ
GET  /api/v1/auth/oauth                       → 200 {"providers":[...]}
GET  /api/v1/auth/oauth/:provider             → 302 ไป provider | 404 provider ปิดอยู่
GET  /api/v1/auth/oauth/:provider/callback    → 302 กลับ frontend พร้อม ?code= หรือ ?error=
POST /api/v1/auth/oauth/exchange {code}       → 200 token คู่ | 401
GET  /api/v1/branches                         → 200 [branch]
GET  /api/v1/branches/:branchID               → 200 branch+images+amenities+nearby | 400 | 404
GET  /api/v1/amenities                        → 200 [amenity]
GET  /api/v1/room-types?branch_id=            → 200 [room_type]
GET  /api/v1/rooms/search?branch_id=&room_type_id=&stay_type=&check_in=&check_out=
     &move_in_date=&min_price=&max_price=&page=&page_size=
                                              → 200 {data,meta} | 400
GET  /api/v1/rooms/:roomID                    → 200 room | 400 | 404

--- ต้องล็อกอิน (ทุกบทบาท) ---
GET  /api/v1/me                               → 200 user (รวม must_change_password) | 401
PUT  /api/v1/me {first_name,last_name,phone,avatar_url}
                                              → 200 | 422
POST /api/v1/me/password {current_password?,new_password,confirm_password}
                                              → 200 (เพิกถอน refresh token ทุกใบ) | 422
     current_password ไม่บังคับถ้าบัญชียังไม่เคยมีรหัสผ่าน (OAuth-only, ดู AC-24)
GET  /api/v1/me/notifications                 → 200 {items(≤30),unread_count}
POST /api/v1/me/notifications/read {id?}      → 204 (ไม่ส่ง id = อ่านทั้งหมด)
GET  /api/v1/bookings/:bookingID              → 200 booking | 404 (ไม่ใช่เจ้าของ/คนละสาขา)
POST /api/v1/bookings/:bookingID/cancel       → 200 | 409 invalid_state | 404

--- Member ---
POST /api/v1/bookings {room_id,stay_type,guest_first_name,guest_last_name,guest_phone,
     emergency_phone?,emergency_relation?,check_in_date+check_out_date|move_in_date+contract_date?}
                                              → 201 | 400 | 404 ไม่พบห้อง | 409 room_unavailable | 422
GET  /api/v1/bookings?status=&stay_type=&page= → 200 {data,meta}
     status ∈ pending_payment | awaiting_review | approved | cancelled
POST /api/v1/bookings/:bookingID/payment      multipart: slip,amount,transferred_at,note
                                              → 200 | 400 ไฟล์ผิด/ใหญ่เกิน | 409 invalid_state | 422

--- Admin (+ Super Admin) ใต้ /api/v1/admin ---
GET    /dashboard                             → 200 {branches:[...],recent_activities:[≤10]}
GET    /activity-logs?actor_role=&actor_id=&action=&branch_id=&search=&page=
                                              → 200 {data,meta} | 400 | 403 ข้ามสาขา
GET    /bookings?status=&stay_type=&search=&branch_id=&page=
                                              → 200 {data,meta} | 403 ข้ามสาขา
POST   /bookings/:bookingID/approve           → 200 | 409 invalid_state | 404 ข้ามสาขา
POST   /bookings/:bookingID/reject {reason}   → 200 (กลับเป็น pending_payment) | 422 ไม่ระบุ reason | 409 | 404
PUT    /bookings/:bookingID/appointment {appointment_at,note}
                                              → 200 | 400 ไม่ใช่รายเดือน | 409 | 422
GET    /members?search=&page=                 → 200 {data,meta}
GET    /members/:memberID/bookings?page=      → 200 {data,meta}
GET    /rooms?branch_id=&...&page=            → 200 {data,meta} | 403
POST   /rooms {branch_id,room_type_id,room_number,building,floor,stay_type,price,
     water_rate,electric_rate,size_sqm,description,image_url,status}
                                              → 201 | 422 เลขห้องซ้ำ/ราคาไม่ถูกต้อง | 403
PUT    /rooms/:roomID                         → 200 | 422 | 403 | 404
POST   /rooms/:roomID/image                   multipart: image → 200 | 400 | 403 | 404
PATCH  /rooms/:roomID/status {status}         → 200 (ไม่จำกัด transition) | 422 สถานะไม่ถูกต้อง | 403
DELETE /rooms/:roomID                         → 204 (ตั้ง is_active=false ไม่เช็คใบจองค้าง) | 403 | 404
PUT    /branch {name,tagline,description,address,phones,line_id,email,latitude,longitude,
     map_url,building_count,floor_count,daily_price_from,monthly_price_min,monthly_price_max,
     water_rate,electric_rate,deposit,advance_payment,contract_fee,cover_image_url}
                                              → 200 | 422 | 403
PUT    /branch/amenities {amenity_ids}        → 200 | 403 (ไม่เขียน activity log — ดู AC-13)
PUT    /branch/nearby {items}                 → 200 | 422 | 403
POST   /branch/cover                          multipart: image → 200 | 400 | 403
POST   /branch/images/upload                  multipart: image,caption,sort_order → 201 | 400 | 403
POST   /branch/images {image_url,caption,sort_order}
                                              → 201 | 422 URL ไม่ถูกต้อง | 403 (ไม่เขียน activity log)
DELETE /branch/images/:imageID                → 204 | 403 | 404

--- Super Admin เท่านั้น ใต้ /api/v1/superadmin ---
GET    /branches                              → 200 [branch] รวมที่ปิดใช้งาน
GET    /staff                                 → 200 [user]
POST   /staff {email,first_name,last_name,phone,branch_id}
                                              → 201 (must_change_password=true, รหัสผ่านจาก STAFF_DEFAULT_PASSWORD) | 422 อีเมลซ้ำ/ไม่พบสาขา
PUT    /staff/:userID {first_name,last_name,phone,branch_id,is_active,password?}
                                              → 200 | 400 ระงับ/ลบตัวเอง | 404 เป้าหมายเป็น member
DELETE /staff/:userID                         → 204 (is_active=false) | 400 ลบตัวเอง | 404

--- ยังไม่มีในโค้ด (ดู AC-16, AC-17, AC-23) ---
POST /api/v1/auth/forgot-password             [ยังไม่ implement]
POST /api/v1/auth/reset-password              [ยังไม่ implement]
GET  /api/v1/bookings/:bookingID/slip         [ยังไม่ implement — ปัจจุบันสลิปเสิร์ฟผ่าน /uploads/* แบบไม่มี auth]
```
