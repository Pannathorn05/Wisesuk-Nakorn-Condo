-- ระบบจองห้องพัก หอพักวิเศษสุขนครคอนโด และหอพักในเครือ
-- Schema เริ่มต้น: ผู้ใช้, สาขา, ห้องพัก, การจอง, การชำระเงิน, activity log

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TYPE user_role       AS ENUM ('member', 'admin', 'superadmin');
CREATE TYPE stay_type       AS ENUM ('daily', 'monthly');
CREATE TYPE room_status     AS ENUM ('available', 'occupied', 'maintenance');
CREATE TYPE booking_status  AS ENUM ('pending_payment', 'awaiting_review', 'approved', 'rejected', 'cancelled', 'completed');
CREATE TYPE payment_status  AS ENUM ('submitted', 'approved', 'rejected');

-- ---------------------------------------------------------------- สาขา
CREATE TABLE branches (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug                TEXT        NOT NULL UNIQUE,
    name                TEXT        NOT NULL,
    tagline             TEXT        NOT NULL DEFAULT '',
    description         TEXT        NOT NULL DEFAULT '',
    address             TEXT        NOT NULL DEFAULT '',
    phones              TEXT[]      NOT NULL DEFAULT '{}',
    line_id             TEXT        NOT NULL DEFAULT '',
    email               TEXT        NOT NULL DEFAULT '',
    latitude            NUMERIC(10, 7),
    longitude           NUMERIC(10, 7),
    map_url             TEXT        NOT NULL DEFAULT '',
    building_count      INT         NOT NULL DEFAULT 1,
    floor_count         INT         NOT NULL DEFAULT 1,
    daily_price_from    NUMERIC(10, 2),
    monthly_price_min   NUMERIC(10, 2),
    monthly_price_max   NUMERIC(10, 2),
    water_rate          NUMERIC(10, 2) NOT NULL DEFAULT 0,   -- บาท/หน่วย
    electric_rate       NUMERIC(10, 2) NOT NULL DEFAULT 0,   -- บาท/หน่วย
    deposit             NUMERIC(10, 2) NOT NULL DEFAULT 0,   -- ค่าประกัน
    advance_payment     NUMERIC(10, 2) NOT NULL DEFAULT 0,   -- ค่าล่วงหน้า
    contract_fee        NUMERIC(10, 2) NOT NULL DEFAULT 0,   -- ค่ายืนยันการทำสัญญา (รายเดือน)
    cover_image_url     TEXT        NOT NULL DEFAULT '',
    is_active           BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE branch_images (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    branch_id   UUID NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
    image_url   TEXT NOT NULL,
    caption     TEXT NOT NULL DEFAULT '',
    sort_order  INT  NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_branch_images_branch ON branch_images (branch_id, sort_order);

-- สิ่งอำนวยความสะดวก (แม่แบบกลาง + ผูกกับสาขา)
CREATE TABLE amenities (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code        TEXT NOT NULL UNIQUE,
    name        TEXT NOT NULL,
    icon        TEXT NOT NULL DEFAULT '',
    sort_order  INT  NOT NULL DEFAULT 0
);

CREATE TABLE branch_amenities (
    branch_id   UUID NOT NULL REFERENCES branches(id)  ON DELETE CASCADE,
    amenity_id  UUID NOT NULL REFERENCES amenities(id) ON DELETE CASCADE,
    PRIMARY KEY (branch_id, amenity_id)
);

-- สถานที่ใกล้เคียง (สถานศึกษา / ห้างสรรพสินค้า / โรงพยาบาล ...)
CREATE TABLE nearby_places (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    branch_id   UUID NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
    category    TEXT NOT NULL,          -- education | shopping | hospital | transport | other
    name        TEXT NOT NULL,
    distance    TEXT NOT NULL DEFAULT '',
    sort_order  INT  NOT NULL DEFAULT 0
);
CREATE INDEX idx_nearby_branch ON nearby_places (branch_id, category, sort_order);

-- ---------------------------------------------------------------- ผู้ใช้งาน
CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email           TEXT        NOT NULL UNIQUE,   -- เก็บเป็นตัวพิมพ์เล็กเสมอ (normalize ที่ชั้น service)
    password_hash   TEXT        NOT NULL,
    first_name      TEXT        NOT NULL,
    last_name       TEXT        NOT NULL,
    phone           TEXT        NOT NULL DEFAULT '',
    role            user_role   NOT NULL DEFAULT 'member',
    branch_id       UUID        REFERENCES branches(id) ON DELETE SET NULL, -- สาขาที่ admin รับผิดชอบ
    avatar_url      TEXT        NOT NULL DEFAULT '',
    is_active       BOOLEAN     NOT NULL DEFAULT TRUE,
    last_login_at   TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    -- admin ต้องผูกกับสาขาเสมอ, member/superadmin ต้องไม่ผูก
    CONSTRAINT admin_requires_branch CHECK (
        (role = 'admin' AND branch_id IS NOT NULL) OR
        (role <> 'admin' AND branch_id IS NULL)
    )
);
CREATE INDEX idx_users_role_branch ON users (role, branch_id);

CREATE TABLE refresh_tokens (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash  TEXT NOT NULL UNIQUE,
    expires_at  TIMESTAMPTZ NOT NULL,
    revoked_at  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_refresh_user ON refresh_tokens (user_id);

-- ---------------------------------------------------------------- ห้องพัก
CREATE TABLE room_types (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    branch_id   UUID NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,                  -- เช่น ห้องพัดลม, ห้องแอร์, ห้องเปล่า
    description TEXT NOT NULL DEFAULT '',
    size_sqm    NUMERIC(6, 2),
    image_url   TEXT NOT NULL DEFAULT '',
    sort_order  INT  NOT NULL DEFAULT 0,
    UNIQUE (branch_id, name)
);

CREATE TABLE rooms (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    branch_id       UUID        NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
    room_type_id    UUID        REFERENCES room_types(id) ON DELETE SET NULL,
    room_number     TEXT        NOT NULL,
    building        TEXT        NOT NULL DEFAULT '1',
    floor           INT         NOT NULL DEFAULT 1,
    stay_type       stay_type   NOT NULL,
    price           NUMERIC(10, 2) NOT NULL,
    water_rate      NUMERIC(10, 2) NOT NULL DEFAULT 0,
    electric_rate   NUMERIC(10, 2) NOT NULL DEFAULT 0,
    size_sqm        NUMERIC(6, 2),
    description     TEXT        NOT NULL DEFAULT '',
    image_url       TEXT        NOT NULL DEFAULT '',
    status          room_status NOT NULL DEFAULT 'available',
    is_active       BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (branch_id, room_number)
);
CREATE INDEX idx_rooms_search ON rooms (branch_id, stay_type, status) WHERE is_active;

CREATE TABLE room_amenities (
    room_id     UUID NOT NULL REFERENCES rooms(id)     ON DELETE CASCADE,
    amenity_id  UUID NOT NULL REFERENCES amenities(id) ON DELETE CASCADE,
    PRIMARY KEY (room_id, amenity_id)
);

CREATE TABLE room_images (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    room_id     UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    image_url   TEXT NOT NULL,
    sort_order  INT  NOT NULL DEFAULT 0
);
CREATE INDEX idx_room_images_room ON room_images (room_id, sort_order);

-- ---------------------------------------------------------------- การจอง
CREATE SEQUENCE booking_code_seq START 1;

CREATE TABLE bookings (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code                TEXT        NOT NULL UNIQUE,
    user_id             UUID        NOT NULL REFERENCES users(id)    ON DELETE RESTRICT,
    branch_id           UUID        NOT NULL REFERENCES branches(id) ON DELETE RESTRICT,
    room_id             UUID        NOT NULL REFERENCES rooms(id)    ON DELETE RESTRICT,
    stay_type           stay_type   NOT NULL,

    -- ผู้เข้าพัก (กรอกในแบบฟอร์มจอง อาจต่างจากเจ้าของบัญชี)
    guest_first_name    TEXT        NOT NULL,
    guest_last_name     TEXT        NOT NULL,
    guest_phone         TEXT        NOT NULL,
    emergency_phone     TEXT        NOT NULL DEFAULT '',
    emergency_relation  TEXT        NOT NULL DEFAULT '',

    -- รายวัน
    check_in_date       DATE,
    check_out_date      DATE,
    nights              INT,
    -- รายเดือน
    move_in_date        DATE,
    contract_date       DATE,                    -- วันที่ต้องการทำสัญญา (ลูกค้าเลือก)
    appointment_at      TIMESTAMPTZ,             -- วันนัดหมายทำสัญญาที่แอดมินกำหนด
    appointment_note    TEXT        NOT NULL DEFAULT '',

    total_amount        NUMERIC(12, 2) NOT NULL,
    status              booking_status NOT NULL DEFAULT 'pending_payment',
    reject_reason       TEXT        NOT NULL DEFAULT '',
    reviewed_by         UUID        REFERENCES users(id) ON DELETE SET NULL,
    reviewed_at         TIMESTAMPTZ,
    cancelled_at        TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT daily_dates_required CHECK (
        stay_type <> 'daily' OR (check_in_date IS NOT NULL AND check_out_date IS NOT NULL AND check_out_date > check_in_date)
    ),
    CONSTRAINT monthly_dates_required CHECK (
        stay_type <> 'monthly' OR move_in_date IS NOT NULL
    )
);
CREATE INDEX idx_bookings_user   ON bookings (user_id, created_at DESC);
CREATE INDEX idx_bookings_branch ON bookings (branch_id, status, created_at DESC);
CREATE INDEX idx_bookings_room   ON bookings (room_id, status);

-- ---------------------------------------------------------------- การชำระเงิน
CREATE TABLE payments (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    booking_id      UUID        NOT NULL REFERENCES bookings(id) ON DELETE CASCADE,
    amount          NUMERIC(12, 2) NOT NULL,
    transferred_at  TIMESTAMPTZ NOT NULL,        -- วันที่+เวลาที่โอน
    slip_url        TEXT        NOT NULL,
    note            TEXT        NOT NULL DEFAULT '',
    status          payment_status NOT NULL DEFAULT 'submitted',
    reviewed_by     UUID        REFERENCES users(id) ON DELETE SET NULL,
    reviewed_at     TIMESTAMPTZ,
    reject_reason   TEXT        NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_payments_booking ON payments (booking_id, created_at DESC);

-- ---------------------------------------------------------------- แจ้งเตือน + activity log
CREATE TABLE notifications (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title       TEXT NOT NULL,
    body        TEXT NOT NULL DEFAULT '',
    link        TEXT NOT NULL DEFAULT '',
    read_at     TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_notifications_user ON notifications (user_id, created_at DESC);

CREATE TABLE activity_logs (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_id    UUID        REFERENCES users(id) ON DELETE SET NULL,
    actor_role  user_role   NOT NULL,
    actor_name  TEXT        NOT NULL DEFAULT '',
    branch_id   UUID        REFERENCES branches(id) ON DELETE SET NULL,
    action      TEXT        NOT NULL,           -- booking.approve, room.update, auth.login ...
    entity_type TEXT        NOT NULL DEFAULT '',
    entity_id   TEXT        NOT NULL DEFAULT '',
    detail      JSONB       NOT NULL DEFAULT '{}'::jsonb,
    ip_address  TEXT        NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_activity_actor  ON activity_logs (actor_id, created_at DESC);
CREATE INDEX idx_activity_branch ON activity_logs (branch_id, created_at DESC);
CREATE INDEX idx_activity_role   ON activity_logs (actor_role, created_at DESC);
