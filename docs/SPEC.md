# SPEC ระบบจองห้องพักวิเศษสุขนครคอนโด (Backend)

> Reverse-engineer จากโค้ดจริงใน `backend/` ณ 2026-09-24 (migration 0001–0007) — ไม่ได้อ้างอิง README หรือ openapi.yaml
> ทุกบรรทัดคือพฤติกรรมที่ยิงผ่าน API หรือ query DB ตรวจได้ ในรูป "ทำแบบนี้ → ได้ผลแบบนี้"
>
> ป้ายกำกับ:
> - **[ยังไม่ implement]** — ตกลงกันว่าจะทำ แต่โค้ดยังไม่มี ห้ามเขียนเทสยืนยันจนกว่าจะทำจริง
> - **[ช่องว่าง]** — โค้ดทำแบบนี้อยู่จริง แต่น่าจะไม่ได้ตั้งใจ หรือขัดกฎใน AGENTS.md — human ต้องตัดสินว่าจะแก้หรือยอมรับ
> - **ยอมรับความเสี่ยง** — ตัดสินใจแล้วว่ารับได้
> - **[ยังไม่ยืนยัน]** — อนุมานจากการอ่านโค้ด ยังไม่ได้รันเทสยืนยัน

## Users
- Guest: ดูสาขา ดูห้อง ค้นหาห้องว่างตามวันที่/ราคา/ประเภท — ไม่ต้องล็อกอิน
- Member: จองห้อง (รายวัน/รายเดือน), แจ้งชำระเงินพร้อมสลิป, ดู/ยกเลิกการจองของตัวเอง, รับแจ้งเตือน, ล็อกอินด้วยอีเมล/รหัสผ่านหรือ Google/Facebook (ดู AC-24)
- Admin: จัดการห้องและรายละเอียดสาขา, อนุมัติ/ปฏิเสธ/ยกเลิกการจอง, นัดทำสัญญา, ดูแดชบอร์ด/activity log/รายชื่อสมาชิก — **เฉพาะสาขาที่ผูกอยู่ 1 สาขา** (ดู AC-4)
- Super Admin: ทุกอย่างที่ Admin ทำได้ในทุกสาขา + เพิ่ม/แก้/ระงับ/ลบบัญชีผู้ดูแล (ดู AC-18)
- สิทธิ์ถูกบังคับที่ DB:
  - `admin_requires_branch` — admin ต้องมี `branch_id` และบทบาทอื่นต้องไม่มี
  - `uq_admin_per_branch` — หนึ่งสาขามี admin ที่ยังไม่ถูกลบได้คนเดียว
  - `uq_single_superadmin` — ทั้งระบบมี superadmin ได้คนเดียว
- token ใช้ยืนยันแค่ว่าผู้เรียกเป็นใคร ส่วน role, `branch_id` และสถานะบัญชีอ่านจาก DB ทุกคำขอ → ลบ ระงับ หรือย้ายสาขาจึงมีผลทันที (ดู AC-3, AC-4)

## Out of scope (สิ่งที่โค้ดปัจจุบันไม่มี)
- Frontend — ระบบนี้มีแค่ REST API
- สร้าง ลบ หรือเปิด/ปิดสาขา — มีแค่ `PUT /admin/branch` แก้สาขาที่มาจาก `cmd/seed` คอลัมน์ `is_active` ของสาขาเปลี่ยนได้ทาง DB เท่านั้น
- จัดการประเภทห้อง (`room_types`) และรายการสิ่งอำนวยความสะดวกกลาง (`amenities`) — ใส่ได้ผ่าน `cmd/seed` เท่านั้น API ทำได้แค่เลือกจากรายการที่มีอยู่แล้ว
- สร้าง Super Admin ผ่าน API — สร้างได้ทาง seed เท่านั้น
- ชำระเงินออนไลน์/ตัดบัตร — ใช้วิธีโอนเงินแล้วแนบสลิปให้แอดมินตรวจ
- คืนเงิน ค่าปรับ ใบเสร็จ/ใบกำกับภาษี
- ยืนยันอีเมล และเปลี่ยนอีเมล
- ระงับหรือลบบัญชี member — ทำได้เฉพาะบัญชี staff (ดู AC-18)
- ส่งอีเมลหรือ SMS — ในระบบไม่มี mailer เลยสักจุด แจ้งเตือนทั้งหมดเก็บในตาราง `notifications` แล้วดึงผ่าน `/me/notifications`
- รีเซ็ตรหัสผ่านของ member — ทำได้เฉพาะบัญชี staff (ดู AC-18) ส่วนลืมรหัสผ่านด้วยตัวเอง **[ยังไม่ implement]** (ดู AC-23)
- Rate limit ที่ endpoint อื่นนอกจาก `/auth/login` (register, refresh, oauth exchange) — มีเฉพาะ login (ดู AC-2)
- endpoint จบสัญญา/เช็คเอาต์ — ปล่อยห้องด้วย `PATCH /admin/rooms/:roomID/status` แทน
- บังคับ `must_change_password` ที่ฝั่ง backend — ธงนี้แค่ส่งออกไปกับ `GET /me` ให้ frontend บังคับเอง ส่วน backend ยังให้ใช้ทุก endpoint ได้ตามปกติ
- `nearby_places.category` เป็น free text ไม่ใช่ enum

## Acceptance Criteria

AC-1: สมัครสมาชิกได้ และได้สิทธิ์ member เสมอ
  - `POST /auth/register` ด้วยอีเมลใหม่และข้อมูลครบ → 201 พร้อม `access_token`, `refresh_token`, `expires_at`, `token_type="Bearer"`, `user.role="member"`
  - ฟิลด์ที่รับมีแค่ `email, password, first_name, last_name, phone` — ส่งฟิลด์อื่นมาด้วย (เช่น `role`) → 422 validation_failed และ `fields.role = "ไม่อนุญาตให้ส่งฟิลด์นี้"` (decoder ปฏิเสธฟิลด์แปลกหน้าทั้งระบบ ดู AC-21)
  - อีเมลถูก trim และเก็บเป็นตัวพิมพ์เล็ก → สมัคร `A@B.com` ซ้ำกับ `a@b.com` → 422 และ `fields.email`
  - อีเมลซ้ำกับบัญชีใดก็ได้ รวมบัญชีที่ถูกระงับหรือ soft delete ไปแล้ว → 422 (เช็คอีเมลซ้ำไม่กรอง `deleted_at`)
  - สองคำขอสมัครด้วยอีเมลเดียวกันพร้อมกัน → ใบหนึ่งได้ 201 อีกใบได้ 409 `conflict` (ชน UNIQUE ที่ DB)
  - กรอกผิดหลายช่อง → 422 และ `fields` มีครบทุกช่องในครั้งเดียว
  - รหัสผ่านต้องยาว 8–72 ไบต์ และมีทั้งตัวอักษรและตัวเลข → ไม่ผ่าน → 422 · ยาว 73 → 422 (ไม่ปล่อยให้ bcrypt ตัดทิ้งเงียบ ๆ)
  - `phone` บังคับกรอก และต้องตรง `^[0-9\-\s()+]{8,20}$` → ไม่ผ่าน → 422
  - `first_name`/`last_name` ห้ามว่าง และยาวไม่เกิน 100 ตัวอักษร → ไม่ผ่าน → 422
  - body พาร์สไม่ได้ → 400 (ดู AC-21)

AC-2: เข้าสู่ระบบด้วย endpoint เดียวทุกบทบาท
  - อีเมล+รหัสถูก และบัญชี active → 200 พร้อม token คู่ และบันทึก `last_login_at` กับ log `auth.login`
  - อีเมลไม่มีในระบบ / รหัสผิด / บัญชี OAuth ที่ยังไม่ตั้งรหัส (`password_hash=''`) / บัญชี staff ที่ถูก soft delete → 401 `invalid_credentials` ข้อความเดียวกันทุกกรณี
  - กรณีไม่พบอีเมลหรือบัญชีไม่มีรหัส ระบบเผาเวลาด้วย dummy bcrypt hash ที่สุ่มตอน boot → จับเวลากรณี "อีเมลไม่มี" เทียบ "รหัสผิด" อย่างละ 50 ครั้ง ค่า median ต้องต่างกันไม่เกิน 20%
  - บัญชี `is_active=false` + **รหัสถูก** → 403 `account_disabled` · รหัสผิด → 401 `invalid_credentials` (เช็คสถานะบัญชีหลังตรวจรหัสผ่านแล้วเท่านั้น)
    **ยอมรับความเสี่ยง:** คนที่รู้รหัสผ่านอยู่แล้วจะรู้ว่าบัญชีนั้นถูกระงับ
  - body `{}` หรือไม่มีฟิลด์ → 401 `invalid_credentials` (login ไม่ตรวจรูปแบบ input)
  - rate limit นับที่ DB (ตาราง `login_attempts`) ในหน้าต่าง 15 นาทีย้อนหลัง:
    - ล้มเหลว 5 ครั้งจากคู่ (อีเมล, IP) เดียวกัน → ครั้งถัดไปจากคู่นั้น → 429 `too_many_requests` พร้อม header `Retry-After` (วินาที)
    - ล้มเหลว 20 ครั้งจาก IP เดียวกัน ไม่ว่าอีเมลไหน → ครั้งถัดไปจาก IP นั้น → 429
    - "ล้มเหลว" = ได้ 401 `invalid_credentials` (รวมอีเมลที่ไม่มีในระบบ) · 403 `account_disabled` ไม่นับ
    - ถูกจำกัดแล้ว → ได้ 429 แม้รหัสผ่านที่ส่งมาจะถูก และได้เหมือนกันไม่ว่าอีเมลจะมีบัญชีหรือไม่
    - อีเมลเดียวกันจาก IP อื่น → ไม่ถูกจำกัด (คนอื่นล็อกบัญชีเหยื่อจากเครื่องตัวเองไม่ได้)
    - เข้าสู่ระบบสำเร็จ → ความล้มเหลวของคู่ (อีเมล, IP) นั้นถูกล้าง · ตัวนับราย IP ไม่ถูกล้าง
    - ความล้มเหลวที่เก่ากว่า 15 นาที → ไม่นับ และถูกลบทิ้งตอนบันทึกความล้มเหลวครั้งใหม่
    - IP มาจากกติกาเดียวกับ activity log (ดู AC-13) → client ปลอม `X-Forwarded-For` เพื่อหลบไม่ได้
    - **[ยังไม่ยืนยัน]** คำขอที่ยิงพร้อมกันจำนวนมากอาจเกินเพดานไปได้เล็กน้อย เพราะนับกับบันทึกไม่ได้อยู่ใน lock เดียวกัน

