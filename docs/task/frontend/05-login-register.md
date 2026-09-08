# Phase FE-5 — Login / Register (Guest)

อ้างอิง: `frontend/prototype/ตัวอย่าง Website v2.pdf` หน้า 14 (รูปภาพที่ 19–20)

| รูปภาพ | หน้า PDF | เนื้อหา |
|---|---|---|
| 19 | 14 | หน้าเข้าสู่ระบบ — 2 การ์ดซ้าย/ขวาในภาพเดียวกันคือ state เดียวกัน (ซ้าย = ซ่อนรหัสผ่าน, ขวา = กดไอคอนตาแล้วเห็นรหัสผ่านเป็นตัวอักษรจริง) ไม่ใช่ 2 หน้าคนละแบบ |
| 20 | 14 | หน้าสมัครสมาชิก — เช่นเดียวกัน 2 การ์ดคือ state เดียวกันของฟอร์มเดียว (toggle แสดง/ซ่อนรหัสผ่าน) |

ตรวจแล้ว: มีภาพเดียว(ต่อหน้า)สำหรับทั้ง flow ไม่มี tab/state อื่นให้แตกเพิ่มนอกจาก toggle แสดงรหัสผ่านที่เห็นในภาพอยู่แล้ว

DoD ทุก task ที่ยิง API ต้องผ่าน **E2E checklist ครบ 4 สถานะ: idle → submitting (ปุ่มโชว์ loading/disabled กันกดซ้ำ) → error (ข้อความจริงจาก backend) → success** — หน้านี้ไม่มี "empty state" แบบหน้าที่ดึงลิสต์ เพราะเป็นฟอร์ม ไม่ใช่หน้าดึงข้อมูลมาแสดง
ขึ้นกับ `docs/c.md`, `.claude/CLAUDE.md`: ห้ามแก้ backend, ห้ามเขียน authentication backend ใหม่, ห้ามสร้าง Member Dashboard/Profile/Booking flow ใด ๆ (นอกขอบเขตรอบนี้ตาม `docs/c.md` ข้อ 5), ต้องใช้ token/session mechanism เดิมที่มีอยู่แล้ว (`src/utils/authStorage.jsx`) ห้ามสร้างระบบเก็บ token ซ้อนใหม่

---

## Section map จาก prototype

**Login** (ภาพ 19):
1. หัวการ์ด: "เข้าสู่ระบบ" + subtext static "เข้าสู่ระบบเพื่อจองห้องพักออนไลน์"
2. ช่อง "อีเมล" — ไอคอนซองจดหมาย, placeholder `your@gmail.com`
3. ช่อง "รหัสผ่าน" — ไอคอนแม่กุญแจ, ไอคอนรูปตา (toggle แสดง/ซ่อน) ชิดขวาในช่อง
4. checkbox "จดจำฉัน" (ซ้าย) + ลิงก์ "ลืมรหัสผ่าน?" (ขวา) แถวเดียวกัน
5. ปุ่ม "เข้าสู่ระบบ" เต็มความกว้าง สีน้ำตาลเข้ม (primary)
6. ลิงก์ท้ายการ์ด: "ยังไม่มีบัญชี?สมัครสมาชิก" → ไป `/register`

**Register** (ภาพ 20):
1. หัวการ์ด: "สมัครสมาชิก" + subtext static "สร้างบัญชีเพื่อจองห้องพักออนไลน์"
2. แถวคู่: "ชื่อ" (placeholder "ชื่อจริง") | "นามสกุล" (placeholder "นามสกุล")
3. ช่อง "เบอร์โทรศัพท์" — ไอคอนโทรศัพท์, placeholder `08xxxxxxxx`
4. ช่อง "อีเมล" — ไอคอนซองจดหมาย, placeholder `your@gmail.com`
5. ช่อง "รหัสผ่าน" — ไอคอนแม่กุญแจ, ไอคอนตา toggle
6. ช่อง "ยืนยันรหัสผ่าน" — ไอคอนแม่กุญแจ, ไอคอนตา toggle
7. ปุ่ม "สมัครสมาชิก" เต็มความกว้าง สีน้ำตาลเข้ม
8. ลิงก์ท้ายการ์ด: "มีบัญชีแล้ว?เข้าสู่ระบบ" → ไป `/login`

