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
		INSERT INTO branches (slug, name, tagline, description, address, phones, line_id,
		                      building_count, floor_count, daily_price_from,
		                      monthly_price_min, monthly_price_max,
		                      water_rate, electric_rate, contract_fee)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		ON CONFLICT (slug) DO UPDATE SET
			name              = EXCLUDED.name,
			tagline           = EXCLUDED.tagline,
			description       = EXCLUDED.description,
			address           = EXCLUDED.address,
			phones            = EXCLUDED.phones,
			line_id           = EXCLUDED.line_id,
			building_count    = EXCLUDED.building_count,
			floor_count       = EXCLUDED.floor_count,
			daily_price_from  = EXCLUDED.daily_price_from,
			monthly_price_min = EXCLUDED.monthly_price_min,
			monthly_price_max = EXCLUDED.monthly_price_max,
			water_rate        = EXCLUDED.water_rate,
			electric_rate     = EXCLUDED.electric_rate,
			contract_fee      = EXCLUDED.contract_fee,
			updated_at        = now()
		RETURNING id`

	type branchSeed struct {
		slug, name, tagline, address, line string
		// ค่าใช้จ่ายที่ยังไม่มีคอลัมน์ของตัวเอง (เช่น ค่าที่จอดรถ) เก็บเป็นข้อความไว้ที่นี่
		description                       string
		phones                            []string
		buildings, floors                 int
		dailyFrom, monthlyMin, monthlyMax float64
		water, electric, contractFee      float64
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
			dailyFrom: 0, monthlyMin: 3500, monthlyMax: 4500,
			water: 18, electric: 5.50, contractFee: 500,
		},
		{
			slug: "charoenkrung-place", name: "เจริญกรุงเพลส",
			tagline: "ใกล้ รพ.เจริญกรุงประชารักษ์ เอเชียทีค Terminal 21 พระราม 3 สาทร",
			address: "10 ซอย เจริญกรุง 6 ถนน มไหสวรรย์ แขวงบางคอแหลม เขตบางคอแหลม กรุงเทพมหานคร 10120",
			phones:  []string{"02-291-8288", "081-859-7398"}, line: "@wisetsuk",
			description: "ค่าที่จอดรถ 1,000 บาท/เดือน",
			buildings:   1, floors: 8,
			dailyFrom: 950, monthlyMin: 7700, monthlyMax: 10000,
			water: 18, electric: 5, contractFee: 500,
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
			b.slug, b.name, b.tagline, b.description, b.address, b.phones, b.line,
			b.buildings, b.floors, dailyFrom, b.monthlyMin, b.monthlyMax,
			b.water, b.electric, b.contractFee,
		).Scan(&id); err != nil {
			return fmt.Errorf("สาขา %s: %w", b.slug, err)
		}
		branchIDs = append(branchIDs, id)
	}

	// ------------------------------------------------------------ สิ่งอำนวยความสะดวก
	amenities := [][3]string{
		{"aircon", "เครื่องปรับอากาศ", "wind"},
		{"parking", "ที่จอดรถ", "car"},
		{"furniture", "เฟอร์นิเจอร์ - ตู้, เตียง", "sofa"},
		{"motorcycle-parking", "ที่จอดรถมอเตอร์ไซด์/จักรยาน", "bike"},
		{"water-heater", "เครื่องทำน้ำอุ่น", "shower"},
		{"fridge", "ตู้เย็น", "refrigerator"},
		{"fan", "พัดลม", "fan"},
		{"lift", "ลิฟต์", "elevator"},
		{"pool", "สระว่ายน้ำ", "pool"},
		{"gym", "โรงยิม / ฟิตเนส", "dumbbell"},
		{"tv", "TV", "tv"},
		{"cable-tv", "เคเบิลทีวี / ดาวเทียม", "satellite"},
		{"keycard", "มีระบบรักษาความปลอดภัย (keycard)", "shield"},
		{"cctv", "กล้องวงจรปิด (CCTV)", "camera"},
		{"security-guard", "รปภ.", "user-shield"},
		{"direct-phone", "โทรศัพท์สายตรง", "phone"},
		{"food-shop", "ร้านขายอาหาร", "utensils"},
		{"convenience-store", "ร้านค้า สะดวกซื้อ", "store"},
		{"wifi", "อินเทอร์เน็ตไร้สาย (WIFI) ในห้อง", "wifi"},
		{"laundry", "ร้านซัก-รีด / มีบริการเครื่องซักผ้า", "washing-machine"},
		{"salon", "ร้านทำผม-เสริมสวย", "scissors"},
	}
	for i, a := range amenities {
		if _, err := tx.Exec(ctx,
			`INSERT INTO amenities (code, name, icon, sort_order) VALUES ($1, $2, $3, $4)
			 ON CONFLICT (code) DO UPDATE SET name       = EXCLUDED.name,
			                                 icon       = EXCLUDED.icon,
			                                 sort_order = EXCLUDED.sort_order`,
			a[0], a[1], a[2], i); err != nil {
			return fmt.Errorf("สิ่งอำนวยความสะดวก %s: %w", a[0], err)
		}
	}

	// สิ่งอำนวยความสะดวกที่แต่ละสาขามีจริง — สาขาที่ไม่ได้ระบุใช้ชุดพื้นฐาน
	defaultBranchAmenities := []string{"wifi", "aircon", "water-heater", "keycard", "cctv", "parking"}
	branchAmenities := map[string][]string{
		"prachauthit-45": {
			"aircon", "parking", "furniture", "motorcycle-parking", "water-heater",
			"fan", "keycard", "cctv", "security-guard", "food-shop",
			"convenience-store", "wifi", "laundry", "salon",
		},
		"bangkhae": {
			"aircon", "parking", "furniture", "motorcycle-parking", "water-heater",
			"fridge", "lift", "tv", "cable-tv", "keycard", "cctv", "security-guard",
			"direct-phone", "food-shop", "convenience-store", "wifi", "laundry", "salon",
		},
		"charoenkrung-place": {
			"aircon", "parking", "furniture", "motorcycle-parking", "water-heater",
			"fridge", "lift", "pool", "gym", "tv", "cable-tv", "keycard", "cctv",
			"security-guard", "food-shop", "convenience-store", "wifi", "laundry", "salon",
		},
	}
	for i, b := range branches {
		codes, ok := branchAmenities[b.slug]
		if !ok {
			codes = defaultBranchAmenities
		}
		// ลบของเดิมก่อนผูกใหม่ เพื่อให้ผลลัพธ์ตรงกับรายการข้างบนเสมอแม้รัน seed ซ้ำ
		if _, err := tx.Exec(ctx,
			`DELETE FROM branch_amenities WHERE branch_id = $1`, branchIDs[i]); err != nil {
			return fmt.Errorf("ล้างสิ่งอำนวยความสะดวกสาขา %s: %w", b.slug, err)
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO branch_amenities (branch_id, amenity_id)
			 SELECT $1, id FROM amenities WHERE code = ANY($2)
			 ON CONFLICT DO NOTHING`, branchIDs[i], codes); err != nil {
			return fmt.Errorf("ผูกสิ่งอำนวยความสะดวกสาขา %s: %w", b.slug, err)
		}
	}

	// ------------------------------------------------------------ สถานที่ใกล้เคียง (ราย branch ตามลำดับใน branches)
	nearbyByBranch := map[int][][3]string{
		// ประชาอุทิศ 45
		0: {
			{"education", "มหาวิทยาลัยเทคโนโลยีพระจอมเกล้าธนบุรี", "600 เมตร"},
			{"education", "วิทยาลัยพณิชยการเชตุพน", "2.4 กิโลเมตร"},
			{"education", "โรงเรียนบางมดวิทยา", "2.4 กิโลเมตร"},
			{"hospital", "โรงพยาบาลบางมด", "4.2 กิโลเมตร"},
			{"hospital", "โรงพยาบาลบางปะกอก 1", "3.4 กิโลเมตร"},
			{"hospital", "โรงพยาบาลบางปะกอก 3", "3.4 กิโลเมตร"},
			{"shopping", "เดอะไบรท์ พระราม 2", "700 เมตร"},
			{"shopping", "เมเจอร์ ฮอลลีวู้ด สุขสวัสดิ์", "2.2 กิโลเมตร"},
			{"shopping", "เทสโก้โลตัส บางปะกอก", "3.3 กิโลเมตร"},
			{"shopping", "บิ๊กซี ราษฎร์บูรณะ", "3.9 กิโลเมตร"},
			{"shopping", "บิ๊กซี เอ็กซ์ตร้า บางปะกอก", "4 กิโลเมตร"},
			{"shopping", "บิ๊กซี สุขสวัสดิ์", "2.4 กิโลเมตร"},
		},
		// บางแค
		1: {
			{"education", "วิทยาลัยเทคโนโลยีกรุงธน บางแค", "2.9 กิโลเมตร"},
			{"education", "มหาวิทยาลัยสยาม", "5.4 กิโลเมตร"},
			{"education", "มหาวิทยาลัยเอเชียอาคเนย์", "5.6 กิโลเมตร"},
			{"education", "วิทยาลัยเทคนิคราชสิทธาราม", "6.6 กิโลเมตร"},
			{"education", "วิทยาลัยการจัดการเพชรเกษม", "8 กิโลเมตร"},
			{"hospital", "โรงพยาบาลเกษมราษฎร์ บางแค", "1.4 กิโลเมตร"},
			{"hospital", "โรงพยาบาลมิตรประชา", "2.9 กิโลเมตร"},
			{"hospital", "โรงพยาบาลบางมด 3", "4.4 กิโลเมตร"},
			{"shopping", "เดอะ มอลล์ บางแค", "1.3 กิโลเมตร"},
			{"shopping", "เทสโก้โลตัส บางแค", "1.8 กิโลเมตร"},
			{"shopping", "แม็คโคร บางบอน", "3.5 กิโลเมตร"},
			{"shopping", "บิ๊กซี เอ็กซ์ตร้า บางบอน", "4 กิโลเมตร"},
			{"shopping", "ตลาดบางแค", "2 กิโลเมตร"},
			{"shopping", "บิ๊กซี กัลปพฤกษ์", "2.9 กิโลเมตร"},
			{"shopping", "โฮมโปร กัลปพฤกษ์", "3.2 กิโลเมตร"},
		},
		// เจริญกรุงเพลส
		2: {
			{"education", "วิทยาลัยพยาบาลกองทัพเรือ", "1.6 กิโลเมตร"},
			{"education", "วิทยาลัยพณิชยการเชตุพน", "3.4 กิโลเมตร"},
			{"education", "มหาวิทยาลัยสยาม", "5.1 กิโลเมตร"},
			{"hospital", "โรงพยาบาลเจริญกรุงประชารักษ์", "380 เมตร"},
			{"hospital", "โรงพยาบาลสมิติเวช ธนบุรี", "1.9 กิโลเมตร"},
			{"hospital", "โรงพยาบาลสมเด็จพระปิ่นเกล้า", "1.7 กิโลเมตร"},
			{"hospital", "โรงพยาบาลบางปะกอก 1", "2.2 กิโลเมตร"},
			{"hospital", "โรงพยาบาลประชาพัฒน์", "2.3 กิโลเมตร"},
			{"hospital", "โรงพยาบาลราษฎร์บูรณะ", "2.4 กิโลเมตร"},
			{"hospital", "โรงพยาบาลสุขสวัสดิ์", "2.5 กิโลเมตร"},
			{"hospital", "โรงพยาบาลบางปะกอก 9 อินเตอร์เนชั่นแนล", "2.9 กิโลเมตร"},
			{"shopping", "บิ๊กซี ดาวคะนอง", "1.3 กิโลเมตร"},
			{"shopping", "บิ๊กซี เอ็กซ์ตร้า บางปะกอก", "1.6 กิโลเมตร"},
			{"shopping", "บิ๊กซี ราษฎร์บูรณะ", "1.7 กิโลเมตร"},
			{"shopping", "เทสโก้โลตัส บางปะกอก", "2.3 กิโลเมตร"},
			{"shopping", "เดอะ มอลล์ ท่าพระ", "2.4 กิโลเมตร"},
			{"shopping", "โฮมโปร พระราม 3", "2.8 กิโลเมตร"},
			{"shopping", "เทสโก้โลตัส พระราม 3", "5 กิโลเมตร"},
			{"shopping", "ตลาดสดบางปะกอก", "2.3 กิโลเมตร"},
		},
	}
	for idx, places := range nearbyByBranch {
		// ลบของเดิมก่อนเพิ่มใหม่ เพื่อให้ผลลัพธ์ตรงกับรายการข้างบนเสมอแม้รัน seed ซ้ำ
		if _, err := tx.Exec(ctx,
			`DELETE FROM nearby_places WHERE branch_id = $1`, branchIDs[idx]); err != nil {
			return fmt.Errorf("ล้างสถานที่ใกล้เคียงสาขา %s: %w", branches[idx].slug, err)
		}
		for i, n := range places {
			if _, err := tx.Exec(ctx,
				`INSERT INTO nearby_places (branch_id, category, name, distance, sort_order)
				 VALUES ($1, $2, $3, $4, $5)`,
				branchIDs[idx], n[0], n[1], n[2], i); err != nil {
				return fmt.Errorf("เพิ่มสถานที่ใกล้เคียงสาขา %s: %w", branches[idx].slug, err)
			}
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
