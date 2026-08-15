# Backend — ระบบจองห้องพัก หอพักวิเศษสุขนครคอนโด และหอพักในเครือ

REST API เขียนด้วย **Go + PostgreSQL** รันด้วย **Docker** สร้างตามสเปกใน `Prototype/ตัวอย่าง Website.pdf`
รองรับสิทธิ์ผู้ใช้ 4 ระดับ: ผู้ใช้ทั่วไป (Guest) / สมาชิก (Member) / ผู้ดูแลระบบ (Admin) / หัวหน้าผู้ดูแลระบบ (Super Admin)

---

## เริ่มใช้งาน (Docker)

```bash
cd backend
cp .env.example .env          # แก้ POSTGRES_PASSWORD และ JWT_SECRET ก่อน
docker compose up -d --build  # ขึ้น API + PostgreSQL
docker compose run --rm seed  # ใส่ข้อมูลตั้งต้น (3 สาขา + บัญชีผู้ดูแล)
```

ตรวจว่าใช้งานได้: <http://localhost:8080/health>
pgAdmin (จัดการฐานข้อมูล): <http://localhost:5050>

**บัญชีตั้งต้น** (รหัสผ่านมาจาก `SEED_DEFAULT_PASSWORD` — เปลี่ยนทันทีหลัง login ครั้งแรก)

| บัญชี | สิทธิ์ | สาขาที่ดูแล |
|---|---|---|
| `super@wisetsuk.com` | Super Admin | ทุกสาขา |
| `admin.1@wisetsuk.com` | Admin | ประชาอุทิศ 45 |
| `admin.2@wisetsuk.com` | Admin | บางแค |
| `admin.3@wisetsuk.com` | Admin | เจริญกรุงเพลส |

คำสั่งที่ใช้บ่อย:

```bash
docker compose logs -f api      # ดู log
docker compose down             # หยุด (ข้อมูลยังอยู่ใน volume)
docker compose down -v          # หยุดและลบข้อมูลทั้งหมด
```

### จัดการฐานข้อมูลด้วย pgAdmin

pgAdmin ขึ้นมาพร้อมกับ `docker compose up` — เปิดที่ <http://localhost:5050>

| | |
|---|---|
| อีเมล | ค่าจาก `PGADMIN_EMAIL` (ค่าเริ่มต้น `admin@wisetsuk.com`) |
| รหัสผ่าน | ค่าจาก `PGADMIN_PASSWORD` ในไฟล์ `.env` |

เซิร์ฟเวอร์ `Wisetsuk (Docker)` ถูกลงทะเบียนไว้ให้อัตโนมัติแล้ว (host `db`, port `5432`,
database `wisetsuk`, user `wisetsuk`) **ครั้งแรกที่กดเข้าจะถามรหัสผ่านฐานข้อมูลหนึ่งครั้ง** —
ใส่ค่า `POSTGRES_PASSWORD` แล้วติ๊ก *Save Password* ครั้งต่อไปจะเข้าได้เลย

> pgAdmin ผูกกับ `127.0.0.1` เท่านั้น เหมือนกับ PostgreSQL — เครื่องมือจัดการฐานข้อมูล
> ไม่ควรเข้าถึงได้จากเครือข่ายภายนอก ถ้าต้องเปิดให้คนอื่นใช้ ให้วางไว้หลัง VPN หรือ reverse proxy
> ที่มีการยืนยันตัวตน อย่าแก้เป็น `0.0.0.0` ตรง ๆ

ถ้าไม่อยากให้ pgAdmin ขึ้นมาด้วย สั่งเฉพาะบริการที่ต้องการได้:

```bash
docker compose up -d db api     # ไม่เอา pgAdmin
docker compose stop pgadmin     # ปิดทีหลัง
```

หรือจะใช้ `psql` ตรง ๆ ก็ได้:

```bash
docker compose exec db psql -U wisetsuk -d wisetsuk
```

### รันแบบไม่ใช้ Docker

