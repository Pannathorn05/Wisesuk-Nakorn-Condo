# TASKS ระบบจองห้องพักวิเศษสุขนครคอนโด
ทำทีละ task ตามลำดับ ห้ามข้าม ห้ามรวม scope
Endpoint ใช้ตาม API Contract ใน SPEC เท่านั้น
Task Test ต้องแดงก่อน แล้ว task Implement ถัดไปทำให้เขียว
TC ใต้แต่ละ task Test คือรายการที่ต้องเขียนให้ครบ ห้ามตกข้อใด

คำสั่งที่ใช้ใน DoD
- CHECK = `cd backend && gofmt -l . && go build ./... && go vet ./...`
- TEST  = `cd backend && go test ./... -race`

## เสร็จแล้ว (ยืนยันจากโค้ดจริง)
- `docs/openapi.yaml` ครบ 39 path / 45 operation ตรง API Contract
- error envelope + 12 code + middleware แปลง error จุดเดียว (AC-21) พร้อมเทส 8 ตัว
- `internal/timex` ตัดสินวันที่ด้วย Asia/Bangkok เสมอ (AC-22) พร้อมเทส 10 ตัว
- โครงเทส 1 test = 1 database + `UPLOAD_DIR` แยก รันขนานได้ อ่าน `.env` เองเมื่อไม่ได้ตั้ง env
- 5 module (account, branch, room, booking, reporting) และ endpoint ครบทุกเส้น
- กันจองซ้อนรายวันแบบ half-open + `SELECT FOR UPDATE` + `booking_code_seq`

---

## Phase 1 ตัดสถานะ rejected กฎเหล็ก (AC-8, AC-10)

T-01 Test: state machine 4 สถานะ + ปฏิเสธแล้วใบกลับ pending_payment โดยห้องยังถูกล็อก
     DoD: `go test ./internal/modules/booking/... -run TestBookingStateMachine` แดง และครอบ AC-10 ครบ
T-02 Implement: migration ตัด `rejected`/`completed` ออกจาก enum + แก้ `booking/model.go` + service
     DoD: TestBookingStateMachine ผ่าน · CHECK สะอาด

## Phase 2 การจองหมดอายุ (AC-16)

T-03 Test: expires_at และการยกเลิกอัตโนมัติ
     DoD: `go test -run TestBookingExpiry` แดง และครอบ AC-16 ครบ
T-04 Implement: คอลัมน์ `expires_at` + งานเบื้องหลังยกเลิกอัตโนมัติ + แจ้งเตือน
     DoD: TestBookingExpiry ผ่าน · CHECK สะอาด

## Phase 3 สลิปต้องตรวจสิทธิ์ กฎเหล็ก (AC-17, AC-14)

T-05 Test: `GET /api/v1/bookings/{bookingID}/slip`
     DoD: `go test -run TestBookingSlip` แดง และครอบ AC-17 ครบ
T-06 Implement: endpoint ใหม่ + ย้ายไฟล์ออกนอก web root + ลบ static route `/uploads/*`
     DoD: TestBookingSlip ผ่าน · CHECK สะอาด

## Phase 4 สิทธิ์สาขา กฎเหล็ก (AC-4)

T-07 Test: ขอบเขตสาขาของ admin
     DoD: `go test -run TestAdminBranchScope` แดง และครอบ AC-4 ครบ
T-08 Implement: เอา `branch_id` ออกจาก JWT claim อ่านจาก DB ทุกคำขอ + แยก 403/404
     DoD: TestAdminBranchScope ผ่าน · CHECK สะอาด

## Phase 5 ประวัติการชำระเงิน (AC-9)

T-09 Test: แจ้งชำระเงินพร้อมสลิป
     DoD: `go test -run TestPaymentAttempts` แดง และครอบ AC-9 ครบ
T-10 Implement: ลบ `bookings.reject_reason` ให้เหลือที่ payment + validation 30 วัน + แสดง amount คู่ total_amount
     DoD: TestPaymentAttempts ผ่าน · CHECK สะอาด

## Phase 6 จัดการห้อง (AC-19)

T-11 Test: วงจรชีวิตของห้อง
     DoD: `go test -run TestRoomLifecycle` แดง และครอบ AC-19 ครบ
T-12 Implement: migration `deleted_at` + partial unique index + service
     DoD: TestRoomLifecycle ผ่าน · CHECK สะอาด

