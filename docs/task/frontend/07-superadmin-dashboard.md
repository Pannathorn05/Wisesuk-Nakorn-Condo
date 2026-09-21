# Phase FE-7 — แดชบอร์ดของหัวหน้าผู้ดูแลระบบ (Super Admin)

อ้างอิง: `frontend/prototype/ตัวอย่าง Website v2.pdf` หน้า 39 (รูปภาพที่ 52)

| รูปภาพ | หน้า PDF | เนื้อหา |
|---|---|---|
| 52 | 39 | หน้าแดชบอร์ดของหัวหน้าผู้ดูแลระบบ — sidebar ซ้าย (ชื่อผู้ใช้ + เมนู 4 รายการ + Logout) และเนื้อหาขวา (การ์ดสรุปรายสาขา 3 สาขา + กล่องกิจกรรมล่าสุด) |
| 51 | 39 | (ทำไปแล้วใน FE-26…FE-29) หน้า login ที่เป็นทางเข้าของหน้านี้ — ดู [06-admin-login.md](06-admin-login.md) |

> ⚠️ **ต้องขยายขอบเขตก่อน implement**
> รอบที่แล้วขยายขอบเขตให้ทำได้**เฉพาะหน้า login + guard** เท่านั้น (`docs/c.md` ข้อ 5 และ
> `.claude/CLAUDE.md` ยังเขียนว่า "หน้าแดชบอร์ด/หน้าจัดการหลัง login ยังห้ามทำ ให้เป็น placeholder ไปก่อน")
> เอกสารนี้เป็นการแตกงานล่วงหน้า — ถ้าจะลงมือจริงต้องแก้ 2 ไฟล์นั้นให้อนุญาตหน้าแดชบอร์ดก่อน

DoD ทุก task ที่ดึงข้อมูลจาก API ต้องผ่าน **E2E checklist ครบ 4 สถานะ: loading / empty / error (มีปุ่มลองใหม่) / success**

---

## Section map จาก prototype (ภาพ 52)

**Sidebar (คงที่ ซ้ายมือ เต็มความสูง)**
1. บนสุด: ชื่อผู้ใช้ตัวหนา "สมชาย วิเศษสุข" + บรรทัดรอง "Superadmin" (มีเส้นคั่นใต้)
2. เมนู 4 รายการ พร้อมไอคอนหน้าข้อความ — **แดชบอร์ด** (active: pill สีน้ำตาลเข้ม ตัวอักษรขาว), จัดการผู้ดูแลระบบ, จัดการสาขา, activity log
3. ล่างสุด (ดันชิดล่าง มีเส้นคั่นเหนือ): ปุ่ม **Logout** พร้อมไอคอนลูกศรออกจากกล่อง

**เนื้อหาหลัก (ขวา)**
1. h1 "แดชบอร์ด" + บรรทัดรอง "สรุปข้อมูลประจำวันที่ 22 มิถุนายน 2569" (วันที่ไทย พ.ศ.)
2. **การ์ดต่อสาขา — 1 ใบต่อ 1 สาขา (ในภาพมี 3 ใบ ตรงกับจำนวนสาขา active)** แต่ละใบ:
   - แถบหัวการ์ด: ชื่อสาขา (เช่น "วิเศษสุขนครคอนโด ประชาอุทิศ 45")
   - ข้างใน 3 ช่องเรียงแนวนอน แต่ละช่อง: ชื่อรายการ (ซ้ายบน) + ไอคอน (ขวาบน) + ตัวเลขใหญ่ (ซ้ายล่าง) + ตัวหาร (ขวาล่าง)
     | ช่อง | ตัวเลขใหญ่ | ตัวหาร |
     |---|---|---|
     | ห้องว่างรายวันวันนี้ | 2 | / 2 ห้อง |
     | ห้องว่างรายเดือนวันนี้ | 12 | / 118 ห้อง |
     | รอตรวจสอบการจอง | 3 | / 12 รายการ |
3. **กล่อง "กิจกรรมล่าสุด"** (เต็มความกว้าง ใต้การ์ดสาขา): หัวกล่อง + รายการแบบ bullet
   แต่ละบรรทัด: ข้อความกิจกรรม ("Admin ประชาอุทิศ 45 อนุมัติการจอง รายวัน") + เวลาใต้ข้อความ ("2026-03-12 13:34")

