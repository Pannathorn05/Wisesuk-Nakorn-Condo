-- AC-8 + AC-10: สถานะการจองเหลือ 4 ค่า ตัด 'rejected' และ 'completed' ทิ้ง
--
-- เหตุผล: รายชื่อสถานะที่ล็อกห้องคือ pending_payment / awaiting_review / approved
-- ถ้ามีสถานะ 'rejected' อยู่ ใบที่ถูกปฏิเสธจะหลุดจากรายชื่อนี้ ห้องจึงว่างให้คนอื่นจอง
-- ทั้งที่เจ้าของใบเดิมยังแนบสลิปใหม่ได้ กลายเป็นจองซ้อนบนห้องเดียวกัน
-- ตอนนี้การปฏิเสธจะเด้งใบกลับไป pending_payment ซึ่งยังล็อกห้องอยู่
--
-- ส่วน 'completed' ไม่มีอะไรในระบบเปลี่ยนสถานะไปเป็นค่านี้เลย จึงตัดออกด้วย

-- ใบที่ค้างอยู่ในสถานะที่กำลังจะหายไป ต้องย้ายก่อนเปลี่ยนชนิดข้อมูล
UPDATE bookings SET status = 'pending_payment' WHERE status = 'rejected';
UPDATE bookings SET status = 'approved'        WHERE status = 'completed';

-- PostgreSQL ลบค่าออกจาก enum ตรง ๆ ไม่ได้ ต้องสร้างชนิดใหม่แล้วย้ายคอลัมน์มาใช้
ALTER TYPE booking_status RENAME TO booking_status_old;

CREATE TYPE booking_status AS ENUM ('pending_payment', 'awaiting_review', 'approved', 'cancelled');

ALTER TABLE bookings ALTER COLUMN status DROP DEFAULT;
ALTER TABLE bookings
    ALTER COLUMN status TYPE booking_status
    USING status::text::booking_status;
ALTER TABLE bookings ALTER COLUMN status SET DEFAULT 'pending_payment';

DROP TYPE booking_status_old;

-- เหตุผลที่ปฏิเสธย้ายไปอยู่ที่แถว payments แทน (AC-9)
-- ใบจองที่ถูกปฏิเสธกลับไปเป็น pending_payment จึงไม่มีเหตุผลค้างอยู่บนตัวใบอีกต่อไป
-- และการเก็บที่ payments ทำให้ย้อนดูได้ว่าการแจ้งชำระเงินครั้งไหนถูกปฏิเสธเพราะอะไร
ALTER TABLE bookings DROP COLUMN reject_reason;
