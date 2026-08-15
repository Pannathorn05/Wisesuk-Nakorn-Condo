// เทสของ AC-10: อนุมัติ/ปฏิเสธการจอง และ state machine 4 สถานะ
//
// checklist เทียบกับ AC-10 ใน docs/SPEC.md ทีละบรรทัด
//
//	อนุมัติใบ awaiting_review -> 200 approved + payment approved  -> approve/สำเร็จ, approve/payment_ล่าสุดเป็น_approved
//	อนุมัติใบรายเดือน -> ห้องเป็น occupied                          -> approve/ใบรายเดือนทำให้ห้อง_occupied
//	ปฏิเสธ -> 200 และใบกลับเป็น pending_payment                    -> reject/กลับไปเป็น_pending_payment
//	ปฏิเสธแล้วห้องยังถูกล็อก                                        -> reject/ห้องยังถูกล็อกอยู่
//	reason เก็บที่แถว payment ไม่ใช่ที่ใบจอง                         -> reject/reason_อยู่ที่_payment_ไม่ใช่ใบจอง
//	ปฏิเสธโดยไม่ระบุ reason -> 422                                 -> reject/ไม่ระบุ_reason
//	อนุมัติ/ปฏิเสธใบที่ไม่ได้อยู่ awaiting_review -> 409            -> invalid_state/*
//	admin สาขา A อนุมัติใบสาขา B -> 404                            -> ข้ามสาขา/*
//	อนุมัติหรือปฏิเสธสำเร็จ -> แจ้งเตือน 1 รายการอ้าง booking_code  -> แจ้งเตือน/*
//	ยิง approve พร้อมกัน 2 request -> สำเร็จ 1 อีกครั้ง 409          -> approve_พร้อมกัน
//	สถานะมีได้แค่ 4 ค่า                                            -> enum_มีแค่_4_สถานะ
package booking_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"backend/internal/testsupport"
	"backend/internal/timex"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	m.Run()
}

// ---------------------------------------------------------------- ตัวช่วย

type client struct {
	app testsupport.App
}

func (c client) do(t *testing.T, method, path, token, body string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	c.app.Handler.ServeHTTP(rec, req)
	return rec
}

