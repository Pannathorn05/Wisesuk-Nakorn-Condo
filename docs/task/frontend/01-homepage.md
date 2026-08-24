# Phase FE-1 — Homepage (Guest)

อ้างอิง: `frontend/prototype/ตัวอย่าง Website v2.pdf` หน้า 1 — รูปภาพที่ 1 ("เป็นหน้า Homepage ที่เมื่อทุกคนกด link เข้ามาก็จะเจอหน้านี้") และรูปภาพที่ 2 (ต่อจากรูปที่ 1)
หมายเหตุ: รูปภาพที่ 1 และ 2 ในไฟล์ PDF เป็นภาพเดียวกันทุก byte (ทุกคอลัมน์ก็เหมือนกันทั้ง 4 คอลัมน์ด้วย) — ให้ยึด mockup เดียวเป็น source of truth พอ ไม่ต้องหาความต่างระหว่างคอลัมน์/รูป

DoD ทุก task ที่ดึงข้อมูลจาก API ต้องผ่าน **E2E checklist 4 สถานะ: loading / empty / error (มีปุ่มลองใหม่) / success**
ขึ้นกับ `docs/c.md` และ `.claude/CLAUDE.md`: ห้ามแก้ backend, ห้าม hardcode ข้อมูลที่ API มีจริง, ห้าม mock แทน API จริง

---

## Section map จาก prototype

1. Header: โลโก้ "วิเศษสุขนครคอนโด และหอพักในเครือ" + nav (หน้าหลัก / สาขาของเรา / ค้นหาห้องพัก / ติดต่อเรา) + ปุ่ม "เข้าสู่ระบบ" / "สมัครสมาชิก"
2. Hero: หัวข้อ "หอพักวิเศษสุขนคร คอนโด" + subtext 3 บรรทัด + ปุ่ม "ค้นหาห้องพัก" / "ดูสาขาทั้งหมด" + carousel รูปห้อง (prev/next + dot 8 จุด) บน background รูปตึก
3. Facilities strip: 4 ไอคอน — คีย์การ์ด / รปภ. 24 ชั่วโมง / ที่จอดรถ / ห้องน้ำในตัว
4. "สาขาของเรา": การ์ดสาขา 3 ใบ (รูป, ชื่อ, ที่อยู่, ลิงก์ "ดูรายละเอียด →")
5. "ประเภทห้องพัก": การ์ด 2 ใบ — ห้องพักรายวัน (500–950 บาท/คืน), ห้องพักรายเดือน (2,200–7,700 บาท/เดือน) พร้อมลิงก์ "ดูรายละเอียด →"
6. CTA แถบสีน้ำตาลเข้ม: "พร้อมจองห้องพักแล้วหรือยัง?" + ปุ่ม "เข้าสู่ระบบ" / "สมัครสมาชิก"
7. Footer: โลโก้+ชื่อ, เมนู, รายชื่อสาขา, ติดต่อเรา (LINE, เบอร์โทรต่อสาขา), ลิขสิทธิ์

## Gap ที่เจอ — ตรวจกับ backend ที่รันจริง (`curl localhost:8080`) ไม่ใช่แค่ `docs/openapi.yaml`

`docs/openapi.yaml` เก่ากว่า backend ที่รันจริงพอสมควร (backend มี field เพิ่มมาเยอะ) — ตามกฎ "Backend ที่มีอยู่แล้วเป็น source of truth" จึงยึด response จริงจาก `GET /api/v1/branches`, `/amenities`, `/room-types`, `/rooms/search` ที่ curl ทดสอบแล้วเป็นหลัก ทุก list endpoint คืนเป็น `{ "data": [...] }` (ไม่ใช่ array เปล่าอย่างที่ openapi ระบุ)

**แก้แล้วจากที่เคยกังวลไว้:**
- `Branch` จริงมี `cover_image_url`, `daily_price_from`, `monthly_price_min`, `monthly_price_max` อยู่แล้ว — **ไม่ต้องรอ backend เพิ่ม field** แต่ข้อมูลจริงตอนนี้ยังว่างเปล่า (`cover_image_url: ""`) เพราะยังไม่มีการอัปโหลดรูป → เป็น empty-state ของรูป ไม่ใช่ field ที่ไม่มี ให้ fallback เป็น placeholder เมื่อค่าว่าง
- ช่วงราคา "รายวัน / รายเดือน" หน้า Home **ไม่ต้องยิง `rooms/search` เพิ่ม** — คำนวณจาก field `daily_price_from`, `monthly_price_min`, `monthly_price_max` ที่ `GET /api/v1/branches` ให้มาแล้ว โดย aggregate ข้าม branch ที่ `is_active=true` ทั้งหมด (min ของ daily_price_from, min ของ monthly_price_min, max ของ monthly_price_max) — ข้อมูลจริงทั้งหมด ไม่ hardcode

