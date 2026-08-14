# Wisetsuk Booking Agent Context

## What this project is
ระบบจองห้องพักหอพักวิเศษสุขนครคอนโดและหอพักในเครือ (หลายสาขา): ค้นห้อง จอง แจ้งชำระเงิน อนุมัติ จัดการสาขา/ห้อง
สิทธิ์ 4 ระดับ: Guest / Member / Admin / Super Admin

กฎเหล็ก:
- Admin ผูกกับสาขาเดียวเสมอ เห็นและแก้ได้เฉพาะข้อมูลสาขาตัวเอง ห้ามข้ามสาขาเด็ดขาด (Super Admin เห็นทุกสาขา)
- ห้อง 1 ห้องมีการจองรายเดือนที่ยังไม่จบได้ทีละ 1 ใบเท่านั้น
- จองรายวันคิดเงิน = ราคาห้อง × จำนวนคืน / จองรายเดือนเก็บแค่ค่าทำสัญญาตอนจอง (ค่าเช่าเก็บวันทำสัญญา)
- อนุมัติจองรายเดือน → ห้องเป็น occupied / ยกเลิก → คืนเป็น available อัตโนมัติ เว้นห้องที่อยู่ maintenance

## Tech stack
- Backend: Go 1.26, Gin v1.12 (+ gin-contrib/cors), PostgreSQL 16 ผ่าน pgx v5 (เขียน SQL เอง ไม่ใช้ ORM)
- Backend Tests: go testing + net/http/httptest
- Runtime: Docker Compose (api, db, pgadmin, seed)
- Frontend: React
- API contract: docs/openapi.yaml คือแหล่งความจริงเดียว

## Project layout
- backend/cmd/api/            main.go entrypoint
- backend/cmd/seed/           ใส่ข้อมูลตั้งต้น (3 สาขา + บัญชีผู้ดูแล)
- backend/internal/routes/    โครง URL ทั้งระบบ + ระดับสิทธิ์ของแต่ละ module
- backend/internal/server/    สร้าง module + ต่อสายพึ่งพา + ตั้งค่า http.Server
- backend/internal/modules/   account, branch, room, booking, reporting (module ละ 5 ไฟล์)
- backend/internal/shared/    types (enum ข้าม module), access (สิทธิ์รายสาขา), audit (activity log)
- backend/internal/database/  pool, TxManager, Binder + migrations/*.sql
- backend/internal/           auth, middleware, config, httpx, storage, validate
- backend/README.md           สถาปัตยกรรม + ตาราง endpoint ทั้งหมด ← อ่านก่อนเริ่มงานทุกครั้ง

## Rules (must follow)
1. Plan ก่อน code ห้ามเขียนโค้ดก่อนเสนอแผนสั้น ๆ ให้ human เห็น
2. ทำทีละ 1 งานตามที่ตกลงกันไว้ ห้ามทำเกินขอบเขต
3. Test ต้องมีก่อนหรือพร้อมโค้ดเสมอ และห้ามลบ/แก้ test เพื่อให้ผ่าน
4. Business rule บังคับใช้ที่ database constraint ไม่ใช่แค่ใน application code
5. ทุก endpoint ต้องตรงกับตารางใน backend/README.md ถ้าต้องเปลี่ยน ให้แก้เอกสารก่อนแล้วถาม human
6. Error response ใช้รูปแบบเดียวทั้งระบบ: {"error": {"code": "...", "message": "...", "fields": {...}}}
7. มีคำถามหรือความกำกวม ให้ถาม human ก่อน ห้ามเดา

## Commands
- Check ก่อนบอกว่าเสร็จ: cd backend && gofmt -l . && go build ./... && go vet ./...
- Run ทั้ง stack:        cd backend && docker compose up -d --build   (api :8080, pgAdmin :5050)
- Rebuild เฉพาะ API:     cd backend && docker compose up -d --build api   (DB/volume ไม่ถูกแตะ)
- Seed ข้อมูลตั้งต้น:     cd backend && docker compose run --rm seed
- Logs:                  cd backend && docker compose logs -f api
- Query DB:              cd backend && docker compose exec -T db psql -U wisetsuk -d wisetsuk -c "SELECT ..."
- **ห้ามรันเอง**: `docker compose down -v` ลบข้อมูลทั้งหมด ต้องถามก่อน

บัญชีทดสอบ: `super@wisetsuk.com` · `admin.1@wisetsuk.com` (ประชาอุทิศ 45) · `admin.2` (บางแค) · `admin.3` (เจริญกรุงเพลส) — รหัสผ่านอยู่ที่ `SEED_DEFAULT_PASSWORD` ใน `.env`