**หมายเหตุ layout**: หน้านี้ใช้โครงคนละแบบกับทั้ง `GuestLayout` และ `AdminAuthLayout` ที่ทำไว้แล้ว
(มี sidebar, ไม่มี header แถบบน, ไม่มี footer) → ต้องทำ layout ที่ 3

---

## API จริงที่มีอยู่แล้ว (`docs/openapi.yaml`) — พอสำหรับทั้งหน้านี้ ไม่ต้องเดา endpoint ใหม่

- **`GET /api/v1/admin/dashboard` ตัวเดียวจบทั้งหน้า** คืน
  `{ branches: BranchSummary[], recent_activities: ActivityLog[] }`
  และ description ใน spec เขียนไว้ตรง ๆ ว่า *"Admin ได้ BranchSummary ของสาขาตัวเองรายการเดียว ส่วน Super Admin ได้ครบทุกสาขา"*
  → **ตรงกับภาพ 52 ที่หัวหน้าผู้ดูแลเห็นครบ 3 สาขาพอดี ไม่ต้องยิงรายสาขาเอง ไม่ต้องรวมข้อมูลเองฝั่ง client**
- `BranchSummary` map ตรงกับการ์ด 1:1 (spec ใส่ description กำกับเองว่า field ไหนเป็นตัวเศษ/ตัวส่วนของการ์ดใบไหน):
  `branch_name` · `daily_rooms_free` / `daily_rooms_total` · `monthly_rooms_free` / `monthly_rooms_total` · `pending_review` / `bookings_total`
- `recent_activities` = **กิจกรรมล่าสุด 10 รายการ** ในสาขาที่มีสิทธิ์เห็น เรียงใหม่→เก่า (มาพร้อม dashboard เลย ไม่ต้องยิง `/admin/activity-logs` ซ้ำในหน้านี้)
- `ActivityLog` มี `actor_name`, `actor_role`, `branch_name` (**join มาให้แล้ว ไม่ต้องยิงถามสาขาซ้ำ**), `action` (enum), `created_at`, `entity_type`, `entity_id`
- ชื่อ + role ที่ sidebar: มีอยู่แล้วใน `authStorage.getCurrentUser()` ตั้งแต่ตอน login (FE-28 เก็บ `user` ไว้แล้ว) — **ไม่ต้องยิง `GET /api/v1/me` ซ้ำตอนเปิดหน้า** (มี endpoint นี้อยู่ ถ้าอยากรีเฟรชข้อมูลโปรไฟล์ทีหลังค่อยใช้)
- Logout: `POST /api/v1/auth/logout` + `clearTokens()` — reuse logic เดียวกับปุ่มใน `Header` (FE-24) ห้ามเขียนใหม่

---

## Gap ที่เจอ — ต้องตัดสินใจก่อน implement ไม่ใช่เดาเอง

