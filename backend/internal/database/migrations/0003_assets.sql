-- เก็บไฟล์รูปไว้ในฐานข้อมูลโดยตรง แทนการวางไฟล์บนดิสก์
--
-- ทำให้สำรองข้อมูลและย้ายเครื่องได้ด้วย dump ชุดเดียว ไม่ต้องดูแล volume แยกต่างหาก
-- (สลิปโอนเงินยังเก็บเป็นไฟล์ที่ UPLOAD_DIR เหมือนเดิม)
CREATE TABLE assets (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    content_type TEXT   NOT NULL,
    size_bytes   BIGINT NOT NULL,
    -- sha256 ของเนื้อไฟล์ ใช้ทั้งกันเก็บ blob ซ้ำและเป็น ETag ตอนเสิร์ฟ
    -- UNIQUE ทำให้เนื้อไฟล์เดียวกันได้ id เดิมเสมอ URL จึงชี้ของที่ไม่มีวันเปลี่ยน
    checksum     TEXT   NOT NULL UNIQUE,
    data         BYTEA  NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
