# SPEC ระบบจองห้องพักวิเศษสุขนครคอนโด (Backend + Frontend)

> Reverse-engineer จากโค้ดจริงทั้ง repo ณ 2026-09-24 (commit `96298ce`) — `backend/` (migration 0001–0008), `frontend/src/`, `backend/cmd/seed/`, `docker-compose.yml`
> เอกสารอื่น (`backend/README.md`, `docs/openapi.yaml`, `docs/PLAN.md`, `docs/TASKS.md`, `docs/task/frontend/*`) ใช้แค่ชี้จุดที่เขียนไม่ตรงกับโค้ด ไม่ได้ใช้เป็นที่มาของพฤติกรรม
> ทุกบรรทัดคือพฤติกรรมที่ยิงผ่าน API, query DB หรือกดบนเบราว์เซอร์ตรวจได้ ในรูป "ทำแบบนี้ → ได้ผลแบบนี้"
>
> ป้ายกำกับ:
> - **[ยังไม่ implement]** — ตกลงกันว่าจะทำ แต่โค้ดยังไม่มี ห้ามเขียนเทสยืนยันจนกว่าจะทำจริง
> - **[ช่องว่าง]** — โค้ดทำแบบนี้อยู่จริง แต่น่าจะไม่ได้ตั้งใจ หรือขัดกฎใน AGENTS.md / ขัดกับเอกสาร — human ต้องตัดสินว่าจะแก้หรือยอมรับ
> - **ยอมรับความเสี่ยง** — ตัดสินใจแล้วว่ารับได้
> - **[ยังไม่ยืนยัน]** — อนุมานจากการอ่านโค้ด ยังไม่ได้รันเทสยืนยัน
>
> ท้ายแต่ละ AC มีบรรทัด "เทส:" บอกไฟล์ `_test.go` ที่คุมข้อนั้นอยู่ตอนนี้ ข้อที่เขียนว่า "ยังไม่มีเทส" คือทดสอบได้แต่ยังไม่มีใครเขียน

## ส่วนประกอบของระบบ
- `api` — Go 1.26 + Gin v1.12 + pgx v5 (SQL เขียนเอง) ฟังที่ `:8080` · migration ฝังในไบนารีและรันเองตอน start
- `db` — PostgreSQL 16 ผูก `127.0.0.1:5432` · `TZ=Asia/Bangkok`
- `pgadmin` — ผูก `127.0.0.1:5050` ลงทะเบียนเซิร์ฟเวอร์ให้เองจาก `docker/pgadmin/servers.json`
- `seed` — profile `tools` รันด้วย `docker compose run --rm seed`
- `frontend` — React 19 + react-router-dom 6 บน Create React App (`react-scripts start` → `:3000`) ไม่อยู่ใน docker compose · ต่อ backend ผ่าน `REACT_APP_API_BASE_URL` (ค่าเริ่มต้น `http://localhost:8080/api/v1`)
- ไฟล์ช่วยตรวจรูป: `backend/scripts/check-images.sql` (นับรูปที่ URL ชี้ไป blob ที่ไม่มีจริง) และ `check-images.html` ที่ root

## Users
- Guest: ดูสาขา ดูห้อง ค้นหาห้องว่างตามวันที่/ราคา/ประเภท สมัครสมาชิก — ไม่ต้องล็อกอิน
- Member: จองห้อง (รายวัน/รายเดือน), แจ้งชำระเงินพร้อมสลิป, ดู/ยกเลิกการจองของตัวเอง, รับแจ้งเตือน, ล็อกอินด้วยอีเมล/รหัสผ่านหรือ Google/Facebook (ดู AC-24) — **ฝั่ง frontend ยังไม่มีหน้าใดของ Member นอกจาก login/register** (ดู AC-31)
- Admin: จัดการห้องและรายละเอียดสาขา, อนุมัติ/ปฏิเสธ/ยกเลิกการจอง, นัดทำสัญญา, ดูแดชบอร์ด/activity log/รายชื่อสมาชิก — **เฉพาะสาขาที่ผูกอยู่ 1 สาขา** (ดู AC-4) — ฝั่ง frontend มีแค่หน้า login แล้วตกไปหน้า placeholder
- Super Admin: ทุกอย่างที่ Admin ทำได้ในทุกสาขา + เพิ่ม/แก้/ระงับ/ลบบัญชีผู้ดูแล (ดู AC-18) — ฝั่ง frontend มีแดชบอร์ดกับหน้าจัดการผู้ดูแล
- สิทธิ์ถูกบังคับที่ DB:
  - `admin_requires_branch` (CHECK) — admin ต้องมี `branch_id` และบทบาทอื่นต้องไม่มี
  - `uq_admin_per_branch` (partial unique `WHERE role='admin' AND deleted_at IS NULL`) — หนึ่งสาขามี admin ที่ยังไม่ถูกลบได้คนเดียว
  - `uq_single_superadmin` (partial unique `WHERE role='superadmin'`) — ทั้งระบบมี superadmin ได้คนเดียว
- token ใช้ยืนยันแค่ว่าผู้เรียกเป็นใคร ส่วน role, `branch_id` และสถานะบัญชีอ่านจาก DB ทุกคำขอ → ลบ ระงับ หรือย้ายสาขาจึงมีผลทันที (ดู AC-3, AC-4)

## Out of scope (สิ่งที่โค้ดปัจจุบันไม่มี)
- สร้าง ลบ หรือเปิด/ปิดสาขา — มีแค่ `PUT /admin/branch` แก้สาขาที่มาจาก `cmd/seed` คอลัมน์ `branches.is_active` เปลี่ยนได้ทาง DB เท่านั้น
- จัดการประเภทห้อง (`room_types`) และรายการสิ่งอำนวยความสะดวกกลาง (`amenities`) — ใส่ได้ผ่าน `cmd/seed` เท่านั้น API ทำได้แค่เลือกจากรายการที่มีอยู่แล้ว
- สร้าง Super Admin ผ่าน API — สร้างได้ทาง seed เท่านั้น
- ชำระเงินออนไลน์/ตัดบัตร — ใช้วิธีโอนเงินแล้วแนบสลิปให้แอดมินตรวจ
- คืนเงิน ค่าปรับ ใบเสร็จ/ใบกำกับภาษี
- ยืนยันอีเมล และเปลี่ยนอีเมล (ทั้ง API และหน้าจัดการผู้ดูแลไม่รับช่องอีเมลตอนแก้ไข)
- ระงับหรือลบบัญชี member — ทำได้เฉพาะบัญชี staff (ดู AC-18)
- ส่งอีเมลหรือ SMS — ไม่มี mailer เลยสักจุด (โฟลเดอร์ `internal/mailer` ที่ README อ้างถึงไม่มีอยู่) แจ้งเตือนทั้งหมดเก็บในตาราง `notifications` แล้วดึงผ่าน `/me/notifications`
- รีเซ็ตรหัสผ่านของ member — ทำได้เฉพาะบัญชี staff (ดู AC-18) ส่วนลืมรหัสผ่านด้วยตัวเอง **[ยังไม่ implement]** (ดู AC-23)
- Rate limit ที่ endpoint อื่นนอกจาก `/auth/login` (register, refresh, oauth exchange) — มีเฉพาะ login (ดู AC-2)
- endpoint จบสัญญา/เช็คเอาต์ — ปล่อยห้องด้วย `PATCH /admin/rooms/:roomID/status` แทน
- บังคับ `must_change_password` ที่ฝั่ง backend — ธงนี้แค่ส่งออกไปกับ `GET /me` และ `user` ใน token pair ส่วน backend ยังให้ใช้ทุก endpoint ได้ตามปกติ และ frontend ก็ยังไม่ได้บังคับ (ดู AC-37)
- `nearby_places.category` เป็น free text ไม่ใช่ enum
- Frontend ที่ยังไม่มี (API ฝั่ง backend มีครบแล้ว):
  - หน้ารายละเอียดห้อง (`/rooms/:roomId` เป็น placeholder)
  - ฟอร์มจอง แจ้งชำระเงิน ประวัติการจอง ยกเลิกการจอง
  - โปรไฟล์ เปลี่ยนรหัสผ่าน กระดิ่งแจ้งเตือน
  - ปุ่มล็อกอินด้วย Google/Facebook และหน้ารับ callback (`/auth/callback`)
  - ทุกหน้าของ Admin สาขา (`/admin` เป็น placeholder)
  - เมนู "จัดการสาขา" และ "activity log" ของ Super Admin (placeholder)
  - ต่ออายุ token อัตโนมัติด้วย refresh token

## Acceptance Criteria — Backend

AC-1: สมัครสมาชิกได้ และได้สิทธิ์ member เสมอ
  - `POST /auth/register` ด้วยอีเมลใหม่และข้อมูลครบ → 201 พร้อม `access_token`, `refresh_token`, `expires_at`, `token_type="Bearer"`, `user.role="member"` และมี log `auth.register`
  - ฟิลด์ที่รับมีแค่ `email, password, first_name, last_name, phone` — ส่งฟิลด์อื่นมาด้วย (เช่น `role`) → 422 `validation_failed` และ `fields.role = "ไม่อนุญาตให้ส่งฟิลด์นี้"` (decoder ปฏิเสธฟิลด์แปลกหน้าทั้งระบบ ดู AC-21)
  - อีเมลถูก trim และเก็บเป็นตัวพิมพ์เล็ก → สมัคร `A@B.com` ซ้ำกับ `a@b.com` → 422 และ `fields.email = "อีเมลนี้ถูกใช้งานแล้ว"`
  - อีเมลต้องตรง `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$` → ไม่ผ่าน → 422 `fields.email`
  - อีเมลซ้ำกับบัญชีใดก็ได้ รวมบัญชีที่ถูกระงับหรือ soft delete ไปแล้ว → 422 (การเช็คอีเมลซ้ำไม่กรอง `deleted_at`)
  - สองคำขอสมัครด้วยอีเมลเดียวกันพร้อมกัน → ใบหนึ่งได้ 201 อีกใบได้ 409 `conflict` (ชน UNIQUE ที่ DB)
  - กรอกผิดหลายช่อง → 422 และ `fields` มีครบทุกช่องในครั้งเดียว (แต่ละช่องเก็บแค่ข้อความแรกที่ผิด)
  - รหัสผ่านต้องยาว 8–72 ไบต์ และมีทั้งตัวอักษรและตัวเลข → ไม่ผ่าน → 422 · ยาว 73 → 422 (ไม่ปล่อยให้ bcrypt ตัดทิ้งเงียบ ๆ)
  - `phone` บังคับกรอก และต้องตรง `^[0-9\-\s()+]{8,20}$` → ไม่ผ่าน → 422
  - `first_name`/`last_name` ห้ามว่าง (หลัง trim) และยาวไม่เกิน 100 ตัวอักษร → ไม่ผ่าน → 422
  - body พาร์สไม่ได้ → 400 (ดู AC-21)
  - เทส: ยังไม่มีเทส

AC-2: เข้าสู่ระบบด้วย endpoint เดียวทุกบทบาท
  - อีเมล+รหัสถูก และบัญชี active → 200 พร้อม token คู่ บันทึก `last_login_at` และ log `auth.login`
  - อีเมลไม่มีในระบบ / รหัสผิด / บัญชี OAuth ที่ยังไม่ตั้งรหัส (`password_hash=''`) / บัญชี staff ที่ถูก soft delete → 401 `invalid_credentials` ข้อความ "อีเมลหรือรหัสผ่านไม่ถูกต้อง" เหมือนกันทุกกรณี
  - กรณีไม่พบอีเมลหรือบัญชีไม่มีรหัส ระบบเผาเวลาด้วย dummy bcrypt hash ที่สุ่มตอน boot → จับเวลากรณี "อีเมลไม่มี" เทียบ "รหัสผิด" อย่างละ 50 ครั้ง ค่า median ต้องต่างกันไม่เกิน 20%
  - บัญชี `is_active=false` + **รหัสถูก** → 403 `account_disabled` · รหัสผิด → 401 `invalid_credentials` (เช็คสถานะบัญชีหลังตรวจรหัสผ่านแล้วเท่านั้น)
    **ยอมรับความเสี่ยง:** คนที่รู้รหัสผ่านอยู่แล้วจะรู้ว่าบัญชีนั้นถูกระงับ
  - body `{}` หรือไม่มีฟิลด์ → 401 `invalid_credentials` (login ไม่ตรวจรูปแบบ input) และนับเป็นความล้มเหลวของคู่ (`""`, IP)
  - rate limit นับที่ DB (ตาราง `login_attempts` จาก migration 0008) ในหน้าต่าง 15 นาทีย้อนหลัง:
    - ล้มเหลว 5 ครั้งจากคู่ (อีเมล, IP) เดียวกัน → ครั้งถัดไปจากคู่นั้น → 429 `too_many_requests` ข้อความ "เข้าสู่ระบบผิดหลายครั้งเกินไป กรุณาลองใหม่ใน N นาที" พร้อม header `Retry-After` (วินาที ≥ 1)
    - ล้มเหลว 20 ครั้งจาก IP เดียวกัน ไม่ว่าอีเมลไหน → ครั้งถัดไปจาก IP นั้น → 429
    - ติดทั้งสองชั้น → `Retry-After` เป็นค่าที่นานกว่า
    - "ล้มเหลว" = ได้ 401 `invalid_credentials` (รวมอีเมลที่ไม่มีในระบบ) · 403 `account_disabled` ไม่นับ
    - ถูกจำกัดแล้ว → ได้ 429 แม้รหัสผ่านที่ส่งมาจะถูก และได้เหมือนกันไม่ว่าอีเมลจะมีบัญชีหรือไม่ (ตรวจเพดานก่อนแตะตาราง `users`)
    - อีเมลเดียวกันจาก IP อื่น → ไม่ถูกจำกัด (คนอื่นล็อกบัญชีเหยื่อจากเครื่องตัวเองไม่ได้)
    - เข้าสู่ระบบสำเร็จ → ความล้มเหลวของคู่ (อีเมล, IP) นั้นถูกลบ · ตัวนับราย IP ของอีเมลอื่นไม่ถูกลบ
    - ความล้มเหลวที่เก่ากว่า 15 นาที → ไม่นับ และถูกลบทิ้งตอนบันทึกความล้มเหลวครั้งใหม่
    - อีเมลใน `login_attempts` เก็บเป็นตัวพิมพ์เล็ก และไม่มี FK ไป `users` (ตั้งใจ เพื่อนับอีเมลที่ไม่มีบัญชีด้วย)
    - IP มาจากกติกาเดียวกับ activity log (ดู AC-13) → client ปลอม `X-Forwarded-For` เพื่อหลบไม่ได้
    - **[ยังไม่ยืนยัน]** คำขอที่ยิงพร้อมกันจำนวนมากอาจเกินเพดานไปได้เล็กน้อย เพราะการนับกับการบันทึกไม่ได้อยู่ใน lock เดียวกัน
  - เทส: `account/ratelimit_test.go` (7 เทส: รายคู่, อีเมลที่ไม่มีบัญชี, IP อื่นไม่โดน, ราย IP, สำเร็จแล้วล้าง, ของเก่าหมดอายุ, บัญชีระงับไม่นับ) · `account/oauth_test.go` `TestPasswordLoginRejectedForOAuthOnlyAccount` · กรณี timing ยังไม่มีเทส

