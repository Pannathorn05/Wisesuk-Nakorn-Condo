# SPEC ระบบจองห้องพักวิเศษสุขนครคอนโด

## Users
- Guest: ดูสาขา, ดูห้องพัก, ค้นหาห้องว่างตามวันที่/ราคา/ประเภท — ไม่ต้องล็อกอิน
- Member: จองห้อง (รายวัน/รายเดือน), แจ้งชำระเงินพร้อมสลิป, ดู/ยกเลิกการจองของตัวเอง, รับแจ้งเตือน
- Admin: จัดการห้องและรายละเอียดสาขา, อนุมัติ/ปฏิเสธการจอง, นัดหมายทำสัญญา, ดูแดชบอร์ดและ activity log — **เฉพาะสาขาที่ตัวเองผูกอยู่ 1 สาขาเท่านั้น**
- Super Admin: ทุกอย่างที่ Admin ทำได้ในทุกสาขา + เพิ่ม/แก้/ลบบัญชีผู้ดูแล
- Admin ผูกกับ 1 สาขาเสมอ ถ้าบัญชีไม่มีสาขา = ใช้ไม่ได้

## Out of scope (v1)
- Frontend (มีแค่ REST API)
- อัปโหลดรูปห้อง/รูปสาขาผ่าน backend (รับเป็น URL จากบริการภายนอก) — สลิปโอนเงินเป็นไฟล์เดียวที่เก็บเอง
- ชำระเงินออนไลน์ / ตัดบัตร (ใช้โอนเงินแล้วแนบสลิปให้แอดมินตรวจ)
- คืนเงิน, ค่าปรับ, ใบเสร็จ/ใบกำกับภาษี
- สร้าง/ลบสาขา, จัดการประเภทห้องและสิ่งอำนวยความสะดวก (ใส่ผ่าน `cmd/seed`)
- รูปห้องหลายรูปและสิ่งอำนวยความสะดวกรายห้อง (ตาราง `room_images`, `room_amenities` มีแล้วแต่ยังไม่มี API)
- ลืมรหัสผ่าน, ยืนยันอีเมล, เปลี่ยนอีเมล
- ระงับบัญชีสมาชิก
- ส่งอีเมล/SMS (แจ้งเตือนเก็บในระบบให้ดึงไปแสดงเท่านั้น)
- **Rate limit / account lockout ที่ `/auth/login`** — ยอมรับความเสี่ยง brute force ใน v1
- **endpoint จบสัญญา/เช็คเอาต์** — ปล่อยห้องด้วย `PATCH /admin/rooms/:roomID/status` แทน

## Acceptance Criteria

AC-1: สมัครสมาชิกได้ และได้สิทธิ์ member เสมอ
  - POST สมัครด้วยอีเมลใหม่ + รหัสถูกกติกา → 201 พร้อม access_token, refresh_token, user.role = "member"
  - ส่งฟิลด์ `role` แถมมาใน body → 422 code "validation_failed" + fields.role (ยกระดับตัวเองไม่ได้)
  - อีเมลซ้ำ (ไม่สนตัวพิมพ์ `A@b.com` = `a@b.com`) → 422 code "validation_failed" + fields.email
  - อีเมลซ้ำกับบัญชี staff ที่ถูกลบไปแล้ว → 422 เช่นกัน (อีเมลห้ามใช้ซ้ำตลอดกาล ดู AC-18)
  - กรอกผิดหลายช่อง → 422 และ fields มีครบทุกช่องในครั้งเดียว ไม่ใช่ทีละช่อง
  - รหัสผ่านต้องยาว 8–72 ตัวและมีทั้งตัวอักษรและตัวเลข ไม่ผ่านข้อใด → 422
  - รหัสยาว 73 ตัว → 422 (ห้ามปล่อยให้ bcrypt ตัดทิ้งเงียบ ๆ)
  - `phone` เป็นตัวเลข 9–10 หลัก, `first_name`/`last_name` ยาว 1–100 ตัว ไม่ผ่าน → 422
  - body พาร์สไม่ได้หรือไม่ใช่ JSON → 400 (ดู AC-21)