- **`action` เป็น enum อังกฤษ (`booking.approve`, `room.update`, `auth.login`, …) แต่ prototype โชว์เป็นไทย** ("อนุมัติการจอง") → ต้องทำตารางแปล enum → ข้อความไทยให้**ครบทุกค่าใน enum** (มี ~5 กลุ่ม: auth / user / booking / room / branch / admin) พร้อม fallback เป็นค่า enum ดิบถ้าเจอค่าใหม่ที่ backend เพิ่มมาทีหลัง **ห้ามโชว์ enum เปล่า ๆ ให้ผู้ใช้เห็น และห้ามแปลมั่วจนความหมายเพี้ยน**
- **ข้อความกิจกรรมในภาพลงท้ายด้วย "รายวัน"** ("Admin ประชาอุทิศ 45 อนุมัติการจอง **รายวัน**") แต่ `ActivityLog` **ไม่มี field ประเภทการเข้าพัก (stay_type) เลย** (เช็คทุก field แล้ว มีแค่ entity_type/entity_id ที่ชี้ไปที่ `bkg-xxx`) → **ห้ามเดาหรือยิง API เพิ่มเพื่อไปหา stay_type ของ booking ทีละใบ** ให้ประกอบข้อความจากเท่าที่มีจริง: `{actor_name} {branch_name} {action ที่แปลแล้ว}` แล้วปล่อยส่วน "รายวัน" ไป (บันทึกเป็น gap ไว้ว่าถ้าอยากได้ต้องขอ backend เพิ่ม field)
- **"สรุปข้อมูลประจำวันที่ 22 มิถุนายน 2569" ไม่มี field จาก API** — เป็นวันที่ที่ผู้ใช้เปิดดูหน้านี้ ใช้วันที่ปัจจุบันฝั่ง client แล้ว format เป็นไทย พ.ศ. (`toLocaleDateString("th-TH", { day: "numeric", month: "long", year: "numeric" })`) **ห้ามไป hardcode วันที่ในภาพ**
- **เมนู sidebar อีก 3 รายการ (จัดการผู้ดูแลระบบ / จัดการสาขา / activity log) ยังไม่มีหน้า** — backend มี endpoint รองรับครบแล้ว (`/superadmin/staff`, `/superadmin/branches`, `/admin/activity-logs`) แต่ prototype ของหน้าพวกนี้อยู่หน้าอื่นและยังไม่ได้แตก task → รอบนี้ทำเป็น **เมนูที่กดได้ไปหน้า placeholder** (ไม่ใช่ปุ่มตาย ๆ กดไม่ได้) แล้วค่อยแตก task ทีละหน้า
- **หน้านี้เป็นของ superadmin — แล้ว admin ล่ะ?** `GET /admin/dashboard` ใช้ endpoint เดียวกันและ admin จะได้ `branches` แค่สาขาตัวเอง → โครงหน้าเดียวกัน**ใช้ซ้ำกับ admin ได้เลย** แต่ prototype ของ admin dashboard อยู่คนละหน้า (ยังไม่ได้เปิดดู) เมนู sidebar ของ admin ก็น่าจะไม่เหมือนกัน → **task ชุดนี้ทำเฉพาะ `/superadmin` ก่อน** ถ้าจะใช้ซ้ำกับ `/admin` ต้องเปิดดู prototype หน้านั้นแล้วแตก task เพิ่ม
- **ตัวเลขในภาพ (2/2, 12/118, 3/12) เป็นตัวเลขตัวอย่างในภาพเท่านั้น** ห้ามยึดเป็นค่าที่ต้องได้ — ต้องมาจาก API จริง และต้องเทสกับข้อมูลจริงที่ backend มีตอนนี้ (ซึ่งอาจเป็น 0 ทั้งหมด) ให้แน่ใจว่าเลข 0 แสดงเป็น "0" ไม่ใช่ช่องว่าง

---

**FE-30 · Layout โซนผู้ดูแลแบบมี sidebar (`AdminShell`) + route**
- ทำ: สร้าง layout ที่ 3 ของโปรเจกต์ — sidebar ซ้ายเต็มความสูง (ชื่อผู้ใช้ + role จาก `authStorage.getCurrentUser()`, เมนู, Logout ชิดล่าง) + พื้นที่เนื้อหาขวาแบบ `<Outlet/>` · ย้าย route `/superadmin` มาใช้ layout นี้แทน `AdminAuthLayout` (ที่ตอนนี้ใช้ครอบ placeholder ชั่วคราวอยู่) · ยังคงครอบด้วย `RequireRole roles={["superadmin"]}` เดิม
- ขึ้นกับ: FE-26…FE-28 (layout/guard/authStorage ที่มี role แล้ว)
- DoD: เข้า `/superadmin` เห็น sidebar ตามภาพ 52 (ชื่อ-role จริงของบัญชีที่ login ไม่ใช่ชื่อในภาพ) · เมนู "แดชบอร์ด" เป็น active state · หน้า Guest และ `/admin/login` ยังใช้ layout เดิมไม่เปลี่ยน · ยังไม่ login เข้า `/superadmin` ยังถูกเด้งไป `/admin/login` เหมือนเดิม