ต้องมี Go 1.26+ และ PostgreSQL 16+ ที่รันอยู่แล้ว

```bash
cp .env.example .env    # ตั้ง DATABASE_URL ให้ชี้ไปที่ฐานข้อมูลของคุณ
go run ./cmd/seed       # migrate + ใส่ข้อมูลตั้งต้น
go run ./cmd/api
```

Migration รันอัตโนมัติทุกครั้งที่ API สตาร์ต — ไฟล์ที่รันแล้วจะถูกข้าม (ดู `internal/database/migrations/`)

---

## โครงสร้างโปรเจกต์

แบ่งเป็น **module ตามเรื่องของธุรกิจ** และ **ในแต่ละ module มีครบทั้ง 5 ชั้น**
งานเรื่องหนึ่งจึงแก้จบในโฟลเดอร์เดียว ไม่ต้องไล่เปิดทีละชั้น

```
backend/
├── cmd/
│   ├── api/                  จุดเริ่มของ API server (graceful shutdown)
│   └── seed/                 ใส่ข้อมูลตั้งต้นตาม prototype
│
├── internal/
│   ├── modules/              ── หนึ่งโฟลเดอร์ = หนึ่งเรื่อง ──
│   │   ├── account/          บัญชีผู้ใช้: สมัคร, เข้าสู่ระบบ, โปรไฟล์,
│   │   │   ├── model.go          แจ้งเตือน, สมาชิก, จัดการผู้ดูแลระบบ
│   │   │   ├── repository.go
│   │   │   ├── service.go
│   │   │   ├── handler.go
│   │   │   └── router.go
│   │   ├── branch/           สาขา: รายละเอียด, รูป, สิ่งอำนวยความสะดวก,
│   │   │   └── (5 ไฟล์เดียวกัน)  สถานที่ใกล้เคียง
│   │   ├── room/             ห้องพัก: ค้นหา, ประเภทห้อง, CRUD, สถานะห้อง
│   │   │   └── (5 ไฟล์เดียวกัน)
│   │   ├── booking/          การจอง + ชำระเงิน + อนุมัติ/ปฏิเสธ + นัดหมาย
│   │   │   └── (5 ไฟล์เดียวกัน)
│   │   └── reporting/        แดชบอร์ด + ประวัติการใช้งาน
│   │       └── (5 ไฟล์เดียวกัน)
│   │
│   ├── routes/               ผูก URL ทั้งระบบ + ระดับสิทธิ์ของแต่ละ module
│   ├── server/               ประกอบทุก module (ต่อสายพึ่งพา) + สร้าง HTTP server
│   │
│   ├── shared/               ── ของกลางที่ทุก module ใช้ ──
│   │   ├── types/            enum ข้าม module (Role, StayType)
│   │   ├── access/           กฎจำกัดสิทธิ์รายสาขา (ใช้ที่เดียวกันทุกที่)
│   │   └── audit/            เขียน activity log (ตัดขวางทุก module)
│   │
│   ├── database/             connection pool, migrate, transaction manager
│   │   └── migrations/       schema (.sql ฝังไว้ใน binary)
│   ├── auth/                 JWT + bcrypt + refresh token
│   ├── middleware/           ตรวจ token, จำกัดสิทธิ์, logging, recover
│   ├── config/               อ่าน env + ตรวจค่าที่จำเป็น
│   ├── httpx/                รูปแบบ response/error + ตัวช่วยอ่าน query
│   ├── storage/              รับไฟล์อัปโหลด — รูปเก็บในฐานข้อมูล, สลิปเก็บลงดิสก์
│   └── validate/             ตรวจข้อมูลเข้า พร้อมข้อความไทยรายฟิลด์
│
├── docker/pgadmin/           ตั้งค่าเซิร์ฟเวอร์ให้ pgAdmin อัตโนมัติ
├── Dockerfile
└── docker-compose.yml
```

### 5 ชั้นในแต่ละ module

