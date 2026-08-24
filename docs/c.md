คุณคือ Senior React Frontend Engineer + UI Implementation Specialist

งานนี้ต้องทำโดยลงมือแก้/สร้างไฟล์ภายใน project ที่มีอยู่จริง ไม่ใช่เพียงให้คำแนะนำหรือสร้างตัวอย่างโค้ดแยกออกมา

==================================================

1. OBJECTIVE หลัก
   ==================================================

สร้าง Frontend สำหรับเว็บไซต์หอพักตาม Prototype ที่มีอยู่ใน project โดยต้องพยายามทำ UI/UX ให้ใกล้เคียง Prototype ที่ให้มา “มากที่สุดเท่าที่ทำได้”

IMPORTANT:

* ใช้ React แบบปกติจาก project ที่มีอยู่
* ห้ามเปลี่ยน project ไปใช้ Vite
* ห้าม migrate project architecture ไป framework อื่น
* ห้ามเปลี่ยน Backend
* ห้ามแก้ไข source code ของ Backend
* ห้ามเปลี่ยน API contract ของ Backend
* ห้ามสร้าง Backend ใหม่
* ห้าม mock ระบบแทน Backend ในส่วนที่มี API จริงอยู่แล้ว
* ต้องเตรียม Frontend ให้เชื่อมต่อ Backend ที่มีอยู่ได้จริง
* ทำเฉพาะส่วน “ผู้ใช้งานทั่วไป (Guest)” ในรอบนี้เท่านั้น
* ห้ามสร้างหน้า Member
* ห้ามสร้างหน้า Admin
* ห้ามสร้างหน้า Super Admin
* ห้ามแก้ไข logic หรือ UI ที่เป็นส่วนของ Member/Admin/Super Admin นอกจากกรณีที่จำเป็นต่อ routing/shared component แต่ต้องไม่ทำให้ส่วนอื่นเสียหาย

==================================================
2. SOURCE OF TRUTH
==================

ก่อนเริ่มเขียน code ให้ตรวจสอบข้อมูลจากแหล่งต่อไปนี้ตามลำดับ:

1. โฟลเดอร์ prototype/
2. ไฟล์ตัวอย่าง Website ที่อยู่ใน project
3. รูปภาพ / screenshot / mockup ใน prototype
4. Existing React project structure
5. Existing API/service/frontend integration
6. Backend API ที่มีอยู่แล้ว เฉพาะเพื่อดูวิธีเรียกใช้งาน
7. package.json และ dependencies ที่ project ใช้อยู่

ให้ถือว่า Prototype เป็น “Source of Truth” ด้าน:

* Layout
* Visual hierarchy
* Spacing
* Typography
* Component positioning
* สี
* ปุ่ม
* Card
* Form
* Navigation
* Modal
* Dropdown
* Filter
* ตาราง
* Image gallery
* Responsive behavior
* Page flow
* UX interaction
* ข้อความบนหน้าจอ
* การเปลี่ยนหน้า
* State ของ component
* Loading / empty / error state ที่ Prototype มี

ห้ามเดาดีไซน์ใหม่เพียงเพราะคิดว่าสวยกว่า

เป้าหมายคือ “implement ตาม Prototype” ไม่ใช่ “ออกแบบเว็บใหม่”

==================================================
3. ต้องอ่าน prototype ก่อนเขียน
===============================

ก่อนสร้างหรือแก้ไฟล์ใด ๆ ให้:

* สำรวจโครงสร้าง project
* สำรวจโฟลเดอร์ prototype/
* อ่านไฟล์ทั้งหมดที่เกี่ยวข้องกับ Prototype
* ตรวจสอบว่ามี screenshot/image/reference ใดบ้าง
* ตรวจสอบชื่อไฟล์และโครงสร้างที่มีอยู่
* ตรวจสอบว่า project ใช้ React configuration แบบใด
* ตรวจสอบ scripts ใน package.json
* ตรวจสอบ existing router
* ตรวจสอบ existing API client/service
* ตรวจสอบ existing authentication utilities
* ตรวจสอบ environment variables
* ตรวจสอบ CSS architecture

ห้ามรีบสร้าง component ก่อนอ่าน Prototype

ถ้ามีหลายไฟล์ที่อธิบายหน้าจอเดียวกัน ให้พิจารณาทั้งหมดร่วมกัน

หากรายละเอียดบางส่วนใน Prototype มีความกำกวม ให้เลือก implementation ที่ใกล้กับ Prototype มากที่สุด และอย่าเปลี่ยน design โดยพลการ