ทั้งสองหน้า: การ์ดสีขาวลอยกลางพื้นหลังครีม (`--color-surface-alt`), header/footer เว็บเหมือนทุกหน้า (reuse `Header`/`Footer` เดิม ไม่ต้องแก้)

---

## API จริงที่มีอยู่แล้ว (`docs/openapi.yaml`) — เชื่อมตรงนี้ ห้ามเดา endpoint อื่น

- `POST /api/v1/auth/register` — body `RegisterRequest {email, password, first_name, last_name, phone}` (**`additionalProperties: false` — ห้ามส่ง field `role` แถมไปเด็ดขาด จะได้ 422**) → สำเร็จ `201` คืน `AuthResult {access_token, refresh_token, user}` (**สมัครแล้ว login ให้อัตโนมัติในตัว ไม่ต้องยิง login ซ้ำ**) · `400` bad_request, `422` validation_failed (มี `error.fields` เป็น map ชื่อ field → ข้อความไทย เช่น `{email: "อีเมลนี้ถูกใช้งานแล้ว"}` โชว์ใต้ช่องนั้นตรง ๆ ได้เลย)
- `POST /api/v1/auth/login` — body `LoginRequest {email, password}` → สำเร็จ `200` คืน `AuthResult` เหมือนกัน · `401` invalid_credentials ("อีเมลหรือรหัสผ่านไม่ถูกต้อง" — ข้อความเดียวกันทั้งอีเมลผิดและรหัสผ่านผิด ห้ามพยายามแยกว่า field ไหนผิดเพราะ backend ไม่บอก) · `403` account_disabled ("บัญชีนี้ถูกปิดใช้งาน")
- validation ฝั่ง backend ที่ต้อง mirror ไว้ฝั่ง client ด้วย (กันข้อมูลไม่ผ่านแล้วค่อยรู้ตอนกด submit): `password` ยาว 8–72 ตัวอักษร, `phone` ตรง pattern `^[0-9]{9,10}$` (ตัวเลขล้วน 9–10 หลัก ไม่มีขีด/เว้นวรรค/+66 — ตรงกับ placeholder `08xxxxxxxx` ใน prototype พอดี), `first_name`/`last_name` ห้ามว่าง
- `src/api/client.jsx` (`apiClient`, `ApiError`, `NetworkError`) มีอยู่แล้ว ห้ามเขียน fetch ใหม่ตรง component — ต้องสร้าง `src/api/authApi.jsx` เป็นจุดเดียวที่เรียก 2 endpoint นี้ (ตาม pattern `branchApi.jsx`/`roomApi.jsx` เดิม)
- `src/utils/authStorage.jsx` (`setTokens`, `clearTokens`, `getAccessToken`, `isLoggedIn`) เป็น token storage เดิมที่ scaffold รอไว้แล้ว (field name `access_token`/`refresh_token` ตรงกับ `AuthResult` เป๊ะ) — เรียกใช้ตรง ๆ **ห้ามสร้างที่เก็บ token ใหม่ซ้อน**
- `POST /api/v1/auth/logout` มีจริง (body `RefreshRequest {refresh_token}`, ตอบ `204` เสมอ) — ใช้ตอนทำปุ่ม logout ใน FE-24

---

## Gap ที่เจอ — ต้องตัดสินใจก่อน implement ไม่ใช่เดาเอง