| ไฟล์ | ทำ | ไม่ทำ |
|---|---|---|
| `model.go` | struct + enum ของเรื่องนี้ | ไม่มี logic |
| `repository.go` | SQL อย่างเดียว | ไม่ตัดสินใจอะไร |
| `service.go` | กฎธุรกิจ, ตรวจสิทธิ์, คุม transaction | ไม่แตะ `http.Request` |
| `handler.go` | อ่าน query/body → เรียก service → ตอบ JSON | ไม่เขียน SQL |
| `router.go` | ประกาศ route ของ module + ประกอบ 4 ชั้นข้างบน | ไม่มี logic |

`internal/routes/routes.go` บอกแค่ว่า **module ไหนขึ้นที่ระดับสิทธิ์ใด**
ส่วน endpoint รายตัวอยู่ใน `router.go` ของ module นั้น — เพิ่ม endpoint ไม่ต้องแก้ routes

`internal/server/server.go` เหลือหน้าที่เดียวคือ **สร้าง module แล้วต่อสายพึ่งพาให้ถูกลำดับ**
บวกกับตั้งค่า `http.Server` (timeout ต่าง ๆ)

### module คุยกันอย่างไร

module **ไม่ import ตัว repository ของกันและกัน** แต่ประกาศ interface สิ่งที่ตัวเองต้องการ
ไว้ฝั่งผู้ใช้งาน แล้วให้ `server.New()` เป็นคนต่อสายให้ เช่น `booking` ประกาศไว้ว่า

```go
type Rooms interface {
    LockForBooking(ctx, roomID) (*room.Room, error)
    IsAvailable(ctx, roomID, checkIn, checkOut) (bool, error)
    MarkOccupied(ctx, roomID, branchID) error
}
```

ทิศทางการพึ่งพาเป็นทางเดียวเสมอ จึงไม่มี import cycle:

```
reporting ──> room, booking
booking   ──> room, branch, account
account   ──> branch
branch    ──> (ไม่พึ่งใคร)
```

### transaction ข้าม module

การจอง 1 ครั้งต้องแตะทั้ง `rooms` (module room), `branches` (module branch)
และ `bookings` (module booking) ให้อยู่ใน transaction เดียวกัน

`database.TxManager` แก้เรื่องนี้ด้วยการ **ส่ง transaction ผ่าน context** —
repository ทุกตัวเรียก `db.Executor(ctx)` ซึ่งคืน transaction ที่เปิดค้างอยู่ถ้ามี
ไม่มีก็ใช้ pool ตามปกติ ทุก module จึงเข้าร่วม transaction เดียวกันได้โดยไม่ต้องรู้จักกัน

```go
s.tx.WithTx(ctx, func(ctx context.Context) error {
    rm, _  := s.rooms.LockForBooking(ctx, roomID)    // module room
    fee, _ := s.branches.ContractFee(ctx, branchID)  // module branch
    return s.repo.Create(ctx, ...)                   // module booking
})   // ล้มขั้นไหน rollback ทั้งหมดพร้อมกัน
```

`WithTx` ซ้อนกันได้ — ถ้า context มี transaction อยู่แล้วจะเข้าร่วมของเดิม ไม่เปิดใหม่

### ทำไม audit ไม่อยู่ใน module reporting

ทุก module ต้องเขียน activity log ได้ ถ้า recorder ไปอยู่ใน `reporting`
ทุก module จะต้อง import `reporting` — แต่ `reporting` ต้อง import `room` กับ `booking`
เพื่อสร้างแดชบอร์ด ก็จะเกิด import cycle ทันที

จึงแยกเป็น: **เขียน** log อยู่ที่ `shared/audit` (ใครก็เขียนได้)
ส่วน **อ่าน** log อยู่ที่ `modules/reporting/repository.go` เพราะเป็นเรื่องของการแสดงผล

---

## API

Base URL: `/api/v1` — ทุก response ห่อด้วย `{"data": ...}` และ error เป็น `{"error": {...}}`
รายการที่แบ่งหน้าจะมี `meta` เพิ่ม (`page`, `page_size`, `total_items`, `total_pages`)