AC-3: อายุและการหมุนเวียนของ token
  - access token เป็น JWT HS256 ที่มี `iss="wisetsuk-api"`, `sub="usr-NNN"`, `jti` สุ่ม อายุ 30 นาที (`ACCESS_TOKEN_TTL_MIN`) → ใช้หลังหมดอายุ, alg อื่น, issuer อื่น หรือ claim `role` ไม่รู้จัก → 401 `unauthorized`
  - header ต้องเป็น `Authorization: Bearer <token>` → ไม่มี prefix `Bearer ` หรือ token ว่าง → 401
  - `expires_at` ใน token pair คือเวลาหมดอายุของ access token
  - refresh token เป็นสตริงสุ่ม 32 ไบต์ (base64url) อายุ 30 วัน (`REFRESH_TOKEN_TTL_DAY`) ตาราง `refresh_tokens` เก็บแค่ sha256 hash
  - `POST /auth/refresh {"refresh_token": "..."}` ด้วยใบที่ยังใช้ได้ → 200 พร้อม token คู่ใหม่ และใบเดิมถูกเพิกถอนในคำสั่ง UPDATE เดียว (rotation)
  - ใช้ใบเดิมซ้ำ / ใบหมดอายุ / ใบที่ logout แล้ว / สตริงมั่ว / ค่าว่าง → 401 `unauthorized`
  - refresh ของบัญชีที่ `is_active=false` → 401 (ใบนั้นถูกใช้ไปแล้วด้วย)
  - token ใบใหม่อ่าน role, ชื่อ และ `branch_id` จาก DB ณ ตอน refresh (ดู AC-4 เรื่องย้ายสาขา)
  - เปลี่ยนรหัสผ่านสำเร็จ (`POST /me/password`) → refresh token ทุกใบของบัญชีนั้นใช้ไม่ได้ทันที รวมใบของ session ปัจจุบันด้วย
    **ยอมรับความเสี่ยง:** access token ใบเดิมยังใช้ได้จนหมดอายุ (สูงสุด 30 นาที)
  - `POST /auth/logout {"refresh_token": "..."}` ด้วยค่าใดก็ได้ (ผิด/หมดอายุ/ไม่มีจริง/ว่าง) → 204 เหมือนกันหมด
  - `POST /auth/logout` แบบไม่มี body → 400 (ต้องส่ง JSON object มาเสมอ)
  - ทุกคำขอที่มี token → middleware ตรวจลายเซ็นกับอายุของ JWT แล้วอ่านบัญชีจาก DB ด้วย `sub` → บัญชีถูก soft delete หรือ `is_active=false` → 401 ทันทีแม้ token ยังไม่หมดอายุ
  - admin ที่ `branch_id` เป็น NULL → 401 (DB มี CHECK กันไว้อยู่แล้ว middleware เช็คซ้ำ)
  - role, `branch_id` และชื่อผู้กระทำมาจาก DB ไม่ใช่ claim ใน JWT (claim `role`/`branch_id`/`name` ยังถูกใส่ใน token ให้ frontend อ่านได้ แต่ backend ใช้แค่ตรวจว่า `role` เป็นค่าที่รู้จัก)
  - ถูกระงับแล้วเปิดใช้งานคืนก่อน access token หมดอายุ → token ใบเดิมกลับมาใช้ได้ (ไม่ได้เพิกถอนรายใบ แค่ตรวจสถานะบัญชี)
  - DB ล่มระหว่างตรวจบัญชี → 500 ไม่ใช่ 401
  - เทส: `account/revocation_test.go` (token ปกติใช้ได้, ลบแล้ว 401 ทันที, ระงับแล้ว 401 ทันที, ย้ายสาขาแล้วเห็นสาขาใหม่ทันที) · rotation/logout ยังไม่มีเทส

AC-4: Admin ถูกล็อกไว้ที่สาขาเดียว (กฎเหล็ก)
  - `branch_id` ของ admin อ่านจาก DB ทุกคำขอ (ดู AC-3) ไม่ใช่จาก claim ใน token
  - endpoint แบบรายการใต้ `/admin` ที่ admin สาขา A ส่ง `?branch_id=<สาขา B>` มา → 403 `forbidden`
  - admin สาขา A ไม่ส่ง `branch_id` → ระบบบังคับใช้สาขา A (ไม่ใช่ "ทุกสาขา") → เช่น `GET /admin/bookings` ทุกแถวมี `branch_id` เป็นสาขา A
  - admin สาขา A ส่ง `?branch_id=<สาขา A>` → ใช้ได้ตามปกติ
  - super admin ไม่ส่ง `branch_id` ใน endpoint แบบรายการ → เห็นทุกสาขา · ใน endpoint ที่แก้ข้อมูลสาขา (`/admin/branch*`, `POST /admin/rooms`) → 400 "กรุณาระบุสาขา (branch_id)"
  - admin สาขา A แตะทรัพยากรรายชิ้นของสาขา B → status ที่ได้ **ไม่เหมือนกันทุก endpoint**:
    - `POST /admin/bookings/:id/approve` และ `/reject` → 404 (แปลงผ่าน `hideCrossBranch`)
    - `DELETE /admin/branch/images/:imageID?branch_id=<A>` ด้วยรหัสรูปของสาขา B → 404 (SQL กรอง `branch_id`)
    - `DELETE /admin/rooms/:roomID/images/:imageID` → 403 (ตรวจจากห้อง)
    - `GET /bookings/:id`, `POST /bookings/:id/cancel`, `PUT /admin/bookings/:id/appointment`, `PUT`/`PATCH`/`DELETE /admin/rooms/:roomID`, `POST /admin/rooms/:roomID/image(s)` → **403**
    - **[ช่องว่าง]** เจตนาในคอมเมนต์ของโค้ดคือ "ทรัพยากรรายชิ้นต้องได้ 404 เพื่อไม่ยืนยันว่ารหัสนั้นมีจริง" แต่มีแค่ approve/reject ที่ทำตาม endpoint อื่นในลิสต์ 403 ข้างบนยังเปิดเผยว่ารหัสนั้นมีอยู่จริง
  - super admin ย้าย admin จากสาขา A ไป B → admin คนนั้นใช้ access token ใบเดิมต่อ → **เห็นสาขา B ทันที** และขอ `?branch_id=<A>` → 403
  - `GET /admin/members` คืนสมาชิก **ทุกคนในระบบ** ไม่กรองตามสาขา (member ไม่ผูกกับสาขา) ส่วน `GET /admin/members/:id/bookings` ของ admin คืนเฉพาะใบจองในสาขาตัวเอง (ดู AC-28)
  - เทส: `shared/access/access_test.go` (`TestBranchScope`, `TestRequireBranch`, `TestIdentityHasBranch`) · `booking/statemachine_test.go` "ข้ามสาขา" · `account/revocation_test.go` `TestMovedAdminSeesNewBranchImmediately` · `storage/asset_test.go` `TestUploadRoomImageAcrossBranchIsForbidden` · `room/images_test.go` `TestAdminCannotDeleteImageOfAnotherBranch`

AC-5: Member เห็นเฉพาะของตัวเอง
  - member A เปิดใบจองของ member B (`GET /bookings/:id`) → 404 (ไม่ใช่ 403 เพื่อไม่เปิดเผยว่ารหัสนั้นมีจริง)
  - member A แจ้งชำระเงินหรือยกเลิกใบของ member B → 404
  - member เรียก endpoint ใต้ `/admin` หรือ `/superadmin` → 403 · admin/superadmin เรียก `POST /bookings`, `GET /bookings` หรือ `POST /bookings/:id/payment` → 403 (endpoint เหล่านี้เปิดให้ member เท่านั้น)
  - `GET /me/notifications` → 200 `{items, unread_count}` โดย `items` เรียงใหม่สุดก่อนไม่เกิน 30 รายการ และ `unread_count` นับรายการที่ยังไม่อ่านทั้งหมด (ไม่จำกัดที่ 30)
  - แต่ละแจ้งเตือนมี `title`, `body`, `link` ในรูป `/bookings/bkg-NNN` และ `read_at` (ไม่มีถ้ายังไม่อ่าน)
  - `POST /me/notifications/read {}` → ทุกรายการที่ยังไม่อ่านของผู้ใช้นั้นเป็นอ่านแล้ว → 204
  - `POST /me/notifications/read {"id":"ntf-001"}` → อ่านรายการเดียว → 204 · ถ้าเป็นของคนอื่นหรือไม่มีจริง → 204 เหมือนกัน โดยไม่เปลี่ยนอะไร
  - `POST /me/notifications/read` แบบไม่มี body → 400 · `id` ผิดรูปแบบ → 400
  - **[ช่องว่าง]** `Notification.ID` ประกาศชนิดเป็น `UserID` (`account/model.go`) → `GET /me/notifications` คืน `id` เป็น `"usr-001"` แต่ `/read` รับแค่ `"ntf-001"` → เอา id ที่ได้จากรายการไปส่งตรง ๆ ได้ 400
  - เทส: `httpx/errorenvelope_test.go` "member เรียก admin" · ที่เหลือยังไม่มีเทส

AC-6: จองรายวันคิดเงินตามจำนวนคืน
  - จำนวนคืน = `int((check_out - check_in).Hours() / 24)` เช่น 5–8 มี.ค. = 3 คืน
  - จอง 3 คืนในห้อง 1,200/คืน → 201 และ `total_amount = 3600`, `nights = 3`, `status = "pending_payment"`, `code` รูปแบบ `PT-NNN`
  - `total_amount` ถูกเขียนลงใบจองครั้งเดียวตอนสร้าง → แก้ราคาห้องเป็น 2,000 ทีหลัง → ใบเดิมยังเป็น 3600
  - `check_out_date <= check_in_date` → 422 `fields.check_out_date` (DB ก็มี CHECK `daily_dates_required` กันไว้อีกชั้น)
  - `check_in_date` เป็นวันที่ผ่านมาแล้วตามปฏิทินไทย → 422 · เป็นวันนี้ → ผ่าน (ดู AC-22)
  - วันที่ต้องอยู่ในรูปแบบ `YYYY-MM-DD` → ผิดรูปแบบหรือไม่ส่งมา → 422
  - `stay_type` ไม่ใช่ `daily`/`monthly` → 422 · `stay_type` ไม่ตรงกับ `stay_type` ของห้อง → 422 `fields.stay_type`
  - ไม่ส่ง `room_id` → 422 · `room_id` ผิดรูปแบบ (ไม่ใช่ `rm-001`) → 400 · ห้องไม่มีจริงหรือ `is_active=false` → 404
  - ฟิลด์บังคับ: `guest_first_name`, `guest_last_name`, `guest_phone` (regex เดียวกับ AC-1) · ฟิลด์ไม่บังคับ: `emergency_phone` (ถ้าส่งมาต้องผ่าน regex), `emergency_relation` (ไม่ตรวจ)
  - ไม่มีเพดานความยาวชื่อผู้เข้าพัก ไม่มีเพดานจำนวนคืน และไม่มีฟิลด์จำนวนผู้เข้าพัก
  - สร้างสำเร็จ → member ได้แจ้งเตือน "สร้างรายการจองสำเร็จ" ที่อ้าง `code` และมี log `booking.create` ผูกกับสาขาของห้อง
  - เทส: ยังไม่มีเทสเฉพาะ (`booking/statemachine_test.go` ใช้การจองเป็นขั้นเตรียมข้อมูลเท่านั้น)

AC-7: จองรายเดือนเก็บแค่ค่าทำสัญญา
  - ห้อง 5,000/เดือน ในสาขาที่มี `branches.contract_fee = 500` → 201 และ `total_amount = 500` (ค่าเช่าไม่ถูกบันทึกที่ใบจอง) · `nights` ไม่มีในผลลัพธ์
  - `contract_fee` ถูกคัดลอกลงใบจองตอนสร้าง → แก้ค่าทำสัญญาของสาขาเป็น 800 ทีหลัง → ใบเดิมยังเป็น 500
  - `move_in_date` บังคับส่ง → ไม่ส่งหรือเป็นวันที่ผ่านมาแล้ว → 422 (DB มี CHECK `monthly_dates_required`)
  - `contract_date` ไม่บังคับ และไม่ตรวจว่าต้องเป็นอนาคตหรือต้องสัมพันธ์กับ `move_in_date` → ถ้าส่งมาผิดรูปแบบ → 422
  - `stay_type=monthly` บนห้องที่เปิดแบบ daily → 422
  - เทส: ยังไม่มีเทส

AC-8: ห้องเดียวจองซ้อนไม่ได้ (กฎเหล็ก)
  - ใบจองที่ถือห้องไว้คือใบสถานะ `pending_payment` / `awaiting_review` / `approved` ส่วน `cancelled` ไม่ถือ → enum ใน DB มีแค่ 4 ค่านี้ (`rejected`/`completed` ถูกลบใน migration 0002)
  - ห้องที่มีใบรายเดือนถืออยู่ 1 ใบ → จองใหม่ทุกประเภทในห้องนั้นไม่ได้เลย ไม่ว่าวันที่ไหน → 409 `room_unavailable`
  - ใบรายวันสองใบทับกันเมื่อ `ใหม่.check_in < เก่า.check_out AND ใหม่.check_out > เก่า.check_in` → จองทับ → 409 `room_unavailable`
  - จองรายวันที่ `check_in` ตรงกับ `check_out` ของใบเดิมพอดี → 201 (ไม่นับว่าทับ)
  - ห้องสถานะ `maintenance` หรือ `occupied` → 409 `room_unavailable`
  - 2 คำขอจองห้องเดียวกันพร้อมกัน → สำเร็จ 1 ใบ เพราะมี `SELECT ... FOR UPDATE` ล็อกแถวห้องไว้ตลอด transaction
  - **[ช่องว่าง]** กฎ "ห้ามทับ" และ "ใบรายเดือนที่ยังไม่จบมีได้ทีละ 1 ใบ" บังคับที่ application (row lock + `IsAvailable`) เท่านั้น → ใน DB **ไม่มี** exclusion constraint หรือ partial unique index รองรับ ขัด AGENTS.md ข้อ 4 → `INSERT` ตรงเข้า DB สร้างใบซ้อนได้
  - **[ช่องว่าง]** ห้องของสาขาที่ `is_active=false` → `POST /bookings` ยังจองได้ตามปกติ (ตรวจแค่ `rooms.is_active`) ทั้งที่หน้าสาธารณะซ่อนห้องนั้นแล้ว (ดู AC-12, AC-27)
  - **[ช่องว่าง]** ใบรายวันที่ค้างอยู่ไม่กันการจองรายเดือน (ตัวตรวจทับของรายวันทำงานเฉพาะเมื่อมี `check_in`/`check_out`) → ถ้าแอดมินเปลี่ยน `stay_type` ของห้องจาก daily เป็น monthly ระหว่างมีใบรายวันค้าง ห้องจะรับใบรายเดือนซ้อนได้ (ดู AC-19)
  - `code` มีรูปแบบ `PT-001` มาจาก sequence `booking_code_seq` และ `LPAD` อย่างน้อย 3 หลัก → เกิน 999 แล้วยาวขึ้นเอง (`PT-1000`)
  - URL ทุกเส้นใช้ id ที่มี prefix เช่น `bkg-001` ไม่ใช่ `code` → `GET /bookings/PT-001` → 400
  - เทส: `booking/statemachine_test.go` "ห้องยังถูกล็อกอยู่" และ "enum มีแค่ 4 สถานะ" · จองพร้อมกันยังไม่มีเทส

AC-9: แจ้งชำระเงินพร้อมสลิป
  - `POST /bookings/:id/payment` แบบ multipart (`slip`, `amount`, `transferred_at`, `note`) บนใบ `pending_payment` ของตัวเอง → 200 ใบเป็น `awaiting_review` และ response มี `latest_payment` ที่เพิ่งสร้าง (`status="submitted"`)
  - ลำดับการตรวจ (ผิดข้อแรกที่เจอ → ตอบทันที):
    1. `:id` ผิดรูปแบบ → 400
    2. body ใหญ่เกิน `MAX_UPLOAD_MB` (ค่าเริ่มต้น 5) + 1MB หรือพาร์ส multipart ไม่ได้ → 400
    3. `amount` ไม่ใช่ตัวเลขหรือว่าง → 422 `fields.amount`
    4. `transferred_at` พาร์สไม่ได้หรือว่าง → 422 `fields.transferred_at`
    5. `note` ไม่ใช่ UTF-8 → 400
    6. ไม่แนบ `slip` → **400** (ไม่ใช่ 422) · ไฟล์ไม่ใช่ JPG/PNG/WEBP ตามเนื้อไฟล์ → 400 · ใหญ่เกินเพดาน → 400
    7. ใบจองไม่มีจริงหรือเป็นของคนอื่น → 404
    8. ใบไม่ได้อยู่ `pending_payment` → 409 `invalid_state`
    9. `amount <= 0` → 422 · `transferred_at` อยู่ในอนาคตเกิน 5 นาที → 422
  - **[ช่องว่าง]** ขั้นที่ 6 เขียนไฟล์สลิปลงดิสก์ก่อนตรวจขั้นที่ 7–9 → คำขอที่ตกขั้น 7–9 ทิ้งไฟล์กำพร้าไว้ใน `UPLOAD_DIR`
  - ไม่มีการเทียบ `amount` กับ `total_amount` → ยอดไม่ตรงก็รับตามปกติ และไม่มีการ flag ใด ๆ แอดมินต้องเทียบเองจาก `total_amount` กับ `latest_payment.amount`
  - `transferred_at` ไม่มีขอบเขตย้อนหลัง (วันโอนเมื่อ 2 ปีก่อนก็ผ่าน) · รูปแบบที่ไม่มี timezone ถูกตีความเป็นเวลาไทย (ดู AC-22)
  - การแจ้งชำระเงินหนึ่งครั้งคือ `payments` หนึ่งแถว (`booking_id` ไม่ unique) → แจ้งใหม่หลังถูกปฏิเสธจะเพิ่มแถวใหม่ ไม่เขียนทับแถวเดิม
  - ไฟล์เก็บที่ `UPLOAD_DIR/slips/YYYY/MM/<uuid>.<ext>` โดยไม่ใช้ชื่อไฟล์จาก client และ `slip_url = {PUBLIC_BASE_URL}/uploads/slips/...`
  - สำเร็จ → member ได้แจ้งเตือน "ส่งหลักฐานการชำระเงินแล้ว" และมี log `booking.submit_payment` ที่ `detail` มี `code` กับ `amount`
  - **[ยังไม่ยืนยัน]** 2 คำขอแจ้งชำระพร้อมกันบนใบเดียวกัน → น่าจะสำเร็จทั้งคู่และได้ `payments` 2 แถว เพราะการตรวจสถานะไม่ได้อยู่ใน transaction และไม่มี lock
  - เทส: ยังไม่มีเทสเฉพาะ (`booking/statemachine_test.go` ใช้เป็นขั้นเตรียมข้อมูล)