AC-3: อายุและการหมุนเวียนของ token
  - access token เป็น JWT HS256 ที่มี `iss="wisetsuk-api"` อายุ 30 นาที (`ACCESS_TOKEN_TTL_MIN`) → ใช้หลังหมดอายุ, alg อื่น, issuer อื่น หรือ role ไม่รู้จัก → 401
  - refresh token อายุ 30 วัน (`REFRESH_TOKEN_TTL_DAY`) ตาราง `refresh_tokens` เก็บแค่ sha256 hash
  - `POST /auth/refresh` ด้วยใบที่ยังใช้ได้ → 200 พร้อม token คู่ใหม่ และใบเดิมถูกเพิกถอนในคำสั่ง UPDATE เดียว (rotation)
  - ใช้ใบเดิมซ้ำ / ใบหมดอายุ / ใบที่ logout แล้ว / สตริงมั่ว / ค่าว่าง → 401 `unauthorized`
  - refresh ของบัญชีที่ `is_active=false` → 401 (ใบนั้นถูกใช้ไปแล้วด้วย)
  - token ใบใหม่อ่าน role, ชื่อ และ `branch_id` จาก DB ณ ตอน refresh (ดู AC-4 เรื่องย้ายสาขา)
  - เปลี่ยนรหัสผ่านสำเร็จ (`POST /me/password`) → refresh token ทุกใบของบัญชีนั้นใช้ไม่ได้ทันที รวมใบของ session ปัจจุบันด้วย
    **ยอมรับความเสี่ยง:** access token ใบเดิมยังใช้ได้จนหมดอายุ (สูงสุด 30 นาที)
  - `POST /auth/logout {"refresh_token": "..."}` ด้วยค่าใดก็ได้ (ผิด/หมดอายุ/ไม่มีจริง/ว่าง) → 204 เหมือนกันหมด
  - `POST /auth/logout` แบบไม่มี body → 400 (ต้องส่ง JSON object มาเสมอ)
  - ทุกคำขอที่มี token → middleware ตรวจลายเซ็นกับอายุของ JWT แล้วอ่านบัญชีจาก DB ด้วย `sub` → บัญชีถูก soft delete หรือ `is_active=false` → 401 ทันทีแม้ token ยังไม่หมดอายุ
  - role, `branch_id` และชื่อผู้กระทำมาจาก DB ไม่ใช่ claim ใน JWT (claim `role`/`branch_id`/`name` ยังถูกใส่ใน token ให้ frontend อ่านได้ แต่ backend ไม่ใช้ตัดสินสิทธิ์)
  - ถูกระงับแล้วเปิดใช้งานคืนก่อน access token หมดอายุ → token ใบเดิมกลับมาใช้ได้ (ไม่ได้เพิกถอนรายใบ แค่ตรวจสถานะบัญชี)
  - DB ล่มระหว่างตรวจบัญชี → 500 ไม่ใช่ 401

AC-4: Admin ถูกล็อกไว้ที่สาขาเดียว (กฎเหล็ก)
  - `branch_id` ของ admin อ่านจาก DB ทุกคำขอ (ดู AC-3) ไม่ใช่จาก claim ใน token
  - endpoint แบบรายการใต้ `/admin` ที่ admin สาขา A ส่ง `?branch_id=<สาขา B>` มา → 403 `forbidden`
  - admin สาขา A ไม่ส่ง `branch_id` → ระบบบังคับใช้สาขา A (ไม่ใช่ "ทุกสาขา") → เช่น `GET /admin/bookings` ทุกแถวมี `branch_id` เป็นสาขา A
  - super admin ไม่ส่ง `branch_id` ใน endpoint แบบรายการ → เห็นทุกสาขา · ใน endpoint ที่แก้ข้อมูลสาขา (`/admin/branch*`, `POST /admin/rooms`) → 400 "กรุณาระบุสาขา"
  - admin สาขา A แตะทรัพยากรรายชิ้นของสาขา B → status ที่ได้ **ไม่เหมือนกันทุก endpoint**:
    - `POST /admin/bookings/:id/approve` และ `/reject` → 404 (แปลงผ่าน `hideCrossBranch`)
    - `DELETE /admin/branch/images/:imageID?branch_id=<A>` ด้วยรหัสรูปของสาขา B → 404 (SQL กรอง `branch_id`)
    - `DELETE /admin/rooms/:roomID/images/:imageID` → 403 (ตรวจจากห้อง)
    - `GET /bookings/:id`, `POST /bookings/:id/cancel`, `PUT /admin/bookings/:id/appointment`, `PUT`/`PATCH`/`DELETE /admin/rooms/:roomID`, `POST /admin/rooms/:roomID/image(s)` → **403**
    - **[ช่องว่าง]** เจตนาในคอมเมนต์ของโค้ดคือ "ทรัพยากรรายชิ้นต้องได้ 404 เพื่อไม่ยืนยันว่ารหัสนั้นมีจริง" แต่มีแค่ approve/reject ที่ทำตาม ส่วน endpoint อื่นในลิสต์ 403 ข้างบนยังเปิดเผยว่ารหัสนั้นมีอยู่จริง
  - super admin ย้าย admin จากสาขา A ไป B → admin คนนั้นใช้ access token ใบเดิมต่อ → **เห็นสาขา B ทันที** และขอ `?branch_id=<A>` → 403
  - `GET /admin/members` คืนสมาชิก **ทุกคนในระบบ** ไม่กรองตามสาขา (member ไม่ผูกกับสาขา) ส่วน `GET /admin/members/:id/bookings` ของ admin คืนเฉพาะใบจองในสาขาตัวเอง (ดู AC-28)

AC-5: Member เห็นเฉพาะของตัวเอง
  - member A เปิดใบจองของ member B (`GET /bookings/:id`) → 404 (ไม่ใช่ 403 เพื่อไม่เปิดเผยว่ารหัสนั้นมีจริง)
  - member A แจ้งชำระเงินหรือยกเลิกใบของ member B → 404
  - member เรียก endpoint ใต้ `/admin` หรือ `/superadmin` → 403 · admin/superadmin เรียก `POST /bookings` หรือ `GET /bookings` หรือ `POST /bookings/:id/payment` → 403 (endpoint เหล่านี้เปิดให้ member เท่านั้น)
  - `GET /me/notifications` → 200 `{items, unread_count}` โดย `items` มีไม่เกิน 30 รายการล่าสุด และ `unread_count` นับรายการที่ยังไม่อ่านทั้งหมด (ไม่จำกัดที่ 30)
  - `POST /me/notifications/read {}` → ทุกรายการที่ยังไม่อ่านของผู้ใช้นั้นเป็นอ่านแล้ว → 204
  - `POST /me/notifications/read {"id":"ntf-001"}` → อ่านรายการเดียว → 204 · ถ้าเป็นของคนอื่นหรือไม่มีจริง → 204 เหมือนกัน โดยไม่เปลี่ยนอะไร
  - `POST /me/notifications/read` แบบไม่มี body → 400 · `id` ผิดรูปแบบ → 400
  - **[ช่องว่าง]** `Notification.ID` ประกาศชนิดเป็น `UserID` → `GET /me/notifications` คืน `id` เป็น `"usr-001"` แต่ `/read` รับแค่ `"ntf-001"` → เอา id ที่ได้จากรายการไปส่งตรง ๆ ได้ 400 (`account/model.go`)

AC-6: จองรายวันคิดเงินตามจำนวนคืน
  - จำนวนคืน = `int((check_out - check_in).Hours() / 24)` เช่น 5–8 มี.ค. = 3 คืน
  - จอง 3 คืนในห้อง 1,200/คืน → 201 และ `total_amount = 3600`, `nights = 3`, `status = "pending_payment"`
  - `total_amount` ถูกเขียนลงใบจองครั้งเดียวตอนสร้าง → แก้ราคาห้องเป็น 2,000 ทีหลัง → ใบเดิมยังเป็น 3600
  - `check_out_date <= check_in_date` → 422 `fields.check_out_date` (DB ก็มี CHECK `daily_dates_required` กันไว้อีกชั้น)
  - `check_in_date` เป็นวันที่ผ่านมาแล้วตามปฏิทินไทย → 422 · เป็นวันนี้ → ผ่าน (ดู AC-22)
  - วันที่ต้องอยู่ในรูปแบบ `YYYY-MM-DD` → ผิดรูปแบบหรือไม่ส่งมา → 422
  - `stay_type` ไม่ใช่ `daily`/`monthly` → 422 · `stay_type` ไม่ตรงกับ `stay_type` ของห้อง → 422 `fields.stay_type`
  - ไม่ส่ง `room_id` → 422 · `room_id` ผิดรูปแบบ (ไม่ใช่ `rm-001`) → 400 · ห้องไม่มีจริงหรือ `is_active=false` → 404
  - ฟิลด์บังคับ: `guest_first_name`, `guest_last_name`, `guest_phone` (regex เดียวกับ AC-1) · ฟิลด์ไม่บังคับ: `emergency_phone` (ถ้าส่งมาต้องผ่าน regex), `emergency_relation` (ไม่ตรวจ)
  - ไม่มีเพดานจำนวนคืน และไม่มีฟิลด์จำนวนผู้เข้าพัก
  - สร้างสำเร็จ → member ได้แจ้งเตือน 1 รายการที่อ้าง `code` และมี log `booking.create` ผูกกับสาขาของห้อง

AC-7: จองรายเดือนเก็บแค่ค่าทำสัญญา
  - ห้อง 5,000/เดือน ในสาขาที่มี `branches.contract_fee = 500` → 201 และ `total_amount = 500` (ค่าเช่าไม่ถูกบันทึกที่ใบจอง)
  - `contract_fee` ถูกคัดลอกลงใบจองตอนสร้าง → แก้ค่าทำสัญญาของสาขาเป็น 800 ทีหลัง → ใบเดิมยังเป็น 500
  - `move_in_date` บังคับส่ง → ไม่ส่งหรือเป็นวันที่ผ่านมาแล้ว → 422
  - `contract_date` ไม่บังคับ และไม่ตรวจว่าต้องเป็นอนาคตหรือต้องสัมพันธ์กับ `move_in_date` → ถ้าส่งมาผิดรูปแบบ → 422
  - `stay_type=monthly` บนห้องที่เปิดแบบ daily → 422