AC-2: เข้าสู่ระบบด้วย endpoint เดียวทุกบทบาท
  - อีเมล+รหัสถูก → 200 พร้อม token คู่
  - อีเมลไม่มีในระบบ **หรือ** รหัสผิด → 401 code "invalid_credentials" ข้อความเหมือนกันทั้งสองกรณี
  - จับเวลา 2 กรณีข้างบนกรณีละ 50 ครั้ง เทียบค่า **median** → ต่างกันไม่เกิน 20%
    (ต้องคำนวณ bcrypt กับ dummy hash แม้ไม่พบอีเมล ไม่งั้นจับเวลาเดาได้ว่าอีเมลใดมีบัญชี)
  - บัญชี is_active = false → 403 code "account_disabled"
    **ยอมรับความเสี่ยง:** โค้ดนี้เปิดเผยว่าอีเมลนั้นมีบัญชีอยู่จริง แลกกับข้อความที่ผู้ใช้เข้าใจได้

AC-3: อายุและการหมุนเวียนของ token
  - access token อายุ **15 นาที** — ใช้หลังหมดอายุ → 401
  - refresh token อายุ **30 วัน**
  - POST refresh ด้วยใบที่ยังไม่หมดอายุ → 200 พร้อม token คู่ใหม่
  - ใช้ใบเดิมซ้ำอีกครั้ง → 401
  - ใบที่หมดอายุแล้ว / ใบที่ logout ไปแล้ว / สตริงที่ไม่ใช่ token → 401 ข้อความเดียวกันทุกกรณี
  - ยิงพร้อมกัน 2 request ด้วยใบเดียวกัน → สำเร็จ 1 ใบเท่านั้น (บังคับที่ database)
  - ตาราง refresh_tokens เก็บเฉพาะ sha256 hash ไม่มี token ดิบ
  - เปลี่ยนรหัสผ่านสำเร็จ → refresh token ทุกใบของบัญชีนั้นใช้ไม่ได้ทันที
    **ยอมรับความเสี่ยง:** access token ใบเดิมยังใช้ได้จนหมดอายุ (สูงสุด 15 นาที)
  - POST logout ด้วย token ที่ผิด/หมดอายุ/ไม่มีจริง → 204 เหมือนกันหมด (ไม่บอกใบ้ว่าใบไหนมีจริง)

AC-4: Admin ถูกล็อกไว้ที่สาขาเดียว (กฎเหล็ก)
  - admin สาขา A เรียก endpoint ใต้ /admin พร้อม `?branch_id=<สาขา B>` → **403**
  - admin สาขา A เปิดทรัพยากรรายชิ้นของสาขา B (เช่น `GET /bookings/:bookingID`) → **404**
    (สองบรรทัดบนต่างกันเพราะรายชื่อสาขาเป็นข้อมูลสาธารณะอยู่แล้ว แต่รหัสการจองไม่ใช่ — ห้ามยืนยันว่ามีอยู่จริง)
  - admin สาขา A ไม่ระบุ branch_id → ระบบบังคับใช้สาขา A ไม่ใช่ "ทุกสาขา"
  - GET /admin/bookings ของ admin สาขา A → ทุกแถวมี branch_id เป็นสาขา A
  - super admin ไม่ระบุ branch_id → เห็นทุกสาขา
  - role = admin แต่บัญชีในฐานข้อมูลไม่มี branch_id → 401
  - super admin ย้าย admin จากสาขา A ไป B แล้ว admin คนนั้นใช้ **token ใบเดิม** ต่อ → เห็นเฉพาะสาขา B ทันที ไม่ต้อง re-login
    (ทำได้เมื่อ JWT เก็บแค่ `user_id` + `role` แล้วอ่าน `branch_id` จากฐานข้อมูลทุกคำขอ — ห้ามฝัง branch_id ใน token)

AC-5: Member เห็นเฉพาะของตัวเอง
  - member A เปิดการจองของ member B → 404 (ไม่ใช่ 403 เพื่อไม่ให้รู้ว่ารหัสจองนั้นมีจริง)
  - member เรียก endpoint ใต้ /admin หรือ /superadmin → 403
  - member A เปิดสลิปของ member B → 404 (ดู AC-17)
  - POST /me/notifications/read ด้วย id ที่ไม่ใช่ของตัวเองหรือไม่มีจริง → 204 โดยไม่เปลี่ยนอะไร (ไม่บอกใบ้)