## Phase 7 Activity log (AC-13)

T-13 Test: บันทึกทุกการกระทำที่เปลี่ยนข้อมูล
     DoD: `go test -run TestActivityLog` แดง และครอบ AC-13 ครบ
T-14 Implement: enum 23 action + เพิ่มค่า `system` ใน `actor_role`
     DoD: TestActivityLog ผ่าน · CHECK สะอาด

## Phase 8 กันจองซ้อนที่ระดับ database (AC-8)

T-15 Test: จองพร้อมกัน
     DoD: `go test -run TestBookingConcurrency -race` แดง และครอบ AC-8 ครบ
T-16 Implement: unique constraint กันใบ active ซ้อนกันต่อห้อง แล้ว map ความผิดพลาดเป็น 409
     DoD: TestBookingConcurrency ผ่าน · CHECK สะอาด

## Phase 9 ปิดเทส AC ที่เหลือ
ทุก task เขียนเทสอย่างเดียว ถ้าแดงเพราะโค้ดไม่ตรง SPEC ให้เพิ่ม task Implement ต่อท้าย task นั้น

T-17 Test AC-1 สมัครสมาชิก + AC-2 เข้าสู่ระบบ
     DoD: `go test -run 'TestRegister|TestLogin'` ครอบ AC-1 และ AC-2 ครบทุกบรรทัด
T-18 Test AC-3 อายุและการหมุนเวียน token
     DoD: `go test -run TestToken -race` ครอบ AC-3 ครบ
T-19 Test AC-5 ขอบเขตข้อมูลของ member
     DoD: `go test -run TestMemberScope` ครอบ AC-5 ครบ
T-20 Test AC-6 จองรายวัน + AC-7 จองรายเดือน
     DoD: `go test -run TestCreateBooking` ครอบ AC-6 และ AC-7 ครบ
T-21 Test AC-11 ยกเลิกและคืนห้อง
     DoD: `go test -run TestCancelBooking` ครอบ AC-11 ครบ
T-22 Test AC-12 ค้นหาห้องว่าง
     DoD: `go test -run TestSearchRooms` ครอบ AC-12 ครบ
T-23 Test AC-18 จัดการผู้ดูแล + AC-20 นัดหมายทำสัญญา
     DoD: `go test -run 'TestStaff|TestAppointment'` ครอบ AC-18 และ AC-20 ครบ
T-24 Test AC-15 ความทนทาน
     DoD: `go test -run TestResilience` ครอบ AC-15 ครบ

## Phase 10 Frontend

T-25 Generate typed client จาก `docs/openapi.yaml` ไป `frontend/src/api/`
     DoD: `npm run generate:api` สำเร็จ · `npm run build` ผ่าน · `grep -r "fetch(" src/` เจอเฉพาะไฟล์ wrapper
T-26 UI ค้นหาห้องว่าง
     DoD: E2E checklist ครบทั้ง 4 ข้อ
T-27 UI เข้าสู่ระบบและสมัครสมาชิก
     DoD: E2E checklist ครบทั้ง 4 ข้อ
T-28 UI จองห้องและแจ้งชำระเงิน
     DoD: E2E checklist ครบทั้ง 4 ข้อ
T-29 UI การจองของฉัน
     DoD: E2E checklist ครบทั้ง 4 ข้อ
T-30 UI แอดมินอนุมัติ/ปฏิเสธ
     DoD: E2E checklist ครบทั้ง 4 ข้อ
T-31 UI จัดการห้องและผู้ดูแล
     DoD: E2E checklist ครบทั้ง 4 ข้อ

## Phase 11 Hardening

T-32 Quality gates บนเครื่อง (gofmt + go build + go vet + go test -race + golangci-lint + govulncheck + redocly lint + npm run build)
     DoD: สคริปต์เดียวรันผ่านครบทุกคำสั่ง และแดงจริงเมื่อจงใจใส่โค้ดผิด
T-33 CI ด้วย GitHub Actions (backend job + frontend job + service container ของ PostgreSQL)
     DoD: workflow เขียวบน GitHub และ PR ที่เทสพังต้องแดง
T-34 Cross-agent review ด้วย session ใหม่ เทียบ SPEC + openapi.yaml + โค้ดจริง
     DoD: มีรายงาน review เป็นข้อ ๆ และแก้หรือบันทึกเหตุผลครบทุกข้อ