==================================================
4. SCOPE รอบนี้ — GUEST ONLY
============================

รอบนี้ให้สร้างเฉพาะหน้าสำหรับ “ผู้ใช้งานทั่วไป (Guest)”

จากเอกสาร Prototype ฟังก์ชันหลักของ Guest ได้แก่:

* เข้าดู Homepage
* ดูข้อมูลหอพักวิเศษสุขนครคอนโดและหอพักในเครือ
* ค้นหาสาขา
* ดูรายละเอียดห้องพัก
* ดูรูปภาพบรรยากาศ
* ดูสิ่งอำนวยความสะดวก
* Filter ห้องตามสาขา
* Filter ตามประเภทห้อง
* Filter รายวัน / รายเดือน
* Filter ตามวันที่ต้องการเข้าพัก
* ดูช่องทางการติดต่อ
* ตรวจสอบสถานะห้องว่างเบื้องต้น
* สมัครสมาชิก
* เข้าสู่ระบบ

อ้างอิงหน้าตาม Prototype Guest:

* Homepage
* ส่วนต่อของ Homepage
* หน้ารวมสาขา
* หน้ารายละเอียดสาขา
* ส่วนรายละเอียดสาขาเพิ่มเติม
* หน้าแผนที่
* หน้าค้นหาห้องพัก
* หน้า Filter ตามสาขา
* หน้า Filter ตามประเภทห้อง
* หน้า Filter ตามวันที่
* หน้า Contact
* หน้า Login
* หน้า Register

เอกสาร Prototype ระบุหน้ากลุ่มนี้ไว้ในภาพที่ 1–20

==================================================
5. IMPORTANT — อย่าทำ MEMBER / ADMIN / SUPER ADMIN
==================================================

ห้ามสร้างในรอบนี้:

* Member Dashboard
* Member Profile
* Booking History สำหรับ Member
* Payment สำหรับ Member
* Member Booking Flow แบบหลัง Login
* Admin Login
* Admin Dashboard
* Room Management ของ Admin
* Branch Management ของ Admin
* Super Admin Dashboard
* Super Admin Management
* Activity Log ของ Admin
* ระบบจัดการสมาชิกสำหรับ Admin

แม้ Prototype จะมีหน้าดังกล่าวอยู่ ให้ใช้เพื่อทำความเข้าใจระบบเท่านั้น และยังไม่ต้อง implement

==================================================
6. REACT REQUIREMENT
====================

ต้องใช้ React project ที่มีอยู่แล้ว

ตรวจสอบก่อนว่า project เป็นแบบใด

ถ้ามี React อยู่แล้ว:

* ใช้ architecture เดิม
* ใช้ package เดิมที่ติดตั้งไว้
* ไม่รื้อ project ใหม่
* ไม่ migrate ไป Vite
* ไม่เปลี่ยน build tool โดยไม่จำเป็น

ห้ามทำ:

* npm create vite
* create-vite
* เปลี่ยน project เป็น Vite
* เปลี่ยน React project เป็น Next.js
* เปลี่ยนเป็น Vue
* เปลี่ยนเป็น Angular

ถ้าต้องติดตั้ง package เพิ่ม ให้เลือก package ที่จำเป็นจริง ๆ และต้องไม่ทำให้ architecture เดิมเสียหาย

==================================================
7. BACKEND PROTECTION — ห้ามแตะเด็ดขาด
======================================

นี่คือข้อกำหนดสำคัญที่สุด

DO NOT MODIFY BACKEND.

Backend เป็นของเดิมที่มีอยู่แล้ว

ห้าม:

* แก้ controller
* แก้ service
* แก้ repository
* แก้ model
* แก้ schema
* แก้ database
* แก้ migration
* แก้ route ของ backend
* แก้ authentication logic ของ backend
* แก้ middleware
* แก้ environment ของ backend
* แก้ business logic
* แก้ API response
* เปลี่ยน HTTP method
* เปลี่ยน endpoint
* เปลี่ยน field name ของ API

ห้าม commit การเปลี่ยนแปลงใน Backend

ถ้า Backend API มีอยู่แล้ว ให้ Frontend เรียกใช้ตาม contract เดิม

หาก API ใดที่จำเป็นยังไม่มี:

* อย่าสร้าง Backend ใหม่
* อย่าแก้ Backend
* ให้สร้าง frontend abstraction/service interface เอาไว้รองรับก่อน
* ระบุจุดที่ต้องรอ API อย่างชัดเจน
* ใช้ fallback/mock เฉพาะใน development เท่านั้น และต้องแยกออกจาก production API flow อย่างชัดเจน

==================================================
8. FRONTEND ↔ BACKEND
=====================

ต้องออกแบบ Frontend ให้พร้อมต่อ Backend ที่มีอยู่แล้ว

สร้าง layer สำหรับเรียก API อย่างเป็นระบบ เช่น:

src/
api/
services/
hooks/
components/
pages/
layouts/
routes/
utils/
assets/
styles/

หรือปรับตาม architecture ที่ project มีอยู่แล้ว ถ้าของเดิมมีโครงสร้างที่เหมาะสมกว่า

สำคัญ:

อย่าเรียก fetch/axios กระจายทั่ว component โดยไม่มี abstraction

ควรแยกเป็น:

* API client
* service
* custom hook
* page/component

ตัวอย่างแนวคิด:

api/
client.js
branchApi.js
roomApi.js
authApi.js

services/
branchService.js
roomService.js
authService.js

hooks/
useBranches.js
useRooms.js
useRoomSearch.js

แต่ให้ปรับตาม existing architecture ของ project

==================================================
9. ROUTING
==========

สร้าง routing สำหรับ Guest ให้เป็นระบบ

ตัวอย่างโครงสร้าง:

/
/branches
/branches/:id
/branches/:id/map
/rooms
/rooms/:id
/contact
/login
/register

แต่ห้ามสมมติว่าต้องใช้ path เหล่านี้ 100%

ให้ตรวจสอบ Prototype และ existing routing ก่อน แล้วใช้ URL structure ที่เหมาะสมกับ project เดิม

Routing ต้องสามารถ:

* ไปหน้า Homepage
* ไปหน้าสาขา
* ไปหน้ารายละเอียดสาขา
* ไปหน้าห้องพัก
* ไปหน้า Search
* ไปหน้า Contact
* ไป Login
* ไป Register

ได้จริง

==================================================
10. COMPONENT ARCHITECTURE
==========================

แยก Component ให้เป็นระบบ

อย่าสร้าง page ใหญ่ไฟล์เดียวหลายพันบรรทัด

ตัวอย่าง:

src/
components/
common/
layout/
navigation/
branch/
room/
search/
auth/

pages/
guest/
Home/
Branches/
BranchDetail/
BranchMap/
Rooms/
Contact/
Login/
Register/

hooks/
services/
api/
utils/
assets/
styles/

แต่ต้องดู architecture เดิมของ project ก่อน

กฎ:

* Component ที่ reuse ได้ต้องแยก
* Component ที่มีเฉพาะ page นั้นสามารถอยู่ใน folder page
* อย่าสร้าง abstraction มากเกินความจำเป็น
* อย่าทำทุกอย่างเป็น component เล็ก ๆ จน maintain ยาก

==================================================
11. PIXEL-ACCURATE IMPLEMENTATION
=================================

ต้องพยายาม reproduce Prototype ให้ใกล้ที่สุด

ตรวจสอบ:

* ขนาด header
* ความกว้าง container
* max-width
* margin
* padding
* gap
* font-size
* font-weight
* line-height
* border-radius
* shadow
* border
* สี background
* สี text
* สี primary
* สี secondary
* icon
* image ratio
* card size
* button size
* input size
* alignment
* vertical rhythm
* responsive behavior

ห้ามปรับดีไซน์เองเพราะคิดว่า “modern กว่า”

ห้ามเปลี่ยนสีเอง

ห้ามเปลี่ยน layout เอง

ห้ามเปลี่ยนตำแหน่งปุ่มเอง

ห้ามลดรายละเอียดของ Prototype เพื่อให้ง่ายต่อการเขียน

==================================================
12. RESPONSIVE
==============

ต้องทำให้ใช้งานได้อย่างน้อย:

* Desktop
* Tablet
* Mobile

แต่ priority ของ visual matching ให้ยึด Desktop Prototype ก่อน หาก Prototype เป็น Desktop

จากนั้นค่อยทำ responsive โดยพยายามรักษา:

* hierarchy
* spacing
* interaction
* visual language

ไม่ใช่แค่เอา desktop ไปย่อขนาด

==================================================
13. IMAGE / ASSET
=================

ใช้ image/assets ที่มีอยู่ใน project ก่อน