AC-6: จองรายวันคิดเงินตามจำนวนคืน
  - **จำนวนคืน = `check_out - check_in`** โดยนับวันเช็คอินรวม ไม่นับวันเช็คเอาต์
    เช่น 5–8 มี.ค. = นอนคืนวันที่ 5, 6, 7 → 3 คืน
  - จอง 3 คืน ห้อง 1,200/คืน → 201 และ total_amount = 3600
  - **ราคาถูกคัดลอกลงใบจองตอนสร้าง** — แก้ราคาห้องเป็น 2,000 หลังจองแล้ว → total_amount ของใบเดิมยังเป็น 3600
  - จอง 400 คืน → 201 และยอดเงินถูกต้องไม่ล้น (เก็บยอดเป็น `numeric` ไม่ใช่ integer, ไม่มีเพดานจำนวนคืน)
  - check_out ≤ check_in → 422
  - check_in เป็นวันที่ผ่านมาแล้ว → 422 (check_in = วันนี้ ผ่าน — ดู AC-22)
  - จองแบบ daily บนห้องที่เปิดขายแบบ monthly → 422
  - room_id ไม่มีในระบบ หรือถูกลบไปแล้ว → 404
  - room_id ไม่ใช่ UUID → 400
  - จองห้องในสาขาที่ปิดใช้งาน → 404

AC-7: จองรายเดือนเก็บแค่ค่าทำสัญญา
  - ห้อง 5,000/เดือน สาขาเก็บค่าทำสัญญา 500 → 201 และ total_amount = **500** ไม่ใช่ 5000
  - **ค่าทำสัญญาถูกคัดลอกลงใบจองตอนสร้าง** — แก้ค่าทำสัญญาของสาขาเป็น 800 หลังจองแล้ว → total_amount ของใบเดิมยังเป็น 500
  - move_in_date เป็นวันที่ผ่านมาแล้ว → 422
  - จองแบบ monthly บนห้องที่เปิดขายแบบ daily → 422

AC-8: ห้องเดียวจองซ้อนไม่ได้ (กฎเหล็ก)
  - **ห้องถูกล็อกเมื่อมีใบสถานะ `pending_payment` / `awaiting_review` / `approved`** — รายชื่อนี้ครบถ้วน ไม่มีสถานะอื่นที่ล็อกห้อง
  - จองห้องที่มีใบ**รายเดือน**สถานะข้างต้นค้างอยู่ → 409 code "room_unavailable"
  - **สองใบรายวันทับกันเมื่อ `ใหม่.check_in < เก่า.check_out AND ใหม่.check_out > เก่า.check_in`**
    จองรายวันทับช่วงของใบที่ล็อกอยู่ → 409 code "room_unavailable"
  - จองรายวันที่ check_in ตรงกับ check_out ของใบเดิมพอดี → **201 สำเร็จ** (ตามสูตรข้างบนถือว่าไม่ทับ ห้องต้องไม่ว่างฟรี 1 คืน)
  - 2 request จองห้องเดียวกันพร้อมกัน → สำเร็จ 1 ใบเท่านั้น (ล็อกแถวห้องตลอด transaction)
  - จองห้องสถานะ maintenance หรือ occupied → 409
  - `booking_code` รูปแบบ `PT-000001` รันต่อเนื่องจาก database sequence (ห้ามใช้ `MAX+1` เพราะจะซ้ำเมื่อยิงพร้อมกัน)
  - จอง **100 ห้องที่ต่างกัน** พร้อมกัน → ได้ booking_code ไม่ซ้ำครบ 100 รหัส
  - **URL ทุกเส้นใช้ UUID ไม่ใช่ booking_code** — `GET /api/v1/bookings/PT-000001` → 400 (booking_code ใช้แสดงผลและค้นหาเท่านั้น)

