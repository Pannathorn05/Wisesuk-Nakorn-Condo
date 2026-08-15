// เทสของ AC-22: เวลาและวันที่
//
// checklist เทียบกับ AC-22 ใน docs/SPEC.md ทีละบรรทัด
//
//	บรรทัดที่ 1  เก็บทุก timestamp เป็น UTC ในฐานข้อมูล        -> TestAC22_StoredAsUTC
//	บรรทัดที่ 2  ตัดสินวันที่ด้วย Asia/Bangkok เสมอ
//	               check_in_date                              -> TestAC22_IsPastDate_CheckInDate
//	               move_in_date                               -> TestAC22_IsPastDate_MoveInDate
//	               transferred_at                             -> TestAC22_ZonelessInputIsBangkok/transferred_at
//	               appointment_at                             -> TestAC22_ZonelessInputIsBangkok/appointment_at
//	บรรทัดที่ 3  เซิร์ฟเวอร์ UTC + จองวันนี้ตอน 00:30 ไทย -> ผ่าน -> TestAC22_ServerInUTCBookingTodayAt0030
//	             และผลต้องไม่ขึ้นกับเขตเวลาของเครื่อง          -> TestAC22_DecisionIndependentOfServerZone
//	บรรทัดที่ 4  วันที่ใน response เป็น ISO 8601 พร้อม offset   -> TestAC22_ResponseIsISO8601WithOffset
//
// ทุกเทสตรึงเวลาปัจจุบันเองผ่านฟังก์ชันที่ลงท้ายด้วย At จึงให้ผลเหมือนเดิมทุกชั่วโมงที่รัน
package timex_test

import (
	"encoding/json"
	"regexp"
	"strings"
	"testing"
	"time"

	"backend/internal/httpx"
	"backend/internal/timex"
)

// เวลาอ้างอิงของทุกเทส: กรุงเทพ 14 ส.ค. 2026 เวลา 00:30 น.
// ขณะนั้น UTC ยังเป็นวันที่ 13 ส.ค. เวลา 17:30 น. — คนละวันกัน
// นี่คือช่วงเวลาเดียวที่จับบั๊กการเทียบวันที่ด้วย UTC ได้ (เวลาไทย 00:00–06:59)
var (
	bangkokEarlyMorning = time.Date(2026, 8, 13, 17, 30, 0, 0, time.UTC) // = 14 ส.ค. 00:30 ตามเวลาไทย
	todayInBangkok      = date(2026, 8, 14)
	yesterdayInBangkok  = date(2026, 8, 13)
	tomorrowInBangkok   = date(2026, 8, 15)
)

// date สร้าง "ป้ายวันที่" แบบเดียวกับที่ได้จากการพาร์ส YYYY-MM-DD ในโค้ดจริง
func date(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// ---------------------------------------------------------------- บรรทัดที่ 2 และ 3

func TestAC22_CalendarDateUsesBangkok(t *testing.T) {
	t.Parallel()

	got := timex.CalendarDate(bangkokEarlyMorning)
	if !got.Equal(todayInBangkok) {
		t.Fatalf("CalendarDate = %s ต้องการ %s — ต้องอ่านวันที่ตามเวลาไทย ไม่ใช่ UTC",
			got.Format("2006-01-02"), todayInBangkok.Format("2006-01-02"))
	}
}

func TestAC22_TodayAtIsBangkokDate(t *testing.T) {
	t.Parallel()

	if got := timex.TodayAt(bangkokEarlyMorning); !got.Equal(todayInBangkok) {
		t.Fatalf("TodayAt = %s ต้องการ %s", got.Format("2006-01-02"), todayInBangkok.Format("2006-01-02"))
	}
}

func TestAC22_IsPastDate_CheckInDate(t *testing.T) {
	t.Parallel()
	assertPastRules(t, "check_in_date")
}

func TestAC22_IsPastDate_MoveInDate(t *testing.T) {
	t.Parallel()
	assertPastRules(t, "move_in_date")
}

// assertPastRules ตรวจกติกาเดียวกันที่ใช้กับทั้ง check_in_date และ move_in_date
func assertPastRules(t *testing.T, field string) {
	t.Helper()

	cases := []struct {
		name     string
		value    time.Time
		wantPast bool
	}{
		{"เมื่อวานตามเวลาไทย เป็นอดีต", yesterdayInBangkok, true},
		{"วันนี้ตามเวลาไทย ไม่เป็นอดีต", todayInBangkok, false},
		{"พรุ่งนี้ตามเวลาไทย ไม่เป็นอดีต", tomorrowInBangkok, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := timex.IsPastDateAt(tc.value, bangkokEarlyMorning)
			if got != tc.wantPast {
				t.Errorf("%s = %s: IsPastDateAt = %v ต้องการ %v",
					field, tc.value.Format("2006-01-02"), got, tc.wantPast)
			}
		})
	}
}