ส่ง token ผ่าน header: `Authorization: Bearer <access_token>`

### สาธารณะ — ผู้ใช้ทั่วไป (Guest)

| Method | Path | คำอธิบาย |
|---|---|---|
| POST | `/auth/register` | สมัครสมาชิก |
| POST | `/auth/login` | เข้าสู่ระบบ (ใช้ร่วมกันทุกสิทธิ์) |
| POST | `/auth/refresh` | ขอ access token ใบใหม่ |
| POST | `/auth/logout` | ออกจากระบบ |
| GET | `/branches` | สาขาในเครือทั้งหมด |
| GET | `/branches/{id}` | รายละเอียดสาขา + รูป + สิ่งอำนวยความสะดวก + สถานที่ใกล้เคียง |
| GET | `/amenities` | รายการสิ่งอำนวยความสะดวกทั้งหมด |
| GET | `/room-types?branch_id=` | ประเภทห้องพัก |
| GET | `/rooms/search` | ค้นหาห้องพัก (ดูตัวกรองด้านล่าง) |
| GET | `/rooms/{id}` | รายละเอียดห้อง |

ตัวกรองของ `/rooms/search`: `branch_id`, `room_type_id`, `stay_type` (`daily`/`monthly`),
`check_in`, `check_out`, `move_in_date` (รูปแบบ `YYYY-MM-DD`), `min_price`, `max_price`, `page`, `page_size`

> คืนเฉพาะห้องที่ **จองได้จริง** — ตัดห้องที่สถานะไม่ว่าง และห้องที่มีการจองทับช่วงวันที่ออกแล้ว

### ต้องเข้าสู่ระบบ (ทุกสิทธิ์)

| Method | Path | คำอธิบาย |
|---|---|---|
| GET | `/me` | ข้อมูลผู้ใช้ปัจจุบัน |
| PUT | `/me` | แก้ไขข้อมูลส่วนตัว |
| POST | `/me/password` | เปลี่ยนรหัสผ่าน (เตะ session อื่นออกทั้งหมด) |
| GET | `/me/notifications` | รายการแจ้งเตือน + จำนวนที่ยังไม่อ่าน |
| POST | `/me/notifications/read` | ทำเครื่องหมายว่าอ่านแล้ว |

### สมาชิก (Member)

| Method | Path | คำอธิบาย |
|---|---|---|
| POST | `/bookings` | ทำรายการจอง (รายวัน / รายเดือน) |
| GET | `/bookings` | ประวัติการจองของตนเอง (กรอง `status`, `stay_type`) |
| GET | `/bookings/{id}` | รายละเอียดการจอง + สลิปล่าสุด |
| POST | `/bookings/{id}/payment` | แจ้งชำระเงิน + อัปโหลดสลิป (`multipart/form-data`) |
| POST | `/bookings/{id}/cancel` | ยกเลิกการจอง |

`POST /bookings/{id}/payment` รับเป็น multipart: `slip` (ไฟล์), `amount`, `transferred_at`, `note`

### ผู้ดูแลระบบ (Admin) และหัวหน้าผู้ดูแลระบบ

ทุก endpoint ใต้ `/admin` — **Admin เห็นและแก้ได้เฉพาะสาขาที่ตนรับผิดชอบ**, Super Admin เห็นทุกสาขา