ตรวจสอบ:

* prototype assets
* public/
* src/assets/
* existing image folder

อย่าสร้าง image placeholder ถ้ามี asset จริงอยู่แล้ว

อย่าเปลี่ยน image โดยไม่มีเหตุผล

ถ้า Prototype ใช้รูปเดียวกัน ให้ใช้ asset เดียวกันถ้ามี

จัด asset ให้เป็นระบบ

ตัวอย่าง:

src/assets/
images/
icons/
logos/
branches/
rooms/

แต่ให้ปรับตาม project ที่มีอยู่แล้ว

==================================================
14. GUEST HOMEPAGE
==================

Implement Homepage ตาม Prototype แบบละเอียด

ต้องตรวจสอบจาก Prototype:

* Header
* Logo
* Navigation
* Hero
* Search/filter area
* Branch section
* Room section
* Accommodation information
* Facilities
* Contact/CTA
* Footer
* Image sections
* spacing ระหว่างแต่ละ section

ทุก section ต้องพยายามตรงกับต้นแบบ

ห้ามตัด section ออกเพียงเพื่อให้เขียนง่าย

==================================================
15. BRANCHES
============

Implement หน้ารวมสาขาตาม Prototype

ต้องรองรับสาขาตามข้อมูลจริงจาก Backend

หน้าควรมี:

* รายการสาขา
* รูป
* ชื่อ
* รายละเอียด
* ปุ่มดูรายละเอียด
* navigation ไป branch detail

หาก Backend ส่งข้อมูลแบบ dynamic ให้ render จาก API

ไม่ hardcode ข้อมูลสาขาหาก API มีข้อมูลเหล่านี้อยู่แล้ว

==================================================
16. BRANCH DETAIL
=================

ต้อง implement หน้ารายละเอียดสาขาให้ใกล้ Prototype

รองรับ:

* branch information
* รูปภาพ
* บรรยากาศ
* facilities
* room information
* nearby places
* contact information
* map
* CTA
* availability information

แบ่ง section ตามที่ Prototype แสดง

หากมี image gallery ให้ interaction ตรงกับต้นแบบเท่าที่ทำได้

==================================================
17. ROOM SEARCH
===============

สร้างหน้าค้นหาห้องพักตาม Prototype

ต้องรองรับ filter อย่างน้อย:

* Branch
* Room Type
* Daily / Monthly
* Desired check-in date

ตัว Filter ต้องใช้งานจริงกับ frontend state และควรเชื่อม API หาก Backend รองรับ

ห้ามทำ UI filter แบบกดแล้วไม่ทำอะไร

ต้อง handle:

* loading
* no result
* error
* result

==================================================
18. SEARCH RESULT
=================

แสดงรายการห้องตามข้อมูลจาก Backend

แต่ละ room card ให้ตรวจสอบ Prototype ว่ามี:

* รูป
* ชื่อ
* ประเภท
* ราคา
* รายละเอียด
* availability
* ปุ่มดูรายละเอียด
* ปุ่มจอง / CTA

อย่าเพิ่มข้อมูลที่ Prototype ไม่มีโดยไม่มีเหตุผล

==================================================
19. GUEST BOOKING RESTRICTION
=============================

จาก Prototype:

ถ้า Guest กด “จอง” โดยยังไม่ได้ Login
ให้แสดง flow ตาม Prototype

ห้ามให้ Guest เข้าสู่ booking flow ของ Member โดยตรง

ถ้ามี modal / dialog / redirect / warning ตาม Prototype ให้ implement ตามนั้น

แต่เนื่องจากรอบนี้ทำ Guest ก่อน:

* implement เฉพาะ behavior ที่ Prototype แสดงสำหรับ Guest
* อย่า implement member booking workflow เต็มรูปแบบ

==================================================
20. LOGIN / REGISTER
====================

ทำเฉพาะหน้าที่จำเป็นจาก Guest flow:

* Login
* Register

UI ต้องอิง Prototype

เชื่อมต่อ API เดิมถ้ามี

ห้ามเขียน authentication backend ใหม่

ตรวจสอบ token/session mechanism ที่ project ใช้อยู่แล้วก่อน

อย่าสร้าง auth system ใหม่ซ้อนระบบเดิม

==================================================
21. CONTACT
===========

Implement หน้า Contact ตาม Prototype

ต้องแสดงข้อมูลของสาขาตามข้อมูลจริงจาก Backend ถ้ามี

รองรับ:

* สาขา
* ที่อยู่
* โทรศัพท์
* ช่องทางติดต่อ
* social/contact links
* map หาก Prototype แสดง

==================================================
22. MAP
=======

ถ้า Prototype มี map:

* ทำ visual และ layout ให้ใกล้ Prototype
* ใช้ map solution ที่ project มีอยู่แล้วก่อน
* อย่าเปลี่ยน backend
* อย่าฝัง API key ลง source code โดยตรง
* ใช้ environment variable หากจำเป็น

==================================================
23. STATE MANAGEMENT
====================

ก่อนติดตั้ง state library ใหม่:

ตรวจสอบ project ว่ามีอยู่แล้วหรือไม่

ถ้า state ไม่ซับซ้อน:

* ใช้ React state
* useContext ตามความเหมาะสม
* custom hooks

อย่าติดตั้ง Redux/Zustand ฯลฯ เพียงเพราะสามารถใช้ได้

==================================================
24. ERROR / LOADING / EMPTY STATE
=================================

ทุกหน้า dynamic data ต้องมี:

* Loading state
* Error state
* Empty state

แต่ visual ของ state เหล่านี้ควรไม่ทำลาย design ของ Prototype

ห้ามปล่อยหน้าขาวหาก API error

==================================================
25. CODE QUALITY
================

โค้ดต้อง:

* readable
* maintainable
* reusable
* consistent
* modular
* มีชื่อไฟล์/ตัวแปรที่เข้าใจง่าย
* ไม่มี duplicate code ที่ไม่จำเป็น
* ไม่มี dead code
* ไม่มี console.log ที่ไม่จำเป็น
* ไม่มี hardcoded secret
* ไม่มี API key ฝังใน source
* ไม่มี backend modification

==================================================
26. FILE / FOLDER ORGANIZATION
==============================

จัดโครงสร้าง project ให้เป็นระเบียบ

ตัวอย่างแนวทาง:

src/
├── api/
├── assets/
│   ├── images/
│   ├── icons/
│   └── logos/
├── components/
│   ├── common/
│   ├── layout/
│   ├── navigation/
│   ├── branch/
│   ├── room/
│   ├── search/
│   └── auth/
├── hooks/
├── layouts/
├── pages/
│   └── guest/
│       ├── Home/
│       ├── Branches/
│       ├── BranchDetail/
│       ├── BranchMap/
│       ├── Rooms/
│       ├── Contact/
│       ├── Login/
│       └── Register/
├── routes/
├── services/
├── utils/
└── styles/

แต่หาก existing project มี architecture อยู่แล้ว
ให้รักษา architecture เดิมและปรับอย่างเป็นระบบ

อย่าสร้าง duplicate architecture ขึ้นมาโดยไม่จำเป็น

==================================================
27. API INTEGRATION RULE
========================

ก่อน implement ให้ตรวจสอบ Backend API ที่มีอยู่

ต้องค้นหา:

* endpoint
* HTTP method
* params
* query
* request body
* response shape
* error response
* authentication
* pagination
* image URL

จากนั้นทำ frontend service ให้ตรงกับ API จริง

ห้ามเดา endpoint ใหม่ถ้า existing backend มี endpoint อยู่แล้ว

ห้ามแก้ endpoint เพื่อให้เข้ากับ frontend

Frontend ต้อง adapt ตาม Backend

==================================================
28. BACKEND SAFETY CHECK
========================

ก่อนเริ่ม:

ระบุ backend directory

หลัง implement:

ตรวจ git diff / file changes

ตรวจสอบว่าไฟล์ Backend ไม่มีการแก้ไข

หากพบว่ามีการแก้ Backend ให้ revert เฉพาะการเปลี่ยนแปลงของ Backend ทันที

สิ่งที่อนุญาตให้เปลี่ยนคือ Frontend เท่านั้น

==================================================
29. IMPLEMENTATION PROCESS
==========================

ทำงานตามลำดับนี้:

STEP 1
สำรวจ project

STEP 2
อ่าน prototype/

STEP 3
ระบุหน้าที่อยู่ใน Guest scope

STEP 4
ตรวจสอบ existing React architecture

STEP 5
ตรวจสอบ existing backend API contract
โดย “อ่าน” เท่านั้น ห้ามแก้

STEP 6
วาง frontend architecture

STEP 7
สร้าง shared layout/component

STEP 8
สร้าง Homepage

STEP 9
สร้าง Branch pages

STEP 10
สร้าง Room Search

