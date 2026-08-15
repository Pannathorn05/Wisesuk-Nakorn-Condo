// Package timex เป็นจุดเดียวในระบบที่ตัดสินว่า "วันนี้ / อดีต / อนาคต" คือเมื่อไร
//
// กติกาตาม AC-22:
//   - เก็บทุก timestamp เป็น UTC ในฐานข้อมูล
//   - ตัดสินวันที่ด้วยเขตเวลา Asia/Bangkok "เสมอ" ไม่ใช่เขตเวลาของเครื่องที่รันอยู่
//
// เหตุผลที่ต้องมี package นี้: ถ้าปล่อยให้โค้ดเรียก time.Now() แล้วอ่านวัน/เดือน/ปีตรง ๆ
// ผลลัพธ์จะขึ้นกับ TZ ของ process ซึ่งถูกเฉพาะตอนรันใน container ที่ตั้ง TZ=Asia/Bangkok
// พอรันบน CI หรือเครื่อง dev ที่เป็น UTC การจอง "วันนี้" ระหว่างเวลาไทย 00:00–06:59
// จะถูกมองว่าเป็นเมื่อวานแล้วโดนปฏิเสธทั้งที่ถูกต้อง
//
// ฟังก์ชันที่ลงท้ายด้วย At รับเวลาปัจจุบันเข้ามาเป็นพารามิเตอร์ เทสจึงตรึงเวลาได้
// โดยไม่ต้องแก้ตัวแปรระดับ package (ซึ่งจะทำให้เทสที่รันขนานกันชนกัน)
package timex

import (
	"time"

	// ฝัง tzdata ไว้ในไบนารี เพื่อให้ LoadLocation ใช้ได้แม้เครื่องปลายทางไม่มีฐานข้อมูลเขตเวลา
	_ "time/tzdata"
)

const zoneName = "Asia/Bangkok"

var location = mustLoadLocation(zoneName)

func mustLoadLocation(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		// เกิดไม่ได้เพราะฝัง tzdata ไว้แล้ว ถ้าเกิดแปลว่าไบนารีถูกประกอบผิด
		panic("timex: โหลดเขตเวลา " + name + " ไม่สำเร็จ: " + err.Error())
	}
	return loc
}

// Location คืนเขตเวลาที่ระบบใช้ตัดสินวันที่
func Location() *time.Location { return location }

// Now คืนเวลาปัจจุบันในเขตเวลาไทย
func Now() time.Time { return time.Now().In(location) }

// CalendarDate ตัดเวลาทิ้งเหลือแต่วันที่ โดยตีความ t ตามเขตเวลาไทย
//
// ค่าที่คืนถูกประทับเป็น UTC midnight เพื่อให้เทียบกันได้ตรง ๆ ด้วย Before/After/Equal
// โดยไม่ต้องกังวลเรื่อง offset — มันคือ "ป้ายวันที่" ไม่ใช่จุดเวลาจริง
func CalendarDate(t time.Time) time.Time {
	b := t.In(location)
	return time.Date(b.Year(), b.Month(), b.Day(), 0, 0, 0, 0, time.UTC)
}

// Today คืนวันที่ของวันนี้ตามเวลาไทย
func Today() time.Time { return TodayAt(time.Now()) }

// TodayAt คือ Today ที่ระบุเวลาปัจจุบันเองได้ ใช้ในเทส
func TodayAt(now time.Time) time.Time { return CalendarDate(now) }

// IsPastDate บอกว่า date เป็นวันที่ผ่านมาแล้วหรือยัง เทียบตามปฏิทินไทย
// วันนี้ไม่ถือว่าเป็นอดีต (AC-6: check_in เป็นวันนี้ใช้ได้)
func IsPastDate(date time.Time) bool { return IsPastDateAt(date, time.Now()) }

// IsPastDateAt คือ IsPastDate ที่ระบุเวลาปัจจุบันเองได้ ใช้ในเทส
func IsPastDateAt(date, now time.Time) bool {
	return CalendarDate(date).Before(CalendarDate(now))
}

// ToUTC แปลงเวลาเป็น UTC ก่อนเก็บลงฐานข้อมูล
func ToUTC(t time.Time) time.Time { return t.UTC() }

// FormatISO คืนสตริง ISO 8601 พร้อม offset สำหรับส่งกลับใน response
func FormatISO(t time.Time) string { return t.In(location).Format(time.RFC3339) }
