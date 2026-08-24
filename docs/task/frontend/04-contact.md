# Phase FE-4 — ติดต่อเรา (Guest)

อ้างอิง: `frontend/prototype/ตัวอย่าง Website v2.pdf` หน้า 13 (รูปภาพที่ 18)

| รูปภาพ | หน้า PDF | เนื้อหา |
|---|---|---|
| 18 | 13 | หน้าติดต่อเรา: header static + การ์ดข้อมูลติดต่อ 3 ใบ (ต่อสาขา) พร้อมแผนที่ embed ต่อใบ + footer |

ตรวจแล้ว: มีภาพเดียวสำหรับทั้งหน้า ไม่มี tab/state อื่นให้แตกเพิ่ม — footer ที่เห็นในภาพตรงกับ `Footer` component ที่ implement ไปแล้วใน FE-01 ทุกจุด (โลโก้, เมนู, สาขา, ติดต่อเรา, ลิขสิทธิ์) ไม่ต้องแก้/สร้างใหม่

DoD ทุก task ที่ดึงข้อมูลจาก API ต้องผ่าน **E2E checklist 4 สถานะ: loading / empty / error (มีปุ่มลองใหม่) / success**
ขึ้นกับ `docs/c.md`, `.claude/CLAUDE.md`: ห้ามแก้ backend, ห้าม hardcode ข้อมูลที่ API มีจริง, ห้าม mock แทน API จริง

---

## Section map จาก prototype

**Header**: title "ติดต่อเรา" + subtext static "พร้อมให้บริการทุกวัน 8:30 - 17:30 น." (ไม่มี field API รองรับเวลาเปิด-ปิดทำการ — ดู Gap)

**Grid การ์ดข้อมูลติดต่อ** (3 คอลัมน์ desktop, 1 การ์ดต่อสาขา active):
1. ชื่อสาขา (หัวข้อการ์ด ตัวหนา)
2. 📍 ที่อยู่
3. 📞 เบอร์โทร (คั่นด้วย comma ถ้ามีหลายเบอร์)
4. 💬 LINE ID
5. 🗺️ ลิงก์ Google Maps เป็นข้อความคลิกได้ (URL เต็ม เปิดแท็บใหม่)
6. ด้านล่างการ์ด: กล่องแผนที่ **embed จริง** (มีปุ่ม zoom +/- ของ Google Maps เห็นในภาพ ไม่ใช่รูปนิ่ง)

**Footer**: ใช้ของเดิมจาก FE-01 ตรงเป๊ะ ไม่ต้องแก้

---

## Gap ที่เจอ — เช็คกับ backend ที่รันจริงแล้ว (curl `localhost:8080`) ไม่ใช่แค่เดาจาก prototype

