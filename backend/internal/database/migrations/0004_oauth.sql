-- เข้าสู่ระบบ/สมัครสมาชิกด้วยบัญชี Google หรือ Facebook
--
-- หลักการ: ตาราง users ยังเป็นเจ้าของ "ตัวตน" ของผู้ใช้เหมือนเดิม
-- ส่วนวิธีพิสูจน์ตัวตนจากภายนอกแยกไปอยู่ user_identities ผู้ใช้หนึ่งคนจึงผูกได้ทั้ง
-- Google และ Facebook พร้อมกัน และยังตั้งรหัสผ่านของตัวเองเพิ่มทีหลังได้

-- ผู้ใช้ที่สมัครผ่าน provider ไม่มีรหัสผ่าน เก็บเป็นค่าว่างแทน NULL เพื่อไม่ให้ทุกจุด
-- ที่อ่าน password_hash อยู่แล้วต้องแก้มารับ NULL (scanUser, VerifyPassword ฯลฯ)
ALTER TABLE users ALTER COLUMN password_hash SET DEFAULT '';

-- ค่าว่างแปลว่า "ยังไม่มีรหัสผ่าน" ค่าอื่นต้องเป็น bcrypt hash จริงเท่านั้น
-- bcrypt.CompareHashAndPassword คืน error เสมอเมื่อ hash เป็นค่าว่าง การ login
-- ด้วยรหัสผ่านของบัญชี OAuth จึงตกเสมอ แม้โค้ดชั้นบนจะลืมเช็ค
ALTER TABLE users
    ADD CONSTRAINT users_password_hash_empty_or_bcrypt
    CHECK (password_hash = '' OR password_hash LIKE '$2%');

-- ---------------------------------------------------------------- ตัวตนจากภายนอก
CREATE TABLE user_identities (
    id               BIGSERIAL PRIMARY KEY,
    user_id          BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider         TEXT NOT NULL CHECK (provider IN ('google', 'facebook')),
    provider_user_id TEXT NOT NULL CHECK (provider_user_id <> ''),
    email            TEXT NOT NULL DEFAULT '',   -- อีเมลที่ provider ส่งมาตอนผูก (ไว้สอบย้อนหลัง)
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    -- บัญชีภายนอกหนึ่งใบผูกกับผู้ใช้ในระบบได้คนเดียวเท่านั้น
    UNIQUE (provider, provider_user_id),
    -- ผู้ใช้หนึ่งคนผูกกับ provider เดิมซ้ำสองใบไม่ได้
    UNIQUE (user_id, provider)
);
CREATE INDEX idx_user_identities_user ON user_identities (user_id);

-- ---------------------------------------------------------------- โค้ดแลก token
-- callback ของ provider วิ่งกลับมาที่ backend แต่ token ต้องไปจบที่ frontend ซึ่งคนละ origin
-- ถ้า redirect พา access token ไปใน URL ตรง ๆ token จะไปติดใน history ของเบราว์เซอร์
-- และ log ของ proxy ทุกตัวกลางทาง จึงส่งเป็นโค้ดใช้ครั้งเดียวอายุสั้นแทน แล้วให้ frontend
-- ยิง POST มาแลกเป็น token pair จริง (เก็บเฉพาะ hash ด้วยเหตุผลเดียวกับ refresh_tokens)
CREATE TABLE oauth_exchange_codes (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code_hash   TEXT NOT NULL UNIQUE,
    expires_at  TIMESTAMPTZ NOT NULL,
    consumed_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_oauth_codes_user ON oauth_exchange_codes (user_id);