**ยังต้องตัดสินใจเป็น static:**
- แถบ facilities 4 ไอคอนบนสุด (คีย์การ์ด / รปภ. 24 ชั่วโมง / ที่จอดรถ / ห้องน้ำในตัว) **ไม่ใช่ข้อมูลชุดเดียวกับ `GET /api/v1/amenities`** — endpoint นั้นเป็น catalog สิ่งอำนวยความสะดวกระดับห้อง/สาขา (tv, aircon, wifi, fridge, water-heater, furniture, keycard, cctv, lift, parking, laundry, convenience-store) ใช้แสดงในหน้ารายละเอียดสาขา/filter ห้อง ไม่ใช่ marketing strip นี้ — 2 ใน 4 อย่างที่ prototype โชว์ ("รปภ. 24 ชั่วโมง", "ห้องน้ำในตัว") ไม่มีอยู่ใน catalog นี้เลย ดึงมาแปะแบบเดายัดใส่กันจะผิดความหมาย → ทำเป็น **static content ตาม prototype ตรง ๆ** (ไม่ผูก API) เพราะไม่มี endpoint ที่ตรงความหมายจริง ๆ ให้ผูก

---

**FE-00 · ตั้งค่า API client พื้นฐาน**
- ทำ: `api/client.js` เดียวคุม base URL (env var) + error handling กลาง, `branchApi.js`, `roomApi.js` ตาม endpoint ที่มีจริงใน `docs/openapi.yaml` (`/branches`, `/room-types`, `/amenities`, `/rooms/search`)
- ขึ้นกับ: —
- DoD: ไม่มีการเรียก `fetch`/`axios` ตรงในหน้า component เลย (ต้องผ่าน api/service เท่านั้น), ยิง `GET /api/v1/branches` จริงผ่าน client แล้วได้ผลลัพธ์

**FE-01 · Layout: Header + Footer ใช้ร่วมทุกหน้า Guest**
- ทำ: component `Header` (โลโก้ + nav 4 เมนู + ปุ่ม เข้าสู่ระบบ/สมัครสมาชิก, active state ตาม route ปัจจุบัน, hamburger บน mobile) และ `Footer` (โลโก้, เมนู, รายชื่อสาขา, ติดต่อเรา, ลิขสิทธิ์) — เบอร์โทรต่อสาขาใน footer ดึงจาก field `phones` ของ `GET /api/v1/branches` จริง
- ขึ้นกับ: FE-00
- DoD: E2E checklist ครบ (loading = skeleton แถบสาขาใน footer, empty = ไม่มีสาขา active ต้องไม่พังเป็นหน้าขาว, error = ปุ่มลองใหม่, success = จำนวน/เบอร์ตรงกับ API) · nav ทุกลิงก์คลิกแล้วเปลี่ยน route จริง (route ที่ยังไม่มีหน้าให้เตรียม placeholder ไว้ก่อน ไม่ error) · ปุ่มเข้าสู่ระบบ/สมัครสมาชิกไป `/login` `/register`

**FE-02 · Hero Section**
- ทำ: หัวข้อ/subtext (static ตาม prototype คำต่อคำ), ปุ่ม "ค้นหาห้องพัก" → หน้าค้นหาห้อง, "ดูสาขาทั้งหมด" → `/branches`, carousel รูปห้องมีปุ่ม prev/next + dot indicator ใช้งานได้จริง
- ขึ้นกับ: FE-01
- DoD: static content ไม่มีสถานะ API — เช็คว่าข้อความ/ปุ่มตรง prototype ทุกคำ · carousel กดปุ่ม/จุดแล้วรูปเปลี่ยนจริง วนลูปได้ทั้งสองทิศทาง · CTA ทั้งสองปุ่มนำทางถูก path · responsive ไม่ล้น/ไม่ทับกันบน mobile