AC-10: อนุมัติ/ปฏิเสธการจอง
  - ลำดับการตรวจ: `:id` ผิดรูปแบบ (400) → ใบไม่มีจริง (404) → ข้ามสาขา (404) → ไม่ได้อยู่ `awaiting_review` (409) → ปฏิเสธโดยไม่มีเหตุผล (422)
  - อนุมัติใบ `awaiting_review` → 200 ใบเป็น `approved` และบันทึก `reviewed_by`/`reviewed_at` · แถว payment ทุกแถวของใบที่ยัง `submitted` เป็น `approved` พร้อม `reviewed_by`/`reviewed_at`
  - อนุมัติใบรายเดือน → ห้องเป็น `occupied` ใน transaction เดียวกัน · อนุมัติใบรายวัน → สถานะห้องไม่เปลี่ยน
  - ปฏิเสธใบ `awaiting_review` พร้อม `reason` → 200 ใบ**กลับไปเป็น `pending_payment`** และห้องยังถูกถือไว้ (ดู AC-8) ส่วน payment เป็น `rejected` และเก็บ `reason` ที่ `payments.reject_reason` (ตาราง `bookings` ไม่มีคอลัมน์ `reject_reason` แล้ว)
  - ปฏิเสธโดยไม่ส่ง `reason`, ไม่มี body หรือ `reason` มีแต่ช่องว่าง → 422 `fields.reason`
  - body ถูกอ่านเฉพาะเมื่อ `Content-Length > 0` → approve ไม่ต้องส่ง body
  - อนุมัติหรือปฏิเสธใบที่ไม่ได้อยู่ `awaiting_review` → 409 `invalid_state` "รายการจองนี้ไม่ได้อยู่ในสถานะรอตรวจสอบ"
  - admin สาขา A อนุมัติหรือปฏิเสธใบของสาขา B → 404
  - สำเร็จ → member ได้แจ้งเตือน 1 รายการที่อ้าง `code` (กรณีปฏิเสธมี `reason` ในข้อความด้วย) และมี log `booking.approve`/`booking.reject`
  - มีเทส "ยิง approve พร้อมกัน 2 คำขอ → สำเร็จ 1 อีกคำขอได้ 409" อยู่แล้ว **[ยังไม่ยืนยัน]** แต่ในโค้ด `UPDATE bookings` ไม่มีเงื่อนไข `status = 'awaiting_review'` และการตรวจสถานะเกิดก่อนเปิด transaction → ผลอาจขึ้นกับจังหวะ ต้องรันเทสซ้ำหลายรอบเพื่อยืนยัน
  - ใน `GET /admin/bookings` ใบที่กลับเป็น `pending_payment` หลังถูกปฏิเสธจะไม่มี `latest_payment` แนบมา (ตัวแนบข้ามใบ pending) → ต้องเปิด `GET /bookings/:id` ถึงจะเห็นเหตุผลที่ถูกปฏิเสธ
  - เทส: `booking/statemachine_test.go` `TestBookingStateMachine` (approve, reject, invalid_state, ข้ามสาขา, แจ้งเตือน, approve พร้อมกัน)

AC-11: ยกเลิกแล้วห้องต้องกลับมาขายได้ (กฎเหล็ก)
  - member ยกเลิกใบ `pending_payment` ของตัวเอง → 200 ใบเป็น `cancelled` และบันทึก `cancelled_at`
  - member ยกเลิกใบสถานะอื่น (`awaiting_review`/`approved`/`cancelled`) → 409 `invalid_state` ข้อความ "รายการจองนี้ไม่สามารถยกเลิกเองได้ กรุณาติดต่อผู้ดูแลระบบของสาขา"
  - admin ในสาขาตัวเองหรือ super admin ยกเลิกใบสถานะใดก็ได้ → 200
  - **[ช่องว่าง]** admin ยกเลิกใบที่ `cancelled` อยู่แล้ว → 200 อีกครั้ง โดย `cancelled_at` ถูกเขียนทับ, member ได้แจ้งเตือนซ้ำ และเกิด log ซ้ำ (ไม่มีการตรวจสถานะสำหรับ admin)
  - **[ช่องว่าง]** `UPDATE` ของการยกเลิกไม่มี `AND branch_id = ...` (ตรวจสิทธิ์สาขาใน application เท่านั้น) ขัดกับที่ `backend/README.md` อ้างว่า "UPDATE/DELETE ของแอดมินมี branch_id เสมอ"
  - ยกเลิกใบรายเดือน → ห้องที่เป็น `occupied` กลับเป็น `available` ใน transaction เดียวกัน **เฉพาะเมื่อ** ไม่มีใบรายเดือนอื่นที่ยังถือห้องอยู่ · ห้องที่เป็น `maintenance` หรือ `available` อยู่แล้ว → ไม่ถูกแตะ
  - ยกเลิกใบรายวัน → สถานะห้องไม่เปลี่ยน (ใบรายวันไม่เคยเปลี่ยนสถานะห้องตั้งแต่แรก)
  - ห้องที่ถูกปล่อยแล้ว → จองใหม่ได้ทันที
  - สำเร็จ → เจ้าของใบได้แจ้งเตือน "รายการจองถูกยกเลิก" และมี log `booking.cancel`
  - เทส: ยังไม่มีเทส

AC-12: ค้นหาห้องว่าง (`GET /rooms/search`)
  - ผลลัพธ์มีเฉพาะห้องที่ `rooms.is_active=true`, `status='available'` **และสาขาของห้อง `is_active=true`**
  - ส่ง `check_in` และ `check_out` มา → ไม่มีห้องที่มีใบรายวันทับช่วงนั้น (สูตรเดียวกับ AC-8) · ห้องที่ `check_out` ของใบเดิมตรงกับ `check_in` ที่ค้นหา → อยู่ในผลลัพธ์ · ส่งมาแค่ตัวเดียว → ไม่กรองใบรายวันเลย
  - ห้องที่มีใบรายเดือนถืออยู่ → ถูกตัดออกเฉพาะเมื่อ `move_in_date` ของใบนั้น ≤ `move_in_date` ที่ค้นหา (ถ้าไม่ส่งมาใช้ `CURRENT_DATE` ของ DB)
  - **[ช่องว่าง]** กติกาข้อข้างบนหลวมกว่าตอนจองจริง → ห้องที่มีใบรายเดือนเข้าอยู่ในอนาคตโผล่ในผลค้นหา แต่กดจองแล้วได้ 409 (ดู AC-8)
  - `stay_type=daily` และ `check_out <= check_in` → 422 `fields.check_out` · ไม่ส่ง `stay_type` → ไม่ตรวจลำดับวันที่
  - ตัวกรองที่รับ: `branch_id`, `room_type_id`, `stay_type`, `check_in`, `check_out`, `move_in_date`, `min_price`, `max_price`, `page`, `page_size`
  - ไม่ระบุ `page` → หน้า 1 ขนาด 20 พร้อม `meta.total_items`/`meta.total_pages` · `page_size=500` → ได้ไม่เกิน 100 และ `meta.page_size=100`
  - `page`/`page_size` เป็น 0, ติดลบ หรือไม่ใช่ตัวเลข → 400 · `stay_type` ไม่รู้จัก → 400 · วันที่ไม่ใช่ `YYYY-MM-DD` → 400 · ราคาไม่ใช่ตัวเลข → 400 · id ผิดรูปแบบ (ต้องเป็น `brn-001`, `rmt-001`) → 400
  - เรียงตาม `branch_id`, `room_number` · แต่ละห้องมี `amenities` ของห้องนั้นเอง (ดึงทั้งหน้าใน query เดียว) · ห้องที่ไม่มี amenity → ไม่มี key `amenities` เลย (`omitempty`) · ไม่มีแกลเลอรี `images`
  - เทส: `room/amenities_test.go` `TestSearchReturnsAmenitiesPerRoom` · `room/closed_test.go` `TestSearchExcludesRoomsInClosedBranch` · `room/images_test.go` `TestSearchDoesNotReturnGallery`

AC-13: บันทึกทุกการกระทำที่เปลี่ยนข้อมูล
  - action ที่มีจริงในโค้ด (29 รายการ):
    `auth.register`, `auth.login`, `auth.oauth_login`, `auth.oauth_link`, `auth.oauth_register`,
    `user.update_profile`, `user.change_password`,
    `admin.create`, `admin.update`, `admin.delete`,
    `booking.create`, `booking.submit_payment`, `booking.cancel`, `booking.approve`, `booking.reject`, `booking.set_appointment`,
    `branch.update`, `branch.set_amenities`, `branch.update_nearby`, `branch.update_cover`, `branch.add_image`, `branch.delete_image`,
    `room.create`, `room.update`, `room.update_image`, `room.add_image`, `room.delete_image`, `room.update_status`, `room.delete`
  - ทำ action ในลิสต์สำเร็จ → `activity_logs` มีแถวใหม่ที่เก็บ `actor_id`, `actor_role`, `actor_name`, `branch_id`, `action`, `entity_type`, `entity_id`, `detail` (jsonb), `ip_address` และ `created_at`
  - `entity_type`: `user` (auth/user/admin), `booking`, `room` (รวม `room.add_image`/`room.delete_image` ที่ `entity_id` เป็นรหัสห้อง), `branch`, `branch_image` (`branch.add_image`/`branch.delete_image`)
  - `ip_address` มาจาก `c.ClientIP()` ของ gin:
    - ไม่ได้ตั้ง `TRUSTED_PROXIES` → ใช้ IP ของ TCP connection และเมิน `X-Forwarded-For`/`X-Real-IP` ทั้งหมด → client ส่ง `X-Forwarded-For: 1.2.3.4` มาก็ปลอม IP ใน log ไม่ได้
    - คำขอมาจาก IP ที่อยู่ใน `TRUSTED_PROXIES` (IP หรือ CIDR คั่นด้วยจุลภาค) → ใช้ IP ขวาสุดใน `X-Forwarded-For` ที่ไม่ใช่ trusted proxy · ค่าปลอมที่ client แอบใส่ไว้ด้านซ้ายไม่มีผล
    - `TRUSTED_PROXIES` มีค่าที่ไม่ใช่ IP หรือ CIDR → server ไม่ start
    - รันผ่าน Docker port publishing โดยไม่มี proxy → IP ที่เห็นอาจเป็น gateway ของ docker ไม่ใช่ IP จริงของผู้ใช้
  - `actor_name` มาจากชื่อใน DB ณ ตอนทำรายการ (ดู AC-3) → แก้ชื่อแล้ว log ถัดไปใช้ชื่อใหม่ทันที
  - `branch_id` ของ log:
    - action ของ booking/room/branch → สาขาของสิ่งที่ถูกกระทำ
    - action ของ auth/user/admin → สาขาของผู้กระทำ (member และ superadmin เป็น `NULL`) → เช่น `admin.create` ที่ super admin สร้างให้สาขา A มี `branch_id = NULL` และสาขาไปอยู่ใน `detail.branch_id` แทน
  - `refresh`, `logout`, `oauth/exchange`, การอ่านแจ้งเตือน และการเข้าสู่ระบบที่ล้มเหลว → ไม่เขียน log
  - เขียน log ไม่สำเร็จ → error ถูกกลืน และธุรกรรมหลักไม่ล้มตาม
  - `GET /admin/activity-logs` → admin เห็นเฉพาะ log ที่ `branch_id` เป็นสาขาตัวเอง (log ของ member/superadmin ที่เป็น `NULL` จึงไม่เห็น) · super admin ไม่ส่ง `branch_id` → เห็นทุกแถวรวมแถวที่ `branch_id` เป็น `NULL`
  - ตัวกรอง: `action` (ตรงตัว), `search` (ILIKE ใน `actor_name`, ชื่อสาขา, `entity_id`), `actor_id`, `actor_role` (ค่าที่ไม่รู้จัก → 400), `branch_id`, `page`/`page_size` (ค่าเริ่มต้น 20) → เรียงใหม่สุดก่อน · แต่ละแถวมี `branch_name` ด้วย
  - เทส: `middleware/clientip_test.go` (`TestClientIP`, `TestActivityLogIgnoresSpoofedForwardedFor`) · `config/config_test.go` (TRUSTED_PROXIES 3 เทส) · `account/oauth_test.go` `TestOAuthLoginIsAudited`

AC-14: รูปภาพสาขา/ห้อง — อัปโหลดไฟล์ หรือใส่ URL ภายนอก
  - อัปโหลดไฟล์ (multipart ช่อง `image`):
    - `POST /admin/branch/cover?branch_id=` → 200 เปลี่ยน `cover_image_url` คืนสาขาทั้งก้อน
    - `POST /admin/branch/images/upload?branch_id=` (+`caption`, `sort_order`) → 201
    - `POST /admin/rooms/:roomID/image` → 200 เปลี่ยนรูปปกห้อง (`rooms.image_url`)
    - `POST /admin/rooms/:roomID/images/upload` (+`sort_order`) → 201 รูปแกลเลอรี
  - ไฟล์ที่อัปโหลดเก็บเป็น blob ในตาราง `assets` (`content_type`, `size_bytes`, `checksum`, `data`) และระบบคืน URL `{PUBLIC_BASE_URL}/files/ast-NNN` (รูปเดโมจาก seed ต่างออกไป เพราะชี้ `/uploads/...` ดูหัวข้อข้อมูลตั้งต้น)
  - อัปโหลดไฟล์เนื้อเดียวกันซ้ำ → ได้ asset id เดิม (dedup ด้วย `UNIQUE(checksum)` sha256)
  - ชนิดไฟล์ตรวจจาก 512 ไบต์แรกของเนื้อไฟล์จริง ต้องเป็น JPG/PNG/WEBP → ไม่ใช่ → 400 · ใหญ่เกิน `MAX_UPLOAD_MB` (ตรวจทั้ง header และจำนวนไบต์ที่อ่านได้จริง) → 400 · ไม่แนบไฟล์ → 400 "กรุณาแนบไฟล์ในช่อง image"
  - `sort_order` ไม่ใช่จำนวนเต็ม → 400 · ไม่ส่ง → 0 · `caption` ไม่ใช่ UTF-8 → 422 · `caption` ยาวเกิน 200 → 422
  - **[ช่องว่าง]** ไฟล์ถูกเขียนลง `assets` ก่อนตรวจสิทธิ์สาขาและก่อนตรวจ `sort_order` → คำขอที่ได้ 400/403/404 ทิ้ง blob กำพร้าไว้ใน DB
  - `GET /files/:assetID` และ `HEAD` → เปิดสาธารณะ ไม่ต้องใช้ token · ตอบ `ETag` (sha256 ในเครื่องหมายคำพูด), `Cache-Control: public, max-age=31536000, immutable`, `X-Content-Type-Options: nosniff` · ส่ง `If-None-Match` ตรงกับ ETag → 304 · ไม่มี asset นั้น → 404 · id ผิดรูปแบบ → 400
  - ใส่ URL ภายนอกแบบ JSON:
    - `POST /admin/branch/images?branch_id= {image_url, caption, sort_order}` → 201
    - `POST /admin/rooms/:roomID/images {image_url, sort_order}` → 201
    - ต้องเป็น `http`/`https` ที่มี host และยาวไม่เกิน 2048 → `javascript:`, `data:`, path สัมพัทธ์ หรือค่าว่าง → 422
  - ช่อง `image_url` ของ `POST`/`PUT /admin/rooms` และช่อง `cover_image_url` ของ `PUT /admin/branch` → ใช้กติกาเดียวกัน แต่ส่งค่าว่างได้
  - แกลเลอรีห้องเป็นคนละช่องกับรูปปก → เพิ่ม/ลบรูปแกลเลอรีไม่แตะ `rooms.image_url` · `GET /rooms/:id` คืน `images` เรียงตาม `sort_order, id` · แกลเลอรีสาขาเรียงตาม `sort_order, created_at`
  - `DELETE /admin/branch/images/:imageID?branch_id=` → 204 | 404 (รูปไม่มีหรือเป็นของสาขาอื่น)
  - `DELETE /admin/rooms/:roomID/images/:imageID` → 204 | 403 (ห้องคนละสาขา) | 404
  - **[ช่องว่าง]** ถ้า `imageID` เป็นรูปของห้องอื่นในสาขาเดียวกัน → ถูกลบด้วย (SQL ผูกแค่สาขา ไม่ผูก `room_id` ใน URL) และ log บันทึก `entity_id` เป็นห้องใน URL ไม่ใช่ห้องเจ้าของรูป
  - เทส: `storage/asset_test.go` (8 เทส: เก็บใน DB, dedup, รูปปกสาขา, ปฏิเสธไฟล์ไม่ใช่รูป, HEAD, caption ไม่ใช่ UTF-8, asset ไม่มีจริง 404, อัปรูปห้องข้ามสาขา 403) · `room/images_test.go` (แกลเลอรีเรียงถูก, ห้องไม่มีแกลเลอรี, เพิ่มไม่แตะรูปปก, ลบ, ลบข้ามสาขา)