AC-8: ห้องเดียวจองซ้อนไม่ได้ (กฎเหล็ก)
  - ใบจองที่ถือห้องไว้คือใบสถานะ `pending_payment` / `awaiting_review` / `approved` ส่วน `cancelled` ไม่ถือ → enum ใน DB มีแค่ 4 ค่านี้ (`rejected`/`completed` ถูกลบใน migration 0002)
  - ห้องที่มีใบรายเดือนถืออยู่ 1 ใบ → จองใหม่ทุกประเภทในห้องนั้นไม่ได้เลย ไม่ว่าวันที่ไหน → 409 `room_unavailable`
  - ใบรายวันสองใบทับกันเมื่อ `ใหม่.check_in < เก่า.check_out AND ใหม่.check_out > เก่า.check_in` → จองทับ → 409 `room_unavailable`
  - จองรายวันที่ `check_in` ตรงกับ `check_out` ของใบเดิมพอดี → 201 (ไม่นับว่าทับ)
  - ห้องสถานะ `maintenance` หรือ `occupied` → 409 `room_unavailable`
  - 2 คำขอจองห้องเดียวกันพร้อมกัน → สำเร็จ 1 ใบ เพราะมี `SELECT ... FOR UPDATE` ล็อกแถวห้องไว้ตลอด transaction
  - **[ช่องว่าง]** กฎ "ห้ามทับ" และ "ใบรายเดือนที่ยังไม่จบมีได้ทีละ 1 ใบ" บังคับที่ application (row lock + `IsAvailable`) เท่านั้น → ใน DB **ไม่มี** exclusion constraint หรือ partial unique index รองรับ ขัด AGENTS.md ข้อ 4 → `INSERT` ตรงเข้า DB สร้างใบซ้อนได้
  - **[ช่องว่าง]** ห้องของสาขาที่ `is_active=false` → ยังจองได้ตามปกติ (ตรวจแค่ `rooms.is_active`)
  - `code` มีรูปแบบ `PT-001` มาจาก sequence และ `LPAD` อย่างน้อย 3 หลัก → เกิน 999 แล้วยาวขึ้นเอง (`PT-1000`)
  - URL ทุกเส้นใช้ id ที่มี prefix เช่น `bkg-001` ไม่ใช่ `code` → `GET /bookings/PT-001` → 400

AC-9: แจ้งชำระเงินพร้อมสลิป
  - `POST /bookings/:id/payment` แบบ multipart (`slip`, `amount`, `transferred_at`, `note`) บนใบ `pending_payment` ของตัวเอง → 200 ใบเป็น `awaiting_review` และ response มี `latest_payment` ที่เพิ่งสร้าง
  - ลำดับการตรวจ (ผิดข้อแรกที่เจอ → ตอบทันที):
    1. body ใหญ่เกิน `MAX_UPLOAD_MB` (ค่าเริ่มต้น 5) + 1MB → 400
    2. `amount` ไม่ใช่ตัวเลขหรือว่าง → 422
    3. `transferred_at` พาร์สไม่ได้หรือว่าง → 422
    4. `note` ไม่ใช่ UTF-8 → 400
    5. ไม่แนบ `slip` → **400** (ไม่ใช่ 422) · ไฟล์ไม่ใช่ JPG/PNG/WEBP ตามเนื้อไฟล์ → 400 · ใหญ่เกินเพดาน → 400
    6. ใบจองไม่มีจริงหรือเป็นของคนอื่น → 404
    7. ใบไม่ได้อยู่ `pending_payment` → 409 `invalid_state`
    8. `amount <= 0` → 422 · `transferred_at` อยู่ในอนาคตเกิน 5 นาที → 422
  - **[ช่องว่าง]** ขั้นที่ 5 เขียนไฟล์สลิปลงดิสก์ก่อนตรวจขั้นที่ 6–8 → คำขอที่ตกขั้น 6–8 ทิ้งไฟล์กำพร้าไว้ใน `UPLOAD_DIR`
  - ไม่มีการเทียบ `amount` กับ `total_amount` → ยอดไม่ตรงก็รับตามปกติ และไม่มีการ flag ใด ๆ แอดมินต้องเทียบเองจาก `total_amount` กับ `latest_payment.amount`
  - `transferred_at` ไม่มีขอบเขตย้อนหลัง (วันโอนเมื่อ 2 ปีก่อนก็ผ่าน) · รูปแบบที่ไม่มี timezone ถูกตีความเป็นเวลาไทย (ดู AC-22)
  - การแจ้งชำระเงินหนึ่งครั้งคือ `payments` หนึ่งแถว (`booking_id` ไม่ unique) → แจ้งใหม่หลังถูกปฏิเสธจะเพิ่มแถวใหม่ ไม่เขียนทับแถวเดิม
  - ไฟล์เก็บที่ `UPLOAD_DIR/slips/YYYY/MM/<uuid>.<ext>` โดยไม่ใช้ชื่อไฟล์จาก client และ `slip_url = {PUBLIC_BASE_URL}/uploads/...`
  - สำเร็จ → member ได้แจ้งเตือน 1 รายการ และมี log `booking.submit_payment`
  - **[ยังไม่ยืนยัน]** 2 คำขอแจ้งชำระพร้อมกันบนใบเดียวกัน → น่าจะสำเร็จทั้งคู่และได้ `payments` 2 แถว เพราะการตรวจสถานะไม่ได้อยู่ใน transaction และไม่มี lock

AC-10: อนุมัติ/ปฏิเสธการจอง
  - อนุมัติใบ `awaiting_review` → 200 ใบเป็น `approved` · แถว payment ที่ยัง `submitted` เป็น `approved` และบันทึก `reviewed_by`/`reviewed_at`
  - อนุมัติใบรายเดือน → ห้องเป็น `occupied` ใน transaction เดียวกัน · อนุมัติใบรายวัน → สถานะห้องไม่เปลี่ยน
  - ปฏิเสธใบ `awaiting_review` พร้อม `reason` → 200 ใบ**กลับไปเป็น `pending_payment`** และห้องยังถูกถือไว้ (ดู AC-8) ส่วน payment เป็น `rejected` และเก็บ `reason` ที่ `payments.reject_reason` (ตาราง `bookings` ไม่มีคอลัมน์ `reject_reason` แล้ว)
  - ปฏิเสธโดยไม่ส่ง `reason`, ไม่มี body หรือ `reason` มีแต่ช่องว่าง → 422 `fields.reason`
  - อนุมัติหรือปฏิเสธใบที่ไม่ได้อยู่ `awaiting_review` → 409 `invalid_state`
  - admin สาขา A อนุมัติหรือปฏิเสธใบของสาขา B → 404
  - สำเร็จ → member ได้แจ้งเตือน 1 รายการที่อ้าง `code` (กรณีปฏิเสธมี `reason` ในข้อความด้วย) และมี log `booking.approve`/`booking.reject`
  - มีเทส "ยิง approve พร้อมกัน 2 คำขอ → สำเร็จ 1 อีกคำขอได้ 409" อยู่แล้ว **[ยังไม่ยืนยัน]** แต่ในโค้ด `UPDATE bookings` ไม่มีเงื่อนไข `status = 'awaiting_review'` และการตรวจสถานะเกิดก่อนเปิด transaction → ผลอาจขึ้นกับจังหวะ ต้องรันเทสซ้ำหลายรอบเพื่อยืนยัน
  - ใน `GET /admin/bookings` ใบที่กลับเป็น `pending_payment` หลังถูกปฏิเสธจะไม่มี `latest_payment` แนบมา (รายการข้ามใบ pending) → ต้องเปิด `GET /bookings/:id` ถึงจะเห็นเหตุผลที่ถูกปฏิเสธ

AC-11: ยกเลิกแล้วห้องต้องกลับมาขายได้ (กฎเหล็ก)
  - member ยกเลิกใบ `pending_payment` ของตัวเอง → 200 ใบเป็น `cancelled` และบันทึก `cancelled_at`
  - member ยกเลิกใบสถานะอื่น (`awaiting_review`/`approved`/`cancelled`) → 409 `invalid_state` ข้อความ "กรุณาติดต่อผู้ดูแลระบบของสาขา"
  - admin ในสาขาตัวเองหรือ super admin ยกเลิกใบสถานะใดก็ได้ → 200
  - **[ช่องว่าง]** admin ยกเลิกใบที่ `cancelled` อยู่แล้ว → 200 อีกครั้ง โดย `cancelled_at` ถูกเขียนทับ, member ได้แจ้งเตือนซ้ำ และเกิด log ซ้ำ (ไม่มีการตรวจสถานะสำหรับ admin)
  - **[ช่องว่าง]** `UPDATE` ของการยกเลิกไม่มี `AND branch_id = ...` (ตรวจสิทธิ์สาขาใน application เท่านั้น) ขัดกับที่ README อ้างว่า "UPDATE ของแอดมินมี branch_id เสมอ"
  - ยกเลิกใบรายเดือน → ห้องที่เป็น `occupied` กลับเป็น `available` ใน transaction เดียวกัน **เฉพาะเมื่อ** ไม่มีใบรายเดือนอื่นที่ยังถือห้องอยู่ · ห้องที่เป็น `maintenance` หรือ `available` อยู่แล้ว → ไม่ถูกแตะ
  - ยกเลิกใบรายวัน → สถานะห้องไม่เปลี่ยน (ใบรายวันไม่เคยเปลี่ยนสถานะห้องตั้งแต่แรก)
  - ห้องที่ถูกปล่อยแล้ว → จองใหม่ได้ทันที
  - สำเร็จ → เจ้าของใบได้แจ้งเตือน 1 รายการ และมี log `booking.cancel`