**FE-03 · Facilities strip**
- ทำ: static content 4 รายการตาม prototype เป๊ะ ๆ (คีย์การ์ด / รปภ. 24 ชั่วโมง / ที่จอดรถ / ห้องน้ำในตัว) — **ไม่ผูก `GET /api/v1/amenities`** เพราะ endpoint นั้นเป็น catalog คนละความหมาย (ดู Gap ด้านบน) ไม่ใช่ marketing strip นี้
- ขึ้นกับ: FE-01
- DoD: ไม่มี API เรียก ไม่ต้องมี 4 สถานะ loading/empty/error/success — เช็คแค่ข้อความ/ไอคอน/ลำดับตรงกับ prototype ทั้ง 4 รายการ และ responsive ไม่ล้นบน mobile

**FE-04 · Section "สาขาของเรา"**
- ทำ: เชื่อม `GET /api/v1/branches` จริง แสดงเฉพาะ `is_active = true`, การ์ด (รูป placeholder ตาม Gap ด้านบน, name, address, ปุ่ม "ดูรายละเอียด →" ไป `/branches/:id`)
- ขึ้นกับ: FE-00
- DoD: E2E checklist ครบ — loading = skeleton การ์ด 3 ใบ, empty = ไม่มีสาขา active ต้องเห็นข้อความชัดเจนไม่ใช่หน้าว่าง, error = ปุ่มลองใหม่, success = จำนวนการ์ด/ชื่อ/ที่อยู่ตรงกับ API จริง (ไม่ hardcode ชื่อสาขา) · คลิก "ดูรายละเอียด" ไปหน้า detail ด้วย `id` ที่ถูกต้อง

**FE-05 · Section "ประเภทห้องพัก" (จริง ๆ คือแยกตาม stay_type: รายวัน/รายเดือน)**
- ทำ: การ์ด 2 ใบ "ห้องพักรายวัน" / "ห้องพักรายเดือน" — ราคาคำนวณจาก field `daily_price_from` / `monthly_price_min` / `monthly_price_max` ของ `GET /api/v1/branches` (aggregate ข้ามสาขา `is_active=true` — ใช้ผลลัพธ์เดียวกับที่ FE-04 ดึงมาแล้ว ไม่ยิงซ้ำ), ปุ่ม "ดูรายละเอียด →" ไปหน้าค้นหาห้องพร้อม query `stay_type=daily` หรือ `monthly`
- ขึ้นกับ: FE-04 (ใช้ผลลัพธ์ branches เดียวกัน)
- DoD: E2E checklist ครบ ผูกกับสถานะเดียวกับ FE-04 (loading/empty/error/success ของ `GET /api/v1/branches`) · ตัวเลขราคาต้องตรงกับ min/max จริงที่คำนวณจากสาขาที่ active ทั้งหมด ไม่ hardcode ตัวเลข · ปุ่มลิงก์ไปหน้าค้นหาพร้อม query filter ที่ถูกต้อง

**FE-06 · CTA "พร้อมจองห้องพักแล้วหรือยัง?"**
- ทำ: static content ตาม prototype, ปุ่ม "เข้าสู่ระบบ" → `/login`, "สมัครสมาชิก" → `/register`; ถ้าผู้ใช้ล็อกอินอยู่แล้ว (มี token ที่ยังไม่หมดอายุ) ให้ปรับ/ซ่อน CTA นี้ให้สมเหตุผลแทนที่จะชวน login ซ้ำ
- ขึ้นกับ: FE-01
- DoD: ปุ่มนำทางถูก path ทั้งสองปุ่ม · เช็ค state ล็อกอินแล้ว/ยังไม่ล็อกอิน ทั้งสองกรณีแสดงผลต่างกันตามที่ตกลง · ไม่มี API เรียกในส่วนนี้นอกจากอ่าน auth state ในเครื่อง

**FE-07 · Responsive + Visual QA ทั้งหน้า Homepage**
- ทำ: ไล่เทียบทุก section (FE-01…FE-06) กับ prototype ทีละจุด — spacing, สี, font-size/weight, border-radius, ขนาดปุ่ม/การ์ด บน breakpoint desktop / tablet / mobile
- ขึ้นกับ: FE-01, FE-02, FE-03, FE-04, FE-05, FE-06
- DoD: checklist เทียบ prototype ผ่านครบทุก section ที่ระบุไว้ใน `.claude/CLAUDE.md` (layout/font/color/image/spacing/button/alignment/card/border/radius/shadow/header/footer/responsive) · `npm run build` ผ่านไม่มี error/warning สำคัญ · ไม่มีไฟล์ backend ถูกแก้ (เช็ค git diff)