- **ไม่มี endpoint "ลืมรหัสผ่าน" ใน `docs/openapi.yaml` เลย** (เช็คทั้งไฟล์แล้ว ไม่มี `forgot`/`reset-password` ใด ๆ) แต่ prototype มีลิงก์ "ลืมรหัสผ่าน?" โชว์อยู่ — **ห้ามเดา endpoint ใหม่หรือ mock flow ลืมรหัสผ่านขึ้นมาเอง** ทางเลือกที่ไม่ผิดกฎ: (ก) ให้ลิงก์นี้พาไปหน้า `/contact` ที่มีอยู่จริงแล้ว (ให้ติดต่อเจ้าหน้าที่โดยตรงแทน) หรือ (ข) แสดงลิงก์ไว้ตามภาพแต่กดแล้วไม่มีผล (disabled + tooltip "ติดต่อเจ้าหน้าที่เพื่อรีเซ็ตรหัสผ่าน") — แนะนำ (ก) เพราะยังพาไปหน้าที่ใช้งานได้จริง ไม่ใช่ dead end แต่ให้ยืนยันกับทีมก่อน implement
- **checkbox "จดจำฉัน" ไม่มี field คู่กันใน `LoginRequest`** (schema มีแค่ `email`/`password`) เป็นแค่ UI จาก prototype ไม่มีผลต่อ backend เลย — ปัจจุบัน `authStorage.jsx` ใช้ `localStorage` อยู่แล้ว (token อยู่ข้ามการปิดเบราว์เซอร์เสมอไม่ว่าจะติ๊กหรือไม่) ตัดสินใจ 2 ทาง: (ก) ใส่ checkbox ไว้เฉย ๆ ตามภาพ ไม่ผูก logic อะไร (ง่ายสุด ตรงกับสิ่งที่ backend รองรับจริงคือ "จำเสมอ") หรือ (ข) ทำให้มีผลจริงโดยตอนไม่ติ๊กให้เก็บ token ใน `sessionStorage` แทน (ต้องแก้ `authStorage.jsx` เพิ่ม parameter เลือก storage — กระทบไฟล์ที่ทุกหน้าพึ่งพา ต้อง regression test ทุกจุดที่เรียก `isLoggedIn()`) — แนะนำ (ก) ก่อนสำหรับรอบนี้ เพราะ (ข) เสี่ยงกระทบวงกว้างเกินขอบเขตที่ขอ
- **หน้า Login/Register ยังไม่มี guard กันคนที่ login อยู่แล้วเข้ามาซ้ำ** — ถ้า login สำเร็จแล้วมี user กด back หรือพิมพ์ `/login` เข้ามาตรง ๆ อีกครั้ง ควร redirect ออกไปหน้าอื่นทันที (เช่นหน้าแรก) ไม่ใช่โชว์ฟอร์ม login ซ้ำ — ไม่มีในภาพ prototype ให้เห็นชัด แต่เป็น behavior พื้นฐานที่ควรมีเพื่อไม่ให้ดูพัง ใส่ไว้ใน FE-22/FE-23 ด้วย
- **หลัง login/register สำเร็จ ควร redirect ไปที่ไหน** — ไม่มี Member Dashboard ให้ไปตาม `docs/c.md` ข้อ 5 (ห้ามสร้างรอบนี้) และปุ่ม login ปัจจุบัน (Header, GuestBookingModal) ยังไม่ได้ส่ง "จะกลับไปหน้าไหนต่อ" มาด้วย — รอบนี้ทำแบบง่ายสุดที่ไม่ผิดขอบเขต: redirect กลับหน้าแรก (`/`) เสมอหลัง login/register สำเร็จ (ยังไม่ทำ "จำหน้าที่มาจาก" — ถ้าต้องการ UX ที่ดีกว่านี้ เช่นกดจองห้อง → เด้ง login → login เสร็จกลับมาที่ห้องเดิม ต้องแตก task เพิ่มแยกต่างหาก ไม่รวมในรอบนี้)
- **Header ("เข้าสู่ระบบ"/"สมัครสมาชิก") ยังไม่มี state ตอน login แล้ว** — ตอนนี้ทั้ง 2 ปุ่มโชว์ตลอดไม่ว่าจะ login อยู่หรือไม่ (เช็คโค้ด `Header.jsx` แล้ว ไม่มีการเช็ค `isLoggedIn()` เลย) ถ้าไม่แก้ ผู้ใช้ login สำเร็จแล้วจะยังเห็นปุ่ม "เข้าสู่ระบบ" ค้างอยู่ ดูเหมือนระบบพัง prototype ไม่ได้โชว์หน้า header ตอน login แล้วให้เห็น (ภาพที่มีอยู่ทั้งหมดเป็น state ยังไม่ login) จึง**ไม่มี design ให้อ้างอิงตรง ๆ** — FE-24 เลยทำแบบขั้นต่ำสุดที่ไม่ต้องออกแบบใหม่: สลับปุ่ม 2 ปุ่มเดิมเป็นปุ่ม "ออกจากระบบ" ปุ่มเดียว (reuse class `.btn.btn-outline` เดิมเป๊ะ ไม่มี dropdown/เมนูสมาชิกใหม่ เพราะนั่นเข้าข่าย Member area ที่ห้ามทำ) ถ้าอยากได้ดีไซน์อื่นต้องขอภาพ prototype เพิ่มก่อน
- **ไอคอนซองจดหมาย/แม่กุญแจ/ตา (แสดง-ซ่อนรหัสผ่าน)/คน ที่เห็นในภาพ ยังไม่มีใน `src/components/icons/index.jsx`** (มีแค่ `IconPhone` ที่ใช้ซ้ำได้กับช่องเบอร์โทร) ต้องเพิ่มใหม่ 4 ตัว (`IconMail`, `IconLockClosed`, `IconEye`/`IconEyeOff`, `IconUser`) ตาม pattern เดิมทุกตัวในไฟล์ (stroke-based, ใช้ `base` object ร่วม, viewBox `0 0 24 24`) **ห้ามติดตั้ง icon library ใหม่** (ตรงกับเหตุผลที่ comment หัวไฟล์ icons เดิมบอกไว้)