AC-12: ค้นหาห้องว่าง (`GET /rooms/search`)
  - ผลลัพธ์มีเฉพาะห้องที่ `is_active=true` และ `status='available'`
  - ส่ง `check_in` และ `check_out` มา → ไม่มีห้องที่มีใบรายวันทับช่วงนั้น (สูตรเดียวกับ AC-8) · ห้องที่ `check_out` ของใบเดิมตรงกับ `check_in` ที่ค้นหา → อยู่ในผลลัพธ์ · ส่งมาแค่ตัวเดียว → ไม่กรองใบรายวันเลย
  - ห้องที่มีใบรายเดือนถืออยู่ → ถูกตัดออกเฉพาะเมื่อ `move_in_date` ของใบนั้น ≤ `move_in_date` ที่ค้นหา (ถ้าไม่ส่งมาใช้วันนี้ตามนาฬิกา DB)
  - **[ช่องว่าง]** กติกาข้อข้างบนหลวมกว่าตอนจองจริง → ห้องที่มีใบรายเดือนเข้าอยู่ในอนาคตโผล่ในผลค้นหา แต่กดจองแล้วได้ 409 (ดู AC-8)
  - **[ช่องว่าง]** ห้องของสาขาที่ `is_active=false` → ยังอยู่ในผลลัพธ์ (search ไม่ join เงื่อนไขสาขา)
  - `stay_type=daily` และ `check_out <= check_in` → 422 `fields.check_out` · ไม่ส่ง `stay_type` → ไม่ตรวจลำดับวันที่
  - ตัวกรองที่รับ: `branch_id`, `room_type_id`, `stay_type`, `check_in`, `check_out`, `move_in_date`, `min_price`, `max_price`, `page`, `page_size`
  - ไม่ระบุ `page` → หน้า 1 ขนาด 20 พร้อม `meta.total_items`/`meta.total_pages` · `page_size=500` → ได้ไม่เกิน 100 และ `meta.page_size=100`
  - `page`/`page_size` เป็น 0, ติดลบ หรือไม่ใช่ตัวเลข → 400 · `stay_type` ไม่รู้จัก → 400 · วันที่ไม่ใช่ `YYYY-MM-DD` → 400 · ราคาไม่ใช่ตัวเลข → 400 · id ผิดรูปแบบ (`brn-001`, `rmt-001`) → 400
  - เรียงตาม `branch_id`, `room_number` · แต่ละห้องมี `amenities` ของห้องนั้นเอง แต่ไม่มีแกลเลอรี `images`

AC-13: บันทึกทุกการกระทำที่เปลี่ยนข้อมูล
  - action ที่มีจริงในโค้ด (29 รายการ):
    `auth.register`, `auth.login`, `auth.oauth_login`, `auth.oauth_link`, `auth.oauth_register`,
    `user.update_profile`, `user.change_password`,
    `admin.create`, `admin.update`, `admin.delete`,
    `booking.create`, `booking.submit_payment`, `booking.cancel`, `booking.approve`, `booking.reject`, `booking.set_appointment`,
    `branch.update`, `branch.set_amenities`, `branch.update_nearby`, `branch.update_cover`, `branch.add_image`, `branch.delete_image`,
    `room.create`, `room.update`, `room.update_image`, `room.add_image`, `room.delete_image`, `room.update_status`, `room.delete`
  - ทำ action ในลิสต์สำเร็จ → `activity_logs` มีแถวใหม่ที่เก็บ `actor_id`, `actor_role`, `actor_name`, `branch_id`, `action`, `entity_type`, `entity_id`, `detail` (jsonb), `ip_address` และ `created_at`
  - `ip_address` มาจาก `c.ClientIP()` ของ gin:
    - ไม่ได้ตั้ง `TRUSTED_PROXIES` → ใช้ IP ของ TCP connection และเมิน `X-Forwarded-For`/`X-Real-IP` ทั้งหมด → client ส่ง `X-Forwarded-For: 1.2.3.4` มาก็ปลอม IP ใน log ไม่ได้
    - คำขอมาจาก IP ที่อยู่ใน `TRUSTED_PROXIES` (IP หรือ CIDR คั่นด้วยจุลภาค) → ใช้ IP ขวาสุดใน `X-Forwarded-For` ที่ไม่ใช่ trusted proxy · ค่าปลอมที่ client แอบใส่ไว้ด้านซ้ายไม่มีผล
    - `TRUSTED_PROXIES` มีค่าที่ไม่ใช่ IP หรือ CIDR → server ไม่ start
    - รันผ่าน Docker port publishing โดยไม่มี proxy → IP ที่เห็นอาจเป็น gateway ของ docker ไม่ใช่ IP จริงของผู้ใช้
  - `actor_name` มาจากชื่อใน DB ณ ตอนทำรายการ (ดู AC-3) → แก้ชื่อแล้ว log ถัดไปใช้ชื่อใหม่ทันที
  - `branch_id` ของ log:
    - action ของ booking/room/branch → สาขาของสิ่งที่ถูกกระทำ
    - action ของ auth/user/admin → สาขาของผู้กระทำ (member และ superadmin เป็น `NULL`)
  - `refresh`, `logout`, `oauth/exchange` และการอ่านแจ้งเตือน → ไม่เขียน log
  - เขียน log ไม่สำเร็จ → error ถูกกลืน และธุรกรรมหลักไม่ล้มตาม
  - `GET /admin/activity-logs` → admin เห็นเฉพาะ log ที่ `branch_id` เป็นสาขาตัวเอง · super admin ไม่ส่ง `branch_id` → เห็นทุกแถวรวมแถวที่ `branch_id` เป็น `NULL`
  - ตัวกรอง: `action` (ตรงตัว), `search` (ILIKE ใน `actor_name`, ชื่อสาขา, `entity_id`), `actor_id`, `actor_role` (ค่าที่ไม่รู้จัก → 400), `branch_id`, `page`/`page_size` → เรียงใหม่สุดก่อน

AC-14: รูปภาพสาขา/ห้อง — อัปโหลดไฟล์ หรือใส่ URL ภายนอก
  - อัปโหลดไฟล์ (multipart ช่อง `image`):
    - `POST /admin/branch/cover?branch_id=` → เปลี่ยน `cover_image_url`
    - `POST /admin/branch/images/upload?branch_id=` (+`caption`, `sort_order`) → 201
    - `POST /admin/rooms/:roomID/image` → เปลี่ยนรูปปกห้อง
    - `POST /admin/rooms/:roomID/images/upload` (+`sort_order`) → 201 รูปแกลเลอรี
  - ไฟล์ที่อัปโหลดเก็บเป็น blob ในตาราง `assets` และระบบคืน URL `{PUBLIC_BASE_URL}/files/ast-NNN` (รูปเดโมจาก seed ต่างออกไป เพราะชี้ `/uploads/...` ดูหัวข้อข้อมูลตั้งต้น)
  - อัปโหลดไฟล์เนื้อเดียวกันซ้ำ → ได้ asset id เดิม (dedup ด้วย `UNIQUE(checksum)` sha256)
  - ชนิดไฟล์ตรวจจากเนื้อไฟล์จริง ต้องเป็น JPG/PNG/WEBP → ไม่ใช่ → 400 · ใหญ่เกิน `MAX_UPLOAD_MB` → 400 · ไม่แนบไฟล์ → 400
  - `sort_order` ไม่ใช่จำนวนเต็ม → 400 · `caption` ไม่ใช่ UTF-8 → 422 · `caption` ยาวเกิน 200 → 422
  - **[ช่องว่าง]** ไฟล์ถูกเขียนลง `assets` ก่อนตรวจสิทธิ์สาขา → คำขอที่ได้ 403/404 ทิ้ง blob กำพร้าไว้ใน DB
  - `GET /files/:assetID` และ `HEAD` → เปิดสาธารณะ ไม่ต้องใช้ token · ตอบ `ETag` (sha256), `Cache-Control: public, max-age=31536000, immutable`, `X-Content-Type-Options: nosniff` · ส่ง `If-None-Match` ตรงกับ ETag → 304 · ไม่มี asset นั้น → 404 · id ผิดรูปแบบ → 400
  - ใส่ URL ภายนอกแบบ JSON:
    - `POST /admin/branch/images?branch_id= {image_url, caption, sort_order}` → 201
    - `POST /admin/rooms/:roomID/images {image_url, sort_order}` → 201
    - ต้องเป็น `http`/`https` ที่มี host และยาวไม่เกิน 2048 → `javascript:`, `data:`, path สัมพัทธ์ หรือค่าว่าง → 422
  - ช่อง `image_url` ของ `POST`/`PUT /admin/rooms` และช่อง `cover_image_url` ของ `PUT /admin/branch` → ใช้กติกาเดียวกัน แต่ส่งค่าว่างได้
  - แกลเลอรีห้องเป็นคนละช่องกับรูปปก → เพิ่ม/ลบรูปแกลเลอรีไม่แตะ `rooms.image_url` · `GET /rooms/:id` คืน `images` เรียงตาม `sort_order, id`
  - `DELETE /admin/branch/images/:imageID?branch_id=` → 204 | 404 (รูปไม่มีหรือเป็นของสาขาอื่น)
  - `DELETE /admin/rooms/:roomID/images/:imageID` → 204 | 403 (ห้องคนละสาขา) | 404
  - **[ช่องว่าง]** ถ้า `imageID` เป็นรูปของห้องอื่นในสาขาเดียวกัน → ถูกลบด้วย (SQL ผูกแค่สาขา ไม่ผูก `room_id` ใน URL)