STEP 11
สร้าง Contact

STEP 12
สร้าง Login/Register

STEP 13
เชื่อม API

STEP 14
ตรวจ responsive

STEP 15
ตรวจ visual consistency กับ Prototype

STEP 16
ตรวจ error/loading/empty state

STEP 17
ตรวจ routing

STEP 18
ตรวจ build

STEP 19
ตรวจว่าห้ามมี backend file ถูกแก้

==================================================
30. IMPORTANT: อย่าหยุดหลังจากวิเคราะห์
=======================================

เมื่ออ่าน Prototype และ project เสร็จแล้ว
ให้ลงมือ implement จริงทันที

อย่าเพียงตอบว่า:

* “ควรทำแบบนี้”
* “แนะนำให้สร้าง component”
* “ตัวอย่าง code คือ...”

ต้องแก้ไฟล์ใน project จริง

==================================================
31. VALIDATION
==============

หลังทำเสร็จ ให้ตรวจ:

1. npm install / dependencies ไม่พัง
2. npm run build หรือ command build เดิมของ project
3. React compile ผ่าน
4. ไม่มี import error
5. ไม่มี broken route
6. ไม่มี missing asset
7. ไม่มี undefined API call
8. ไม่มี runtime error สำคัญ
9. ตรวจ Guest flow
10. ตรวจ responsive
11. ตรวจ Backend diff

==================================================
32. VISUAL QA
=============

ใช้ Prototype เปรียบเทียบหน้าจอทีละหน้า

ตรวจ:

* Layout
* Font
* Color
* Image
* Spacing
* Button
* Alignment
* Card
* Border
* Radius
* Shadow
* Header
* Footer
* Responsive

ถ้าจุดใดต่างจาก Prototype ให้แก้ให้ใกล้ขึ้น

อย่าจบงานเพียงเพราะ “ใช้งานได้”

เป้าหมายคือ:

Functional + Visual Accuracy

==================================================
33. PRIORITY
============

ถ้าต้องเลือกระหว่าง:

A. เขียนง่าย
B. สวยขึ้น
C. ตรง Prototype
D. maintainable

ให้เรียง priority:

1. ห้ามกระทบ Backend
2. ตรง Prototype
3. ใช้งานจริง
4. Maintainable
5. Performance / optimization

อย่าเลือก design ใหม่แทน Prototype

==================================================
34. OUTPUT ตอนท้าย
==================

เมื่อทำเสร็จ ให้สรุปเพียง:

1. ไฟล์/โฟลเดอร์ Frontend ที่เพิ่มหรือแก้
2. Guest pages ที่ implement สำเร็จ
3. API ที่ Frontend เชื่อมต่อ
4. ส่วนใดที่ยังรอ Backend API
5. คำสั่งรัน project
6. ผลการ build/test
7. ยืนยันว่า Backend ไม่ได้ถูกแก้ไข

ห้ามสรุปว่าส่วน Member/Admin/Super Admin เสร็จแล้ว
เพราะรอบนี้ยังไม่ใช่ scope

==================================================
35. FINAL HARD RULES
====================

RULE 1:
ห้ามแก้ Backend เด็ดขาด

RULE 2:
ห้ามเปลี่ยนเป็น Vite

RULE 3:
ห้ามเปลี่ยน framework

RULE 4:
ห้าม implement Member/Admin/Super Admin ในรอบนี้

RULE 5:
Prototype เป็น source of truth สำหรับ UI/UX

RULE 6:
Backend ที่มีอยู่แล้วเป็น source of truth สำหรับ API

RULE 7:
อย่า hardcode data หาก API มีข้อมูลจริง

RULE 8:
อย่าสร้าง mock API แทน API จริง

RULE 9:
อย่ารื้อ existing project โดยไม่จำเป็น

RULE 10:
ต้องลงมือแก้ project จริง

RULE 11:
ต้องจัดโครงสร้างไฟล์ให้เป็นระบบ

RULE 12:
ต้องตรวจ build และ runtime

RULE 13:
ก่อนจบต้องตรวจอีกครั้งว่า Backend ไม่มีไฟล์ถูกแก้

เริ่มจากการสำรวจ project และอ่าน prototype/
ก่อนลงมือเขียน code

อย่าถามคำถามที่สามารถตรวจสอบจาก project เองได้
ให้ตรวจสอบไฟล์จริงก่อน แล้วจึงตัดสินใจ implementation ที่เหมาะสม
