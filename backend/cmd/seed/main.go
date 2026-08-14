// Command seed ใส่ข้อมูลตั้งต้นตาม prototype (3 สาขา, สิ่งอำนวยความสะดวก,
// ประเภทห้อง, ห้องตัวอย่าง, บัญชี superadmin/admin)
//
// รันซ้ำได้ปลอดภัย — ใช้ ON CONFLICT DO NOTHING/UPDATE ทุกจุด
//
// หมายเหตุ: ราคาและรายละเอียดบางส่วนเป็นข้อมูลสมมุติจาก prototype
// เมื่อเริ่มใช้งานจริงให้แก้ผ่านหน้า "จัดการรายละเอียดสาขา" ของแอดมิน
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"backend/internal/auth"
	"backend/internal/config"
	"backend/internal/database"
)

func main() {
	password := flag.String("password", "", "รหัสผ่านตั้งต้นของบัญชีผู้ดูแล (ค่าเริ่มต้นอ่านจาก SEED_DEFAULT_PASSWORD)")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	if *password == "" {
		*password = cfg.SeedAdminSecret
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	pool, err := database.Connect(ctx, cfg.DatabaseURL, cfg.DBMaxConns, cfg.DBMinConns, cfg.DBConnectTimout)
	if err != nil {
		log.Fatalf("เชื่อมต่อฐานข้อมูลไม่สำเร็จ: %v", err)
	}
	defer pool.Close()

	if err := database.Migrate(ctx, pool); err != nil {
		log.Fatalf("migrate ไม่สำเร็จ: %v", err)
	}

	authMgr := auth.NewManager(cfg.JWTSecret, cfg.AccessTokenTTL, cfg.RefreshTokenTTL, cfg.BcryptCost)
	hash, err := authMgr.HashPassword(*password)
	if err != nil {
		log.Fatalf("hash รหัสผ่านไม่สำเร็จ: %v", err)
	}

	if err := seed(ctx, pool, hash); err != nil {
		log.Fatalf("seed ไม่สำเร็จ: %v", err)
	}

	fmt.Println("ใส่ข้อมูลตั้งต้นเรียบร้อย")
	fmt.Println("  superadmin : super@wisetsuk.com")
	fmt.Println("  admin      : admin.1@wisetsuk.com (ประชาอุทิศ 45)")
	fmt.Println("  admin      : admin.2@wisetsuk.com (บางแค)")
	fmt.Println("  admin      : admin.3@wisetsuk.com (เจริญกรุงเพลส)")
	fmt.Printf("  รหัสผ่าน   : %s\n", *password)
	fmt.Fprintln(os.Stderr, "\nคำเตือน: เปลี่ยนรหัสผ่านทันทีหลังเข้าสู่ระบบครั้งแรก")
}

func seed(ctx context.Context, pool *pgxpool.Pool, passwordHash string) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// ------------------------------------------------------------ สาขา
	const upsertBranch = `
		INSERT INTO branches (slug, name, tagline, address, phones, line_id,
		                      building_count, floor_count, daily_price_from,
		                      monthly_price_min, monthly_price_max,
		                      water_rate, electric_rate, contract_fee)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		ON CONFLICT (slug) DO UPDATE SET name = EXCLUDED.name, updated_at = now()
		RETURNING id`

	type branchSeed struct {
		slug, name, tagline, address, line string
		phones                             []string
		buildings, floors                  int
		dailyFrom, monthlyMin, monthlyMax  float64
		water, electric, contractFee       float64
	}

	branches := []branchSeed{
		{
			slug: "prachauthit-45", name: "วิเศษสุขนครคอนโด ประชาอุทิศ 45",
			tagline: "ห่างจากมหาวิทยาลัยเทคโนโลยีพระจอมเกล้าธนบุรี เพียง 600 เมตร",
			address: "679, 223 ซอย ประชาอุทิศ 45 แขวงบางมด เขตทุ่งครุ กรุงเทพมหานคร 10140",
			phones:  []string{"02-872-7800", "084-702-8970"}, line: "@wisetsuk",
			buildings: 2, floors: 5,
			dailyFrom: 500, monthlyMin: 2200, monthlyMax: 3400,
			water: 17, electric: 7, contractFee: 500,
		},
		{
			slug: "bangkhae", name: "วิเศษสุขนครคอนโด บางแค",
			tagline: "ห่างจาก The Mall บางแค 2 กิโลเมตร มีเฟอร์นิเจอร์ครบครัน",
			address: "111 106 แขวงบางแค เขตบางแค กรุงเทพมหานคร 10160",
			phones:  []string{"02-454-0888", "085-047-5904"}, line: "@wisetsuk",
			buildings: 1, floors: 8,
			dailyFrom: 0, monthlyMin: 3000, monthlyMax: 4500,
			water: 17, electric: 7, contractFee: 500,
		},
		{
			slug: "charoenkrung-place", name: "เจริญกรุงเพลส",
			tagline: "ใกล้ รพ.เจริญกรุงประชารักษ์ เอเชียทีค Terminal 21 พระราม 3 สาทร",
			address: "10 ซอย เจริญกรุง 6 ถนน มไหสวรรย์ แขวงบางคอแหลม เขตบางคอแหลม กรุงเทพมหานคร 10120",
			phones:  []string{"02-291-8288", "081-859-7398"}, line: "@wisetsuk",
			buildings: 1, floors: 7,
			dailyFrom: 600, monthlyMin: 3400, monthlyMax: 4800,
			water: 17, electric: 7, contractFee: 500,
		},
	}

	branchIDs := make([]string, 0, len(branches))
	for _, b := range branches {
		var id string
		var dailyFrom any
		if b.dailyFrom > 0 {
			dailyFrom = b.dailyFrom
		}
		if err := tx.QueryRow(ctx, upsertBranch,
			b.slug, b.name, b.tagline, b.address, b.phones, b.line,
			b.buildings, b.floors, dailyFrom, b.monthlyMin, b.monthlyMax,
			b.water, b.electric, b.contractFee,
		).Scan(&id); err != nil {
			return fmt.Errorf("สาขา %s: %w", b.slug, err)
		}
		branchIDs = append(branchIDs, id)
	}

	// ------------------------------------------------------------ สิ่งอำนวยความสะดวก
	amenities := [][3]string{
		{"tv", "TV", "tv"},
		{"aircon", "เครื่องปรับอากาศ", "wind"},
		{"wifi", "ไวไฟ", "wifi"},
		{"fridge", "ตู้เย็น", "refrigerator"},
		{"water-heater", "เครื่องทำน้ำอุ่น", "shower"},
		{"furniture", "เฟอร์นิเจอร์ครบครัน", "sofa"},
		{"keycard", "ระบบรักษาความปลอดภัย (keycard)", "shield"},
		{"cctv", "กล้องวงจรปิด", "camera"},
		{"lift", "ลิฟต์", "elevator"},
		{"parking", "ที่จอดรถ", "car"},
		{"laundry", "เครื่องซักผ้าหยอดเหรียญ", "washing-machine"},
		{"convenience-store", "ร้านสะดวกซื้อ", "store"},
	}
	for i, a := range amenities {
		if _, err := tx.Exec(ctx,
			`INSERT INTO amenities (code, name, icon, sort_order) VALUES ($1, $2, $3, $4)
			 ON CONFLICT (code) DO UPDATE SET name = EXCLUDED.name`,
			a[0], a[1], a[2], i); err != nil {
			return fmt.Errorf("สิ่งอำนวยความสะดวก %s: %w", a[0], err)
		}
	}
	// ผูกสิ่งอำนวยความสะดวกพื้นฐานให้ทุกสาขา
	for _, branchID := range branchIDs {
		if _, err := tx.Exec(ctx,
			`INSERT INTO branch_amenities (branch_id, amenity_id)
			 SELECT $1, id FROM amenities
			 WHERE code IN ('wifi', 'aircon', 'water-heater', 'keycard', 'cctv', 'parking')
			 ON CONFLICT DO NOTHING`, branchID); err != nil {
			return err
		}
	}

	// ------------------------------------------------------------ สถานที่ใกล้เคียง (สาขาประชาอุทิศ 45)
	nearby := [][3]string{
		{"education", "มหาวิทยาลัยเทคโนโลยีพระจอมเกล้าธนบุรี", "600 เมตร"},
		{"education", "โรงเรียนบางมดวิทยา", "2.4 กิโลเมตร"},
		{"shopping", "เดอะไบรท์ พระราม 2", "700 เมตร"},
		{"hospital", "โรงพยาบาลบางมด", "4.2 กิโลเมตร"},
	}
	for i, n := range nearby {
		if _, err := tx.Exec(ctx,
			`INSERT INTO nearby_places (branch_id, category, name, distance, sort_order)
			 SELECT $1, $2, $3, $4, $5
			 WHERE NOT EXISTS (
			   SELECT 1 FROM nearby_places WHERE branch_id = $1 AND name = $3
			 )`,
			branchIDs[0], n[0], n[1], n[2], i); err != nil {
			return err
		}
	}

	// ------------------------------------------------------------ ประเภทห้อง + ห้องตัวอย่าง
	roomTypes := []struct{ name, desc string }{
		{"ห้องแอร์", "ห้องพักติดเครื่องปรับอากาศ พร้อมเฟอร์นิเจอร์"},
		{"ห้องพัดลม", "ห้องพักพัดลม พร้อมเฟอร์นิเจอร์พื้นฐาน"},
		{"ห้องเปล่า", "ห้องเปล่า ตกแต่งเองได้"},
	}
	for _, branchID := range branchIDs {
		for i, rt := range roomTypes {
			if _, err := tx.Exec(ctx,
				`INSERT INTO room_types (branch_id, name, description, sort_order)
				 VALUES ($1, $2, $3, $4) ON CONFLICT (branch_id, name) DO NOTHING`,
				branchID, rt.name, rt.desc, i); err != nil {
				return err
			}
		}
	}

	// ห้องตัวอย่างของสาขาประชาอุทิศ 45 ตามเลขห้องที่ปรากฏใน prototype
	sampleRooms := []struct {
		number, building string
		floor            int
		stayType         string
		price            float64
		sizeSqm          float64
		typeName         string
	}{
		{"201", "1", 2, "daily", 500, 21, "ห้องแอร์"},
		{"210", "1", 2, "daily", 500, 21, "ห้องแอร์"},
		{"302", "1", 3, "monthly", 2900, 21, "ห้องแอร์"},
		{"409", "2", 4, "monthly", 3200, 21, "ห้องแอร์"},
		{"502", "2", 5, "daily", 500, 21, "ห้องพัดลม"},
		{"511", "2", 5, "monthly", 2900, 21, "ห้องเปล่า"},
	}
	for _, rm := range sampleRooms {
		if _, err := tx.Exec(ctx,
			`INSERT INTO rooms (branch_id, room_type_id, room_number, building, floor,
			                    stay_type, price, water_rate, electric_rate, size_sqm)
			 VALUES ($1,
			         (SELECT id FROM room_types WHERE branch_id = $1 AND name = $2),
			         $3, $4, $5, $6, $7, 17, 7, $8)
			 ON CONFLICT (branch_id, room_number) DO NOTHING`,
			branchIDs[0], rm.typeName, rm.number, rm.building, rm.floor,
			rm.stayType, rm.price, rm.sizeSqm); err != nil {
			return fmt.Errorf("ห้อง %s: %w", rm.number, err)
		}
	}

	// ------------------------------------------------------------ บัญชีผู้ดูแล
	if _, err := tx.Exec(ctx,
		`INSERT INTO users (email, password_hash, first_name, last_name, phone, role)
		 VALUES ('super@wisetsuk.com', $1, 'สมชาย', 'วิเศษสุข', '02-872-7800', 'superadmin')
		 ON CONFLICT (email) DO NOTHING`, passwordHash); err != nil {
		return fmt.Errorf("superadmin: %w", err)
	}

	adminSeeds := []struct{ email, first, last string }{
		{"admin.1@wisetsuk.com", "วิรัช", "มั่นคง"},
		{"admin.2@wisetsuk.com", "ศิริพร", "แสงทอง"},
		{"admin.3@wisetsuk.com", "สมหญิง", "งานดี"},
	}
	for i, a := range adminSeeds {
		if _, err := tx.Exec(ctx,
			`INSERT INTO users (email, password_hash, first_name, last_name, role, branch_id)
			 VALUES ($1, $2, $3, $4, 'admin', $5)
			 ON CONFLICT (email) DO NOTHING`,
			a.email, passwordHash, a.first, a.last, branchIDs[i]); err != nil {
			return fmt.Errorf("admin %s: %w", a.email, err)
		}
	}

	return tx.Commit(ctx)
}
