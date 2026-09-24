// เทส "ปิดแล้วปิดเลย" ฝั่งสาขา: สาขาที่ปิดใช้งานต้องเปิดดูผ่าน GET /branches/{id} ไม่ได้
// ทั้งเรียกด้วย id และ slug แต่แอดมินของสาขานั้นยังต้องแก้สาขาตัวเองได้ตามปกติ
package branch_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/internal/shared/types"
	"backend/internal/testsupport"
)

type env struct {
	app testsupport.App
}

func newEnv(t *testing.T) env {
	t.Helper()
	return env{app: testsupport.NewApp(t)}
}

func (e env) do(t *testing.T, method, path, body, token string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(method, path, bytes.NewReader([]byte(body)))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	rec := httptest.NewRecorder()
	e.app.Handler.ServeHTTP(rec, req)
	return rec
}

func deactivateBranch(t *testing.T, e env, id types.BranchID) {
	t.Helper()
	if _, err := e.app.Pool.Exec(context.Background(),
		`UPDATE branches SET is_active = FALSE WHERE id = $1`, id); err != nil {
		t.Fatalf("ปิดสาขาไม่สำเร็จ: %v", err)
	}
}

func assertNotFound(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d ต้องเป็น 404 (body=%s)", rec.Code, rec.Body.String())
	}
	var body struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("อ่าน response ไม่ได้: %v (body=%s)", err, rec.Body.String())
	}
	if body.Error.Code != "not_found" {
		t.Errorf("error.code = %q ต้องเป็น not_found", body.Error.Code)
	}
}

func TestGetOpenBranchByIDAndSlug(t *testing.T) {
	e := newEnv(t)

	for _, key := range []string{e.app.Fixture.BranchAID.String(), "branch-a"} {
		rec := e.do(t, http.MethodGet, "/api/v1/branches/"+key, "", "")
		if rec.Code != http.StatusOK {
			t.Errorf("GET /branches/%s status = %d ต้องเป็น 200 (body=%s)", key, rec.Code, rec.Body.String())
		}
	}
}

func TestGetClosedBranchByIDReturns404(t *testing.T) {
	e := newEnv(t)
	deactivateBranch(t, e, e.app.Fixture.BranchAID)

	assertNotFound(t, e.do(t, http.MethodGet, "/api/v1/branches/"+e.app.Fixture.BranchAID.String(), "", ""))
}

// slug เป็นอีกทางเข้าหนึ่งของ endpoint เดียวกัน ต้องปิดตามด้วย ไม่ใช่รั่วทางนี้
func TestGetClosedBranchBySlugReturns404(t *testing.T) {
	e := newEnv(t)
	deactivateBranch(t, e, e.app.Fixture.BranchAID)

	assertNotFound(t, e.do(t, http.MethodGet, "/api/v1/branches/branch-a", "", ""))
}

// ปิดสาขาหนึ่งต้องไม่กระทบสาขาอื่น
func TestClosingBranchDoesNotHideOthers(t *testing.T) {
	e := newEnv(t)
	deactivateBranch(t, e, e.app.Fixture.BranchAID)

	rec := e.do(t, http.MethodGet, "/api/v1/branches/"+e.app.Fixture.BranchBID.String(), "", "")
	if rec.Code != http.StatusOK {
		t.Errorf("สาขา B ยังเปิดอยู่ status = %d ต้องเป็น 200 (body=%s)", rec.Code, rec.Body.String())
	}
}

// PUT /admin/branch คืนรายละเอียดสาขาหลังแก้ ต้องไม่โดนตัวกรองฝั่งสาธารณะจนกลายเป็น 404
func TestAdminStillUpdatesClosedBranch(t *testing.T) {
	e := newEnv(t)
	deactivateBranch(t, e, e.app.Fixture.BranchAID)
	token := testsupport.AccessToken(t, e.app.Pool, e.app.Fixture.AdminAID)

	body := `{"name":"สาขาทดสอบ A (ปิดปรับปรุง)","phones":[],"building_count":1,"floor_count":5}`
	rec := e.do(t, http.MethodPut, "/api/v1/admin/branch", body, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d ต้องเป็น 200 (body=%s)", rec.Code, rec.Body.String())
	}
}