AC-9: แจ้งชำระเงินพร้อมสลิป
  - การจองสถานะ pending_payment + แนบสลิป → 200 และสถานะเป็น awaiting_review
  - การจองสถานะ awaiting_review แจ้งซ้ำ → 409 code "invalid_state" (รอแอดมินตรวจอยู่)
  - การจองสถานะ approved หรือ cancelled แจ้งซ้ำ → 409 code "invalid_state"
  - การจองที่เลย expires_at ไปแล้ว → 409 code "invalid_state"
  - แนบไฟล์ .txt ที่เปลี่ยนนามสกุลเป็น .jpg → 400 (ตรวจจากเนื้อไฟล์ ไม่เชื่อ Content-Type)
  - ไฟล์ใหญ่เกิน MAX_UPLOAD_MB → 400 และต้องไม่มีไฟล์ถูกเขียนลงดิสก์
  - transferred_at เป็นอนาคตเกิน 5 นาที → 422
  - transferred_at ย้อนหลังเกิน 30 วัน → 422
  - ไฟล์ที่บันทึกต้องตั้งชื่อด้วย UUID ไม่ใช้ชื่อจาก client และเก็บ **นอก** directory ที่เว็บเสิร์ฟได้
  - `amount` ที่แจ้งไม่ตรงกับ total_amount → **รับเข้าระบบตามปกติ** ให้แอดมินเป็นผู้ตัดสิน แต่ response ฝั่งแอดมินต้องแสดงทั้งสองค่า
  - **การแจ้งชำระเงินเก็บเป็นประวัติหลายแถวต่อ 1 การจอง** ห้ามเขียนทับแถวเดิม (ต้องย้อนดูได้ว่าเคยถูกปฏิเสธเพราะอะไร)

AC-10: อนุมัติ/ปฏิเสธการจอง
  - อนุมัติใบสถานะ awaiting_review → 200 สถานะเป็น approved และ payment ล่าสุดเป็น approved
  - อนุมัติใบ **รายเดือน** → ห้องเปลี่ยนเป็น occupied
  - **ปฏิเสธใบสถานะ awaiting_review → 200 และใบกลับไปสถานะ `pending_payment` โดยห้องยังถูกล็อกอยู่**
    - `expires_at` ถูกตั้งใหม่เป็น **+24 ชั่วโมง** นับจากเวลาที่ปฏิเสธ (คนละค่ากับ 10 นาทีตอนสร้างใบ — คนนี้โอนเงินมาแล้วจริง)
    - `reason` เก็บที่แถว payment attempt นั้น ไม่ใช่ที่ใบจอง
  - ปฏิเสธโดยไม่ระบุ reason → 422
  - อนุมัติหรือปฏิเสธใบที่ไม่ได้อยู่สถานะ awaiting_review → 409 code "invalid_state"
  - admin สาขา A อนุมัติใบของสาขา B → 404 (เป็นทรัพยากรรายชิ้น ดู AC-4)
  - อนุมัติหรือปฏิเสธสำเร็จ → สมาชิกได้แจ้งเตือน 1 รายการที่อ้าง booking_code และกรณีปฏิเสธต้องมี reason อยู่ในแจ้งเตือน
  - ยิง approve พร้อมกัน 2 request บนใบเดียว → สำเร็จ 1 ครั้ง อีกครั้งได้ 409

AC-11: ยกเลิกแล้วห้องต้องกลับมาขายได้ (กฎเหล็ก)
  - member ยกเลิกใบสถานะ pending_payment ของตัวเอง → 200 สถานะเป็น cancelled
  - **member ยกเลิกใบสถานะ awaiting_review หรือ approved เอง → 409** (ต้องให้แอดมินจัดการ)
  - admin ยกเลิกใบสถานะ pending_payment / awaiting_review / approved ในสาขาตัวเอง → 200
  - ยกเลิกใบที่ยกเลิกไปแล้ว → 409 code "invalid_state"
  - admin ยกเลิกใบรายเดือนที่อนุมัติแล้ว → ห้องกลับเป็น available
  - ห้องอยู่สถานะ maintenance แล้วใบถูกยกเลิก → ห้อง**ยังคง** maintenance ไม่ถูกทับเป็น available
  - คืนห้องไม่สำเร็จ → การจองต้องไม่ถูกยกเลิกด้วย (transaction เดียวกัน)
  - ห้องที่ถูกปล่อยแล้ว → จองใหม่ได้ทันที
  - ยิง cancel พร้อมกัน 2 request บนใบเดียว → สำเร็จ 1 ครั้ง อีกครั้งได้ 409