AC-15: ความปลอดภัยและความทนทานพื้นฐาน
  - ทุก query ใช้ placeholder (`$n`) ผ่าน `database.Binder` และไม่มีการต่อค่าจากผู้ใช้เข้า SQL — **ตรวจด้วย code review ไม่ใช่เทสผ่าน API**
  - error 5xx → body เป็น `{"error":{"code":"internal_error","message":"เกิดข้อผิดพลาดภายในระบบ"}}` เสมอ ไม่มี stack trace หรือข้อความ SQL (รายละเอียดจริง log ไว้ฝั่ง server)
  - panic ใน handler → server ไม่ล้ม และ client ได้ 500 เป็น JSON
  - ข้อความ error ที่ผู้ใช้เห็นเป็นภาษาไทยทั้งหมด รวม 404 เส้นทางที่ไม่มีจริง และ 405 method ผิด
  - ไม่มี `DATABASE_URL` หรือ `JWT_SECRET` สั้นกว่า 32 ตัว → server ไม่ start (fail fast)
  - ทุก response มี header `X-Request-ID` → ถ้า client ส่งมาจะใช้ค่านั้น ถ้าไม่ส่งจะสุ่ม UUID ให้
  - CORS อนุญาตเฉพาะ origin ใน `ALLOWED_ORIGINS` (ค่าเริ่มต้น `http://localhost:3000,http://localhost:5173`) พร้อม credentials
  - JSON body ใหญ่เกิน 1MB → 400
  - `GET /uploads/` หรือ path ที่ลงท้ายด้วย `/` → 404 (ไม่เปิดให้ไล่ดูรายชื่อไฟล์)
  - รหัสผ่านเก็บเป็น bcrypt (cost 12 จาก `BCRYPT_COST` · DB มี CHECK ว่าต้องขึ้นต้นด้วย `$2` หรือเป็นค่าว่าง)
  - migration รันอัตโนมัติตอน API start (ไฟล์ละ 1 transaction และข้ามไฟล์ที่รันแล้ว) · ได้รับ SIGTERM → รอคำขอที่ค้างอยู่ได้สูงสุด 20 วินาที

AC-16: การจองที่ไม่จ่ายเงินหมดอายุเอง **[ยังไม่ implement]**
  - สถานะโค้ดตอนนี้: ตาราง `bookings` ไม่มีคอลัมน์ `expires_at` และไม่มี worker หรือ goroutine คอยตัดใบที่หมดอายุ → ใบ `pending_payment` ค้างถือห้องไว้ตลอดไปจนกว่าจะมีคนยกเลิก
  - พฤติกรรมที่ตกลงจะทำต่อ:
    - สร้างใบจอง → `expires_at` = ตอนนี้ + 10 นาที
    - ถูกปฏิเสธแล้วกลับเป็น `pending_payment` → `expires_at` = เวลาที่ปฏิเสธ + 24 ชั่วโมง
    - ใบที่เลย `expires_at` และยังเป็น `pending_payment` → ระบบเปลี่ยนเป็น `cancelled` และห้องถูกปล่อยตาม AC-11
    - log action `booking.auto_expire` โดย `actor_id = NULL` (คอลัมน์ nullable อยู่แล้ว) — แต่ `actor_role` เป็น enum `user_role` ซึ่งไม่มีค่า `system` → ต้องเพิ่มค่าใน enum หรือเปลี่ยนแบบ
  - ห้ามเขียนเทสยืนยันเรื่องนี้จนกว่าจะ implement จริง

AC-17: ไฟล์สลิปโอนเงิน — ยังไม่มีการตรวจสิทธิ์ก่อนเสิร์ฟ
  - สถานะโค้ดตอนนี้: สลิปเสิร์ฟผ่าน `GET /uploads/*filepath` ซึ่งอยู่นอก `/api/v1` และไม่มี auth → ใครที่รู้ URL (ชื่อไฟล์เป็น UUID) ก็เปิดดูได้
  - path ของไฟล์มาจากโค้ดเท่านั้น (folder คงที่ + UUID) → client ไม่มีทางส่ง path มาต่อเองได้ จึงกัน path traversal ได้ แต่ไม่ได้กันคนที่ไม่มีสิทธิ์
  - **[ยังไม่ implement]** `GET /api/v1/bookings/:bookingID/slip`:
    - เจ้าของใบ / admin สาขาเดียวกัน / super admin → 200
    - คนอื่น หรือใบที่ไม่มีจริง → 404 เหมือนกัน
    - ไม่มี token → 401
    - `docs/openapi.yaml` มี path นี้แล้ว แต่ router ยังไม่มี

AC-18: Super Admin จัดการบัญชีผู้ดูแล
  - `GET /superadmin/staff` → admin และ superadmin ทุกคนที่ยังไม่ถูกลบ รวมคนที่ถูกระงับ (`is_active=false`) และมี `branch_name` มาด้วย
  - `POST /superadmin/staff {email, first_name, last_name, phone, branch_id}` → 201 ได้ role `admin` และ `must_change_password=true`
  - ข้อมูลที่ POST ส่งไม่ได้ / ถูกตั้งให้อัตโนมัติ:
    - ไม่รับ `password` และไม่รับฟิลด์อื่นนอกจาก 5 ตัวข้างบน → ส่งมา → 422
    - รหัสผ่านตั้งจาก `STAFF_DEFAULT_PASSWORD` → ถ้าไม่ตั้งใช้ `SEED_DEFAULT_PASSWORD` → ถ้าไม่ตั้งทั้งคู่ใช้ `"Wisetsuk!2026"`
    - `phone` ไม่บังคับ แต่ถ้าส่งมาต้องผ่าน regex
  - POST ผิดเงื่อนไข:
    - ไม่ส่ง `branch_id` → 422
    - `branch_id` ผิดรูปแบบ → 400
    - ไม่มีสาขานั้น → 422 `fields.branch_id`
    - สาขามี admin อยู่แล้ว → 422 `fields.branch_id` (ดัก `uq_admin_per_branch`)
  - อีเมลซ้ำ รวมอีเมลของบัญชีที่ถูกลบไปแล้ว → 422 `fields.email`
  - `PUT /superadmin/staff/:userID {first_name, last_name, phone, branch_id, is_active, password?}` เป็นการแทนที่ทั้งชุด → ไม่ส่ง `first_name`/`last_name` → 422 · ไม่ส่ง `is_active` → คงค่าเดิม
    - เป้าหมายเป็น admin แต่ไม่ส่ง `branch_id` → 422 · เป้าหมายเป็น superadmin แต่ส่ง `branch_id` มา → 422
    - ย้ายไปสาขาที่มี admin แล้ว → 422 `fields.branch_id`
    - **[ช่องว่าง]** ย้ายไปสาขาที่ไม่มีอยู่จริง → **500** (FK violation ไม่ถูกแปลง ต่างจาก POST ที่เช็คก่อน)
    - ระงับตัวเอง (`is_active=false` บน id ของตัวเอง) → 400
    - เป้าหมายเป็น member, ไม่มีจริง หรือถูกลบแล้ว → 404
    - ส่ง `password` มา → ตั้งรหัสใหม่ทันที, `must_change_password=true` และเพิกถอน refresh token ทุกใบของบัญชีนั้น — เป็นช่องทางเดียวที่ super admin รีเซ็ตรหัสให้คนอื่นได้
    - ระงับ (`is_active=false`) → access token ใบเดิมได้ 401 ทันที และ refresh ได้ 401 (ดู AC-3) — refresh token ไม่ถูกเพิกถอน จึงใช้ได้อีกถ้าเปิดบัญชีคืน
    - ย้ายสาขา → ไม่เพิกถอน token แต่ token ใบเดิมเห็นสาขาใหม่ทันที (ดู AC-4)
  - `DELETE /superadmin/staff/:userID` → 204 เป็น soft delete (`deleted_at = now()`, `is_active = false`) และเพิกถอน refresh token ทุกใบ
    - ลบตัวเอง → 400 · เป้าหมายไม่ใช่ admin (member/superadmin), ไม่มีจริง หรือถูกลบไปแล้ว → 404
    - แถวยังอยู่ใน DB → `activity_logs.actor_id`, `bookings.reviewed_by`, `payments.reviewed_by` ไม่หลุด
    - บัญชีที่ถูกลบหายจากการ login (401), `GET /me` (404) และรายชื่อทุกหน้า
    - สาขาของคนที่ถูกลบว่างลงทันที → สร้าง admin คนใหม่ลงสาขาเดิมได้ (`uq_admin_per_branch` ไม่นับแถวที่ `deleted_at IS NOT NULL`)
  - DB ปฏิเสธการมี admin คนที่สองในสาขาเดียวกัน และ superadmin คนที่สอง แม้ INSERT ตรงโดยข้าม service

AC-19: จัดการห้อง
  - `POST /admin/rooms` → 201 · field: `branch_id, room_type_id, room_number, building, floor, stay_type, price, water_rate, electric_rate, size_sqm, description, image_url, status, amenity_ids`
    - admin ไม่ส่ง `branch_id` → ใช้สาขาตัวเอง · ส่งสาขาอื่น → 403 · super admin ไม่ส่ง → 400
  - `PUT /admin/rooms/:roomID` → 200 เป็นการแทนที่ทั้งชุดด้วย field ชุดเดียวกัน (สาขายึดตามห้องเดิม และ `branch_id` ใน body ถูกเมิน)
  - กติกาของค่า:
    - `room_number` บังคับ และยาวไม่เกิน 20
    - `price > 0`, `floor > 0`, `water_rate >= 0`, `electric_rate >= 0`
    - `stay_type` ∈ `daily`/`monthly`
    - `status` ว่าง = `available` · ค่าอื่นนอกจาก 3 ค่า → 422
    - `building` ว่าง = `"1"`
    - ผิดข้อใด → 422
  - `room_number` ซ้ำในสาขาเดียวกัน → 422 `fields.room_number` — ชนกับห้องที่ถูกลบไปแล้วด้วย เพราะ `UNIQUE(branch_id, room_number)` ไม่มีเงื่อนไข จึงเอาเลขห้องที่ลบไปแล้วกลับมาใช้ไม่ได้
  - `amenity_ids`: ไม่ส่ง = ไม่แตะของเดิม · `[]` = ล้างทั้งหมด · id ที่ไม่มีจริง → 422 `fields.amenity_ids` · id ซ้ำใน array → รับได้ (DISTINCT)
  - **[ช่องว่าง]** `room_type_id` ที่ไม่มีจริง → **500** (FK violation ไม่ถูกแปลง) · `room_type_id` ของอีกสาขา → รับได้ (ไม่ตรวจว่าอยู่สาขาเดียวกัน)
  - `PATCH /admin/rooms/:roomID/status {status}` → 200 · ไม่จำกัดว่าเปลี่ยนจากสถานะไหนไปไหน และไม่ตรวจว่ามีใบจองค้างอยู่ · ค่าไม่อยู่ใน 3 ค่า → 422
  - `DELETE /admin/rooms/:roomID` → 204 เป็น soft delete (`is_active=false`) โดยไม่ตรวจใบจองที่ยังค้าง · ลบห้องที่ลบไปแล้วซ้ำ → 204 อีกครั้ง
  - **[ช่องว่าง]** `PUT`/`PATCH`/`DELETE`/อัปโหลดรูป บนห้องที่ถูกลบแล้ว → ยังทำงานได้ (การหาห้องไม่กรอง `is_active`)
  - **[ช่องว่าง]** `PUT` เปลี่ยน `stay_type` หรือ `status` ของห้องที่มีใบจองค้างอยู่ → ได้เลย ไม่มีการตรวจ
  - `GET /admin/rooms` → ห้องทุกสถานะที่ `is_active=true` ของสาขาที่ดูแล · ตัวกรอง `branch_id, room_type_id, stay_type, min_price, max_price, page, page_size` (ส่ง `check_in`/`check_out`/`move_in_date` มาได้แต่ไม่มีผล) · ไม่มี `amenities` แนบมา
  - ห้องคนละสาขา → 403 · ห้องไม่มีจริง → 404
  - สำเร็จ → มี log `room.create`/`room.update`/`room.update_status`/`room.delete` ผูกสาขาของห้อง

