-- AC-2: rate limit ของ POST /auth/login นับที่ฐานข้อมูล
--
-- นับที่ DB ไม่ใช่ในหน่วยความจำ เพื่อให้ตัวนับไม่หายตอน restart และใช้ร่วมกันได้
-- ถ้าวันหนึ่งรัน API หลาย instance
--
-- เก็บเฉพาะครั้งที่ล้มเหลว (ได้ invalid_credentials) — login สำเร็จจะล้างแถวของคู่
-- (email, ip) นั้นทิ้ง ส่วนแถวที่เก่ากว่าหน้าต่างเวลาถูกลบตอนบันทึกความล้มเหลวครั้งใหม่
--
-- email ไม่มี foreign key ไปที่ users โดยตั้งใจ เพราะต้องนับอีเมลที่ไม่มีบัญชีด้วย
-- ไม่งั้นผู้โจมตีแยกออกได้ว่าอีเมลไหนมีบัญชีจากการที่ตัวไหนโดนจำกัด ตัวไหนไม่โดน
CREATE TABLE login_attempts (
    id          BIGSERIAL PRIMARY KEY,
    email       TEXT        NOT NULL,   -- ตัวพิมพ์เล็กเสมอ เหมือน users.email
    ip          TEXT        NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- query นับทุกครั้งกรองด้วย ip + ช่วงเวลา (ตัวนับราย IP และรายคู่ใช้ index เดียวกัน)
CREATE INDEX idx_login_attempts_ip   ON login_attempts (ip, created_at);
-- ใช้ตอนลบแถวที่หมดอายุ
CREATE INDEX idx_login_attempts_time ON login_attempts (created_at);