AC-12: ค้นหาห้องว่าง
  - ระบุ check_in/check_out → ไม่มีห้องที่มีการจองรายวันช่วงคาบเกี่ยวกันในผลลัพธ์ (ใช้สูตรทับกันเดียวกับ AC-8)
  - ห้องที่ check_out ของใบเดิมตรงกับ check_in ที่ค้นหา → **ต้องอยู่ในผลลัพธ์**
  - ห้องสถานะ maintenance / occupied, ห้องที่ถูกลบไปแล้ว, ห้องของสาขาที่ปิดใช้งาน → ไม่อยู่ในผลลัพธ์
  - page_size=500 → คืนไม่เกิน 100 รายการ
  - ไม่ระบุ page → หน้า 1 ขนาด 20 พร้อม meta.total_items และ meta.total_pages
  - page หรือ page_size เป็น 0 / ติดลบ / ไม่ใช่ตัวเลข → 400
  - page เกิน total_pages → 200 พร้อม data เป็น array ว่าง (ไม่ใช่ 404)
  - stay_type เป็นค่าที่ไม่รู้จัก → 400
  - branch_id / room_id ที่ไม่ใช่ UUID → 400

AC-13: บันทึกทุกการกระทำที่เปลี่ยนข้อมูล
  - ทำ action ใด ๆ ใน **23 รายการข้างล่าง** → activity_logs มีแถวใหม่ครบ: ผู้ทำ (id+ชื่อ+role), สาขา, action, entity, IP, เวลา
  - `booking.auto_expire` เป็นรายการเดียวที่ผู้ทำเป็นระบบ → `actor_role = "system"`, `actor_id` เป็น NULL, `actor_name = "system"`
    (แปลว่าคอลัมน์ `actor_id` ต้องเป็น nullable ตั้งแต่ migration แรก)
  - เขียน log ล้มเหลว → ธุรกรรมหลักที่สำเร็จแล้วต้องไม่ rollback ตาม
  - admin สาขา A เรียก GET /admin/activity-logs → เห็นเฉพาะ log สาขา A

  | # | action | endpoint |
  |---|---|---|
  | 1 | `user.register` | POST /auth/register |
  | 2 | `profile.update` | PUT /me |
  | 3 | `password.change` | POST /me/password |
  | 4 | `booking.create` | POST /bookings |
  | 5 | `payment.submit` | POST /bookings/:bookingID/payment |
  | 6 | `booking.cancel_by_member` | POST /bookings/:bookingID/cancel (member) |
  | 7 | `booking.cancel_by_admin` | POST /bookings/:bookingID/cancel (admin) |
  | 8 | `booking.approve` | POST /admin/bookings/:bookingID/approve |
  | 9 | `booking.reject` | POST /admin/bookings/:bookingID/reject |
  | 10 | `booking.appointment_set` | PUT /admin/bookings/:bookingID/appointment |
  | 11 | `room.create` | POST /admin/rooms |
  | 12 | `room.update` | PUT /admin/rooms/:roomID |
  | 13 | `room.status_change` | PATCH /admin/rooms/:roomID/status |
  | 14 | `room.delete` | DELETE /admin/rooms/:roomID |
  | 15 | `branch.update` | PUT /admin/branch |
  | 16 | `branch.amenities_update` | PUT /admin/branch/amenities |
  | 17 | `branch.nearby_update` | PUT /admin/branch/nearby |
  | 18 | `branch.image_add` | POST /admin/branch/images |
  | 19 | `branch.image_delete` | DELETE /admin/branch/images/:imageID |
  | 20 | `staff.create` | POST /superadmin/staff |
  | 21 | `staff.update` | PUT /superadmin/staff/:userID |
  | 22 | `staff.delete` | DELETE /superadmin/staff/:userID |
  | 23 | `booking.auto_expire` | ระบบ (AC-16) |

AC-14: รูปภาพรับเป็น URL ภายนอกเท่านั้น
  - image_url เป็น `javascript:...` หรือ `data:...` → 422
  - image_url ไม่มี host (เช่น `/images/a.jpg`) → 422
  - image_url ที่ scheme ไม่ใช่ http/https → 422
  - **ไม่มี static file route ในระบบ** — `GET /uploads/*` ทุกรูปแบบ → 404 (สลิปเสิร์ฟผ่าน AC-17 เท่านั้น)