AC-20: นัดหมายทำสัญญา
  - `PUT /admin/bookings/:bookingID/appointment {appointment_at, note}` บนใบรายเดือน → 200 และเขียนทับนัดเดิมได้ทุกครั้ง
  - บนใบรายวัน → 400 (ตรวจก่อนตรวจ `appointment_at`)
  - `appointment_at` รับ RFC3339 หรือ `YYYY-MM-DD[ T]HH:MM[:SS]` (ไม่มี timezone = เวลาไทย) → พาร์สไม่ได้ → 422 · ไม่อยู่ในอนาคต → 422
  - admin คนละสาขา → 403 (ไม่ใช่ 404 ดู AC-4) · super admin → ทำได้ทุกสาขา
  - **[ช่องว่าง]** ไม่ตรวจสถานะใบจอง → ตั้งนัดบนใบ `pending_payment`, `awaiting_review` หรือ `cancelled` ได้ทั้งหมด (ยังไม่มีเทสคุมเงื่อนไข "ต้อง approved แล้ว")
  - สำเร็จ → member ได้แจ้งเตือน 1 รายการที่อ้าง `code` และวันเวลาในรูป `DD-MM-YYYY HH:MM`
  - **[ช่องว่าง]** วันเวลาในแจ้งเตือนแสดงตาม timezone ที่ client ส่งมา → ส่ง `2026-09-30T03:00:00Z` → ข้อความเป็น `30-09-2026 03:00` ไม่ใช่เวลาไทย 10:00

AC-21: รหัสสถานะและรูปแบบ error สอดคล้องกันทั้งระบบ
  - error: `{"error": {"code": "...", "message": "...", "fields": {...}}}` → `fields` ปรากฏเฉพาะเมื่อ code = `validation_failed`
  - สำเร็จ: `{"data": ...}` · รายการแบ่งหน้ามี `meta: {page, page_size, total_items, total_pages}` เพิ่ม · 204 ไม่มี body · ข้อยกเว้นเดียวคือ `GET /health` ที่ตอบ `{"status":"healthy","time":...}` โดยไม่ห่อ `data`
  - error code มีแค่ 13 ตัว: `unauthorized`, `forbidden`, `not_found`, `conflict`, `bad_request`, `validation_failed`, `method_not_allowed`, `internal_error`, `invalid_credentials`, `account_disabled`, `room_unavailable`, `invalid_state`, `too_many_requests`
  - 429 `too_many_requests` → มี header `Retry-After` (วินาที) เสมอ (ปัจจุบันใช้ที่ `/auth/login` ที่เดียว ดู AC-2)
  - 400 `bad_request`:
    - body พาร์สไม่ได้, มี JSON มากกว่า 1 object หรือชนิดข้อมูลผิด
    - id ใน URL/query/body ผิดรูปแบบ (เช่น `branch_id` ไม่ใช่ `brn-001`)
    - query param ผิดรูปแบบ
    - ไฟล์อัปโหลดผิดหรือใหญ่เกิน
    - ข้อความใน multipart ไม่ใช่ UTF-8
  - 422 `validation_failed`: ค่าถูกชนิดแต่ผิดกติกา (เช่น `check_out <= check_in`, รหัสผ่านสั้น) **หรือ** body มีฟิลด์ที่ไม่รู้จัก → `fields.<ชื่อฟิลด์>`
  - ไม่มี token หรือ token ใช้ไม่ได้ → 401 `unauthorized` · มี token แต่บทบาทไม่พอ → 403 `forbidden`
  - ชน unique constraint ที่ service ไม่ได้ดักไว้เอง → 409 `conflict` "ข้อมูลนี้มีอยู่ในระบบแล้ว"
  - **[ช่องว่าง]** ชน foreign key → ไม่มีการแปลง กลายเป็น 500 (ต้นเหตุของ 500 ใน AC-18, AC-19, AC-26)
  - error ทุกตัวรวม 404/405/500 ใช้รูปแบบเดียวกัน

AC-22: เวลาและวันที่
  - ทุก timestamp เก็บเป็น `TIMESTAMPTZ` (UTC) ส่วนวันที่ล้วน (`check_in_date`, `check_out_date`, `move_in_date`, `contract_date`) เก็บเป็น `DATE`
  - "วันนี้ / อดีต / อนาคต" ของ `check_in_date` และ `move_in_date` ตัดสินด้วยปฏิทิน Asia/Bangkok เสมอ ไม่ขึ้นกับ TZ ของเครื่อง (`internal/timex` ฝัง tzdata ไว้ในไบนารี) → server ที่เป็น UTC รับจอง "วันนี้" ตอน 00:30 เวลาไทยได้
  - `transferred_at` และ `appointment_at` ที่ไม่มี offset → ตีความเป็นเวลาไทย · มี offset มา → ใช้ตามนั้น
  - การเทียบ "อนาคต" ของ `transferred_at` (+5 นาที) และ `appointment_at` เทียบเป็นจุดเวลาจริง ไม่ใช่วันที่
  - timestamp ใน response เป็น RFC 3339 ที่มี offset · ฟิลด์วันที่ล้วนออกมาเป็นเที่ยงคืน UTC เช่น `"2026-08-14T00:00:00Z"` ไม่ใช่ `"2026-08-14"` **[ยังไม่ยืนยัน]**

AC-23: ลืมรหัสผ่าน — ขอลิงก์รีเซ็ตทางอีเมล **[ยังไม่ implement]**
  - สถานะโค้ดตอนนี้: ไม่มี route `/auth/forgot-password` หรือ `/auth/reset-password`, ไม่มี mailer และไม่มีตารางเก็บ reset token → member ที่ลืมรหัสกู้คืนเองไม่ได้ และ super admin รีเซ็ตให้ได้เฉพาะบัญชี staff (ดู AC-18)
  - `docs/openapi.yaml` และ `backend/README.md` เขียนไว้แล้วว่ามี 2 endpoint นี้ **ซึ่งไม่ตรงกับโค้ด**
  - พฤติกรรมที่ตกลงจะทำต่อ:
    - `POST /auth/forgot-password {email}` → 204 เหมือนกันทุกกรณี และเผาเวลาเท่ากันด้วย dummy hash
    - บัญชี `is_active=false` → 204 เหมือนกัน แต่ไม่ส่งอีเมลจริง
    - token เก็บแค่ sha256 hash, อายุ 15 นาที, ใช้ได้ครั้งเดียว และขอใบใหม่แล้วใบเก่าใช้ไม่ได้ทันที
    - ขอได้ไม่เกิน 5 ครั้ง/ชม./อีเมล นับที่ DB
    - ลิงก์ชี้ไปหน้า frontend ด้วย env ตัวใหม่ แยกจาก `PUBLIC_BASE_URL` ที่เป็นโดเมน API
    - `POST /auth/reset-password {token, new_password, confirm_password}` สำเร็จ → 204, เพิกถอน refresh token ทุกใบ และบันทึก log `user.reset_password`
    - `APP_ENV=production` แต่ไม่ได้ตั้งค่า SMTP → server ไม่ start
  - ห้ามเขียนเทสยืนยันเรื่องนี้จนกว่าจะ implement จริง

AC-24: เข้าสู่ระบบด้วย Google/Facebook OAuth
  - provider ที่รองรับ: `google`, `facebook` (DB CHECK) → provider ที่ตั้ง client id/secret ไม่ครบจะถูกปิดเงียบ ๆ โดยไม่ทำให้ server ล้ม
  - `GET /auth/oauth` → `{"data":{"providers":[...]}}` เฉพาะ provider ที่เปิดอยู่ เรียง google ก่อน facebook
  - `GET /auth/oauth/:provider` (ชื่อไม่สนตัวพิมพ์) → provider ปิดอยู่ → 404 JSON · เปิดอยู่ → ตั้ง cookie `wisetsuk_oauth_state` และ `wisetsuk_oauth_verifier` แล้ว 302 ไปหน้ายินยอม (PKCE S256 และ Google บังคับ `prompt=select_account`)
    - cookie ตั้งเป็น HttpOnly, SameSite=Lax, path `/api/v1/auth/oauth`, อายุ 10 นาที และตั้ง Secure เฉพาะ `APP_ENV=production`
  - `GET /auth/oauth/:provider/callback` → provider ปิดอยู่ → 404 JSON · กรณีอื่นลบ cookie ทิ้งแล้ว **302 ไป `FRONTEND_OAUTH_CALLBACK_URL` เสมอ**:
    - สำเร็จ → `?code=<exchange code>`
    - ผู้ใช้กดยกเลิก / `state` ไม่ตรงกับ cookie / ไม่มี cookie / ไม่มี `code` / provider ไม่ส่งอีเมล / อีเมลยังไม่ยืนยัน / แลก token ไม่สำเร็จ / บัญชีถูกระงับ → `?error=<ข้อความไทย>`
  - การจับคู่บัญชี (ทั้งหมดอยู่ใน transaction เดียว):
    1. มี `user_identities` ของ provider+subject นี้แล้ว → เข้าบัญชีเดิม (`auth.oauth_login`) — ใช้ subject จับคู่ ไม่ใช้อีเมล
    2. ไม่มี แต่มีบัญชีอีเมลเดียวกันที่ยังไม่ถูกลบ → ผูก identity กับบัญชีนั้น (`auth.oauth_link`) ทุก role รวมถึง admin/superadmin
    3. ไม่มีทั้งคู่ → สร้าง member ใหม่ที่ `password_hash=''`, `phone=''` และ `first_name` มาจาก provider (ถ้าไม่มีใช้ส่วนหน้า `@` ของอีเมล) (`auth.oauth_register`)
  - **[ช่องว่าง]** อีเมลตรงกับบัญชี staff ที่ถูก soft delete → ข้อ 2 หาไม่เจอ ข้อ 3 ชน UNIQUE ของอีเมล → redirect พร้อม `?error=ข้อมูลนี้มีอยู่ในระบบแล้ว`
  - `POST /auth/oauth/exchange {code}` → 200 token คู่ · code ว่าง/ไม่มีจริง/หมดอายุ (2 นาที)/ใช้แล้ว → 401 `unauthorized` · บัญชีถูกระงับ → 403 `account_disabled` · ตาราง `oauth_exchange_codes` เก็บแค่ hash
  - บัญชีที่ยังไม่มีรหัสผ่านเรียก `POST /me/password` → ไม่ต้องส่ง `current_password` (เป็นการตั้งรหัสครั้งแรก) · ตั้งแล้วครั้งถัดไปต้องส่ง `current_password` ตามปกติ
  - บัญชีที่สร้างผ่าน OAuth มี `phone=''` → `PUT /me` บังคับ `phone` ดังนั้นต้องกรอกเบอร์ก่อนถึงจะแก้โปรไฟล์ได้

