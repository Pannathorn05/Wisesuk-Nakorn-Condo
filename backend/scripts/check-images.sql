-- ตรวจว่ารูปสาขาเข้าฐานข้อมูลจริงหรือยัง
--
-- รูปเก็บเป็น blob ในตาราง assets ส่วน branch_images เก็บแค่ image_url ที่เป็นข้อความ
-- ("http://.../files/<assetID>") ไม่มี foreign key ผูกไว้ การนับแถวใน branch_images
-- อย่างเดียวจึงไม่พอ — ต้องแกะ assetID จาก URL แล้วเช็คว่ามีแถวใน assets จริง
--
-- เป็น SQL ล้วน ไม่มีคำสั่ง \ ของ psql จึงวางลงหน้าต่าง Query ของ pgAdmin ได้ตรง ๆ
-- หรือจะรันจากบรรทัดคำสั่งก็ได้ (ยืนที่โฟลเดอร์ backend ใน Git Bash):
--
--     docker exec -i wisetsuk-db psql -U wisetsuk -d wisetsuk < scripts/check-images.sql
--
-- อ่านผลอย่างไร:
--   ส่วนที่ 1-2  missing ต้องเป็น 0 ทุกแถว  ถ้าไม่ใช่ = มีรูปที่ URL ชี้ไปหา blob ที่ไม่มีอยู่
--   ส่วนที่ 3    blob_ok = 1 คือรูปปกใช้ได้  0 คือยังไม่ได้ตั้งหรือ blob หาย
--   ส่วนที่ 4    blob ที่ไม่มีรูปสาขาไหนอ้างถึงแล้ว — กินที่เปล่า ๆ ลบทิ้งได้ถ้าต้องการ

WITH img AS (
    -- แกะ assetID ออกจาก image_url เฉพาะ URL ที่เป็นรูปแบบ /files/<uuid>
    -- รูปที่โฮสต์ไว้ที่อื่น (เพิ่มผ่าน POST /admin/branch/images) จะได้ NULL
    -- แล้วถูกนับเป็น missing ตามที่ควร เพราะ blob ไม่ได้อยู่ในฐานข้อมูลนี้
    SELECT bi.id,
           bi.branch_id,
           bi.image_url,
           CASE WHEN bi.image_url ~ '/files/[0-9a-fA-F-]{36}$'
                THEN substring(bi.image_url from '/files/(.*)$')::uuid
           END AS asset_id
    FROM branch_images bi
),
joined AS (
    SELECT b.slug, img.id, img.asset_id, a.id AS blob_id, a.size_bytes
    FROM branches b
    LEFT JOIN img    ON img.branch_id = b.id
    LEFT JOIN assets a ON a.id = img.asset_id
)
SELECT '1. รูปต่อสาขา'   AS section,
       slug              AS item,
       count(id)::int    AS images,
       count(blob_id)::int AS blob_ok,
       (count(id) - count(blob_id))::int AS missing,
       pg_size_pretty(coalesce(sum(size_bytes), 0)) AS detail
FROM joined
GROUP BY slug

UNION ALL
SELECT '2. รวมทุกสาขา',
       'ทั้งหมด',
       count(id)::int,
       count(blob_id)::int,
       (count(id) - count(blob_id))::int,
       pg_size_pretty(coalesce(sum(size_bytes), 0))
FROM joined

UNION ALL
SELECT '3. รูปหน้าปก',
       b.slug,
       (b.cover_image_url <> '')::int,
       (EXISTS(SELECT 1 FROM assets a
               WHERE a.id = substring(b.cover_image_url from '/files/(.*)$')::uuid))::int,
       0,
       coalesce(nullif(b.cover_image_url, ''), 'ยังไม่ได้ตั้งรูปปก')
FROM branches b

UNION ALL
SELECT '4. blob กำพร้า',
       a.id::text,
       0, 0, 0,
       pg_size_pretty(a.size_bytes::bigint) || ' · ' || a.created_at::date
FROM assets a
WHERE NOT EXISTS (SELECT 1 FROM img WHERE img.asset_id = a.id)

ORDER BY section, item;