AC-15: ความปลอดภัยและความทนทานพื้นฐาน
  - ทุก query ใช้ parameterized ไม่มีการต่อ string ค่าเข้า SQL — **ตรวจด้วย code review ไม่ใช่เทสผ่าน API**
  - error 5xx → body ต้องไม่มี stack trace, ชื่อตาราง หรือข้อความ SQL
  - panic ใน handler → server ไม่ล้ม และ client ได้ 500 เป็น JSON
  - ข้อความ error ทุกตัวที่ผู้ใช้เห็นเป็นภาษาไทย รวมถึง error ระดับ router (404 เส้นทางไม่มีจริง, 405 method ผิด, body พาร์สไม่ได้)

AC-16: การจองที่ไม่จ่ายเงินต้องหมดอายุเอง
  - สร้างใบจอง → `expires_at` = เวลาปัจจุบัน + **10 นาที**
  - ถูกปฏิเสธแล้วกลับเป็น pending_payment → `expires_at` = เวลาที่ปฏิเสธ + **24 ชั่วโมง** (AC-10)
  - `expires_at` เป็นคอลัมน์รายใบ ห้ามคำนวณจาก `created_at`
  - ใบที่เลย `expires_at` และยังเป็น pending_payment → สถานะเป็น cancelled และห้องถูกปล่อยตามกฎ AC-11
  - แนบสลิปสำเร็จ (เข้าสู่ awaiting_review) → หยุดนับเวลา รอแอดมินตรวจไม่จำกัดเวลา
  - ห้องที่ถูกปล่อยเพราะหมดอายุ → จองใหม่ได้ทันที
  - สมาชิกได้แจ้งเตือน 1 รายการเมื่อใบถูกยกเลิกอัตโนมัติ

AC-17: ไฟล์สลิปต้องตรวจสิทธิ์ก่อนเสิร์ฟ (กฎเหล็ก)
  - `GET /api/v1/bookings/:bookingID/slip` ไม่มี token → 401
  - เจ้าของใบจอง / admin สาขาเดียวกับใบนั้น / super admin → 200 พร้อมไฟล์และ Content-Type ที่ถูกต้อง
  - member คนอื่น หรือ admin คนละสาขา หรือ bookingID ไม่มีจริง → 404 เหมือนกันหมด
  - ใบจองที่ยังไม่เคยแนบสลิป → 404
  - ไฟล์เก็บนอก directory ที่เว็บเสิร์ฟได้ และ path ที่ใช้อ่านไฟล์ต้องมาจากฐานข้อมูลเท่านั้น ห้ามนำ input จาก client มาต่อ path (กัน path traversal)

AC-18: Super Admin จัดการบัญชีผู้ดูแล
  - POST /superadmin/staff ด้วยอีเมลใหม่ + branch_id ที่มีจริง → 201 และ role เป็น admin
  - **อีเมลห้ามใช้ซ้ำตลอดกาล** — สมัครด้วยอีเมลของ staff ที่ลบไปแล้ว → 422
  - branch_id ไม่พบ → 422
  - PUT /superadmin/staff/:userID ระงับตัวเอง (is_active=false) → 400
  - DELETE /superadmin/staff/:userID ลบตัวเอง → 400
  - เป้าหมายเป็น member หรือไม่มีจริง → 404
  - **ลบหรือระงับ staff สำเร็จ → refresh token ทุกใบของบัญชีนั้นถูกเพิกถอนทันที**
  - **ลบ staff เป็น soft delete** → ตั้ง `deleted_at` และ `is_active = false` แต่ไม่ลบแถวจริง
    เรียก GET /superadmin/staff → ไม่เห็นคนที่ถูกลบ แต่ activity_logs เดิมยังแสดงชื่อผู้ทำได้ครบ
  - ย้ายสาขา staff → **ไม่**เพิกถอน token และผู้ใช้เห็นสาขาใหม่ทันทีในคำขอถัดไป (AC-4)