AC-25: แดชบอร์ดสำหรับ Admin
  - `GET /admin/dashboard` → `{branches: [...], recent_activities: [...]}`
  - `branches` มีสถิติรายสาขา: `branch_id, branch_name, daily_rooms_free, daily_rooms_total, monthly_rooms_free, monthly_rooms_total, pending_review, bookings_total`
    - นับเฉพาะห้องที่ `is_active=true` · "free" = `status='available'` (ไม่ได้ดูใบจองรายวันที่ค้างอยู่)
    - `pending_review` = จำนวนใบ `awaiting_review` · `bookings_total` = ใบจองทุกสถานะ
    - ไม่มีตัวเลขรวมทุกสาขาแยกต่างหาก และรวมสาขาที่ `is_active=false` ด้วย
  - `recent_activities` = activity log ล่าสุดไม่เกิน 10 รายการ ในขอบเขตเดียวกับ AC-13
  - admin → เห็นเฉพาะสาขาตัวเอง · super admin → เห็นทุกสาขา และส่ง `?branch_id=` เพื่อเจาะสาขาเดียวไม่ได้ (endpoint นี้ไม่อ่าน query นั้น)
  - module `reporting` ยังไม่มีไฟล์ `_test.go` → ทุกข้อใน AC นี้ทดสอบได้แต่ยังไม่มีเทสคุม

AC-26: จัดการรายละเอียดสาขา (ทุก endpoint ส่งสาขาผ่าน `?branch_id=` ไม่ใช่ใน body)
  - `PUT /admin/branch?branch_id=` → 200 คืนสาขาพร้อม images/amenities/nearby · แทนที่ทั้งชุด: `name, tagline, description, address, phones, line_id, email, latitude, longitude, map_url, building_count, floor_count, daily_price_from, monthly_price_min, monthly_price_max, water_rate, electric_rate, deposit, advance_payment, contract_fee, cover_image_url`
    - `name` บังคับและยาวไม่เกิน 200 · `building_count > 0` · `floor_count > 0` · `water_rate`/`electric_rate >= 0` · `monthly_price_min <= monthly_price_max` · `cover_image_url` ต้องเป็น http(s) → ผิดข้อใด → 422
    - ไม่ส่ง `latitude`/`longitude`/ราคา → ค่าเดิมกลายเป็น `NULL` · ไม่ส่ง `cover_image_url` → คงรูปเดิม
    - **[ช่องว่าง]** `contract_fee`, `deposit`, `advance_payment` ติดลบได้ → `contract_fee = -500` ทำให้ใบจองรายเดือนใหม่มี `total_amount = -500` · `email`/`phones` ไม่ถูกตรวจรูปแบบ
  - `PUT /admin/branch/amenities?branch_id= {amenity_ids}` → 200 คืนรายการใหม่ แทนที่ทั้งชุด
    - **[ช่องว่าง]** ไม่ได้อยู่ใน transaction และไม่แปลง error → id ที่ไม่มีจริง → 500 · id ซ้ำ → 409 · ทั้งสองกรณี**ลบของเดิมไปแล้ว**
  - `PUT /admin/branch/nearby?branch_id= {items:[{category, name, distance, sort_order}]}` → 200 แทนที่ทั้งชุด · `name` ว่างในรายการใดก็ได้ → 422 `fields.items` · `sort_order=0` → ใช้ลำดับใน array แทน
    - **[ช่องว่าง]** ไม่ได้อยู่ใน transaction → INSERT พังกลางทาง → รายการเดิมหายไปบางส่วน
  - admin ส่ง `branch_id` สาขาอื่น → 403 · super admin ไม่ส่ง → 400
  - สำเร็จ → log `branch.update`/`branch.set_amenities`/`branch.update_nearby` ผูกกับสาขานั้น

AC-27: ข้อมูลสาธารณะของสาขาและห้อง (ไม่ต้องล็อกอิน)
  - `GET /branches` → เฉพาะสาขาที่ `is_active=true` เรียงตาม `created_at` แต่ละสาขามี `images`, `amenities`, `nearby_places`
  - `GET /superadmin/branches` → ทุกสาขารวมที่ปิดใช้งาน (super admin เท่านั้น)
  - `GET /branches/:branchID` รับได้ทั้ง `brn-001` และ slug (เช่น `prachauthit-45`) → 200 · ค่าที่ไม่ใช่ id ถูกตีความเป็น slug → ไม่เจอ → **404** (ไม่ใช่ 400)
  - **[ช่องว่าง]** `GET /branches/:branchID` ของสาขาที่ `is_active=false` → ยังคืน 200
  - `GET /amenities` → รายการกลางทั้งหมด เรียง `sort_order, name`
  - `GET /room-types?branch_id=` → ประเภทห้อง (ไม่ส่ง = ทุกสาขา) · `branch_id` ผิดรูปแบบ → 400
  - `GET /rooms/:roomID` → ห้องพร้อม `amenities` และ `images` · ห้องไม่มีจริง → 404
  - **[ช่องว่าง]** `GET /rooms/:roomID` ของห้องที่ถูกลบ (`is_active=false`) หรือสถานะไม่ว่าง → ยังคืน 200 ตามปกติ

AC-28: รายชื่อสมาชิกและประวัติรายคน (Admin + Super Admin)
  - `GET /admin/members?search=&page=` → member ที่ยังไม่ถูกลบ ค้นด้วย ILIKE ใน `first_name`, `last_name`, `email`, `phone` เรียงคนที่สมัครล่าสุดก่อน → ทั้ง admin และ super admin เห็นสมาชิกทั้งระบบ
  - `GET /admin/members/:memberID/bookings?page=` → admin เห็นเฉพาะใบในสาขาตัวเอง · super admin เห็นทุกสาขา (ส่ง `branch_id` ไม่ได้)
    - `memberID` ผิดรูปแบบ → 400 · ไม่มีจริงหรือไม่ใช่ member → 200 รายการว่าง (ไม่ตรวจว่าผู้ใช้มีจริง)
  - `GET /admin/bookings` ค้น `search` ด้วย ILIKE ใน `code`, `room_number`, `first_name`, `last_name` ของเจ้าของบัญชี (ไม่รวมอีเมลและชื่อผู้เข้าพัก) · `status=all` = ไม่กรอง · `status`/`stay_type` ที่ไม่รู้จัก → 400

AC-29: โปรไฟล์ของตัวเอง
  - `GET /me` → ข้อมูลผู้ใช้รวม `must_change_password`, `branch_id` และ `branch_name` (ถ้าเป็น admin) โดยไม่มี `password_hash`
  - `PUT /me {first_name, last_name, phone, avatar_url?}` → 200 · `first_name`/`last_name`/`phone` บังคับตามกติกาเดียวกับ AC-1 · ไม่ส่ง `avatar_url` → คงค่าเดิม
  - **[ช่องว่าง]** `avatar_url` ไม่ถูกตรวจเลย → `javascript:alert(1)` ถูกบันทึกได้ (ช่อง URL อื่นในระบบมีการตรวจ ดู AC-14)
  - `POST /me/password {current_password, new_password, confirm_password}` → 200 `{"data":{"message":"เปลี่ยนรหัสผ่านเรียบร้อยแล้ว กรุณาเข้าสู่ระบบใหม่"}}` และปลด `must_change_password`
    - `current_password` ไม่ถูกต้อง → 422 `fields.current_password`
    - รหัสใหม่เหมือนรหัสเดิม → 422 `fields.new_password`
    - `confirm_password` ไม่ตรง → 422 `fields.confirm_password`
    - รหัสใหม่ไม่ผ่านกติกา AC-1 → 422

## ข้อมูลตั้งต้น (`cmd/seed`)
- รันซ้ำได้ ใช้ upsert ทุกจุด → ได้ 3 สาขา (`prachauthit-45`, `bangkhae`, `charoenkrung-place`) ที่มี `contract_fee = 500`, สิ่งอำนวยความสะดวก 24 รายการ, ประเภทห้อง 3 แบบต่อสาขา และห้องตัวอย่าง 9 ห้อง
- บัญชีที่ได้ (รหัสผ่านทุกบัญชี = `SEED_DEFAULT_PASSWORD` หรือ `--password`):
  - superadmin: `super@wisetsuk.com`
  - admin: `adminpracha@wisetsuk.com` (ประชาอุทิศ 45) · `adminbangkae@wisetsuk.com` (บางแค) · `admincharoenkrung@wisetsuk.com` (เจริญกรุงเพลส)
  - member: `member.1@wisetsuk.com`, `member.2@wisetsuk.com`, `member.3@wisetsuk.com`