---

**FE-21 · Auth API layer + shared form pieces**
- ทำ: สร้าง `src/api/authApi.jsx` (`register(body)`, `login(body)`, `logout(refreshToken)` — เรียกผ่าน `apiClient` ที่มีอยู่แล้ว ตาม pattern `branchApi.jsx`) · เพิ่มไอคอนที่ขาด 4 ตัวใน `src/components/icons/index.jsx` ตาม Gap ด้านบน · สร้าง `src/components/auth/` (โฟลเดอร์ใหม่ตาม `docs/c.md` ข้อ 10) เก็บ input ที่ใช้ซ้ำทั้ง 2 หน้า เช่น `PasswordInput` (ช่อง password + ปุ่มไอคอนตา toggle type text/password ในตัวเดียว ใช้ซ้ำได้ทั้ง login/register/confirm password)
- ขึ้นกับ: FE-00 (`apiClient`), `authStorage.jsx` เดิม
- DoD: `authApi.jsx` คืน error ผ่าน `ApiError`/`NetworkError` เดิมไม่ครอบ/เปลี่ยนรูปแบบเอง (ที่อื่นในโปรเจกต์ใช้ pattern นี้อยู่แล้ว) · ไอคอนใหม่ทั้ง 4 ใช้ `currentColor`/`strokeWidth` เดียวกับไอคอนเดิมในไฟล์ ไม่หลุด style · `PasswordInput` toggle แสดง/ซ่อนได้จริงตรงกับภาพ 19/20 (state ซ้าย/ขวา)

**FE-22 · หน้า Login — แทนที่ `LoginPage` placeholder**
- ทำ: ฟอร์มตาม section map (อีเมล, รหัสผ่าน + toggle, จดจำฉัน, ลืมรหัสผ่าน?, ปุ่มเข้าสู่ระบบ, ลิงก์ไป register) → validate ฝั่ง client ก่อนยิง (อีเมล format, ทั้งคู่ห้ามว่าง) → เรียก `authApi.login` → สำเร็จ: `setTokens()` + redirect ตามที่ตกลงใน Gap (`/` เป็นค่าเริ่มต้น) → ผิดพลาด: โชว์ข้อความจาก `ApiError.message` จริงเหนือปุ่ม submit (401/403 ใช้ข้อความที่ backend ส่งมาตรง ๆ ไม่เขียนเอง) → ระหว่างยิง disable ปุ่ม + โชว์ loading กันกดซ้ำ → ถ้า `isLoggedIn()` อยู่แล้วตอนเข้าหน้านี้ redirect ออกทันที (ตาม Gap)
- ขึ้นกับ: FE-21
- DoD: E2E ครบ idle/submitting/error(401 invalid_credentials, 403 account_disabled, NetworkError)/success · toggle แสดง/ซ่อนรหัสผ่านตรงภาพ 19 ทั้ง 2 state · กด "ยังไม่มีบัญชี?สมัครสมาชิก" ไป `/register` จริง · login สำเร็จแล้ว token เก็บถูกคีย์ตรงกับที่ FE-24 (header) และหน้าอื่นที่เรียก `isLoggedIn()` อ่านเจอ (ทดสอบข้าม reload หน้า)