AC-15: ความปลอดภัยและความทนทานพื้นฐาน
  - ทุก query ใช้ placeholder (`$n`) ผ่าน `database.Binder` และไม่มีการต่อค่าจากผู้ใช้เข้า SQL — **ตรวจด้วย code review ไม่ใช่เทสผ่าน API**
  - error 5xx → body เป็น `{"error":{"code":"internal_error","message":"เกิดข้อผิดพลาดภายในระบบ"}}` เสมอ ไม่มี stack trace หรือข้อความ SQL (รายละเอียดจริง log ไว้ฝั่ง server)
  - panic ใน handler → server ไม่ล้ม และ client ได้ 500 เป็น JSON
  - ข้อความ error ที่ผู้ใช้เห็นเป็นภาษาไทยทั้งหมด รวม 404 เส้นทางที่ไม่มีจริง ("ไม่พบข้อมูลที่ต้องการ") และ 405 method ผิด ("ไม่รองรับ HTTP method นี้")
  - ไม่มี `DATABASE_URL` หรือ `JWT_SECRET` สั้นกว่า 32 ตัว → server ไม่ start (fail fast)
  - ทุก response มี header `X-Request-ID` → ถ้า client ส่งมาจะใช้ค่านั้น ถ้าไม่ส่งจะสุ่ม UUID ให้
  - CORS อนุญาตเฉพาะ origin ใน `ALLOWED_ORIGINS` (ค่าเริ่มต้น `http://localhost:3000,http://localhost:5173`) พร้อม credentials · method `GET, POST, PUT, PATCH, DELETE, OPTIONS` · header `Accept, Authorization, Content-Type, X-Request-ID` · preflight cache 300 วินาที
  - JSON body ใหญ่เกิน 1MB → 400 · body มี JSON มากกว่า 1 object → 400
  - `GET /uploads/` หรือ path ที่ลงท้ายด้วย `/` → 404 (ไม่เปิดให้ไล่ดูรายชื่อไฟล์)
  - รหัสผ่านเก็บเป็น bcrypt (cost 12 จาก `BCRYPT_COST` ค่านอกช่วงที่ bcrypt รับถูกปัดเป็น 12 · DB มี CHECK ว่าต้องขึ้นต้นด้วย `$2` หรือเป็นค่าว่าง)
  - migration รันอัตโนมัติตอน API start (ไฟล์ละ 1 transaction เรียงตามชื่อไฟล์ บันทึกใน `schema_migrations` และข้ามไฟล์ที่รันแล้ว) · ได้รับ SIGTERM/SIGINT → รอคำขอที่ค้างอยู่ได้สูงสุด 20 วินาที
  - http.Server ตั้ง timeout กัน slowloris: ReadHeader 10s, Read 30s, Write 60s, Idle 120s, header ไม่เกิน 1MB
  - `APP_ENV=production` → gin release mode และ log เป็น JSON ระดับ info · ค่าอื่น → log เป็น text ระดับ debug
  - container รันด้วย user `appuser` (uid 10001) และมี `HEALTHCHECK` ยิง `/health` ทุก 30 วินาที
  - เทส: `httpx/errorenvelope_test.go` `TestAC21_PanicReturns500JSON`, `TestAC21_RouterErrorsUseSameEnvelope` · ที่เหลือยังไม่มีเทส

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
  - path ของไฟล์มาจากโค้ดเท่านั้น (folder คงที่ + ปี/เดือน + UUID) → client ไม่มีทางส่ง path มาต่อเองได้ จึงกัน path traversal ได้ แต่ไม่ได้กันคนที่ไม่มีสิทธิ์
  - route `/uploads/*` เดียวกันนี้ใช้เสิร์ฟรูปเดโมของสาขาด้วย (ดูหัวข้อข้อมูลตั้งต้น) → ปิด route นี้ทิ้งเฉย ๆ จะทำให้รูปเดโมหายไปด้วย
  - **[ช่องว่าง]** `backend/.gitignore` ไม่ได้ ignore `uploads/` แล้ว และ docker compose bind-mount `./uploads` เข้า container → สลิปจริงที่ลูกค้าอัปโหลดตกลงไปที่ `backend/uploads/slips/...` ซึ่ง `git add .` จะดึงเข้า repo ได้
  - **[ยังไม่ implement]** `GET /api/v1/bookings/:bookingID/slip`:
    - เจ้าของใบ / admin สาขาเดียวกัน / super admin → 200
    - คนอื่น หรือใบที่ไม่มีจริง → 404 เหมือนกัน
    - ไม่มี token → 401
    - `docs/openapi.yaml` มี path นี้แล้ว แต่ router ยังไม่มี
  - เทส: ยังไม่มีเทส

AC-18: Super Admin จัดการบัญชีผู้ดูแล
  - `GET /superadmin/staff` → admin และ superadmin ทุกคนที่ยังไม่ถูกลบ รวมคนที่ถูกระงับ (`is_active=false`) เรียง superadmin ก่อนแล้วตามวันที่สร้าง และมี `branch_name` มาด้วย
  - `POST /superadmin/staff {email, first_name, last_name, phone, branch_id}` → 201 ได้ role `admin` และ `must_change_password=true`
  - ข้อมูลที่ POST ส่งไม่ได้ / ถูกตั้งให้อัตโนมัติ:
    - ไม่รับ `password` และไม่รับฟิลด์อื่นนอกจาก 5 ตัวข้างบน → ส่งมา → 422
    - รหัสผ่านตั้งจาก `STAFF_DEFAULT_PASSWORD` → ถ้าไม่ตั้งใช้ `SEED_DEFAULT_PASSWORD` → ถ้าไม่ตั้งทั้งคู่ใช้ `"Wisetsuk!2026"`
    - `phone` ไม่บังคับ แต่ถ้าส่งมาต้องผ่าน regex · ไม่มีเพดานความยาวชื่อ (ต่างจาก register)
  - POST ผิดเงื่อนไข:
    - ไม่ส่ง `branch_id` → 422 "กรุณาเลือกสาขาที่รับผิดชอบ"
    - `branch_id` ผิดรูปแบบ → 400
    - ไม่มีสาขานั้น → 422 `fields.branch_id = "ไม่พบสาขาที่เลือก"`
    - สาขามี admin อยู่แล้ว → 422 `fields.branch_id = "สาขานี้มีผู้ดูแลอยู่แล้ว หนึ่งสาขามีผู้ดูแลได้คนเดียว"` (ดัก `uq_admin_per_branch`)
  - อีเมลซ้ำ รวมอีเมลของบัญชีที่ถูกลบไปแล้ว → 422 `fields.email`
  - สำเร็จ → log `admin.create` ที่ `detail` มี `email` และ `branch_id`
  - `PUT /superadmin/staff/:userID {first_name, last_name, phone, branch_id, is_active, password?}` เป็นการแทนที่ทั้งชุด → ไม่ส่ง `first_name`/`last_name` → 422 · ไม่ส่ง `phone` → ถูกล้างเป็นค่าว่าง · ไม่ส่ง `is_active` → คงค่าเดิม
    - หาเป้าหมายก่อนตรวจ body → เป้าหมายเป็น member, ไม่มีจริง หรือถูกลบแล้ว → 404
    - เป้าหมายเป็น admin แต่ไม่ส่ง `branch_id` → 422 · เป้าหมายเป็น superadmin แต่ส่ง `branch_id` มา → 422
    - ย้ายไปสาขาที่มี admin แล้ว → 422 `fields.branch_id`
    - **[ช่องว่าง]** ย้ายไปสาขาที่ไม่มีอยู่จริง → **500** (FK violation ไม่ถูกแปลง ต่างจาก POST ที่เช็คก่อน)
    - ระงับตัวเอง (`is_active=false` บน id ของตัวเอง) → 400 "ไม่สามารถระงับบัญชีของตนเองได้"
    - ส่ง `password` มา → ต้องผ่านกติกา AC-1 · ตั้งรหัสใหม่ทันที, `must_change_password=true` และเพิกถอน refresh token ทุกใบของบัญชีนั้น — เป็นช่องทางเดียวที่ super admin รีเซ็ตรหัสให้คนอื่นได้
    - ระงับ (`is_active=false`) → access token ใบเดิมได้ 401 ทันที และ refresh ได้ 401 (ดู AC-3) — refresh token ไม่ถูกเพิกถอน จึงใช้ได้อีกถ้าเปิดบัญชีคืน
    - ย้ายสาขา → ไม่เพิกถอน token แต่ token ใบเดิมเห็นสาขาใหม่ทันที (ดู AC-4)
    - สำเร็จ → log `admin.update` ที่ `detail.is_active` เป็นค่าหลังแก้
  - `DELETE /superadmin/staff/:userID` → 204 เป็น soft delete (`deleted_at = now()`, `is_active = false`) และเพิกถอน refresh token ทุกใบ · log `admin.delete`
    - ลบตัวเอง → 400 "ไม่สามารถลบบัญชีของตนเองได้" · เป้าหมายไม่ใช่ admin (member/superadmin), ไม่มีจริง หรือถูกลบไปแล้ว → 404
    - แถวยังอยู่ใน DB → `activity_logs.actor_id`, `bookings.reviewed_by`, `payments.reviewed_by` ไม่หลุด
    - บัญชีที่ถูกลบหายจากการ login (401), token เดิม (401), `GET /me` และรายชื่อทุกหน้า
    - สาขาของคนที่ถูกลบว่างลงทันที → สร้าง admin คนใหม่ลงสาขาเดิมได้ (`uq_admin_per_branch` ไม่นับแถวที่ `deleted_at IS NOT NULL`)
  - DB ปฏิเสธการมี admin คนที่สองในสาขาเดียวกัน และ superadmin คนที่สอง แม้ INSERT ตรงโดยข้าม service
  - เทส: `account/staff_test.go` (12 เทส: ผูกสาขาเดียว, รหัสตั้งต้น+บังคับเปลี่ยน, ไม่ส่งสาขา, สาขาไม่มีจริง, สาขามีคนแล้ว, ย้ายสาขา, ย้ายเข้าสาขาที่มีคน, superadmin ส่งสาขา, soft delete, ลบแล้วสาขาว่าง, DB กัน admin คนที่สอง, DB กัน superadmin คนที่สอง)

AC-19: จัดการห้อง
  - `POST /admin/rooms` → 201 · field: `branch_id, room_type_id, room_number, building, floor, stay_type, price, water_rate, electric_rate, size_sqm, description, image_url, status, amenity_ids`
    - admin ไม่ส่ง `branch_id` → ใช้สาขาตัวเอง · ส่งสาขาอื่น → 403 · super admin ไม่ส่ง → 400
  - `PUT /admin/rooms/:roomID` → 200 เป็นการแทนที่ทั้งชุดด้วย field ชุดเดียวกัน (สาขายึดตามห้องเดิม และ `branch_id` ใน body ถูกเมิน) · ไม่ส่ง `image_url` → คงรูปเดิม · ไม่ส่ง `room_type_id`/`size_sqm` → กลายเป็น `NULL`
  - กติกาของค่า:
    - `room_number` บังคับ (หลัง trim) และยาวไม่เกิน 20
    - `price > 0`, `floor > 0`, `water_rate >= 0`, `electric_rate >= 0`
    - `stay_type` ∈ `daily`/`monthly`
    - `status` ว่าง = `available` · ค่าอื่นนอกจาก 3 ค่า → 422
    - `building` ว่าง = `"1"`
    - ผิดข้อใด → 422
  - `room_number` ซ้ำในสาขาเดียวกัน → 422 `fields.room_number = "เลขห้องนี้มีอยู่ในสาขาแล้ว"` — ชนกับห้องที่ถูกลบไปแล้วด้วย เพราะ `UNIQUE(branch_id, room_number)` ไม่มีเงื่อนไข จึงเอาเลขห้องที่ลบไปแล้วกลับมาใช้ไม่ได้
  - `amenity_ids`: ไม่ส่ง = ไม่แตะของเดิม · `[]` = ล้างทั้งหมด · id ที่ไม่มีจริง → 422 `fields.amenity_ids` และห้องไม่ถูกสร้าง (อยู่ใน transaction เดียวกัน) · id ซ้ำใน array → รับได้ (DISTINCT)
  - **[ช่องว่าง]** `room_type_id` ที่ไม่มีจริง → **500** (FK violation ไม่ถูกแปลง) · `room_type_id` ของอีกสาขา → รับได้ (ไม่ตรวจว่าอยู่สาขาเดียวกัน)
  - `PATCH /admin/rooms/:roomID/status {status}` → 200 · ไม่จำกัดว่าเปลี่ยนจากสถานะไหนไปไหน และไม่ตรวจว่ามีใบจองค้างอยู่ · ค่าไม่อยู่ใน 3 ค่า → 422
  - `DELETE /admin/rooms/:roomID` → 204 เป็น soft delete (`is_active=false`) โดยไม่ตรวจใบจองที่ยังค้าง · ลบห้องที่ลบไปแล้วซ้ำ → 204 อีกครั้ง
  - **[ช่องว่าง]** `PUT`/`PATCH`/`DELETE`/อัปโหลดรูป บนห้องที่ถูกลบแล้ว → ยังทำงานได้ (การหาห้องไม่กรอง `is_active`) แต่ห้องนั้นไม่โผล่ใน `GET /admin/rooms` อีกแล้ว
  - **[ช่องว่าง]** `PUT` เปลี่ยน `stay_type` หรือ `status` ของห้องที่มีใบจองค้างอยู่ → ได้เลย ไม่มีการตรวจ (ดู AC-8)
  - `GET /admin/rooms` → ห้องทุกสถานะที่ `is_active=true` ของสาขาที่ดูแล รวมห้องในสาขาที่ปิดใช้งาน · ตัวกรอง `branch_id, room_type_id, stay_type, min_price, max_price, page, page_size` (ส่ง `check_in`/`check_out`/`move_in_date` มาได้และถูกตรวจรูปแบบ แต่ไม่มีผลกรอง) · ไม่มี `amenities` แนบมา
  - ห้องคนละสาขา → 403 · ห้องไม่มีจริง → 404
  - สำเร็จ → มี log `room.create`/`room.update`/`room.update_status`/`room.delete` ผูกสาขาของห้อง
  - เทส: `room/amenities_test.go` (สร้างพร้อม amenity, แก้แทนที่, แก้ไม่ส่งแล้วคงเดิม, ส่ง `[]` ล้าง, id ไม่มีจริง, id ซ้ำ) · `room/closed_test.go` `TestAdminStillManagesRoomsInClosedBranch`

AC-20: นัดหมายทำสัญญา
  - `PUT /admin/bookings/:bookingID/appointment {appointment_at, note}` บนใบรายเดือน → 200 และเขียนทับนัดเดิมได้ทุกครั้ง
  - `appointment_at` พาร์สไม่ได้ → 422 (ตรวจใน handler ก่อนหาใบจอง)
  - บนใบรายวัน → 400 "วันนัดหมายทำสัญญาใช้ได้เฉพาะการจองรายเดือน"
  - `appointment_at` รับ RFC3339 หรือ `YYYY-MM-DD[ T]HH:MM[:SS]` หรือ `YYYY-MM-DD` (ไม่มี timezone = เวลาไทย) → ไม่อยู่ในอนาคต → 422
  - admin คนละสาขา → 403 (ไม่ใช่ 404 ดู AC-4) · super admin → ทำได้ทุกสาขา
  - **[ช่องว่าง]** ไม่ตรวจสถานะใบจอง → ตั้งนัดบนใบ `pending_payment`, `awaiting_review` หรือ `cancelled` ได้ทั้งหมด
  - สำเร็จ → member ได้แจ้งเตือน "นัดหมายทำสัญญา" ที่อ้าง `code` และวันเวลาในรูป `DD-MM-YYYY HH:MM` · log `booking.set_appointment`
  - **[ช่องว่าง]** วันเวลาในแจ้งเตือนแสดงตาม timezone ที่ client ส่งมา → ส่ง `2026-09-30T03:00:00Z` → ข้อความเป็น `30-09-2026 03:00` ไม่ใช่เวลาไทย 10:00
  - เทส: ยังไม่มีเทส