- รูปเดโม: seed อ่านไฟล์ใน `uploads/<slug>/` (flag `--uploads` ค่าเริ่มต้นคือ `UPLOAD_DIR`) แล้วเขียนแค่ URL `{PUBLIC_BASE_URL}/uploads/<slug>/<path>` ลงฐานข้อมูล ไม่ได้เก็บเป็น blob ใน `assets` รูปเสิร์ฟผ่าน route `/uploads/*`
  - แกลเลอรีสาขา = ไฟล์รูปทุกไฟล์ใต้ `uploads/<slug>/` (`.jpg/.jpeg/.png/.webp`) เรียงตาม path ซึ่งกลายเป็น `sort_order` 0, 1, 2… และ `caption.txt` (UTF-8) ในโฟลเดอร์คือคำบรรยายของทุกรูปในโฟลเดอร์นั้น
  - รัน seed ซ้ำ → แถวใน `branch_images` ที่ URL ขึ้นต้นด้วย `{PUBLIC_BASE_URL}/uploads/<slug>/` ถูกแทนด้วยไฟล์ชุดปัจจุบัน (ไฟล์ที่ลบออกจาก repo หายจากแกลเลอรีด้วย) ส่วนรูปที่อัปผ่าน API (`/files/ast-…`) อยู่ครบ
  - รูปปกห้องมาจากฟิลด์ `cover` ใน `roomSeed` → ถูกตั้งค่าเฉพาะเมื่อ `rooms.image_url` ว่าง หรือเป็น URL เดโมอยู่แล้ว รูปปกที่แอดมินอัปเองไม่ถูกทับ · ไฟล์ที่ `cover` อ้างถึงไม่มีอยู่ → seed ล้มทั้ง transaction
  - ไม่มีโฟลเดอร์ `uploads` → ข้ามส่วนรูปทั้งหมดพร้อมพิมพ์เตือน ข้อมูลส่วนอื่นยังลงครบ
  - เปลี่ยน `PUBLIC_BASE_URL` → ต้องรัน seed ใหม่ เพราะ URL ถูกเก็บแบบเต็ม
- **[ช่องว่าง]** อีเมล admin ใน seed **ไม่ตรง** กับที่ `AGENTS.md` และ `backend/README.md` บอกไว้ (`admin.1@` / `admin.2@` / `admin.3@`)

## Error format (ทั้งระบบ)
```
{"error": {"code": "machine_readable_code", "message": "ข้อความไทย", "fields": {"ชื่อฟิลด์": "ข้อความไทย"}}}
```
`fields` มีเฉพาะเมื่อ code = `validation_failed`
code ที่ใช้มี 13 ตัวเท่านั้น: `unauthorized` · `forbidden` · `not_found` · `conflict` · `bad_request` · `validation_failed` ·
`method_not_allowed` · `internal_error` · `invalid_credentials` · `account_disabled` · `room_unavailable` · `invalid_state` · `too_many_requests`

## API Contract (ตามโค้ดจริง)

```
GET  /health                                  → 200 {"status":"healthy","time"} (ไม่ห่อ data)
GET  /files/:assetID   (+HEAD)                → 200 ไฟล์ (public, ETag+immutable) | 304 | 400 | 404
GET  /uploads/*filepath                       → 200 ไฟล์สลิป (public ไม่มี auth ดู AC-17) | 404

--- Guest ---
POST /api/v1/auth/register {email,password,first_name,last_name,phone}
                                              → 201 token คู่ | 400 | 409 | 422
POST /api/v1/auth/login {email,password}      → 200 | 401 invalid_credentials | 403 account_disabled
                                              | 429 too_many_requests (+Retry-After)
POST /api/v1/auth/refresh {refresh_token}     → 200 | 401
POST /api/v1/auth/logout {refresh_token}      → 204 | 400 ไม่มี body
GET  /api/v1/auth/oauth                       → 200 {"providers":[...]}
GET  /api/v1/auth/oauth/:provider             → 302 ไป provider | 404 provider ปิดอยู่
GET  /api/v1/auth/oauth/:provider/callback    → 302 กลับ frontend ?code= | ?error= | 404 provider ปิดอยู่
POST /api/v1/auth/oauth/exchange {code}       → 200 token คู่ | 401 | 403 account_disabled
GET  /api/v1/branches                         → 200 [branch+images+amenities+nearby_places] (active เท่านั้น)
GET  /api/v1/branches/:branchIDหรือslug        → 200 | 404
GET  /api/v1/amenities                        → 200 [amenity]
GET  /api/v1/room-types?branch_id=            → 200 [room_type] | 400
GET  /api/v1/rooms/search?branch_id=&room_type_id=&stay_type=&check_in=&check_out=
     &move_in_date=&min_price=&max_price=&page=&page_size=
                                              → 200 {data,meta} | 400 | 422
GET  /api/v1/rooms/:roomID                    → 200 room+amenities+images | 400 | 404

--- ต้องล็อกอิน (ทุกบทบาท) ---
GET  /api/v1/me                               → 200 user | 401 | 404 บัญชีถูกลบ
PUT  /api/v1/me {first_name,last_name,phone,avatar_url?}
                                              → 200 | 422
POST /api/v1/me/password {current_password?,new_password,confirm_password}
                                              → 200 {message} (เพิกถอน refresh token ทุกใบ) | 422
GET  /api/v1/me/notifications                 → 200 {items(≤30),unread_count}
POST /api/v1/me/notifications/read {id?}      → 204 | 400 (ไม่มี body / id ผิดรูป)
GET  /api/v1/bookings/:bookingID              → 200 booking+latest_payment | 400 | 403 admin คนละสาขา | 404
POST /api/v1/bookings/:bookingID/cancel       → 200 | 403 admin คนละสาขา | 404 | 409 invalid_state (member)

--- Member เท่านั้น ---
POST /api/v1/bookings {room_id,stay_type,guest_first_name,guest_last_name,guest_phone,
     emergency_phone?,emergency_relation?,check_in_date+check_out_date | move_in_date+contract_date?}
                                              → 201 | 400 | 404 ไม่พบห้อง | 409 room_unavailable | 422
GET  /api/v1/bookings?status=&stay_type=&page=&page_size=
                                              → 200 {data,meta} | 400
POST /api/v1/bookings/:bookingID/payment      multipart: slip,amount,transferred_at,note
                                              → 200 | 400 | 404 | 409 invalid_state | 422

--- Admin + Super Admin ใต้ /api/v1/admin ---
GET    /dashboard                             → 200 {branches:[...],recent_activities:[≤10]}
GET    /activity-logs?actor_role=&actor_id=&action=&branch_id=&search=&page=&page_size=
                                              → 200 {data,meta} | 400 | 403
GET    /bookings?status=&stay_type=&search=&branch_id=&page=&page_size=
                                              → 200 {data,meta} | 400 | 403
POST   /bookings/:bookingID/approve           → 200 | 404 (รวมข้ามสาขา) | 409 invalid_state
POST   /bookings/:bookingID/reject {reason}   → 200 (กลับเป็น pending_payment) | 404 | 409 | 422
PUT    /bookings/:bookingID/appointment {appointment_at,note}
                                              → 200 | 400 ใบรายวัน | 403 | 404 | 422
GET    /members?search=&page=&page_size=      → 200 {data,meta}
GET    /members/:memberID/bookings?page=      → 200 {data,meta} | 400
GET    /rooms?branch_id=&room_type_id=&stay_type=&min_price=&max_price=&page=
                                              → 200 {data,meta} | 400 | 403
POST   /rooms {branch_id,room_type_id,room_number,building,floor,stay_type,price,
     water_rate,electric_rate,size_sqm,description,image_url,status,amenity_ids}
                                              → 201 | 400 | 403 | 422 | 500 room_type_id ไม่มีจริง
PUT    /rooms/:roomID  (field ชุดเดียวกัน)      → 200 | 403 | 404 | 422
PATCH  /rooms/:roomID/status {status}         → 200 | 403 | 404 | 422
DELETE /rooms/:roomID                         → 204 (is_active=false) | 403 | 404
POST   /rooms/:roomID/image                   multipart: image → 200 | 400 | 403 | 404
POST   /rooms/:roomID/images {image_url,sort_order}
                                              → 201 | 403 | 404 | 422
POST   /rooms/:roomID/images/upload           multipart: image,sort_order → 201 | 400 | 403 | 404
DELETE /rooms/:roomID/images/:imageID         → 204 | 403 | 404
PUT    /branch?branch_id= {name,tagline,description,address,phones,line_id,email,latitude,
     longitude,map_url,building_count,floor_count,daily_price_from,monthly_price_min,
     monthly_price_max,water_rate,electric_rate,deposit,advance_payment,contract_fee,cover_image_url}
                                              → 200 | 400 | 403 | 422
PUT    /branch/amenities?branch_id= {amenity_ids}
                                              → 200 | 400 | 403 | 409 | 500
PUT    /branch/nearby?branch_id= {items}      → 200 | 400 | 403 | 422
POST   /branch/cover?branch_id=               multipart: image → 200 | 400 | 403
POST   /branch/images/upload?branch_id=       multipart: image,caption,sort_order → 201 | 400 | 403 | 422
POST   /branch/images?branch_id= {image_url,caption,sort_order}
                                              → 201 | 400 | 403 | 422
DELETE /branch/images/:imageID?branch_id=     → 204 | 400 | 403 | 404

--- Super Admin เท่านั้น ใต้ /api/v1/superadmin ---
GET    /branches                              → 200 [branch] รวมที่ปิดใช้งาน
GET    /staff                                 → 200 [user] (ไม่รวมคนที่ถูกลบ)
POST   /staff {email,first_name,last_name,phone?,branch_id}
                                              → 201 (must_change_password=true) | 400 | 422
PUT    /staff/:userID {first_name,last_name,phone,branch_id,is_active?,password?}
                                              → 200 | 400 ระงับตัวเอง | 404 | 422 | 500 branch_id ไม่มีจริง
DELETE /staff/:userID                         → 204 (soft delete) | 400 ลบตัวเอง | 404

--- ยังไม่มีในโค้ด (แต่ openapi.yaml/README เขียนไว้แล้ว ดู AC-17, AC-23) ---
POST /api/v1/auth/forgot-password             [ยังไม่ implement]
POST /api/v1/auth/reset-password              [ยังไม่ implement]
GET  /api/v1/bookings/:bookingID/slip         [ยังไม่ implement]
```