// TestAC22_ServerInUTCBookingTodayAt0030 คือเคสหลักของ AC-22
//
// จำลองเซิร์ฟเวอร์ที่ไม่ได้ตั้ง TZ (นาฬิกาเป็น UTC) ขณะเวลาไทย 00:30
// การจองวันนี้ตามปฏิทินไทยต้องผ่าน ไม่ใช่ถูกตีกลับว่าเป็นวันที่ผ่านมาแล้ว
func TestAC22_ServerInUTCBookingTodayAt0030(t *testing.T) {
	t.Parallel()

	nowOnUTCServer := bangkokEarlyMorning.In(time.UTC)
	if nowOnUTCServer.Format("2006-01-02") == todayInBangkok.Format("2006-01-02") {
		t.Fatal("เคสนี้ต้องเป็นช่วงที่วันที่ UTC ต่างจากวันที่ไทย ไม่งั้นไม่ได้ทดสอบอะไร")
	}

	if timex.IsPastDateAt(todayInBangkok, nowOnUTCServer) {
		t.Error("จอง check_in = วันนี้ตามเวลาไทย ตอน 00:30 น. ถูกมองว่าเป็นอดีต " +
			"(เทียบวันที่ด้วย UTC แทนที่จะเป็น Asia/Bangkok)")
	}
}

// TestAC22_DecisionIndependentOfServerZone ยืนยันคำว่า "เสมอ" ใน AC-22
// จุดเวลาเดียวกันที่แสดงในคนละเขตเวลา ต้องให้คำตอบเหมือนกันทุกครั้ง
func TestAC22_DecisionIndependentOfServerZone(t *testing.T) {
	t.Parallel()

	zones := map[string]*time.Location{
		"UTC":          time.UTC,
		"Asia/Bangkok": timex.Location(),
		"UTC-11":       time.FixedZone("UTC-11", -11*3600),
		"UTC+13":       time.FixedZone("UTC+13", 13*3600),
	}

	for _, value := range []time.Time{yesterdayInBangkok, todayInBangkok, tomorrowInBangkok} {
		want := timex.IsPastDateAt(value, bangkokEarlyMorning)
		for name, loc := range zones {
			got := timex.IsPastDateAt(value, bangkokEarlyMorning.In(loc))
			if got != want {
				t.Errorf("วันที่ %s: เครื่องอยู่โซน %s ให้ผล %v แต่โซนอื่นให้ %v — การตัดสินต้องไม่ขึ้นกับเขตเวลาของเครื่อง",
					value.Format("2006-01-02"), name, got, want)
			}
		}
	}
}

func TestAC22_LocationIsBangkok(t *testing.T) {
	t.Parallel()

	_, offset := bangkokEarlyMorning.In(timex.Location()).Zone()
	const wantOffset = 7 * 3600
	if offset != wantOffset {
		t.Fatalf("offset = %d วินาที ต้องการ %d (UTC+7)", offset, wantOffset)
	}
}

// ---------------------------------------------------------------- บรรทัดที่ 2 (สองฟิลด์ที่รับเวลา)