AC-21: รหัสสถานะและรูปแบบ error สอดคล้องกันทั้งระบบ
  - error: `{"error": {"code": "...", "message": "...", "fields": {...}}}` → `fields` ปรากฏเฉพาะเมื่อ code = `validation_failed`
  - สำเร็จ: `{"data": ...}` · รายการแบ่งหน้ามี `meta: {page, page_size, total_items, total_pages}` เพิ่ม · 204 ไม่มี body · ข้อยกเว้นเดียวคือ `GET /health` ที่ตอบ `{"status":"healthy","time":...}` โดยไม่ห่อ `data`
  - error code มีแค่ 13 ตัว: `unauthorized`, `forbidden`, `not_found`, `conflict`, `bad_request`, `validation_failed`, `method_not_allowed`, `internal_error`, `invalid_credentials`, `account_disabled`, `room_unavailable`, `invalid_state`, `too_many_requests`
  - 429 `too_many_requests` → มี header `Retry-After` (วินาที) เสมอ (ปัจจุบันใช้ที่ `/auth/login` ที่เดียว ดู AC-2)
  - 400 `bad_request`:
    - body พาร์สไม่ได้, มี JSON มากกว่า 1 object หรือชนิดข้อมูลผิด
    - id ใน URL/query/body ผิดรูปแบบ (เช่น `branch_id` ไม่ใช่ `brn-001`) — prefix ไม่สนตัวพิมพ์ (`BRN-1` ใช้ได้) และตัดศูนย์นำหน้าเอง แต่ตัวเลขเปล่า/prefix ของตารางอื่น/ค่า ≤ 0 → 400
    - query param ผิดรูปแบบ
    - ไฟล์อัปโหลดผิดหรือใหญ่เกิน
    - ข้อความใน multipart ไม่ใช่ UTF-8
  - 422 `validation_failed`: ค่าถูกชนิดแต่ผิดกติกา (เช่น `check_out <= check_in`, รหัสผ่านสั้น) **หรือ** body มีฟิลด์ที่ไม่รู้จัก → `fields.<ชื่อฟิลด์>`
  - ไม่มี token หรือ token ใช้ไม่ได้ → 401 `unauthorized` ("กรุณาเข้าสู่ระบบ") · มี token แต่บทบาทไม่พอ → 403 `forbidden` ("คุณไม่มีสิทธิ์เข้าถึงส่วนนี้")
  - ชน unique constraint ที่ service ไม่ได้ดักไว้เอง → 409 `conflict` "ข้อมูลนี้มีอยู่ในระบบแล้ว"
  - **[ช่องว่าง]** ชน foreign key → ไม่มีการแปลง กลายเป็น 500 (ต้นเหตุของ 500 ใน AC-18, AC-19, AC-26) · ฟังก์ชัน `database.IsForeignKeyViolation` มีแล้วแต่ใช้แค่ที่ `amenity_ids` ของห้อง
  - error ทุกตัวรวม 404/405/500 ใช้รูปแบบเดียวกัน
  - เทส: `httpx/errorenvelope_test.go` (8 เทส: 400, 422, fields เฉพาะ 422, 401 vs 403, 404/405 ของ router, panic → 500, ทุก error ใช้ envelope เดียว, helper ตรงสเปก)

AC-22: เวลาและวันที่
  - ทุก timestamp เก็บเป็น `TIMESTAMPTZ` (UTC) ส่วนวันที่ล้วน (`check_in_date`, `check_out_date`, `move_in_date`, `contract_date`) เก็บเป็น `DATE`
  - "วันนี้ / อดีต / อนาคต" ของ `check_in_date` และ `move_in_date` ตัดสินด้วยปฏิทิน Asia/Bangkok เสมอ ไม่ขึ้นกับ TZ ของเครื่อง (`internal/timex` ฝัง tzdata ไว้ในไบนารี) → server ที่เป็น UTC รับจอง "วันนี้" ตอน 00:30 เวลาไทยได้
  - `transferred_at` และ `appointment_at` ที่ไม่มี offset → ตีความเป็นเวลาไทย · มี offset มา → ใช้ตามนั้น
  - การเทียบ "อนาคต" ของ `transferred_at` (+5 นาที) และ `appointment_at` เทียบเป็นจุดเวลาจริง ไม่ใช่วันที่
  - ค่าเริ่มต้นของ `move_in_date` ในการค้นหา (AC-12) ใช้ `CURRENT_DATE` ของ DB ไม่ใช่ `timex` → ขึ้นกับ TZ ของ session ฐานข้อมูล (container `db` ตั้ง `TZ=Asia/Bangkok`)
  - **[ยังไม่ยืนยัน]** timestamp ใน response เป็น RFC 3339 ที่ offset ขึ้นกับ TZ ของ process API (container ตั้ง `TZ=Asia/Bangkok` → `+07:00`) · ฟิลด์วันที่ล้วนออกมาเป็นเที่ยงคืน UTC เช่น `"2026-08-14T00:00:00Z"` ไม่ใช่ `"2026-08-14"` · `timex.FormatISO` มีอยู่แต่ไม่ได้ถูกใช้ตอน serialize response
  - เทส: `timex/timex_test.go` (11 เทส `TestAC22_*`)

AC-23: ลืมรหัสผ่าน — ขอลิงก์รีเซ็ตทางอีเมล **[ยังไม่ implement]**
  - สถานะโค้ดตอนนี้: ไม่มี route `/auth/forgot-password` หรือ `/auth/reset-password`, ไม่มี mailer และไม่มีตารางเก็บ reset token → member ที่ลืมรหัสกู้คืนเองไม่ได้ และ super admin รีเซ็ตให้ได้เฉพาะบัญชี staff (ดู AC-18)
  - `docs/openapi.yaml` และ `backend/README.md` เขียนไว้แล้วว่ามี 2 endpoint นี้ **ซึ่งไม่ตรงกับโค้ด** · ฝั่ง frontend ลิงก์ "ลืมรหัสผ่าน?" พาไปหน้า `/contact` (ดู AC-36)
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
    - redirect URI ที่ส่งให้ provider = `{PUBLIC_BASE_URL}/api/v1/auth/oauth/<provider>/callback`
    - scope: Google `openid email profile` · Facebook `email public_profile` (Graph API v21.0)
    - cookie ตั้งเป็น HttpOnly, SameSite=Lax, path `/api/v1/auth/oauth`, อายุ 10 นาที และตั้ง Secure เฉพาะ `APP_ENV=production`
  - `GET /auth/oauth/:provider/callback` → provider ปิดอยู่ → 404 JSON · กรณีอื่นลบ cookie ทิ้งแล้ว **302 ไป `FRONTEND_OAUTH_CALLBACK_URL` เสมอ** (ค่าเริ่มต้น `http://localhost:3000/auth/callback`):
    - สำเร็จ → `?code=<exchange code>`
    - ผู้ใช้กดยกเลิก / `state` ไม่ตรงกับ cookie / ไม่มี cookie / ไม่มี `code` / provider ไม่ส่งอีเมล / อีเมลยังไม่ยืนยัน (Facebook ถือว่ายืนยันแล้วถ้าส่งอีเมลมา) / แลก token ไม่สำเร็จ / บัญชีถูกระงับ → `?error=<ข้อความไทย>`
  - การจับคู่บัญชี (ทั้งหมดอยู่ใน transaction เดียว):
    1. มี `user_identities` ของ provider+subject นี้แล้ว → เข้าบัญชีเดิม (`auth.oauth_login`) — ใช้ subject จับคู่ ไม่ใช้อีเมล
    2. ไม่มี แต่มีบัญชีอีเมลเดียวกันที่ยังไม่ถูกลบ → ผูก identity กับบัญชีนั้น (`auth.oauth_link`) ทุก role รวมถึง admin/superadmin
    3. ไม่มีทั้งคู่ → สร้าง member ใหม่ที่ `password_hash=''`, `phone=''` และ `first_name` มาจาก provider (ถ้าไม่มีใช้ส่วนหน้า `@` ของอีเมล) (`auth.oauth_register`)
  - log ของทั้งสามทางมี `detail.provider` · บันทึก `last_login_at` ด้วย
  - **[ช่องว่าง]** อีเมลตรงกับบัญชี staff ที่ถูก soft delete → ข้อ 2 หาไม่เจอ ข้อ 3 ชน UNIQUE ของอีเมล → redirect พร้อม `?error=ข้อมูลนี้มีอยู่ในระบบแล้ว`
  - `POST /auth/oauth/exchange {code}` → 200 token คู่ · code ว่าง/ไม่มีจริง/หมดอายุ (2 นาที)/ใช้แล้ว → 401 `unauthorized` "ลิงก์เข้าสู่ระบบหมดอายุแล้ว กรุณาลองใหม่อีกครั้ง" · บัญชีถูกระงับ → 403 `account_disabled` · ตาราง `oauth_exchange_codes` เก็บแค่ hash
  - บัญชีที่ยังไม่มีรหัสผ่านเรียก `POST /me/password` → ไม่ต้องส่ง `current_password` (เป็นการตั้งรหัสครั้งแรก) · ตั้งแล้วครั้งถัดไปต้องส่ง `current_password` ตามปกติ
  - บัญชีที่สร้างผ่าน OAuth มี `phone=''` → `PUT /me` บังคับ `phone` ดังนั้นต้องกรอกเบอร์ก่อนถึงจะแก้โปรไฟล์ได้
  - DB CHECK `users_must_change_password_needs_password` → บัญชีที่ไม่มีรหัสผ่านถูกปักธง `must_change_password` ไม่ได้
  - **[ช่องว่าง]** `backend/.env.example` ตั้ง `GOOGLE_CLIENT_ID=...` และ `GOOGLE_CLIENT_SECRET=...` เป็นสตริง `"..."` ซึ่งไม่ว่าง → copy ไปใช้ตรง ๆ แล้ว Google ถูกนับว่าเปิดอยู่ทั้งที่ไม่มี credential จริง · และตั้ง `FRONTEND_OAUTH_CALLBACK_URL=http://localhost:8080/health`
  - เทส: `account/oauth_test.go` (12 เทส) · `oauth/oauth_test.go` (16 เทส)

AC-25: แดชบอร์ดสำหรับ Admin
  - `GET /admin/dashboard` → `{branches: [...], recent_activities: [...]}`
  - `branches` มีสถิติรายสาขาเรียงตาม `created_at`: `branch_id, branch_name, daily_rooms_free, daily_rooms_total, monthly_rooms_free, monthly_rooms_total, pending_review, bookings_total`
    - นับเฉพาะห้องที่ `is_active=true` · "free" = `status='available'` (ไม่ได้ดูใบจองรายวันที่ค้างอยู่)
    - `pending_review` = จำนวนใบ `awaiting_review` · `bookings_total` = ใบจองทุกสถานะ
    - ไม่มีตัวเลขรวมทุกสาขาแยกต่างหาก และรวมสาขาที่ `is_active=false` ด้วย
  - `recent_activities` = activity log ล่าสุดไม่เกิน 10 รายการ ในขอบเขตเดียวกับ AC-13
  - admin → เห็นเฉพาะสาขาตัวเอง · super admin → เห็นทุกสาขา และส่ง `?branch_id=` เพื่อเจาะสาขาเดียวไม่ได้ (endpoint นี้ไม่อ่าน query นั้น ขัดกับที่ `backend/README.md` บอกไว้)
  - เทส: module `reporting` ยังไม่มีไฟล์ `_test.go` เลย

AC-26: จัดการรายละเอียดสาขา (ทุก endpoint ส่งสาขาผ่าน `?branch_id=` ไม่ใช่ใน body)
  - `PUT /admin/branch?branch_id=` → 200 คืนสาขาพร้อม images/amenities/nearby · แทนที่ทั้งชุด: `name, tagline, description, address, phones, line_id, email, latitude, longitude, map_url, building_count, floor_count, daily_price_from, monthly_price_min, monthly_price_max, water_rate, electric_rate, deposit, advance_payment, contract_fee, cover_image_url`
    - `name` บังคับและยาวไม่เกิน 200 · `building_count > 0` · `floor_count > 0` · `water_rate`/`electric_rate >= 0` · `monthly_price_min <= monthly_price_max` (ตรวจเมื่อส่งมาทั้งคู่ ผิดแล้วลง `fields.monthly_price_max`) · `cover_image_url` ต้องเป็น http(s) → ผิดข้อใด → 422
    - ไม่ส่ง `latitude`/`longitude`/ราคา → ค่าเดิมกลายเป็น `NULL` · ไม่ส่ง `cover_image_url` → คงรูปเดิม · ส่ง `phones: []` → เบอร์ถูกล้าง
    - **[ช่องว่าง] [ยังไม่ยืนยัน]** ไม่ส่ง `phones` หรือส่ง `null` → น่าจะได้ 500 เพราะ pgx แปลง slice ที่เป็น nil เป็น `NULL` แล้วชน `NOT NULL` ของ `branches.phones`
    - **[ช่องว่าง]** `contract_fee`, `deposit`, `advance_payment` ติดลบได้ → `contract_fee = -500` ทำให้ใบจองรายเดือนใหม่มี `total_amount = -500` · `email`/`phones`/`map_url` ไม่ถูกตรวจรูปแบบ
    - แก้สาขาที่ `is_active=false` ได้ตามปกติ
  - `PUT /admin/branch/amenities?branch_id= {amenity_ids}` → 200 คืนรายการใหม่ แทนที่ทั้งชุด
    - **[ช่องว่าง]** ไม่ได้อยู่ใน transaction และไม่แปลง error → id ที่ไม่มีจริง → 500 · id ซ้ำ → 409 · ทั้งสองกรณี**ลบของเดิมไปแล้ว**
  - `PUT /admin/branch/nearby?branch_id= {items:[{category, name, distance, sort_order}]}` → 200 แทนที่ทั้งชุด คืนรายการเรียง `category, sort_order` · `name` ว่างในรายการใดก็ได้ → 422 `fields.items` "รายการที่ N ยังไม่ได้กรอกชื่อสถานที่" · `sort_order=0` → ใช้ลำดับใน array แทน · ไม่ส่ง `items` → ลบทั้งหมด
    - **[ช่องว่าง]** ไม่ได้อยู่ใน transaction → INSERT พังกลางทาง → รายการเดิมหายไปบางส่วน
  - admin ส่ง `branch_id` สาขาอื่น → 403 · super admin ไม่ส่ง → 400 · `branch_id` ผิดรูปแบบ → 400
  - สำเร็จ → log `branch.update` (`detail.name`)/`branch.set_amenities` (`detail.count`)/`branch.update_nearby` (`detail.count`) ผูกกับสาขานั้น
  - เทส: `branch/closed_test.go` `TestAdminStillUpdatesClosedBranch` · `storage/asset_test.go` `TestUploadBranchCover`

AC-27: ข้อมูลสาธารณะของสาขาและห้อง (ไม่ต้องล็อกอิน)
  - `GET /branches` → เฉพาะสาขาที่ `is_active=true` เรียงตาม `created_at` แต่ละสาขามี `images`, `amenities`, `nearby_places` (key ที่ว่างถูกตัดทิ้งทั้ง key เพราะ `omitempty`)
  - ราคาช่วงของสาขา (`daily_price_from`, `monthly_price_min`, `monthly_price_max`) และพิกัด ที่เป็น `NULL` → ไม่มี key นั้นใน JSON (frontend ใช้การ "มี key" ตัดสินว่าสาขามีห้องประเภทนั้น ดู AC-33)
  - `GET /superadmin/branches` → ทุกสาขารวมที่ปิดใช้งาน (super admin เท่านั้น)
  - `GET /branches/:branchID` รับได้ทั้ง `brn-001` และ slug (เช่น `prachauthit-45`) → 200 · ค่าที่ไม่ใช่ id ถูกตีความเป็น slug → ไม่เจอ → **404** (ไม่ใช่ 400)
  - `GET /branches/:branchID` ของสาขาที่ `is_active=false` → 404 ทั้งทาง id และ slug · ปิดสาขาหนึ่งไม่กระทบสาขาอื่น
  - `GET /amenities` → รายการกลางทั้งหมด เรียง `sort_order, name`
  - `GET /room-types?branch_id=` → ประเภทห้อง เรียง `sort_order, name` (ไม่ส่ง = ทุกสาขาปนกัน) · `branch_id` ผิดรูปแบบ → 400
  - `GET /rooms/:roomID` → ห้องพร้อม `amenities` และ `images` · ห้องไม่มีจริง, ถูกลบ (`is_active=false`) หรืออยู่ในสาขาที่ปิด → 404
  - **[ช่องว่าง]** `GET /rooms/:roomID` ของห้องที่สถานะ `occupied`/`maintenance` → ยังคืน 200 ตามปกติ (ไม่ได้บอกว่าจองไม่ได้)
  - เทส: `branch/closed_test.go` (id/slug ปกติ, ปิดแล้ว 404 ทั้ง id และ slug, ไม่กระทบสาขาอื่น) · `room/closed_test.go` (ห้องถูกลบ 404, ห้องในสาขาที่ปิด 404) · `room/amenities_test.go` (amenity เรียงถูก, ห้องไม่มี amenity)

AC-28: รายชื่อสมาชิกและประวัติรายคน (Admin + Super Admin)
  - `GET /admin/members?search=&page=&page_size=` → member ที่ยังไม่ถูกลบ ค้นด้วย ILIKE ใน `first_name`, `last_name`, `email`, `phone` เรียงคนที่สมัครล่าสุดก่อน → ทั้ง admin และ super admin เห็นสมาชิกทั้งระบบ
  - `GET /admin/members/:memberID/bookings?page=` → admin เห็นเฉพาะใบในสาขาตัวเอง · super admin เห็นทุกสาขา (ส่ง `branch_id` ไม่ได้ endpoint นี้ไม่อ่าน)
    - `memberID` ผิดรูปแบบ → 400 · ไม่มีจริงหรือไม่ใช่ member → 200 รายการว่าง (ไม่ตรวจว่าผู้ใช้มีจริง)
  - `GET /admin/bookings` ค้น `search` ด้วย ILIKE ใน `code`, `room_number`, `first_name`, `last_name` ของเจ้าของบัญชี (ไม่รวมอีเมลและชื่อผู้เข้าพัก) · `status=all` = ไม่กรอง · `status`/`stay_type` ที่ไม่รู้จัก → 400 · เรียงใหม่สุดก่อน
  - ทุกรายการใบจองมี `branch_name`, `room_number`, `member_name`, `member_email` จากการ join และมี `latest_payment` เฉพาะใบที่ไม่ใช่ `pending_payment`
  - เทส: ยังไม่มีเทส