**FE-23 · หน้า Register — แทนที่ `RegisterPage` placeholder**
- ทำ: ฟอร์มตาม section map (ชื่อ, นามสกุล, เบอร์โทร, อีเมล, รหัสผ่าน + toggle, ยืนยันรหัสผ่าน + toggle, ปุ่มสมัครสมาชิก, ลิงก์ไป login) → validate ฝั่ง client mirror กฎ backend ทั้งหมดที่ระบุไว้ในหัวข้อ API ด้านบน (ความยาว password, pattern เบอร์โทร, ชื่อ/นามสกุลห้ามว่าง) บวก "ยืนยันรหัสผ่านต้องตรงกับรหัสผ่าน" (เช็คฝั่ง client เท่านั้น ไม่มีใน backend schema) → เรียก `authApi.register` (**ห้ามส่ง field `role` เด็ดขาด**) → สำเร็จ: `setTokens()` จาก response ตรง ๆ (ไม่ต้อง login ซ้ำ) + redirect เหมือน FE-22 → error 422: โชว์ `error.fields` เป็น inline error ใต้ช่องนั้น ๆ ตรงชื่อ field ที่ backend ส่งมา ถ้าไม่มี `fields` (เช่น 400) โชว์ข้อความรวมเหนือปุ่ม → ถ้า `isLoggedIn()` อยู่แล้ว redirect ออกทันทีเหมือนกัน
- ขึ้นกับ: FE-21
- DoD: E2E ครบ idle/submitting/error(422 พร้อม field-level error เช่น "อีเมลนี้ถูกใช้งานแล้ว" ใต้ช่องอีเมลจริง, 400, NetworkError)/success · client validate ครบทุกกฎที่ backend บังคับก่อนค่อยยิง request (ลดจำนวน round-trip ที่ต้อง fail จริง) · กด "มีบัญชีแล้ว?เข้าสู่ระบบ" ไป `/login` จริง

**FE-24 · Header แสดงสถานะ login/logout**
- ทำ: `Header.jsx` เช็ค `isLoggedIn()` — login แล้วสลับปุ่ม "เข้าสู่ระบบ"/"สมัครสมาชิก" เป็นปุ่ม "ออกจากระบบ" ปุ่มเดียว (reuse `.btn.btn-outline`) กดแล้วเรียก `authApi.logout(refreshToken)` (best-effort ไม่ต้องรอผลลัพธ์นาน เพราะ backend ตอบ `204` เสมอ) → `clearTokens()` → redirect `/` เสมอ ไม่ว่า API logout จะสำเร็จหรือ fail (เพื่อไม่ให้ผู้ใช้ค้าง — token ฝั่ง client ถูกลบไปแล้วถือว่า logout สำเร็จจากมุมมอง UI)
- ขึ้นกับ: FE-21, FE-22/FE-23 (ให้มี token จริงให้ทดสอบ)
- DoD: reload หน้าใด ๆ หลัง login แล้ว header ยังโชว์สถานะ login ถูกต้อง (อ่านจาก `localStorage` ไม่ใช่ React state ล้วนที่หายตอน reload) · กด "ออกจากระบบ" แล้วปุ่มกลับเป็น 2 ปุ่มเดิมทันที และเข้าหน้าที่ต้อง login (ถ้ามี) ไม่ได้อีก

**FE-25 · Responsive + Visual QA**
- ทำ: ไล่เทียบกับ prototype ภาพ 19/20 ทีละจุดตาม checklist เดียวกับไฟล์ก่อนหน้า ทั้ง Login และ Register — การ์ดฟอร์มต้องอยู่กลางจอ ไม่ overflow แนวนอนบน mobile, แถวคู่ "ชื่อ/นามสกุล" ใน Register ต้องยุบเป็น 1 คอลัมน์บน mobile
- ขึ้นกับ: FE-22, FE-23, FE-24
- DoD: ตรวจ responsive จริงที่ 390/820/1440px (device metrics override ไม่ใช่แค่ย่อหน้าต่าง) · `npm run build` ผ่านไม่มี error/warning สำคัญ · ไม่มีไฟล์ backend ถูกแก้ (เช็ค git diff) · ตรวจ git diff ว่าไม่มี token/secret ใด ๆ ถูก commit ติดไปด้วย (เช่น .env ที่ทดสอบ login จริง)