AC-19: จัดการห้องและการปล่อยห้อง
  - POST /admin/rooms เลขห้องซ้ำในสาขาเดียวกัน → 422
  - **ลบห้องเป็น soft delete** → ตั้ง `deleted_at` ไม่ลบแถวจริง ห้องหายจากผลค้นหาแต่ใบจองเก่ายังอ้างถึงห้องนั้นได้
  - สร้างห้องด้วยเลขที่ซ้ำกับห้องที่ถูกลบไปแล้ว → 201 สำเร็จ
    (ต้องใช้ partial unique index `UNIQUE (branch_id, room_number) WHERE deleted_at IS NULL`)
  - price ≤ 0 หรือ stay_type ไม่ตรงกับราคาที่ส่งมา → 422
  - PATCH /admin/rooms/:roomID/status เปลี่ยนห้องจาก occupied เป็น available → 200 (ทางออกจาก occupied ใน v1)
  - PATCH เปลี่ยนเป็น maintenance ขณะมีใบ active ค้างอยู่ → 409 code "conflict"
  - DELETE ห้องที่มีใบสถานะ pending_payment / awaiting_review / approved ค้างอยู่ → 409
  - status ที่ไม่อยู่ใน 3 ค่าที่กำหนด → 422

AC-20: นัดหมายทำสัญญา
  - PUT /admin/bookings/:bookingID/appointment บนใบรายวัน → 400
  - บนใบรายเดือนที่ยังไม่ถึงสถานะ approved → 409 code "invalid_state"
  - appointment_at เป็นเวลาในอดีต → 422
  - นัดหมายสำเร็จ → สมาชิกได้แจ้งเตือน 1 รายการที่อ้าง booking_code

AC-21: รหัสสถานะและรูปแบบ error สอดคล้องกันทั้งระบบ
  - **400 bad_request** ใช้เมื่อ: พาร์ส body ไม่ได้, ชนิดข้อมูลผิด, query param ผิดรูป (branch_id ไม่ใช่ UUID, page ไม่ใช่ตัวเลข), ไฟล์อัปโหลดไม่ถูกต้อง
  - **422 validation_failed** ใช้เมื่อ: ค่าถูกชนิดแต่ผิดกติกาธุรกิจ (รหัสผ่านสั้นไป, check_out ≤ check_in, ส่งฟิลด์ต้องห้าม)
  - `fields` ปรากฏเฉพาะตอน code = "validation_failed" เท่านั้น ห้ามมีใน error ตัวอื่น
  - ไม่มี token หรือ token ใช้ไม่ได้ → 401 · มี token แต่บทบาทไม่พอ → 403
  - error ทุกตัวรวมถึง 404/405/500 ใช้ format เดียวกันหมด ไม่มี response ที่หลุดออกนอกรูปแบบนี้

AC-22: เวลาและวันที่
  - เก็บทุก timestamp เป็น **UTC** ในฐานข้อมูล
  - ตัดสิน "วันนี้ / อดีต / อนาคต" ด้วย **Asia/Bangkok** เสมอ ทั้ง check_in, move_in_date, transferred_at, appointment_at
  - ตั้งนาฬิกาเซิร์ฟเวอร์เป็น UTC แล้วจอง check_in = วันนี้ตามเวลาไทย ตอน **00:30 น. เวลาไทย** → **201 ไม่ใช่ 422**
    (ต้องเป็นช่วงเวลาไทย 00:00–06:59 เท่านั้น เพราะตอนนั้น UTC ยังเป็นเมื่อวาน
    โค้ดที่เทียบวันที่ด้วย UTC ตรง ๆ จะเห็นว่าเป็นอดีตแล้วตอบ 422 ผิด ๆ
    ส่วนเวลาไทย 07:00–23:59 วันที่ของสองเขตเวลาตรงกัน จับบั๊กนี้ไม่ได้)
  - วันที่ใน response เป็น ISO 8601 พร้อม offset

## Error format (ทั้งระบบ)
{"error": {"code": "machine_readable_code", "message": "ข้อความไทย", "fields": {"ชื่อฟิลด์": "ข้อความไทย"}}}

`fields` มีเฉพาะตอน code = "validation_failed"
code ที่ใช้: unauthorized · forbidden · not_found · conflict · bad_request · validation_failed ·
method_not_allowed · internal_error · invalid_credentials · account_disabled · room_unavailable · invalid_state

## API Contract

