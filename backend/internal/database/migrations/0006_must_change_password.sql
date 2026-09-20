-- บัญชีผู้ดูแลที่หัวหน้าผู้ดูแลสร้างให้ ไม่ได้ตั้งรหัสผ่านเอง
--
-- หน้า "เพิ่มผู้ดูแลใหม่" ไม่มีช่องรหัสผ่าน ระบบจึงตั้งรหัสผ่านตั้งต้นร่วมให้จาก
-- STAFF_DEFAULT_PASSWORD แล้วปักธงนี้ไว้ เพื่อให้ frontend บังคับตั้งรหัสใหม่ก่อนใช้งานจริง
-- ค่าเริ่มต้นเป็น FALSE เพราะบัญชีที่มีอยู่แล้วทุกใบตั้งรหัสผ่านของตัวเองมาแล้ว
ALTER TABLE users ADD COLUMN must_change_password BOOLEAN NOT NULL DEFAULT FALSE;

-- บัญชีที่ยังไม่มีรหัสผ่าน (สมัครผ่าน Google/Facebook) ไม่ควรถูกบังคับให้ "เปลี่ยน"
-- รหัสผ่านที่ไม่มีอยู่ กันไว้ที่ DB เลยจะได้ไม่ต้องไว้ใจว่าทุกจุดในโค้ดจำกฎนี้ได้
ALTER TABLE users
    ADD CONSTRAINT users_must_change_password_needs_password
    CHECK (NOT must_change_password OR password_hash <> '');
