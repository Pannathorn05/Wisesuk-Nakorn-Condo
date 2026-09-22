-- ลบบัญชีผู้ดูแล (superadmin/staff/{id}) ต้องเป็น soft delete ตามที่ openapi.yaml ระบุไว้
-- แต่ของเดิมยิง DELETE ตรง ๆ ลบแถวถาวร กู้คืนไม่ได้ และพารายการที่อ้างอิงไว้
-- (activity_logs.actor_id, bookings.reviewed_by, payments.reviewed_by) หลุดหายตามไปด้วย

ALTER TABLE users ADD COLUMN deleted_at TIMESTAMPTZ;

-- ต้องสร้าง index ใหม่ก่อนแล้วค่อยลบตัวเก่า ไม่งั้นจะมีช่วงเวลาที่ไม่มี index
-- คุ้มครองกฎ "หนึ่งสาขามีผู้ดูแลคนเดียว" อยู่เลย
--
-- เงื่อนไข deleted_at IS NULL ทำให้แถวที่ถูก soft delete ไม่ถูกนับ จึงสร้าง
-- ผู้ดูแลคนใหม่ลงสาขาเดิมได้ แม้แถวเก่าจะยังเก็บ branch_id ไว้เพื่อ audit อยู่ก็ตาม
-- (แก้ branch_id เป็น NULL ไม่ได้เพราะชน CHECK admin_requires_branch)
CREATE UNIQUE INDEX uq_admin_per_branch_v2
    ON users (branch_id)
    WHERE role = 'admin' AND deleted_at IS NULL;

DROP INDEX uq_admin_per_branch;
ALTER INDEX uq_admin_per_branch_v2 RENAME TO uq_admin_per_branch;