AC-29: โปรไฟล์ของตัวเอง
  - `GET /me` → ข้อมูลผู้ใช้รวม `must_change_password`, `is_active`, `last_login_at`, `branch_id` และ `branch_name` (ถ้าเป็น admin) โดยไม่มี `password_hash`
  - `PUT /me {first_name, last_name, phone, avatar_url?}` → 200 · `first_name`/`last_name`/`phone` บังคับตามกติกาเดียวกับ AC-1 · ไม่ส่ง `avatar_url` → คงค่าเดิม · log `user.update_profile`
  - **[ช่องว่าง]** `avatar_url` ไม่ถูกตรวจเลย → `javascript:alert(1)` ถูกบันทึกได้ (ช่อง URL อื่นในระบบมีการตรวจ ดู AC-14)
  - `POST /me/password {current_password, new_password, confirm_password}` → 200 `{"data":{"message":"เปลี่ยนรหัสผ่านเรียบร้อยแล้ว กรุณาเข้าสู่ระบบใหม่"}}`, ปลด `must_change_password` และมี log `user.change_password`
    - ไม่ส่ง `current_password` (บัญชีที่มีรหัสแล้ว) → 422 `fields.current_password`
    - รหัสใหม่เหมือนรหัสเดิม (เทียบสตริงที่ส่งมา) → 422 `fields.new_password`
    - `confirm_password` ไม่ตรง → 422 `fields.confirm_password`
    - รหัสใหม่ไม่ผ่านกติกา AC-1 → 422
    - `current_password` ไม่ถูกต้อง → 422 `fields.current_password = "รหัสผ่านเดิมไม่ถูกต้อง"` (ตรวจหลังข้อข้างบนทั้งหมดผ่านแล้วเท่านั้น)
  - เทส: `account/oauth_test.go` `TestOAuthUserCanSetFirstPassword`, `TestChangePasswordStillRequiresCurrentAfterFirstSet`

## Acceptance Criteria — Frontend (`frontend/`)

AC-30: ชั้นเรียก API และที่เก็บ token
  - ทุกคำขอไป backend ผ่าน `src/api/client.jsx` ที่เดียว (`apiClient.get/post/put/patch/delete`) → ไม่มี component ไหนเรียก `fetch` เอง
  - base URL = `REACT_APP_API_BASE_URL` (ตัด `/` ท้าย) ไม่ตั้ง → `http://localhost:8080/api/v1` · `API_ORIGIN` = base URL ที่ตัด `/api/v1` ออก ใช้ประกอบ URL ของ `/uploads/*`
  - มี access token ใน `localStorage` → แนบ `Authorization: Bearer <token>` ทุกคำขอ รวม endpoint สาธารณะ
  - query param ที่เป็น `undefined`/`null`/`""` → ไม่ถูกส่ง
  - ตอบ 204 → คืน `null` · ตอบ 2xx → คืน JSON ทั้งก้อน (ผู้เรียกต้อง unwrap `.data` เอง)
  - ตอบไม่ใช่ 2xx → throw `ApiError {code, message, fields, status}` ใช้ `error.message` จาก backend ก่อน ถ้าไม่มีใช้ตารางข้อความสำรองตาม `code` (ตารางมี 12 code ขาด `too_many_requests`) ถ้าไม่รู้จักอีกใช้ "เกิดข้อผิดพลาด กรุณาลองใหม่"
  - body ที่ไม่ใช่ JSON (เช่น HTML 502 จาก proxy) → `ApiError` code `internal_error` ข้อความ "เกิดข้อผิดพลาดที่เซิร์ฟเวอร์ กรุณาลองใหม่"
  - ยิงไม่ถึง backend (เน็ตหลุด/CORS/backend ปิด) → `NetworkError` ข้อความ "เชื่อมต่อเซิร์ฟเวอร์ไม่ได้ กรุณาตรวจสอบอินเทอร์เน็ตแล้วลองใหม่" · ยกเลิกด้วย `AbortController` → ส่ง `AbortError` ต่อไปเงียบ ๆ
  - token เก็บใน `localStorage` คีย์ `wisetsuk_access_token`, `wisetsuk_refresh_token` และ `wisetsuk_user` (JSON ของ `user` จาก token pair) → login/logout ยิง event `wisetsuk_auth_changed` ในแท็บเดียวกัน และฟัง event `storage` เพื่อ sync ข้ามแท็บ
  - `wisetsuk_user` เสีย (JSON พัง) → ถือว่าไม่มี user แทนที่ทั้งแอปจะพัง
  - "ล็อกอินอยู่" ตัดสินจากการมี access token ใน `localStorage` อย่างเดียว ไม่ได้ดูวันหมดอายุ
  - **[ช่องว่าง]** ไม่มีการต่ออายุ token → หลัง access token หมดอายุ (30 นาที) ทุกหน้าที่ต้องล็อกอินได้ 401 แสดงเป็น error state ขณะที่ header ยังขึ้นว่าล็อกอินอยู่ ต้องกดออกจากระบบเองแล้วเข้าใหม่ (`refresh_token` ถูกเก็บไว้แต่ใช้แค่ตอน logout)
  - **[ช่องว่าง]** role ที่ใช้พาไปหน้าต่าง ๆ มาจาก `wisetsuk_user` ที่เก็บตอนล็อกอิน → super admin ย้ายสาขาหรือเปลี่ยนชื่อ frontend ยังแสดงค่าเดิมจนกว่าจะล็อกอินใหม่
  - hook `useAsync` เป็นที่เดียวที่จัดการ loading/error/data และ `refetch` → เปลี่ยน dependency หรือ unmount ระหว่างโหลด → คำขอเดิมถูก abort
  - เทส: ยังไม่มีเทส

AC-31: เส้นทางหน้าเว็บและ layout
  - โซน Guest ใช้ `GuestLayout` (header เมนู 4 ลิงก์ "หน้าหลัก / สาขาของเรา / ค้นหาห้องพัก / ติดต่อเรา" + footer):
    | path | หน้า |
    |---|---|
    | `/` | หน้าแรก (AC-32) |
    | `/branches` | สาขาในเครือทั้งหมด (AC-33) |
    | `/branches/:branchId` | รายละเอียดสาขา แท็บรูปภาพ (AC-33) |
    | `/branches/:branchId/map` | รายละเอียดสาขา แท็บแผนที่ (AC-33) |
    | `/rooms` | ค้นหาห้องพัก (AC-34) |
    | `/rooms/:roomId` | placeholder "รายละเอียดห้องพัก" |
    | `/contact` | ติดต่อเรา (AC-35) |
    | `/login`, `/register` | เข้าสู่ระบบ / สมัครสมาชิก (AC-36) |
    | `*` | placeholder "ไม่พบหน้าที่ต้องการ" |
  - `/admin/login` ใช้ `AdminAuthLayout` (header โลโก้อย่างเดียว ไม่มี footer) (AC-37)
  - `/admin` ต้องเป็น role `admin` → placeholder "ระบบผู้ดูแลระบบ" ใน `AdminAuthLayout` ที่มีปุ่มออกจากระบบ
  - `/superadmin`, `/superadmin/staff`, `/superadmin/branches`, `/superadmin/activity-logs` ต้องเป็น role `superadmin` ใช้ `AdminShell` (sidebar) → สองหน้าหลังเป็น placeholder "จัดการสาขา" และ "Activity log"
  - placeholder ทุกหน้าแสดงหัวข้อ + "หน้านี้กำลังอยู่ระหว่างพัฒนา" + ปุ่ม "กลับหน้าหลัก"
  - **[ช่องว่าง]** ไม่มี route `/auth/callback` ที่ backend redirect มาหลัง OAuth (AC-24) และไม่มี route `/bookings/:id` ที่ `link` ของแจ้งเตือนชี้ไป (AC-5) → ทั้งสองตกไปหน้า "ไม่พบหน้าที่ต้องการ"
  - เทส: `src/App.test.js` (2 เทส: หน้าแรกมีหัวข้อหลัก, เมนูนำทางครบ 4 รายการ)

AC-32: หน้าแรก (`/`)
  - ดึง `GET /branches` ครั้งเดียวแล้วแบ่งให้ทั้ง section "สาขาของเรา" และ "ประเภทห้องพัก" · กรองเฉพาะ `is_active=true` ซ้ำอีกชั้นฝั่ง client
  - Hero: หัวข้อ "หอพักวิเศษสุขนคร คอนโด" + ปุ่ม "ค้นหาห้องพัก" (`/rooms`) และ "ดูสาขาทั้งหมด" (`/branches`) · carousel สุ่มรูป 15 รูปจากรายการไฟล์ใน `src/assets/branchPhotos.jsx` ครั้งเดียวตอนเข้าหน้า เลื่อนอัตโนมัติทุก 5 วินาที วนไม่สิ้นสุด · กดเลื่อนเอง → นับเวลาใหม่
  - แถบสิ่งอำนวยความสะดวกเป็นข้อความคงที่ 4 ช่อง "คีย์การ์ด / รปภ. 24 ชั่วโมง / ที่จอดรถ / ห้องน้ำในตัว" ไม่ได้มาจาก API
  - "สาขาของเรา": การ์ดละสาขา (รูปปก, ชื่อ, ที่อยู่) → กดแล้วไป `/branches/<slug>`
  - รูปปกสาขา = `cover_image_url` ถ้ามี ไม่มี → รูปแรกของรายการไฟล์ในข้อความ AC-33 · ไม่มีทั้งคู่หรือโหลดรูปไม่ขึ้น → ไอคอนอาคารแทน
  - "ประเภทห้องพัก": การ์ดรายวันแสดงเมื่อมีสาขาที่ `daily_price_from > 0` → "ราคาเริ่มต้น <ค่าน้อยสุด> บาท/คืน" กดแล้วไป `/rooms?stay_type=daily` · การ์ดรายเดือนแสดงเมื่อมีสาขาที่ `monthly_price_min > 0` → "ราคา <min ของ min> - <max ของ max> บาท/เดือน" กดแล้วไป `/rooms?stay_type=monthly` · ไม่มีทั้งคู่ → "ยังไม่มีข้อมูลราคาห้องพัก"
  - CTA ท้ายหน้า: ยังไม่ล็อกอิน → ปุ่ม "เข้าสู่ระบบ" + "สมัครสมาชิก" · ล็อกอินแล้ว → ปุ่ม "ค้นหาห้องพัก"
  - ราคาแสดงแบบคั่นหลักพันไม่มีทศนิยม (`th-TH`) · ค่าที่ไม่มี → "-"
  - ทุก section ที่ดึง API มี 4 สถานะ: loading (skeleton), error (ข้อความ + ปุ่มลองใหม่), empty, success
  - เทส: `src/App.test.js` (หัวข้อหลัก)

AC-33: หน้ารวมสาขาและรายละเอียดสาขา
  - `/branches` แสดงสาขาที่เปิดอยู่เป็นแถว: รูปปก, ชื่อ, ที่อยู่, เบอร์โทร (คั่นจุลภาค), LINE, chip สูงสุด 3 อัน, tagline, ข้อความสถานะห้อง, ปุ่ม "ดูรายละเอียด →" ไป `/branches/<slug>`
    - chip แรก = "<building_count> อาคาร <floor_count> ชั้น" (ถ้ามีทั้งคู่) ตามด้วย amenity ของสาขาตามลำดับ `convenience-store, parking, elevator, furniture, laundry, wifi, aircon`
    - **[ช่องว่าง]** chip "มีลิฟต์" ผูกกับ code `elevator` แต่ seed ใช้ code `lift` → ไม่มีสาขาไหนได้ chip นี้
    - ข้อความสถานะห้องตัดสินจากการ**มี key** ใน JSON ไม่ใช่ค่ามากกว่า 0: มี `daily_price_from` และ `monthly_price_min` → "มีห้องพักทั้งรายวัน และรายเดือน" · มีแค่รายเดือน → "มีห้องพักรายเดือนพร้อมเข้าพัก" · มีแค่รายวัน → "มีห้องพักรายวันพร้อมเข้าพัก" · ไม่มีทั้งคู่ → "ติดต่อสอบถามห้องว่าง"
  - `/branches/:branchId` รับทั้ง slug และ `brn-NNN` → ดึง `GET /branches/:branchId`
    - error `bad_request` → "รหัสสาขาไม่ถูกต้อง" ไม่มีปุ่มลองใหม่ · `not_found` → "ไม่พบสาขาที่ต้องการ" (รวมสาขาที่ปิด) · อื่น ๆ → ข้อความจาก backend
    - แท็บ "รูปภาพ"/"แผนที่" เปลี่ยน URL ระหว่าง `/branches/:id` กับ `/branches/:id/map` → เปิด URL `/map` ตรง ๆ → แท็บแผนที่ active ทันที
    - **[ช่องว่าง]** แกลเลอรีไม่ได้ใช้ `images` จาก API แต่ใช้รายการไฟล์ที่ hardcode ใน `src/assets/branchPhotos.jsx` ตาม slug (บางแค 45 รูปเริ่มที่ `bk-23`, เจริญกรุงเพลส 44 รูปเริ่มที่ `ckp-41`, ประชาอุทิศ 45 15 รูป) → รูปที่แอดมินอัปผ่าน API ไม่โผล่ และสาขาใหม่ที่ slug ไม่อยู่ในรายการ → "ยังไม่มีรูปภาพสำหรับสาขานี้"
    - แผนที่ iframe เลือกตามลำดับ: `map_url` จาก API → พิกัดที่ hardcode ใน `src/utils/mapEmbed.jsx` ตาม slug (3 สาขา) → ค้นจากข้อความที่อยู่ · ลิงก์ "เปิดหน้าสถานที่เต็มใน Maps" ใช้ `map_url` → place URL ที่ hardcode → URL เดียวกับ iframe · `map_url` ที่ไม่มี `/maps/embed` หรือ `output=embed` → แสดงเป็นปุ่ม "เปิดใน Google Maps" แทน iframe
    - **[ช่องว่าง]** `latitude`/`longitude` ของสาขาจาก API ไม่ถูกใช้เลย (ใช้พิกัด hardcode แทน)
    - การ์ดราคา 2 ใบ (รายวัน/รายเดือน) ใช้กติกา "มี key" เดียวกับหน้ารวม → ไม่มีประเภทนั้น → "ไม่มีห้องพักรายวัน"/"ไม่มีห้องพักรายเดือน" · ไม่มีทั้งคู่ → ไม่แสดงการ์ด
    - การ์ด "รายละเอียดเพิ่มเติม": เงินประกัน (`deposit` > 0 → "<deposit> บาท" · ไม่งั้น `advance_payment` > 0 → "จ่ายล่วงหน้า <advance_payment> บาท" · ไม่งั้น → ข้อความคงที่ "จ่ายล่วงหน้า 3 เดือนของราคาห้องพัก"), ค่าน้ำ, ค่าไฟ (บาท/ยูนิต), ที่จอดรถ (มี amenity code `parking` → "มีที่จอดรถ" ไม่งั้น "ไม่มีที่จอดรถ"), "<building_count> ตึก <floor_count> ชั้น/ตึก"
    - **[ช่องว่าง]** seed ทุกสาขามี `deposit = advance_payment = 0` → เงินประกันทุกสาขาแสดงข้อความคงที่ "จ่ายล่วงหน้า 3 เดือนของราคาห้องพัก" ซึ่งไม่ได้มาจากข้อมูลจริง
    - ปุ่ม "จองห้องพัก" → `/rooms?branch_id=brn-NNN` · ปุ่มที่สอง: มี `line_id` → "แอดไลน์ <line_id>" เปิด `https://line.me/R/ti/p/%40<id>` (เติม `@` ให้ถ้าไม่มี) · ไม่มี LINE แต่มีเบอร์ → `tel:<เบอร์แรก>` · ไม่มีทั้งคู่ → `/contact`
    - การ์ดสิ่งอำนวยความสะดวกใช้ `amenities` ของสาขา ไอคอนเลือกจาก field `icon` (รู้จัก 12 ค่า ที่เหลือใช้ไอคอนสำรอง) · ว่าง → "ยังไม่มีข้อมูลสิ่งอำนวยความสะดวกสำหรับสาขานี้"
    - สถานที่ใกล้เคียงจัดกลุ่มตาม `category` (`education`→สถานศึกษา, `hospital`→โรงพยาบาล, `shopping`→ห้างสรรพสินค้าและแหล่งช้อปปิ้ง, `park`→สวนสาธารณะ, ค่าอื่นแสดงตามตัว) เรียงในกลุ่มด้วย `sort_order`
    - "สาขาอื่นๆ ของเรา" แสดงสาขาที่เปิดอยู่อื่นไม่เกิน 2 สาขา (ไม่แสดง section ถ้าไม่มี)
  - เทส: ยังไม่มีเทส