```
GET  /health                                  → 200 {"status":"healthy","time"}

--- Guest ---
POST /api/v1/auth/register {email,password,first_name,last_name,phone}
                                              → 201 | 400 | 422
POST /api/v1/auth/login {email,password}      → 200 | 401 invalid_credentials | 403 account_disabled
POST /api/v1/auth/refresh {refresh_token}     → 200 | 401
POST /api/v1/auth/logout {refresh_token}      → 204 เสมอ
GET  /api/v1/branches                         → 200 [branch]
GET  /api/v1/branches/:branchID               → 200 branch+images+amenities+nearby | 400 | 404
GET  /api/v1/amenities                        → 200 [amenity]
GET  /api/v1/room-types?branch_id=            → 200 [room_type]
GET  /api/v1/rooms/search?branch_id=&room_type_id=&stay_type=&check_in=&check_out=
     &move_in_date=&min_price=&max_price=&page=&page_size=
                                              → 200 {data,meta} | 400
GET  /api/v1/rooms/:roomID                    → 200 room | 400 | 404

--- ต้องล็อกอิน (ทุกบทบาท) ---
GET  /api/v1/me                               → 200 user | 401
PUT  /api/v1/me {first_name,last_name,phone,avatar_url}
                                              → 200 | 422
POST /api/v1/me/password {current_password,new_password,confirm_password}
                                              → 200 (เพิกถอน refresh token ทุกใบ) | 422
GET  /api/v1/me/notifications                 → 200 {items,unread_count}
POST /api/v1/me/notifications/read {id?}      → 204
GET  /api/v1/bookings/:bookingID              → 200 booking | 404 (ไม่ใช่เจ้าของ/คนละสาขา)
GET  /api/v1/bookings/:bookingID/slip         → 200 ไฟล์ | 401 | 404
POST /api/v1/bookings/:bookingID/cancel       → 200 | 409 invalid_state | 404

--- Member ---
POST /api/v1/bookings {room_id,stay_type,guest_*,check_in_date|move_in_date,...}
                                              → 201 | 400 | 404 ไม่พบห้อง | 409 room_unavailable | 422
GET  /api/v1/bookings?status=&stay_type=&page= → 200 {data,meta}
     status ∈ pending_payment | awaiting_review | approved | cancelled
POST /api/v1/bookings/:bookingID/payment      multipart: slip,amount,transferred_at,note
                                              → 200 | 400 ไฟล์ผิด | 409 invalid_state | 422

--- Admin (+ Super Admin) ใต้ /api/v1/admin ---
GET    /dashboard                             → 200 {branches,recent}
GET    /activity-logs?actor_role=&actor_id=&action=&branch_id=&page=
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
POST   /rooms {branch_id,room_number,stay_type,price,...}
                                              → 201 | 422 เลขห้องซ้ำ/ราคาไม่ถูกต้อง | 403
PUT    /rooms/:roomID                         → 200 | 422 | 403 | 404
PATCH  /rooms/:roomID/status {status}         → 200 | 409 มีใบ active ค้าง | 422 สถานะไม่ถูกต้อง | 403
DELETE /rooms/:roomID                         → 204 (soft delete) | 409 มีใบ active ค้าง | 403 | 404
PUT    /branch?branch_id= {name,phones,rates,...}
                                              → 200 | 422 | 403
PUT    /branch/amenities?branch_id= {amenity_ids}
                                              → 200 | 403
PUT    /branch/nearby?branch_id= {items}      → 200 | 422 | 403
POST   /branch/images?branch_id= {image_url,caption,sort_order}
                                              → 201 | 422 URL ไม่ถูกต้อง | 403
DELETE /branch/images/:imageID?branch_id=     → 204 | 403 | 404

--- Super Admin เท่านั้น ใต้ /api/v1/superadmin ---
GET    /branches                              → 200 [branch] รวมที่ปิดใช้งาน
GET    /staff                                 → 200 [user]
POST   /staff {email,password,first_name,last_name,phone,branch_id}
                                              → 201 | 422 อีเมลซ้ำ/ไม่พบสาขา
PUT    /staff/:userID {first_name,last_name,phone,branch_id,is_active,password?}
                                              → 200 | 400 ระงับตัวเอง | 404 เป้าหมายเป็น member
DELETE /staff/:userID                         → 204 (soft delete) | 400 ลบตัวเอง | 404
```