| Method | Path | คำอธิบาย |
|---|---|---|
| GET | `/admin/dashboard` | ห้องว่างรายวัน/รายเดือน, รายการรอตรวจสอบ, กิจกรรมล่าสุด |
| GET | `/admin/bookings` | จัดการการจอง (กรอง `status`, `stay_type`, `search`) |
| POST | `/admin/bookings/{id}/approve` | อนุมัติการจอง |
| POST | `/admin/bookings/{id}/reject` | ปฏิเสธ (ต้องระบุ `reason`) |
| PUT | `/admin/bookings/{id}/appointment` | กำหนดวันนัดหมายทำสัญญา (รายเดือน) |
| GET | `/admin/rooms` | รายการห้องทุกสถานะ |
| POST | `/admin/rooms` | เพิ่มห้องพัก |
| PUT | `/admin/rooms/{id}` | แก้ไขห้องพัก |
| PATCH | `/admin/rooms/{id}/status` | อัปเดตสถานะห้องแบบ real-time |
| DELETE | `/admin/rooms/{id}` | ลบห้อง (soft delete เพื่อรักษาประวัติการจอง) |
| PUT | `/admin/branch` | แก้ไขรายละเอียดสาขา |
| PUT | `/admin/branch/amenities` | แก้ไขสิ่งอำนวยความสะดวก |
| PUT | `/admin/branch/nearby` | แก้ไขสถานที่ใกล้เคียง |
| POST | `/admin/branch/cover` | อัปโหลดรูปปกสาขา (multipart: `image`) |
| POST | `/admin/branch/images/upload` | อัปโหลดรูปสาขา (multipart: `image`, `caption`, `sort_order`) |
| POST | `/admin/branch/images` | เพิ่มรูปสาขาจาก URL ภายนอก (JSON: `image_url`, `caption`, `sort_order`) |
| DELETE | `/admin/branch/images/{id}` | ลบรูปสาขา |
| POST | `/admin/rooms/{id}/image` | อัปโหลดรูปห้อง (multipart: `image`) |
| GET | `/admin/members` | รายชื่อสมาชิก (ค้นด้วย `search`) |
| GET | `/admin/members/{id}/bookings` | ประวัติการจองของสมาชิกรายคน |
| GET | `/admin/activity-logs` | ประวัติการใช้งาน |

**เรื่องรูปภาพ:** รูปสาขาและรูปห้อง**เก็บเป็นเนื้อไฟล์ในฐานข้อมูล** (ตาราง `assets`) สำรองข้อมูล
ด้วย dump ชุดเดียวจึงได้ทั้งข้อมูลและรูป ไม่ต้องดูแล volume แยกต่างหาก อัปโหลดผ่าน endpoint
ข้างบน (multipart ช่อง `image` — JPG/PNG/WEBP ไม่เกิน `MAX_UPLOAD_MB`) แล้วระบบบันทึก
`image_url` / `cover_image_url` เป็น `{PUBLIC_BASE_URL}/files/{assetID}` ให้เอง

- `GET /files/{assetID}` เปิดสาธารณะ ใช้เป็น `src` ของ `<img>` ได้ตรง ๆ ตอบ `ETag` และ
  `Cache-Control: immutable` เพราะเนื้อไฟล์ผูกกับ id ตายตัว
- ชนิดไฟล์ถูกตรวจจากเนื้อไฟล์จริง ไม่เชื่อนามสกุลหรือ `Content-Type` ที่ client ส่งมา
- ไฟล์เนื้อเดียวกันที่อัปโหลดซ้ำใช้แถวเดิม (dedup ด้วย sha256) ไม่เก็บ blob ซ้ำ
- `POST /admin/branch/images` แบบ JSON ยังใช้ได้ สำหรับรูปที่โฮสต์ไว้ที่อื่นอยู่แล้ว —
  ตรวจแค่ว่าเป็น URL แบบ `http`/`https` ที่มี host จริง (กัน `javascript:` / `data:`)

**สลิปโอนเงิน** ยังเก็บเป็นไฟล์บนดิสก์ที่ `UPLOAD_DIR` และเสิร์ฟผ่าน `/uploads/*` เหมือนเดิม
เพราะเป็นหลักฐานการเงินที่ต้องคุมเองว่ามีจริงและไม่หาย

### หัวหน้าผู้ดูแลระบบ (Super Admin) เท่านั้น