- **`map_url` ว่างเปล่าทุกสาขาในข้อมูลจริงตอนนี้** เหมือน gap เดียวกับที่เจอใน `02-branches.md` (FE-12) — ต้องมี empty-state "ยังไม่มีแผนที่สำหรับสาขานี้" ไม่ใช่กล่องว่าง/error **ห้ามเขียน logic embed/fallback ใหม่** ให้ reuse `BranchMapPanel` + `isEmbeddableMapUrl` ที่มีอยู่แล้วตรง ๆ (ส่ง `mapUrl={branch.map_url}` `branchName={branch.name}`) ประหยัดงานและได้พฤติกรรมเดียวกันทั้งเว็บ
- **ที่อยู่/เบอร์โทร/LINE ID ตรงกับ prototype เป๊ะทุกตัวอักษร** เทียบ curl จริงกับภาพ 18 แล้ว (ทั้ง 3 สาขา) รวมถึง `line_id: "@wisetsuk"` ที่เหมือนกันทั้ง 3 สาขาในข้อมูลจริง — **ไม่ใช่บั๊ก** ตรงกับที่ prototype โชว์ซ้ำกันทั้ง 3 การ์ดเช่นกัน ไม่ต้องพยายามหาค่าที่ต่างกันมาใส่
- **ไม่มี field เวลาเปิด-ปิดทำการ** ใน Branch object เลย (เช็ค response ทุก key แล้วไม่มี) → บรรทัด "พร้อมให้บริการทุกวัน 8:30 - 17:30 น." เป็น business copy คงที่ ไม่ผูกกับสาขาไหนเป็นพิเศษ ใส่เป็นข้อความ static ที่หัวหน้าเดียว (เหมือนกรณี "รปภ."/สูตรเงินประกันใน `02-branches.md`) **ห้ามไปหา field เวลาเปิดปิดที่ไม่มีจริงมาแทน**
- **ไอคอนหน้าลิงก์ Google Maps (บรรทัด 5) กับไอคอนที่อยู่ (บรรทัด 2)** ในภาพต้นฉบับดูเป็นสัญลักษณ์ pin คล้ายกันทั้งคู่ แยกไม่ออกชัดจากความละเอียดภาพ — ใช้ `IconPin` (มีอยู่แล้วใน `src/components/icons`) ซ้ำได้ทั้ง 2 บรรทัด ไม่ต้องสร้างไอคอนใหม่ (ตรงกับ `IconPhone`/`IconLine` ที่ก็มีอยู่แล้วสำหรับบรรทัดเบอร์โทร/LINE)
- **`email` field มีอยู่ใน response แต่ว่างเปล่าทุกสาขา และ prototype ไม่โชว์บรรทัดอีเมลในการ์ดเลย** — ไม่ต้องเพิ่ม UI สำหรับ field นี้ในรอบนี้

---

**FE-19 · หน้าติดต่อเรา — แทนที่ `ContactPage` placeholder**
- ทำ: header (title + subtext static ตาม section map) → เชื่อม `GET /api/v1/branches` (reuse hook `useBranches` เดิมจาก FE-08 ตัวเดียวกับที่ Footer ใช้) → grid การ์ดต่อสาขา active: ชื่อ, ที่อยู่ (`IconPin`), เบอร์โทร join ด้วย `", "` (`IconPhone`), LINE ID ถ้ามีค่า (`IconLine`), ลิงก์ `map_url` เป็นข้อความคลิกได้เปิดแท็บใหม่ด้วย `IconPin` (ถ้า `map_url` ว่างไม่ต้องโชว์บรรทัดนี้ในการ์ด — ปล่อยให้กล่อง `BranchMapPanel` ด้านล่างเป็นคนแสดง empty-state แทน) → ใต้การ์ดแต่ละใบ reuse `BranchMapPanel` (component เดิมจาก FE-12) ส่ง `mapUrl`/`branchName` ของสาขานั้นตรง ๆ
- ขึ้นกับ: FE-00 (api client), FE-08 (`useBranches`), FE-12 (`BranchMapPanel`/`isEmbeddableMapUrl`)
- DoD: E2E checklist ครบ — loading = skeleton 3 การ์ด, error = ปุ่มลองใหม่, empty = ไม่มีสาขา active เห็นข้อความชัดเจน (ไม่ใช่หน้าขาว), success = จำนวนการ์ด/ที่อยู่/เบอร์/LINE ตรงกับ backend จริงทั้ง 3 สาขา (เทียบตัวอักษรกับภาพ 18) · แต่ละการ์ดที่ `map_url` ว่างต้องเห็น empty-state "ยังไม่มีแผนที่สำหรับสาขานี้" จาก `BranchMapPanel` ไม่ใช่กล่องว่างเปล่า

**FE-20 · Responsive + Visual QA**
- ทำ: ไล่เทียบกับ prototype ภาพ 18 ทีละจุดตาม checklist เดียวกับไฟล์ก่อนหน้า — grid 3 คอลัมน์ต้องยุบเป็น 1 คอลัมน์บน mobile โดยไม่ overflow แนวนอน (`minmax(0, 1fr)` ตามบทเรียนจาก `01-homepage.md`)
- ขึ้นกับ: FE-19
- DoD: ตรวจ responsive จริงที่ 390/820/1440px (device metrics override ไม่ใช่แค่ย่อหน้าต่าง) · `npm run build` ผ่านไม่มี error/warning สำคัญ · ไม่มีไฟล์ backend ถูกแก้ (เช็ค git diff)
