// เทสว่าการลบ / ระงับ / ย้ายสาขาผู้ดูแลมีผลทันที แม้ผู้ถูกกระทำยังถือ access token ใบเดิมอยู่
//
// checklist เทียบกับ AC-3, AC-4, AC-18 ใน docs/SPEC.md
//
//	ลบผู้ดูแลแล้ว token เดิม -> 401 ทันที                      -> TestDeletedAdminTokenRejectedImmediately
//	ระงับผู้ดูแลแล้ว token เดิม -> 401 ทันที                    -> TestSuspendedAdminTokenRejectedImmediately
//	ย้ายสาขาแล้ว token เดิมเห็นสาขาใหม่ทันที ไม่เห็นสาขาเดิม    -> TestMovedAdminSeesNewBranchImmediately
//	บัญชีปกติยังใช้ token ได้ตามเดิม                            -> TestActiveAccountTokenStillWorks
package account_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"backend/internal/shared/types"
	"backend/internal/testsupport"
)

func call(t *testing.T, app testsupport.App, method, path, token, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	app.Handler.ServeHTTP(rec, req)
	return rec
}

func expectStatus(t *testing.T, rec *httptest.ResponseRecorder, want int) {
	t.Helper()
	if rec.Code != want {
		t.Fatalf("status = %d ต้องการ %d (body=%s)", rec.Code, want, rec.Body.String())
	}
}

func TestActiveAccountTokenStillWorks(t *testing.T) {
	t.Parallel()
	app := testsupport.NewApp(t)
	adminA := testsupport.AccessToken(t, app.Pool, app.Fixture.AdminAID)

	expectStatus(t, call(t, app, http.MethodGet, "/api/v1/me", adminA, ""), http.StatusOK)
	expectStatus(t, call(t, app, http.MethodGet, "/api/v1/admin/bookings", adminA, ""), http.StatusOK)
}

func TestDeletedAdminTokenRejectedImmediately(t *testing.T) {
	t.Parallel()
	app := testsupport.NewApp(t)
	super := testsupport.AccessToken(t, app.Pool, app.Fixture.SuperAdminID)
	adminA := testsupport.AccessToken(t, app.Pool, app.Fixture.AdminAID)

	expectStatus(t, call(t, app, http.MethodDelete,
		"/api/v1/superadmin/staff/"+app.Fixture.AdminAID.String(), super, ""), http.StatusNoContent)

	// token ยังไม่หมดอายุ แต่บัญชีถูกลบแล้ว ต้องใช้ไม่ได้ทุก endpoint
	expectStatus(t, call(t, app, http.MethodGet, "/api/v1/admin/bookings", adminA, ""), http.StatusUnauthorized)
	expectStatus(t, call(t, app, http.MethodGet, "/api/v1/me", adminA, ""), http.StatusUnauthorized)
}

func TestSuspendedAdminTokenRejectedImmediately(t *testing.T) {
	t.Parallel()
	app := testsupport.NewApp(t)
	super := testsupport.AccessToken(t, app.Pool, app.Fixture.SuperAdminID)
	adminA := testsupport.AccessToken(t, app.Pool, app.Fixture.AdminAID)

	body := `{"first_name":"ผู้ดูแล","last_name":"สาขาเอ","phone":"","branch_id":"` +
		app.Fixture.BranchAID.String() + `","is_active":false}`
	expectStatus(t, call(t, app, http.MethodPut,
		"/api/v1/superadmin/staff/"+app.Fixture.AdminAID.String(), super, body), http.StatusOK)

	expectStatus(t, call(t, app, http.MethodGet, "/api/v1/admin/bookings", adminA, ""), http.StatusUnauthorized)
}

func TestMovedAdminSeesNewBranchImmediately(t *testing.T) {
	t.Parallel()
	app := testsupport.NewApp(t)
	super := testsupport.AccessToken(t, app.Pool, app.Fixture.SuperAdminID)
	adminA := testsupport.AccessToken(t, app.Pool, app.Fixture.AdminAID)

	// สาขาใหม่ที่ยังไม่มีผู้ดูแล (1 สาขามีผู้ดูแลได้คนเดียว)
	var branchC types.BranchID
	if err := app.Pool.QueryRow(context.Background(),
		`INSERT INTO branches (slug, name) VALUES ('branch-c', 'สาขาทดสอบ C') RETURNING id`).Scan(&branchC); err != nil {
		t.Fatalf("สร้างสาขาไม่สำเร็จ: %v", err)
	}

	body := `{"first_name":"ผู้ดูแล","last_name":"สาขาเอ","phone":"","branch_id":"` + branchC.String() + `"}`
	expectStatus(t, call(t, app, http.MethodPut,
		"/api/v1/superadmin/staff/"+app.Fixture.AdminAID.String(), super, body), http.StatusOK)

	// สาขาเดิมต้องเข้าไม่ได้แล้ว แม้ token จะออกตอนยังอยู่สาขา A
	expectStatus(t, call(t, app, http.MethodGet,
		"/api/v1/admin/bookings?branch_id="+app.Fixture.BranchAID.String(), adminA, ""), http.StatusForbidden)

	// ไม่ระบุสาขา = สาขาปัจจุบันจาก DB คือ C
	rec := call(t, app, http.MethodGet, "/api/v1/admin/dashboard", adminA, "")
	expectStatus(t, rec, http.StatusOK)
	var dash struct {
		Data struct {
			Branches []struct {
				BranchID string `json:"branch_id"`
			} `json:"branches"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &dash); err != nil {
		t.Fatalf("อ่าน dashboard ไม่ได้: %v", err)
	}
	if len(dash.Data.Branches) != 1 || dash.Data.Branches[0].BranchID != branchC.String() {
		t.Errorf("dashboard เห็นสาขา %+v ต้องการเฉพาะ %s", dash.Data.Branches, branchC)
	}
}