| Method | Path | คำอธิบาย |
|---|---|---|
| GET | `/superadmin/staff` | รายชื่อผู้ดูแลระบบทั้งหมด |
| POST | `/superadmin/staff` | เพิ่มผู้ดูแลระบบ + กำหนดสาขาที่รับผิดชอบ |
| PUT | `/superadmin/staff/{id}` | แก้ไข / ระงับบัญชี / เปลี่ยนรหัสผ่าน |
| DELETE | `/superadmin/staff/{id}` | ลบผู้ดูแลระบบ |
| GET | `/superadmin/branches` | สาขาทั้งหมดรวมที่ปิดใช้งาน |

---

## สถานะการจอง

```
pending_payment   จองแล้ว รอแจ้งชำระเงิน
        ↓ สมาชิกอัปโหลดสลิป
awaiting_review   รอผู้ดูแลระบบตรวจสอบ
        ↓ แอดมินกดอนุมัติ / ปฏิเสธ
approved  หรือ  rejected
```

เพิ่มเติม: `cancelled` (ยกเลิก) และ `completed` (เข้าพักครบแล้ว)
เมื่อ **อนุมัติการจองรายเดือน** ระบบจะตั้งสถานะห้องเป็น `occupied` ให้อัตโนมัติ

การคิดเงิน:
- **รายวัน** — ราคาห้อง × จำนวนคืน
- **รายเดือน** — เก็บค่ายืนยันการทำสัญญาตอนจอง (`branches.contract_fee`) ส่วนค่าเช่าเก็บวันทำสัญญา

---

## สิ่งที่ทำไว้ด้านความปลอดภัย

- รหัสผ่านเก็บเป็น **bcrypt** (cost 12) ไม่มีการเก็บรหัสผ่านจริงที่ใดเลย
- **Refresh token rotation** — เก็บเฉพาะ SHA-256 hash ลง DB, ใช้ครั้งเดียวแล้วถูกเพิกถอนทันที
- เปลี่ยนรหัสผ่านแล้ว **session อื่นทั้งหมดถูกเตะออก**
- ข้อความ login ผิดเหมือนกันทุกกรณี + เผาเวลาเท่ากันเมื่อไม่พบอีเมล (กัน user enumeration)
- **จำกัดสิทธิ์รายสาขาที่ระดับ query** — คำสั่ง UPDATE/DELETE ของแอดมินมี `AND branch_id = ...` เสมอ
  ไม่ได้พึ่งการเช็คใน handler อย่างเดียว
- สมาชิกที่เรียกดูการจองของคนอื่นได้ **404** (ไม่ใช่ 403) เพื่อไม่ให้รู้ว่ารหัสจองนั้นมีจริง
- การจองใช้ **transaction + row lock** (`SELECT ... FOR UPDATE`) กันสองคนจองห้องเดียวกันพร้อมกัน
- ตรวจชนิดไฟล์อัปโหลดจาก **เนื้อไฟล์จริง** ไม่เชื่อ `Content-Type` ที่ client ส่งมา, ตั้งชื่อไฟล์ใหม่ด้วย UUID
- query ทั้งหมดใช้ **parameterized placeholder** ไม่มีการต่อ string ค่าจากผู้ใช้เข้า SQL
- error 5xx ไม่ส่งรายละเอียดภายในออกไปหา client (log ไว้ฝั่ง server เท่านั้น)
- container รันด้วย **non-root user**, พอร์ต PostgreSQL ผูกกับ `127.0.0.1` เท่านั้น

---

## หมายเหตุเรื่องข้อมูล

ข้อมูลจาก `cmd/seed` (ราคา, จำนวนชั้น, ห้องตัวอย่าง, สถานที่ใกล้เคียง) อ้างอิงจาก prototype
ซึ่งเอกสารระบุเองว่า *"ข้อมูลบางอย่างเป็นข้อมูลที่สมมุติขึ้นมา"* — ที่อยู่และเบอร์โทรของทั้ง 3 สาขาใช้ตามเอกสาร
ส่วนราคาและรายละเอียดห้องควรแก้ให้ตรงความจริงผ่านหน้า "จัดการรายละเอียดสาขา" ก่อนใช้งานจริง