// data แกะ envelope {"data": ...} ที่ httpx.OK และ httpx.Created ห่อไว้
func data(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()

	var body struct {
		Data map[string]any `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("อ่าน response ไม่ได้: %v (body=%s)", err, rec.Body.String())
	}
	return body.Data
}

func errorCode(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()

	var body struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("อ่าน error ไม่ได้: %v (body=%s)", err, rec.Body.String())
	}
	return body.Error.Code
}

func assertStatus(t *testing.T, rec *httptest.ResponseRecorder, want int) {
	t.Helper()
	if rec.Code != want {
		t.Fatalf("status = %d ต้องการ %d (body=%s)", rec.Code, want, rec.Body.String())
	}
}

// insertRoom ใส่ห้องตรงลงฐานข้อมูล เพราะการสร้างห้องผ่าน API เป็นขอบเขตของ AC-19
func insertRoom(t *testing.T, c client, branchID uuid.UUID, stayType string, price float64) uuid.UUID {
	t.Helper()

	var id uuid.UUID
	const q = `
		INSERT INTO rooms (branch_id, room_number, stay_type, price)
		VALUES ($1, $2, $3, $4)
		RETURNING id`
	number := "R" + strings.ReplaceAll(uuid.NewString()[:8], "-", "")
	if err := c.app.Pool.QueryRow(context.Background(), q, branchID, number, stayType, price).Scan(&id); err != nil {
		t.Fatalf("สร้างห้องไม่สำเร็จ: %v", err)
	}
	return id
}

func day(offset int) string {
	return timex.Today().AddDate(0, 0, offset).Format("2006-01-02")
}

// createBooking สร้างการจองผ่าน API จริงแล้วคืน id กับ code
func createBooking(t *testing.T, c client, token string, roomID uuid.UUID, stayType string, from int) (string, string) {
	t.Helper()

	var body string
	if stayType == "monthly" {
		body = fmt.Sprintf(`{"room_id":%q,"stay_type":"monthly","guest_first_name":"ทดสอบ",`+
			`"guest_last_name":"ระบบ","guest_phone":"0800000001","move_in_date":%q}`,
			roomID, day(from))
	} else {
		body = fmt.Sprintf(`{"room_id":%q,"stay_type":"daily","guest_first_name":"ทดสอบ",`+
			`"guest_last_name":"ระบบ","guest_phone":"0800000001","check_in_date":%q,"check_out_date":%q}`,
			roomID, day(from), day(from+3))
	}

	rec := c.do(t, http.MethodPost, "/api/v1/bookings", token, body)
	assertStatus(t, rec, http.StatusCreated)

	d := data(t, rec)
	id, _ := d["id"].(string)
	code, _ := d["code"].(string)
	if id == "" {
		t.Fatalf("ไม่ได้รหัสการจองกลับมา: %s", rec.Body.String())
	}
	return id, code
}

// submitPayment แนบสลิปผ่าน API จริงเพื่อผลักใบจองเข้าสถานะ awaiting_review
func submitPayment(t *testing.T, c client, token, bookingID string) {
	t.Helper()

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	_ = w.WriteField("amount", "3600")
	_ = w.WriteField("transferred_at", timex.Now().Add(-time.Hour).Format("2006-01-02 15:04"))
	_ = w.WriteField("note", "โอนแล้ว")

	part, err := w.CreateFormFile("slip", "slip.jpg")
	if err != nil {
		t.Fatalf("สร้าง multipart ไม่สำเร็จ: %v", err)
	}
	// ไบต์นำของ JPEG จริง เพื่อให้ผ่านการตรวจชนิดไฟล์จากเนื้อไฟล์
	if _, err := part.Write(append([]byte{0xFF, 0xD8, 0xFF, 0xE0}, bytes.Repeat([]byte{0x00}, 256)...)); err != nil {
		t.Fatalf("เขียนไฟล์ทดสอบไม่สำเร็จ: %v", err)
	}
	_ = w.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/bookings/"+bookingID+"/payment", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)

	rec := httptest.NewRecorder()
	c.app.Handler.ServeHTTP(rec, req)
	assertStatus(t, rec, http.StatusOK)
}

// awaitingReview เตรียมใบจองที่พร้อมให้แอดมินตรวจ
func awaitingReview(t *testing.T, c client, token string, roomID uuid.UUID, stayType string, from int) (string, string) {
	t.Helper()
	id, code := createBooking(t, c, token, roomID, stayType, from)
	submitPayment(t, c, token, id)
	return id, code
}

func statusOf(t *testing.T, c client, bookingID string) string {
	t.Helper()
	var s string
	if err := c.app.Pool.QueryRow(context.Background(),
		`SELECT status::text FROM bookings WHERE id = $1`, bookingID).Scan(&s); err != nil {
		t.Fatalf("อ่านสถานะการจองไม่ได้: %v", err)
	}
	return s
}

func roomStatusOf(t *testing.T, c client, roomID uuid.UUID) string {
	t.Helper()
	var s string
	if err := c.app.Pool.QueryRow(context.Background(),
		`SELECT status::text FROM rooms WHERE id = $1`, roomID).Scan(&s); err != nil {
		t.Fatalf("อ่านสถานะห้องไม่ได้: %v", err)
	}
	return s
}

// ---------------------------------------------------------------- AC-10

func TestBookingStateMachine(t *testing.T) {
	t.Parallel()

	setup := func(t *testing.T) (client, string, string, string) {
		t.Helper()
		c := client{app: testsupport.NewApp(t)}
		member := testsupport.AccessToken(t, c.app.Pool, c.app.Fixture.MemberID)
		adminA := testsupport.AccessToken(t, c.app.Pool, c.app.Fixture.AdminAID)
		adminB := testsupport.AccessToken(t, c.app.Pool, c.app.Fixture.AdminBID)
		return c, member, adminA, adminB
	}

	t.Run("approve", func(t *testing.T) {
		t.Parallel()
		c, member, adminA, _ := setup(t)

		t.Run("สำเร็จ", func(t *testing.T) {
			room := insertRoom(t, c, c.app.Fixture.BranchAID, "daily", 1200)
			id, _ := awaitingReview(t, c, member, room, "daily", 1)

			rec := c.do(t, http.MethodPost, "/api/v1/admin/bookings/"+id+"/approve", adminA, "")
			assertStatus(t, rec, http.StatusOK)

			if got := statusOf(t, c, id); got != "approved" {
				t.Errorf("สถานะ = %q ต้องการ approved", got)
			}
		})

		t.Run("payment ล่าสุดเป็น approved", func(t *testing.T) {
			room := insertRoom(t, c, c.app.Fixture.BranchAID, "daily", 1200)
			id, _ := awaitingReview(t, c, member, room, "daily", 10)

			c.do(t, http.MethodPost, "/api/v1/admin/bookings/"+id+"/approve", adminA, "")

			var s string
			if err := c.app.Pool.QueryRow(context.Background(),
				`SELECT status::text FROM payments WHERE booking_id = $1 ORDER BY created_at DESC LIMIT 1`,
				id).Scan(&s); err != nil {
				t.Fatalf("อ่านสถานะการชำระเงินไม่ได้: %v", err)
			}
			if s != "approved" {
				t.Errorf("payment.status = %q ต้องการ approved", s)
			}
		})

		t.Run("ใบรายเดือนทำให้ห้อง occupied", func(t *testing.T) {
			room := insertRoom(t, c, c.app.Fixture.BranchAID, "monthly", 5000)
			id, _ := awaitingReview(t, c, member, room, "monthly", 1)

			c.do(t, http.MethodPost, "/api/v1/admin/bookings/"+id+"/approve", adminA, "")

			if got := roomStatusOf(t, c, room); got != "occupied" {
				t.Errorf("สถานะห้อง = %q ต้องการ occupied", got)
			}
		})
	})

	t.Run("reject", func(t *testing.T) {
		t.Parallel()
		c, member, adminA, _ := setup(t)

		t.Run("กลับไปเป็น pending_payment", func(t *testing.T) {
			room := insertRoom(t, c, c.app.Fixture.BranchAID, "daily", 1200)
			id, _ := awaitingReview(t, c, member, room, "daily", 1)

			rec := c.do(t, http.MethodPost, "/api/v1/admin/bookings/"+id+"/reject", adminA,
				`{"reason":"สลิปไม่ชัด"}`)
			assertStatus(t, rec, http.StatusOK)

			if got := statusOf(t, c, id); got != "pending_payment" {
				t.Errorf("สถานะหลังถูกปฏิเสธ = %q ต้องการ pending_payment", got)
			}
		})

		t.Run("ห้องยังถูกล็อกอยู่", func(t *testing.T) {
			room := insertRoom(t, c, c.app.Fixture.BranchAID, "daily", 1200)
			id, _ := awaitingReview(t, c, member, room, "daily", 20)

			c.do(t, http.MethodPost, "/api/v1/admin/bookings/"+id+"/reject", adminA, `{"reason":"ยอดไม่ตรง"}`)

			// จองห้องเดิมช่วงเดิมซ้ำ ต้องถูกปฏิเสธเพราะใบเดิมยังถือห้องอยู่
			body := fmt.Sprintf(`{"room_id":%q,"stay_type":"daily","guest_first_name":"ทดสอบ",`+
				`"guest_last_name":"ระบบ","guest_phone":"0800000002","check_in_date":%q,"check_out_date":%q}`,
				room, day(20), day(23))
			rec := c.do(t, http.MethodPost, "/api/v1/bookings", member, body)

			assertStatus(t, rec, http.StatusConflict)
			if code := errorCode(t, rec); code != "room_unavailable" {
				t.Errorf("code = %q ต้องการ room_unavailable", code)
			}
		})

		t.Run("reason อยู่ที่ payment ไม่ใช่ใบจอง", func(t *testing.T) {
			room := insertRoom(t, c, c.app.Fixture.BranchAID, "daily", 1200)
			id, _ := awaitingReview(t, c, member, room, "daily", 30)

			const reason = "ยอดโอนไม่ตรงกับที่ต้องชำระ"
			c.do(t, http.MethodPost, "/api/v1/admin/bookings/"+id+"/reject", adminA,
				fmt.Sprintf(`{"reason":%q}`, reason))

			var onPayment string
			if err := c.app.Pool.QueryRow(context.Background(),
				`SELECT reject_reason FROM payments WHERE booking_id = $1 ORDER BY created_at DESC LIMIT 1`,
				id).Scan(&onPayment); err != nil {
				t.Fatalf("อ่านเหตุผลจาก payments ไม่ได้: %v", err)
			}
			if onPayment != reason {
				t.Errorf("payments.reject_reason = %q ต้องการ %q", onPayment, reason)
			}

			// ใบจองต้องไม่เก็บเหตุผลไว้เอง เพราะสถานะกลับไปรอชำระเงินแล้ว
			var exists bool
			if err := c.app.Pool.QueryRow(context.Background(),
				`SELECT EXISTS (
					SELECT 1 FROM information_schema.columns
					WHERE table_name = 'bookings' AND column_name = 'reject_reason'
				)`).Scan(&exists); err != nil {
				t.Fatalf("ตรวจคอลัมน์ไม่ได้: %v", err)
			}
			if exists {
				t.Error("ตาราง bookings ต้องไม่มีคอลัมน์ reject_reason แล้ว เหตุผลอยู่ที่แถว payment")
			}
		})

		t.Run("ไม่ระบุ reason", func(t *testing.T) {
			room := insertRoom(t, c, c.app.Fixture.BranchAID, "daily", 1200)
			id, _ := awaitingReview(t, c, member, room, "daily", 40)

			rec := c.do(t, http.MethodPost, "/api/v1/admin/bookings/"+id+"/reject", adminA, `{"reason":"  "}`)
			assertStatus(t, rec, http.StatusUnprocessableEntity)
			if code := errorCode(t, rec); code != "validation_failed" {
				t.Errorf("code = %q ต้องการ validation_failed", code)
			}
		})
	})

	t.Run("invalid_state", func(t *testing.T) {
		t.Parallel()
		c, member, adminA, _ := setup(t)

		cases := []struct {
			name   string
			action string
			prep   func(t *testing.T, id string)
		}{
			{"อนุมัติใบ pending_payment", "approve", nil},
			{"ปฏิเสธใบ pending_payment", "reject", nil},
			{"อนุมัติใบที่อนุมัติไปแล้ว", "approve", func(t *testing.T, id string) {
				submitPayment(t, c, member, id)
				c.do(t, http.MethodPost, "/api/v1/admin/bookings/"+id+"/approve", adminA, "")
			}},
		}

		for i, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				room := insertRoom(t, c, c.app.Fixture.BranchAID, "daily", 1200)
				id, _ := createBooking(t, c, member, room, "daily", 100+i*10)
				if tc.prep != nil {
					tc.prep(t, id)
				}

				body := ""
				if tc.action == "reject" {
					body = `{"reason":"เหตุผล"}`
				}
				rec := c.do(t, http.MethodPost, "/api/v1/admin/bookings/"+id+"/"+tc.action, adminA, body)

				assertStatus(t, rec, http.StatusConflict)
				if code := errorCode(t, rec); code != "invalid_state" {
					t.Errorf("code = %q ต้องการ invalid_state", code)
				}
			})
		}
	})

	t.Run("ข้ามสาขา", func(t *testing.T) {
		t.Parallel()
		c, member, _, adminB := setup(t)

		room := insertRoom(t, c, c.app.Fixture.BranchAID, "daily", 1200)
		id, _ := awaitingReview(t, c, member, room, "daily", 1)

		for _, action := range []string{"approve", "reject"} {
			t.Run(action, func(t *testing.T) {
				body := ""
				if action == "reject" {
					body = `{"reason":"เหตุผล"}`
				}
				rec := c.do(t, http.MethodPost, "/api/v1/admin/bookings/"+id+"/"+action, adminB, body)

				// ต้องเป็น 404 ไม่ใช่ 403 เพราะรหัสการจองไม่ใช่ข้อมูลสาธารณะ (AC-4)
				assertStatus(t, rec, http.StatusNotFound)
				if code := errorCode(t, rec); code != "not_found" {
					t.Errorf("code = %q ต้องการ not_found", code)
				}
			})
		}
	})

	t.Run("แจ้งเตือน", func(t *testing.T) {
		t.Parallel()
		c, member, adminA, _ := setup(t)

		countNotifications := func(t *testing.T) int {
			t.Helper()
			var n int
			if err := c.app.Pool.QueryRow(context.Background(),
				`SELECT count(*) FROM notifications WHERE user_id = $1`,
				c.app.Fixture.MemberID).Scan(&n); err != nil {
				t.Fatalf("นับแจ้งเตือนไม่ได้: %v", err)
			}
			return n
		}

		latestNotification := func(t *testing.T) string {
			t.Helper()
			var title, body string
			if err := c.app.Pool.QueryRow(context.Background(),
				`SELECT title, body FROM notifications WHERE user_id = $1 ORDER BY created_at DESC LIMIT 1`,
				c.app.Fixture.MemberID).Scan(&title, &body); err != nil {
				t.Fatalf("อ่านแจ้งเตือนไม่ได้: %v", err)
			}
			return title + " " + body
		}

		t.Run("อนุมัติ", func(t *testing.T) {
			room := insertRoom(t, c, c.app.Fixture.BranchAID, "daily", 1200)
			id, code := awaitingReview(t, c, member, room, "daily", 1)

			before := countNotifications(t)
			c.do(t, http.MethodPost, "/api/v1/admin/bookings/"+id+"/approve", adminA, "")

			if got := countNotifications(t) - before; got != 1 {
				t.Fatalf("ได้แจ้งเตือน %d รายการ ต้องการ 1", got)
			}
			if msg := latestNotification(t); !strings.Contains(msg, code) {
				t.Errorf("แจ้งเตือนต้องอ้างรหัสการจอง %q แต่ได้ %q", code, msg)
			}
		})

		t.Run("ปฏิเสธ", func(t *testing.T) {
			room := insertRoom(t, c, c.app.Fixture.BranchAID, "daily", 1200)
			id, code := awaitingReview(t, c, member, room, "daily", 50)

			const reason = "สลิปอ่านไม่ออก"
			before := countNotifications(t)
			c.do(t, http.MethodPost, "/api/v1/admin/bookings/"+id+"/reject", adminA,
				fmt.Sprintf(`{"reason":%q}`, reason))

			if got := countNotifications(t) - before; got != 1 {
				t.Fatalf("ได้แจ้งเตือน %d รายการ ต้องการ 1", got)
			}
			msg := latestNotification(t)
			if !strings.Contains(msg, code) {
				t.Errorf("แจ้งเตือนต้องอ้างรหัสการจอง %q แต่ได้ %q", code, msg)
			}
			if !strings.Contains(msg, reason) {
				t.Errorf("แจ้งเตือนกรณีปฏิเสธต้องมีเหตุผล %q แต่ได้ %q", reason, msg)
			}
		})
	})

	t.Run("approve พร้อมกัน", func(t *testing.T) {
		t.Parallel()
		c, member, adminA, _ := setup(t)

		room := insertRoom(t, c, c.app.Fixture.BranchAID, "daily", 1200)
		id, _ := awaitingReview(t, c, member, room, "daily", 1)

		var (
			wg      sync.WaitGroup
			mu      sync.Mutex
			results []int
		)
		for range 2 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				rec := c.do(t, http.MethodPost, "/api/v1/admin/bookings/"+id+"/approve", adminA, "")
				mu.Lock()
				results = append(results, rec.Code)
				mu.Unlock()
			}()
		}
		wg.Wait()

		var ok, conflict int
		for _, code := range results {
			switch code {
			case http.StatusOK:
				ok++
			case http.StatusConflict:
				conflict++
			}
		}
		if ok != 1 || conflict != 1 {
			t.Errorf("ยิงพร้อมกัน 2 request ได้ผล %v — ต้องการสำเร็จ 1 และ 409 หนึ่งครั้ง", results)
		}
	})

	t.Run("enum มีแค่ 4 สถานะ", func(t *testing.T) {
		t.Parallel()
		c := client{app: testsupport.NewApp(t)}

		// ส่งค่าเป็นพารามิเตอร์ ไม่ใช่ต่อสตริงเข้า SQL
		castable := func(t *testing.T, value string) bool {
			t.Helper()
			var out string
			err := c.app.Pool.QueryRow(context.Background(),
				`SELECT $1::booking_status::text`, value).Scan(&out)
			return err == nil
		}

		for _, removed := range []string{"rejected", "completed"} {
			t.Run(removed, func(t *testing.T) {
				if castable(t, removed) {
					t.Errorf("ฐานข้อมูลยังรับสถานะ %q อยู่ ทั้งที่ SPEC เหลือ 4 สถานะ", removed)
				}
			})
		}

		for _, kept := range []string{"pending_payment", "awaiting_review", "approved", "cancelled"} {
			t.Run(kept, func(t *testing.T) {
				if !castable(t, kept) {
					t.Errorf("ฐานข้อมูลต้องรับสถานะ %q ได้", kept)
				}
			})
		}
	})
}