**FE-31 · การ์ดสรุปรายสาขา — เชื่อม `GET /api/v1/admin/dashboard`**
- ทำ: hook `useAdminDashboard` (ผ่าน `api/adminApi.jsx` ใหม่ + `useAsync` เดิม ตาม pattern `useBranches`) → หัวข้อ "แดชบอร์ด" + วันที่ไทยฝั่ง client → render การ์ด 1 ใบต่อ 1 รายการใน `branches` แต่ละใบมี 3 ช่องตาม section map (map field ตาม description ใน spec ตรง ๆ ห้ามสลับตัวเศษ/ตัวส่วน)
- ขึ้นกับ: FE-30
- DoD: E2E ครบ — loading = skeleton การ์ด, error = ข้อความ + ปุ่มลองใหม่, empty = `branches` ว่าง เห็นข้อความชัดเจนไม่ใช่หน้าเปล่า, success = ตัวเลขทุกช่องตรงกับ response จริง (เทียบ curl) · หัวหน้าผู้ดูแลเห็นครบทุกสาขา · เลข 0 แสดงเป็น "0"

**FE-32 · กล่อง "กิจกรรมล่าสุด" + แปล action enum เป็นไทย**
- ทำ: render `recent_activities` จาก response เดียวกัน (ไม่ยิง API ซ้ำ) เป็นรายการ: ข้อความกิจกรรม + เวลาใต้ข้อความ · สร้าง map `ActivityAction` → ข้อความไทยให้ครบทุกค่าใน enum + fallback · ประกอบข้อความจาก field ที่มีจริงเท่านั้นตาม Gap (ไม่ใส่ "รายวัน" ที่ไม่มีข้อมูลรองรับ) · format `created_at` เป็น `YYYY-MM-DD HH:mm` ตามภาพ
- ขึ้นกับ: FE-31
- DoD: E2E ครบ — empty = ยังไม่มีกิจกรรม เห็นข้อความแทนกล่องว่าง · success = จำนวนรายการและข้อความตรงกับ response จริง · ทุกค่า `action` ที่ backend ส่งมาจริงต้องแปลเป็นไทยได้ (ไล่เช็คกับ enum ในสเปกให้ครบ ไม่ใช่เช็คแค่ค่าที่บังเอิญมีในข้อมูลตอนนี้)

**FE-33 · เมนู sidebar + Logout + หน้า placeholder ของเมนูที่เหลือ**
- ทำ: ผูกเมนู 4 รายการกับ route จริง (`/superadmin` = แดชบอร์ด, อีก 3 เมนูไปหน้า placeholder ที่ยังไม่ได้แตก task) โดย active state ตาม route ปัจจุบัน · ปุ่ม Logout ใช้ logic เดียวกับใน `Header` (`authApi.logout` best-effort → `clearTokens()` → ไป `/admin/login`)
- ขึ้นกับ: FE-30
- DoD: กดเมนูแล้วเปลี่ยนหน้าได้จริงทุกปุ่ม ไม่มีปุ่มตาย · active state ถูกต้องตามหน้าที่อยู่ · กด Logout แล้ว token ถูกล้างหมด (รวม user) กลับไปหน้า login และเข้า `/superadmin` ซ้ำไม่ได้อีก

**FE-34 · Responsive + Visual QA**
- ทำ: ไล่เทียบกับ prototype ภาพ 52 ทีละจุด · sidebar บนจอแคบต้องยุบ/ซ่อนอย่างมีเหตุผล (prototype ไม่ได้ให้ภาพ mobile ของหน้านี้ — เลือกวิธีที่ไม่ทำให้เนื้อหาล้น แล้วบันทึกเหตุผลไว้) · การ์ด 3 ช่องต่อสาขาต้องยุบเป็นแนวตั้งบน mobile โดยไม่ overflow
- ขึ้นกับ: FE-30…FE-33
- DoD: ตรวจ responsive จริงที่ 390/820/1440px (device metrics override ไม่ใช่แค่ย่อหน้าต่าง) ไม่มี horizontal scroll · `npm run build` ผ่านไม่มี warning · `npm test` เดิมยังผ่าน · ไม่มีไฟล์ backend ถูกแก้ (เช็ค git diff) · ไม่มี token/รหัสผ่านบัญชีทดสอบหลุดเข้า git