// TestAC22_ZonelessInputIsBangkok ครอบ transferred_at และ appointment_at
// ทั้งคู่รับค่าจากผู้ใช้ผ่าน httpx.ParseFlexibleTime ซึ่งเป็นจุดเดียวที่ตีความเวลาที่ไม่มี offset
func TestAC22_ZonelessInputIsBangkok(t *testing.T) {
	t.Parallel()

	// ผู้ใช้กรอก "14 ส.ค. 2026 เวลา 00:30" ซึ่งหมายถึงเวลาไทย
	// จุดเวลาจริงจึงเท่ากับ 13 ส.ค. 17:30 UTC
	cases := []struct {
		name string
		raw  string
	}{
		{"transferred_at", "2026-08-14 00:30"},
		{"appointment_at", "2026-08-14T00:30"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := httpx.ParseFlexibleTime(tc.raw)
			if err != nil {
				t.Fatalf("พาร์ส %q ไม่สำเร็จ: %v", tc.raw, err)
			}
			if !got.Equal(bangkokEarlyMorning) {
				t.Errorf("%q ถูกตีความเป็น %s ต้องการ %s — เวลาที่ไม่มี offset ต้องถือเป็นเวลาไทย",
					tc.raw, got.UTC().Format(time.RFC3339), bangkokEarlyMorning.Format(time.RFC3339))
			}
			if _, offset := got.Zone(); offset != 7*3600 {
				t.Errorf("%q ได้ offset %d วินาที ต้องการ %d", tc.raw, offset, 7*3600)
			}
		})
	}
}

func TestAC22_ZonedInputIsRespected(t *testing.T) {
	t.Parallel()

	got, err := httpx.ParseFlexibleTime("2026-08-13T17:30:00Z")
	if err != nil {
		t.Fatalf("พาร์สค่าที่มี offset ไม่สำเร็จ: %v", err)
	}
	if !got.Equal(bangkokEarlyMorning) {
		t.Errorf("ค่าที่ระบุ offset มาเองต้องถูกเคารพ ได้ %s", got.UTC().Format(time.RFC3339))
	}
}

// ---------------------------------------------------------------- บรรทัดที่ 1

func TestAC22_StoredAsUTC(t *testing.T) {
	t.Parallel()

	bangkokValue := bangkokEarlyMorning.In(timex.Location())
	stored := timex.ToUTC(bangkokValue)

	if _, offset := stored.Zone(); offset != 0 {
		t.Errorf("ค่าที่จะเก็บลงฐานข้อมูลต้องเป็น UTC แต่ offset = %d", offset)
	}
	if !stored.Equal(bangkokEarlyMorning) {
		t.Errorf("การแปลงเป็น UTC ต้องไม่เปลี่ยนจุดเวลา: ได้ %s ต้องการ %s",
			stored.Format(time.RFC3339), bangkokEarlyMorning.Format(time.RFC3339))
	}
	if got := stored.Format("2006-01-02"); got != "2026-08-13" {
		t.Errorf("จุดเวลาเดียวกันใน UTC ต้องเป็นวันที่ 2026-08-13 แต่ได้ %s", got)
	}
}

// ---------------------------------------------------------------- บรรทัดที่ 4

var iso8601WithOffset = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?(Z|[+-]\d{2}:\d{2})$`)

func TestAC22_ResponseIsISO8601WithOffset(t *testing.T) {
	t.Parallel()

	t.Run("FormatISO", func(t *testing.T) {
		got := timex.FormatISO(bangkokEarlyMorning)
		if !iso8601WithOffset.MatchString(got) {
			t.Fatalf("FormatISO = %q ไม่ใช่ ISO 8601 ที่มี offset", got)
		}
		if !strings.HasSuffix(got, "+07:00") {
			t.Errorf("FormatISO = %q ต้องลงท้ายด้วย +07:00", got)
		}
	})

	// ค่าที่ handler ส่งกลับผ่าน encoding/json ต้องอยู่ในรูปแบบเดียวกัน
	t.Run("ผ่าน encoding/json", func(t *testing.T) {
		payload := struct {
			CheckInDate   time.Time `json:"check_in_date"`
			TransferredAt time.Time `json:"transferred_at"`
		}{
			CheckInDate:   todayInBangkok,
			TransferredAt: bangkokEarlyMorning.In(timex.Location()),
		}

		raw, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("marshal ไม่สำเร็จ: %v", err)
		}

		var decoded map[string]string
		if err := json.Unmarshal(raw, &decoded); err != nil {
			t.Fatalf("unmarshal ไม่สำเร็จ: %v", err)
		}
		for field, value := range decoded {
			if !iso8601WithOffset.MatchString(value) {
				t.Errorf("%s = %q ไม่ใช่ ISO 8601 ที่มี offset", field, value)
			}
		}
	})
}