AC-34: ค้นหาห้องพัก (`/rooms`)
  - ตัวกรอง 3 ตัว: สาขา (จาก `GET /branches` + "ทุกสาขา"), ประเภท ("ทุกประเภท / รายวัน / รายเดือน" คือ `stay_type` ไม่ใช่ `room_type_id`), วันที่
  - ค่าตัวกรองอ่านเริ่มต้นจาก query `branch_id`, `stay_type`, `date_from`, `date_to`, `page` และเขียนกลับลง URL ทุกครั้งที่เปลี่ยน (แบบ replace) · เปลี่ยนตัวกรองใด ๆ → กลับไปหน้า 1 · เปลี่ยนประเภท → ล้างวันที่
  - ตัวกรองวันที่: ยังไม่เลือกประเภท → ปิดใช้งาน ป้าย "เลือกประเภทก่อน" · รายวัน → เลือกช่วง 2 คลิก (คลิกที่สองก่อนวันเริ่ม = เริ่มช่วงใหม่) · รายเดือน → เลือกวันเดียว · ปฏิทินเริ่มวันจันทร์ ป้ายแสดง `วว/ดด/ปป` เป็น พ.ศ. 2 หลัก · มีปุ่ม "ล้างวันที่"
  - แปลงเป็น query ของ `GET /rooms/search`: `branch_id`, `stay_type`, `page` · รายวัน → `check_in`/`check_out` · รายเดือน → `move_in_date` · ไม่ส่ง `page_size` (ได้ 20 ต่อหน้า) และไม่ส่ง `room_type_id`/`min_price`/`max_price`
  - **[ช่องว่าง]** เลือกวันออกเป็นวันเดียวกับวันเข้าได้ → ส่ง `check_in == check_out` → backend 422 → หน้าแสดง error "ข้อมูลที่กรอกไม่ถูกต้อง" แทนการกันไว้ที่ปฏิทิน · เลือกวันในอดีตได้ด้วย
  - error `bad_request` (เช่น `?stay_type=abc` ใน URL) → "เงื่อนไขค้นหาไม่ถูกต้อง" · ไม่พบห้อง → "ไม่พบห้องพักที่ตรงกับเงื่อนไขที่เลือก ลองปรับตัวกรองแล้วค้นหาใหม่อีกครั้ง"
  - การ์ดห้อง: หัวข้อ "ห้องพักรายวัน"/"ห้องพักรายเดือน", ชื่อสาขา, chip, ราคา "฿<price>/วัน" หรือ "/เดือน", "ค่าน้ำ X บาท/ยูนิต, ค่าไฟ Y บาท/ยูนิต", ปุ่ม "จองเลย"
    - รูปการ์ด = `image_url` ของห้อง ไม่มี → ไอคอนอาคาร (ไม่ตกไปใช้รูปสาขา)
    - chip = `room_type_name`, "<size_sqm> ตร.ม." แล้วตามด้วย **amenity ทั้งหมดของสาขา**
    - **[ช่องว่าง]** backend ส่ง `amenities` ของห้องมาแล้ว (AC-12) แต่ frontend ยังใช้ amenity ของสาขา → ห้องที่ไม่มีแอร์ก็ขึ้น chip "เครื่องปรับอากาศ" ถ้าสาขามี
    - เลข/อาคาร/ชั้นของห้องไม่ถูกแสดง
  - กด "จองเลย": ยังไม่ล็อกอิน → modal "โปรดลงทะเบียน / เข้าสู่ระบบ เพื่อทำการจอง" พร้อมปุ่มไป `/login` และ `/register` (ปิดด้วยคลิกนอกกล่อง/Escape/ปุ่ม X) · ล็อกอินแล้ว (ทุก role) → ไป `/rooms/:roomId` ซึ่งเป็น placeholder
  - แบ่งหน้าด้วยปุ่มก่อนหน้า/ถัดไป + "หน้า X จาก Y" จาก `meta` · เปลี่ยนหน้า → เลื่อนขึ้นบนสุด
  - เทส: ยังไม่มีเทส

AC-35: ติดต่อเราและ footer
  - `/contact` แสดงข้อความคงที่ "พร้อมให้บริการทุกวัน 8:30 - 17:30 น." (ไม่มี field ใน API) และการ์ดละสาขาที่เปิดอยู่: ชื่อ, ที่อยู่, เบอร์, LINE, `map_url` (ถ้ามี) และแผนที่ตามกติกาเดียวกับ AC-33
  - footer ทุกหน้า Guest: เมนู 4 ลิงก์, รายชื่อสาขาที่เปิดอยู่ (ลิงก์ไป `/branches/<slug>`), LINE ของสาขาแรกเท่านั้น, เบอร์โทรของทุกสาขาพร้อมชื่อสาขา, "© <ปี ค.ศ. ปัจจุบัน> วิเศษสุขนคร คอนโด และหอพักในเครือ สงวนลิขสิทธิ์"
  - เทส: ยังไม่มีเทส

AC-36: สมัครสมาชิกและเข้าสู่ระบบ (Member)
  - ล็อกอินอยู่แล้วเปิด `/login` หรือ `/register` → redirect ไป `/` ทันที (ทุก role)
  - `/login` ตรวจฝั่ง client ก่อนยิง: อีเมลว่าง → "กรุณากรอกอีเมล" · ไม่ตรง `^[^\s@]+@[^\s@]+\.[^\s@]+$` → "รูปแบบอีเมลไม่ถูกต้อง" · รหัสว่าง → "กรุณากรอกรหัสผ่าน"
  - `/login` สำเร็จ → เก็บ token + user แล้วพาไปหน้าแรกของ role: member → `/`, admin → `/admin`, superadmin → `/superadmin`
  - `/login` ไม่สำเร็จ → แสดง `error.message` ของ backend เป็นแถบบนฟอร์ม (เช่น "อีเมลหรือรหัสผ่านไม่ถูกต้อง", ข้อความ 429 ของ AC-2)
  - checkbox "จดจำฉัน" ไม่มีผลใด ๆ (token อยู่ใน `localStorage` เสมอ)
  - ลิงก์ "ลืมรหัสผ่าน?" → `/contact` (ดู AC-23)
  - `/register` ตรวจฝั่ง client: ชื่อ/นามสกุลว่าง, เบอร์ต้องเป็นตัวเลข 9–10 หลักล้วน, อีเมลตาม regex เดียวกับ login, รหัสผ่านยาว 8–72 ตัวอักษร, ยืนยันรหัสผ่านต้องตรง
  - **[ช่องว่าง]** กติกาเบอร์ฝั่ง client (`^[0-9]{9,10}$`) เข้มกว่า backend (`^[0-9\-\s()+]{8,20}$`) → "02-872-7800" ถูกปฏิเสธที่หน้าเว็บทั้งที่ backend รับ · ส่วนรหัสผ่านฝั่ง client ไม่ตรวจ "ต้องมีทั้งตัวอักษรและตัวเลข" → ไปโดน 422 จาก backend แทน
  - `/register` ส่งแค่ `first_name, last_name, phone, email, password` (trim ทุกช่องยกเว้นรหัสผ่าน ไม่ส่ง `confirm_password`) → สำเร็จ → เก็บ token แล้วไป `/` (ไม่แยกตาม role) · ได้ 422 → `fields` ไปแสดงใต้ช่องที่ชื่อตรงกัน · error อื่น → แถบบนฟอร์ม
  - header เปลี่ยนจากปุ่ม "เข้าสู่ระบบ"/"สมัครสมาชิก" เป็นปุ่ม "ออกจากระบบ" ทันทีที่ล็อกอิน (ไม่ต้อง reload) · กดออกจากระบบ → ยิง `POST /auth/logout` แบบไม่รอผล ล้าง token แล้วไป `/`
  - เทส: ยังไม่มีเทส

AC-37: เข้าสู่ระบบผู้ดูแลและการกันหน้า
  - `/admin/login` ฟอร์มเดียวกับ `/login` แต่ไม่มีลิงก์สมัครสมาชิก และท้ายการ์ดเขียน "ลืมรหัสผ่าน? กรุณาติดต่อหัวหน้าผู้ดูแลระบบเพื่อรีเซ็ตรหัสผ่าน"
  - ล็อกอินอยู่แล้วเปิด `/admin/login` → redirect ไปหน้าแรกของ role ที่ล็อกอินอยู่
  - บัญชี member ล็อกอินที่ `/admin/login` → ไม่เก็บ token แสดง "บัญชีนี้ไม่มีสิทธิ์เข้าใช้ระบบผู้ดูแล กรุณาเข้าสู่ระบบที่หน้าสำหรับสมาชิก"
    - **[ช่องว่าง]** backend ออก refresh token และบันทึก `auth.login` ให้ member คนนั้นไปแล้ว frontend ทิ้ง token โดยไม่ยิง logout → refresh token ค้างใช้ได้ใน DB 30 วัน
  - admin/superadmin สำเร็จ → ไปหน้าที่ถูกกันไว้ก่อนหน้า (`state.from`) ถ้ามี ไม่งั้นหน้าแรกของ role
  - การกันหน้า (`RequireRole`) อ่านจาก `localStorage` จึงรอดจากการ reload:
    - ไม่มี token → ไป `/admin/login` พร้อมจำ path เดิม
    - มี token แต่ role ไม่ตรง → ไปหน้าแรกของ role ตัวเอง (member เปิด `/superadmin` → `/` · admin เปิด `/superadmin` → `/admin` · superadmin เปิด `/admin` → `/superadmin`)
    - เป็นแค่ UX — สิทธิ์จริงถูกตรวจที่ backend ทุก endpoint (AC-4, AC-5)
  - ออกจากระบบจาก `AuthHeader` (หน้า `/admin`) หรือ sidebar ของ `AdminShell` → ยิง logout แบบไม่รอผล ล้าง token แล้วไป `/admin/login`
  - **[ช่องว่าง]** `user.must_change_password=true` (บัญชีที่ super admin เพิ่งสร้างหรือรีเซ็ต) → frontend ไม่บังคับเปลี่ยนรหัส และยังไม่มีหน้าเปลี่ยนรหัสให้ใช้
  - เทส: ยังไม่มีเทส

AC-38: แดชบอร์ดหัวหน้าผู้ดูแลระบบ (`/superadmin`)
  - ข้อมูลทั้งหน้ามาจาก `GET /admin/dashboard` ครั้งเดียว (AC-25)
  - หัวหน้า "แดชบอร์ด" + "สรุปข้อมูลประจำวันที่ <วันที่ปัจจุบันของเบราว์เซอร์แบบไทย พ.ศ.>"
  - การ์ดละสาขา 3 ช่อง: "ห้องว่างรายวันวันนี้" `daily_rooms_free / daily_rooms_total ห้อง` · "ห้องว่างรายเดือนวันนี้" `monthly_rooms_free / monthly_rooms_total ห้อง` · "รอตรวจสอบการจอง" `pending_review / bookings_total รายการ`
  - กล่อง "กิจกรรมล่าสุด" บรรทัดละรายการ: "<ป้าย role> <branch_name> <ป้าย action>" + เวลา `YYYY-MM-DD HH:mm` ตามเขตเวลาของเบราว์เซอร์ · ป้าย role: member→สมาชิก, admin→Admin, superadmin→Superadmin · ไม่มีกิจกรรม → "ยังไม่มีกิจกรรมในระบบ"
  - **[ช่องว่าง]** ตารางป้าย action มี 27 ค่า ขาด `room.add_image` และ `room.delete_image` → สองค่านี้แสดงเป็นรหัสภาษาอังกฤษดิบ
  - sidebar: ชื่อ-นามสกุลผู้ล็อกอิน + ป้าย role, เมนู "แดชบอร์ด / จัดการผู้ดูแลระบบ / จัดการสาขา / activity log", ปุ่ม Logout
  - เทส: ยังไม่มีเทส

AC-39: จัดการผู้ดูแลระบบ (`/superadmin/staff`)
  - ดึง `GET /superadmin/staff` และ `GET /superadmin/branches` พร้อมกัน → ตาราง: ผู้ดูแล (ชื่อ + อีเมล), บทบาท (ป้าย "Superadmin" หรือ "Admin"), สิทธิ์ดูแลสาขา (superadmin → "ทุกสาขา" · admin → ชื่อสาขาที่แปลงจาก `branch_id` ด้วยรายการสาขา ไม่พบ → "—"), ปุ่ม "แก้ไข" และไอคอนลบ
  - ไอคอนลบไม่แสดงในแถว superadmin และแถวของตัวเอง
  - modal "เพิ่มผู้ดูแลใหม่"/"แก้ไขผู้ดูแล" (ฟอร์มเดียวกัน):
    - ตรวจฝั่ง client: ชื่อ/นามสกุลว่าง · เบอร์ถ้ากรอกต้องตรง `^[0-9\-\s()+]{8,20}$` (ตรงกับ backend) · อีเมลบังคับเฉพาะตอนเพิ่ม · ต้องเลือกสาขาเมื่อเป้าหมายไม่ใช่ superadmin
    - สาขาเลือกได้ทีละ 1 (radio) · สาขาที่มี admin คนอื่นแล้วถูกปิดพร้อมบอกชื่อคนดูแล (สาขาของคนที่กำลังแก้ยังเลือกได้) · ทุกสาขามีคนดูแลครบ → ปุ่มบันทึกถูกปิดพร้อมข้อความ "ทุกสาขามีผู้ดูแลครบแล้ว ต้องย้ายหรือลบผู้ดูแลเดิมออกก่อนจึงจะเพิ่มคนใหม่ได้"
    - เป้าหมายเป็น superadmin → ซ่อนตัวเลือกสาขาและไม่ส่ง `branch_id`
    - ตอนแก้ไข ช่องอีเมลถูกปิด
    - เพิ่ม → `POST /superadmin/staff {email, first_name, last_name, phone, branch_id}` · แก้ → `PUT /superadmin/staff/:id {first_name, last_name, phone, branch_id?}`
    - ได้ 422 → `fields` แสดงใต้ช่อง (เช่น อีเมลซ้ำ, สาขามีผู้ดูแลแล้ว) · error อื่น → แถบบน modal · สำเร็จ → ปิด modal แล้วโหลดตารางใหม่
    - ปิด modal ด้วยคลิกนอกกล่อง/Escape/ปุ่ม X
  - **[ช่องว่าง]** หน้าแก้ไขไม่มีช่องระงับบัญชี (`is_active`) และตั้งรหัสผ่านใหม่ (`password`) → ทำสองอย่างนี้จาก UI ไม่ได้ทั้งที่ API รองรับ (AC-18) · ตารางไม่แสดงว่าใครถูกระงับอยู่
  - modal ลบ: "ต้องการลบบัญชีของ <ชื่อ> (<อีเมล>) ใช่หรือไม่?" → `DELETE /superadmin/staff/:id` → สำเร็จ → โหลดตารางใหม่ · error → แสดงในกล่อง
  - **[ช่องว่าง]** ข้อความยืนยันเขียนว่า "บัญชีจะถูกลบออกจากระบบถาวรและกู้คืนไม่ได้" และคอมเมนต์ในโค้ดอ้างว่า backend ลบแถวถาวร แต่ backend ปัจจุบันเป็น soft delete (AC-18)
  - เทส: ยังไม่มีเทส

## ข้อมูลตั้งต้น (`cmd/seed`)
- รันซ้ำได้ ทั้งหมดอยู่ใน transaction เดียว และรัน migration ก่อนเสมอ → ได้ 3 สาขา, สิ่งอำนวยความสะดวก 24 รายการ, ประเภทห้อง 3 แบบต่อสาขา ("ห้องแอร์", "ห้องพัดลม", "ห้องเปล่า") และห้องตัวอย่าง 9 ห้อง
- สาขา (upsert ด้วย `slug`):
  | slug | ชื่อ | อาคาร×ชั้น | รายวันเริ่ม | รายเดือน | น้ำ/ไฟ | ค่าทำสัญญา |
  |---|---|---|---|---|---|---|
  | `prachauthit-45` | วิเศษสุขนครคอนโด ประชาอุทิศ 45 | 2×5 | 500 | 2,200–3,400 | 17/7 | 500 |
  | `bangkhae` | วิเศษสุขนครคอนโด บางแค | 1×8 | NULL (ไม่รับรายวัน) | 3,500–4,500 | 18/5.50 | 500 |
  | `charoenkrung-place` | เจริญกรุงเพลส | 1×8 | 950 | 7,700–10,000 | 18/5 | 500 |
  - ทุกสาขา `line_id = "@wisetsuk"` มีเบอร์ 2 เบอร์ และ `deposit = advance_payment = 0`, ไม่มีพิกัด, ไม่มี `map_url`, ไม่มี `email`
  - รัน seed ซ้ำ → ทับเฉพาะ field ที่ seed ตั้ง (`name, tagline, description, address, phones, line_id, building_count, floor_count, ราคา, น้ำ/ไฟ, contract_fee`) ส่วน field อื่นที่แอดมินแก้ไว้คงอยู่
- สิ่งอำนวยความสะดวกของสาขา, ของห้องตัวอย่างทั้ง 9 ห้อง และสถานที่ใกล้เคียง (ประชาอุทิศ 12, บางแค 15, เจริญกรุงเพลส 19 รายการ) ถูก**ลบแล้วใส่ใหม่ทั้งชุด**ทุกครั้งที่รัน seed → ที่แอดมินแก้ผ่าน API หายหมด
- ห้องตัวอย่าง (upsert ด้วย `branch_id, room_number` · ไม่แตะ `status`):
  - ประชาอุทิศ 45: `201` รายวัน 500 · `302` รายเดือน 3,400 · `409` รายเดือน 2,900 · `511` รายเดือน 2,200
  - บางแค: `301`/`302`/`303` รายเดือน 3,500/4,000/4,500 ไม่ระบุประเภทห้อง
  - เจริญกรุงเพลส: `201` รายวัน 950 · `301` รายเดือน 7,700 ไม่ระบุขนาด
  - ห้อง `210`, `502` ของประชาอุทิศ 45 (ชุดเก่า) ถูกลบถ้ายังไม่มีใบจองอ้างถึง
- บัญชีที่ได้ (รหัสผ่านทุกบัญชี = `--password` หรือ `SEED_DEFAULT_PASSWORD` หรือ `"Wisetsuk!2026"` · `ON CONFLICT (email) DO NOTHING` จึงไม่ทับรหัสที่เปลี่ยนแล้ว):
  - superadmin: `super@wisetsuk.com` (สมชาย วิเศษสุข)
  - admin: `adminpracha@wisetsuk.com` (ประชาอุทิศ 45) · `adminbangkae@wisetsuk.com` (บางแค) · `admincharoenkrung@wisetsuk.com` (เจริญกรุงเพลส) — เบอร์สมมุติ `080-000-000N`
  - member: `member.1@wisetsuk.com`, `member.2@wisetsuk.com`, `member.3@wisetsuk.com` — **ไม่มีเบอร์โทร** จึงต้องกรอกเบอร์ก่อน `PUT /me` ได้ (AC-29)
  - บัญชีจาก seed ไม่ถูกปักธง `must_change_password`
- **[ช่องว่าง]** DB ที่เคย seed ด้วยอีเมล admin ชุดเก่า (`admin.1@` / `admin.2@` / `admin.3@`) หรือมี superadmin อีเมลอื่นอยู่แล้ว → INSERT บัญชีใหม่ชน `uq_admin_per_branch`/`uq_single_superadmin` (ซึ่ง `ON CONFLICT (email)` ไม่ครอบ) → seed ล้มทั้ง transaction **[ยังไม่ยืนยัน]**
- รูปเดโม: seed อ่านไฟล์ใน `uploads/<slug>/` (flag `--uploads` ค่าเริ่มต้นคือ `UPLOAD_DIR`) แล้วเขียนแค่ URL `{PUBLIC_BASE_URL}/uploads/<slug>/<path>` ลงฐานข้อมูล ไม่ได้เก็บเป็น blob ใน `assets` · รูปเสิร์ฟผ่าน route `/uploads/*`
  - แกลเลอรีสาขา = ไฟล์รูปทุกไฟล์ใต้ `uploads/<slug>/` (`.jpg/.jpeg/.png/.webp` ไม่สนตัวพิมพ์) เรียงตาม path ซึ่งกลายเป็น `sort_order` 0, 1, 2… · `caption.txt` (UTF-8 ตัด BOM ให้) ในโฟลเดอร์คือคำบรรยายของทุกรูปในโฟลเดอร์นั้น · `caption.txt` ไม่ใช่ UTF-8 → seed ล้ม
  - ชื่อไฟล์ถูก escape ทีละท่อน → `m-45-3 (1).jpg` กลายเป็น `m-45-3%20%281%29.jpg`
  - รัน seed ซ้ำ → แถวใน `branch_images` ที่ URL ขึ้นต้นด้วย `{PUBLIC_BASE_URL}/uploads/<slug>/` ถูกแทนด้วยไฟล์ชุดปัจจุบัน (ไฟล์ที่ลบออกจาก repo หายจากแกลเลอรีด้วย) ส่วนรูปที่อัปผ่าน API (`/files/ast-…`) อยู่ครบ
  - รูปปกห้องมาจากฟิลด์ `cover` ใน `roomSeed` → ถูกตั้งค่าเฉพาะเมื่อ `rooms.image_url` ว่าง หรือเป็น URL เดโมอยู่แล้ว รูปปกที่แอดมินอัปเองไม่ถูกทับ · ไฟล์ที่ `cover` อ้างถึงไม่มีอยู่ → seed ล้มทั้ง transaction
  - ไม่มีโฟลเดอร์ `uploads` → ข้ามส่วนรูปทั้งหมดพร้อมพิมพ์เตือน ข้อมูลส่วนอื่นยังลงครบ · มีโฟลเดอร์แต่ไม่มีโฟลเดอร์ของสาขา → สาขานั้นไม่มีรูป ไม่ใช่ error
  - รูปปกสาขา (`cover_image_url`) ไม่ถูก seed ตั้งให้ → หน้าเว็บตกไปใช้รูปแรกของรายการไฟล์ใน frontend (AC-32)
  - เปลี่ยน `PUBLIC_BASE_URL` → ต้องรัน seed ใหม่ เพราะ URL ถูกเก็บแบบเต็ม
- **[ช่องว่าง]** `backend/README.md` ยังเขียนอีเมล admin เป็น `admin.1@` / `admin.2@` / `admin.3@` ไม่ตรงกับ seed (ส่วน `AGENTS.md` ตรงแล้ว)
- เทส: `cmd/seed/images_test.go` (8 เทส: อ่านรูปทั้งโฟลเดอร์, ไม่มีโฟลเดอร์, escape URL, URL เสิร์ฟผ่าน `/uploads` ได้จริง, แกลเลอรีสาขา, ไฟล์ที่ลบหายจากแกลเลอรี, รูปปกห้อง, ไฟล์รูปปกไม่มีอยู่)

## ค่าตั้งค่า (env ของ `cmd/api`)
| ตัวแปร | ค่าเริ่มต้น | ผล |
|---|---|---|
| `APP_ENV` | `development` | `production` → gin release mode, log JSON, cookie OAuth ตั้ง Secure |
| `PORT` | `8080` | |
| `DATABASE_URL` | — | ไม่ตั้ง → ไม่ start |
| `DB_MAX_CONNS` / `DB_MIN_CONNS` / `DB_CONNECT_TIMEOUT_SEC` | 10 / 2 / 10 | |
| `JWT_SECRET` | — | สั้นกว่า 32 ตัว → ไม่ start |
| `ACCESS_TOKEN_TTL_MIN` / `REFRESH_TOKEN_TTL_DAY` | 30 / 30 | |
| `BCRYPT_COST` | 12 | |
| `UPLOAD_DIR` | `./uploads` | สร้างให้ถ้าไม่มี (สิทธิ์ 0750) |
| `PUBLIC_BASE_URL` | `http://localhost:8080` | ตัด `/` ท้าย · ใช้ประกอบ URL รูป/สลิป และ redirect URI ของ OAuth |
| `MAX_UPLOAD_MB` | 5 | |
| `ALLOWED_ORIGINS` | `http://localhost:3000,http://localhost:5173` | |
| `TRUSTED_PROXIES` | ว่าง | ค่าที่ไม่ใช่ IP/CIDR → ไม่ start |
| `SEED_DEFAULT_PASSWORD` | `Wisetsuk!2026` | |
| `STAFF_DEFAULT_PASSWORD` | = `SEED_DEFAULT_PASSWORD` | |
| `GOOGLE_CLIENT_ID/SECRET`, `FACEBOOK_CLIENT_ID/SECRET` | ว่าง | ตั้งไม่ครบคู่ → provider นั้นปิด |
| `FRONTEND_OAUTH_CALLBACK_URL` | `http://localhost:3000/auth/callback` | |
- ค่าตัวเลขที่พาร์สไม่ได้ → ใช้ค่าเริ่มต้นเงียบ ๆ · ค่าว่าง → ใช้ค่าเริ่มต้น
- `docker-compose.yml` ของ `api` ไม่ส่ง `BCRYPT_COST`, `STAFF_DEFAULT_PASSWORD` และ `DB_*` เข้า container → ใช้ค่าเริ่มต้นเสมอ
- frontend ใช้ตัวแปรเดียว `REACT_APP_API_BASE_URL`

## Error format (ทั้งระบบ)
```
{"error": {"code": "machine_readable_code", "message": "ข้อความไทย", "fields": {"ชื่อฟิลด์": "ข้อความไทย"}}}
```
`fields` มีเฉพาะเมื่อ code = `validation_failed`
code ที่ใช้มี 13 ตัวเท่านั้น: `unauthorized` · `forbidden` · `not_found` · `conflict` · `bad_request` · `validation_failed` ·
`method_not_allowed` · `internal_error` · `invalid_credentials` · `account_disabled` · `room_unavailable` · `invalid_state` · `too_many_requests`

## API Contract (ตามโค้ดจริง)

★ = frontend เรียกใช้อยู่แล้ว

```
GET  /health                                  → 200 {"status":"healthy","time"} (ไม่ห่อ data)
GET  /files/:assetID   (+HEAD)                → 200 ไฟล์ (public, ETag+immutable) | 304 | 400 | 404
GET  /uploads/*filepath                       ★ 200 ไฟล์สลิป + รูปเดโม (public ไม่มี auth ดู AC-17) | 404

--- Guest ---
POST /api/v1/auth/register {email,password,first_name,last_name,phone}
                                              ★ 201 token คู่ | 400 | 409 | 422
POST /api/v1/auth/login {email,password}      ★ 200 | 401 invalid_credentials | 403 account_disabled
                                              | 429 too_many_requests (+Retry-After)
POST /api/v1/auth/refresh {refresh_token}     → 200 | 401
POST /api/v1/auth/logout {refresh_token}      ★ 204 | 400 ไม่มี body
GET  /api/v1/auth/oauth                       → 200 {"providers":[...]}
GET  /api/v1/auth/oauth/:provider             → 302 ไป provider | 404 provider ปิดอยู่
GET  /api/v1/auth/oauth/:provider/callback    → 302 กลับ frontend ?code= | ?error= | 404 provider ปิดอยู่
POST /api/v1/auth/oauth/exchange {code}       → 200 token คู่ | 401 | 403 account_disabled
GET  /api/v1/branches                         ★ 200 [branch+images+amenities+nearby_places] (active เท่านั้น)
GET  /api/v1/branches/:branchIDหรือslug        ★ 200 | 404 (รวมสาขาที่ปิด)
GET  /api/v1/amenities                        → 200 [amenity] (frontend มีฟังก์ชันแต่ไม่มีหน้าไหนเรียก)
GET  /api/v1/room-types?branch_id=            → 200 [room_type] | 400
GET  /api/v1/rooms/search?branch_id=&room_type_id=&stay_type=&check_in=&check_out=
     &move_in_date=&min_price=&max_price=&page=&page_size=
                                              ★ 200 {data,meta} | 400 | 422
GET  /api/v1/rooms/:roomID                    → 200 room+amenities+images | 400 | 404 (รวมห้องที่ลบ/สาขาที่ปิด)
                                              (frontend มีฟังก์ชันแต่ไม่มีหน้าไหนเรียก)

--- ต้องล็อกอิน (ทุกบทบาท) ---
GET  /api/v1/me                               → 200 user | 401
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
GET    /dashboard                             ★ 200 {branches:[...],recent_activities:[≤10]}
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
PUT    /rooms/:roomID  (field ชุดเดียวกัน)      → 200 | 403 | 404 | 422 | 500 room_type_id ไม่มีจริง
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
GET    /branches                              ★ 200 [branch] รวมที่ปิดใช้งาน
GET    /staff                                 ★ 200 [user] (ไม่รวมคนที่ถูกลบ)
POST   /staff {email,first_name,last_name,phone?,branch_id}
                                              ★ 201 (must_change_password=true) | 400 | 422
PUT    /staff/:userID {first_name,last_name,phone,branch_id,is_active?,password?}
                                              ★ 200 | 400 ระงับตัวเอง | 404 | 422 | 500 branch_id ไม่มีจริง
DELETE /staff/:userID                         ★ 204 (soft delete) | 400 ลบตัวเอง | 404
```

### รูปร่างข้อมูลหลักใน response
- id ทุกตัวเป็นสตริงมี prefix อย่างน้อย 3 หลัก: `brn` สาขา · `bimg` รูปสาขา · `amt` สิ่งอำนวยความสะดวก · `np` สถานที่ใกล้เคียง · `usr` ผู้ใช้ · `rmt` ประเภทห้อง · `rm` ห้อง · `rimg` รูปห้อง · `bkg` ใบจอง · `pay` การชำระเงิน · `ntf` แจ้งเตือน · `log` activity log · `ast` ไฟล์
- token pair: `access_token, refresh_token, expires_at, token_type, user`
- user: `id, email, first_name, last_name, phone, role, branch_id?, branch_name?, must_change_password, avatar_url, is_active, last_login_at?, created_at, updated_at`
- branch: `id, slug, name, tagline, description, address, phones[], line_id, email, latitude?, longitude?, map_url, building_count, floor_count, daily_price_from?, monthly_price_min?, monthly_price_max?, water_rate, electric_rate, deposit, advance_payment, contract_fee, cover_image_url, is_active, created_at, updated_at, images?[], amenities?[], nearby_places?[]`
- room: `id, branch_id, branch_name?, room_type_id?, room_type_name?, room_number, building, floor, stay_type, price, water_rate, electric_rate, size_sqm?, description, image_url, status, is_active, amenities?[], images?[], created_at, updated_at`
- booking: `id, code, user_id, branch_id, room_id, stay_type, guest_first_name, guest_last_name, guest_phone, emergency_phone, emergency_relation, check_in_date?, check_out_date?, nights?, move_in_date?, contract_date?, appointment_at?, appointment_note, total_amount, status, reviewed_by?, reviewed_at?, cancelled_at?, created_at, updated_at, branch_name?, room_number?, member_name?, member_email?, latest_payment?`
- payment: `id, booking_id, amount, transferred_at, slip_url, note, status, reviewed_by?, reviewed_at?, reject_reason, created_at`
- activity log: `id, actor_id?, actor_role, actor_name, branch_id?, branch_name?, action, entity_type, entity_id, detail, ip_address, created_at`
- `?` = key หายไปทั้ง key เมื่อค่าว่าง/NULL (`omitempty`) ไม่ใช่ส่ง `null`

### เอกสารที่ไม่ตรงกับโค้ด
- มีใน `docs/openapi.yaml` / `backend/README.md` แต่ไม่มีในโค้ด:
  - `POST /api/v1/auth/forgot-password` **[ยังไม่ implement]** (AC-23)
  - `POST /api/v1/auth/reset-password` **[ยังไม่ implement]** (AC-23)
  - `GET /api/v1/bookings/:bookingID/slip` **[ยังไม่ implement]** (AC-17)
  - package `internal/mailer` และ "reset token" ใน `internal/auth` (README)
- มีในโค้ดแต่ไม่มีใน `docs/openapi.yaml`: `GET /uploads/*filepath` และ OAuth ทั้ง 4 เส้น (`/auth/oauth`, `/auth/oauth/:provider`, `/auth/oauth/:provider/callback`, `/auth/oauth/exchange`)
- `backend/README.md` ที่ยังเขียนตามของเก่า:
  - แผนภาพสถานะการจองยังมี `rejected` และ `completed` (ตัดไปแล้วใน migration 0002 ดู AC-8, AC-10)
  - "Super Admin ใส่ `branch_id` เพื่อเจาะดูสาขาเดียว" ใช้ไม่ได้กับ `/admin/dashboard` และ `/admin/members/:id/bookings` (AC-25, AC-28)
  - "UPDATE/DELETE ของแอดมินมี `AND branch_id = ...` เสมอ" ไม่จริงสำหรับการยกเลิกใบจอง (AC-11)
  - "ปิดแล้วปิดเลย" ยังจองห้องในสาขาที่ปิดได้ผ่าน `POST /bookings` (AC-8)
  - อีเมล admin ใน seed (ดูหัวข้อข้อมูลตั้งต้น)
- `docs/TASKS.md` บันทึกว่า error code มี 12 ตัว แต่โค้ดมี 13 (เพิ่ม `too_many_requests`) และยังมี task ที่ทำเสร็จแล้วในโค้ด (T-01/T-02 ตัด rejected, T-08 อ่านสาขาจาก DB) ค้างอยู่เป็นงานที่ยังไม่ทำ
- `docs/PLAN.md` บอก runtime ใช้ SQLite แต่โค้ดจริงใช้ PostgreSQL ทั้ง runtime และเทส
- คอมเมนต์ใน `frontend/src/pages/superadmin/Staff/ConfirmDeleteModal.jsx` และ `utils/nearbyPlaces.jsx` อ้างพฤติกรรม backend ชุดเก่า (ลบถาวร, "2 ใน 3 สาขาไม่มี nearby") ซึ่งไม่จริงแล้ว
