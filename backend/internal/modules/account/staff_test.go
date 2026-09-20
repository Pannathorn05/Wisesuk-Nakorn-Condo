// เทสหน้า "จัดการผู้ดูแลระบบ" ของหัวหน้าผู้ดูแล
//
// กฎความเป็นเจ้าของที่ต้องยืนยันมีสองข้อ: 1 สาขามีผู้ดูแลคนเดียว และทั้งระบบมี
// หัวหน้าผู้ดูแลคนเดียว ทั้งคู่บังคับไว้สองชั้น — ชั้น service ให้ข้อความอ่านรู้เรื่อง
// และชั้นฐานข้อมูลเป็น unique index กันไว้อีกที เทสชุดนี้จึงยิงทั้งผ่าน service
// และยิง SQL ตรง ๆ เพื่อพิสูจน์ว่าชั้น DB กันได้จริง ไม่ใช่แค่โค้ดที่เผลอลืมเช็คเมื่อไหร่ก็หลุด
package account_test

import (
	"context"
	"strings"
	"testing"

	"backend/internal/httpx"
	"backend/internal/middleware"
	"backend/internal/modules/account"
	"backend/internal/shared/types"
)

// identityFor สร้าง Identity ของผู้ใช้ที่เพิ่งถูกสร้าง เหมือนที่ middleware ทำจาก token
func identityFor(u *account.User) middleware.Identity {
	return middleware.Identity{
		UserID: u.ID, Role: u.Role, Name: u.FullName(), BranchID: u.BranchID,
	}
}

func superIdentity(e env) middleware.Identity {
	return middleware.Identity{
		UserID: e.fixture.SuperAdminID,
		Role:   types.RoleSuperAdmin,
		Name:   "หัวหน้าผู้ดูแลทดสอบ",
	}
}

// newBranch สร้างสาขาเปล่าที่ยังไม่มีผู้ดูแล — fixture ตั้งต้นใช้สาขาไปครบทั้งสองแล้ว
func newBranch(t *testing.T, e env, slug, name string) types.BranchID {
	t.Helper()
	var id types.BranchID
	err := e.pool.QueryRow(context.Background(),
		`INSERT INTO branches (slug, name) VALUES ($1, $2) RETURNING id`, slug, name).Scan(&id)
	if err != nil {
		t.Fatalf("สร้างสาขา %s ไม่สำเร็จ: %v", slug, err)
	}
	return id
}

func createStaff(t *testing.T, e env, email string, branch types.BranchID) *account.User {
	t.Helper()
	user, err := e.svc.CreateAdmin(context.Background(), superIdentity(e), account.CreateAdminInput{
		Email:     email,
		FirstName: "ทดสอบ",
		LastName:  "ผู้ดูแล",
		Phone:     "0891112222",
		BranchID:  branch,
	}, "127.0.0.1")
	if err != nil {
		t.Fatalf("CreateAdmin คืน error: %v", err)
	}
	return user
}

func TestCreateStaffBindsToOneBranch(t *testing.T) {
	e := newEnv(t)
	branch := newBranch(t, e, "branch-c", "สาขาทดสอบ C")

	user := createStaff(t, e, "one@wisetsuk.test", branch)

	if user.Role != types.RoleAdmin {
		t.Errorf("role = %q ต้องเป็น admin เสมอ", user.Role)
	}
	if user.BranchID == nil || *user.BranchID != branch {
		t.Fatalf("สาขา = %v ต้องเป็น %v", user.BranchID, branch)
	}
	// ชื่อสาขาต้องติดมาด้วย ไม่งั้นหน้าตารางต้องยิงถามรายการสาขาซ้ำเพื่อแปลง id เป็นชื่อ
	if user.BranchName == "" {
		t.Error("ต้องมี branch_name ติดมาด้วยสำหรับตารางหน้าจัดการผู้ดูแล")
	}
}

// ฟอร์มไม่มีช่องรหัสผ่าน ระบบต้องตั้งรหัสตั้งต้นให้แล้วบังคับให้เปลี่ยนเอง
func TestCreateStaffUsesDefaultPasswordAndForcesChange(t *testing.T) {
	e := newEnv(t)
	branch := newBranch(t, e, "branch-pw", "สาขาทดสอบรหัสผ่าน")

	user := createStaff(t, e, "default-pw@wisetsuk.test", branch)
	if !user.MustChangePassword {
		t.Error("บัญชีที่ได้รหัสตั้งต้นต้องถูกบังคับให้เปลี่ยนรหัสผ่าน")
	}

	// ต้องเข้าสู่ระบบด้วยรหัสตั้งต้นได้จริง ไม่งั้นผู้ดูแลคนใหม่เข้าระบบไม่ได้เลย
	ctx := context.Background()
	if _, err := e.svc.Login(ctx, account.LoginInput{
		Email:    "default-pw@wisetsuk.test",
		Password: testStaffDefaultPassword,
	}, "127.0.0.1"); err != nil {
		t.Fatalf("เข้าสู่ระบบด้วยรหัสตั้งต้นไม่สำเร็จ: %v", err)
	}

	// เปลี่ยนรหัสเองแล้วธงต้องถูกปลด
	if err := e.svc.ChangePassword(ctx, identityFor(user), account.ChangePasswordInput{
		CurrentPassword: testStaffDefaultPassword,
		NewPassword:     "NewPassword!2026",
		ConfirmPassword: "NewPassword!2026",
	}, "127.0.0.1"); err != nil {
		t.Fatalf("เปลี่ยนรหัสผ่านไม่สำเร็จ: %v", err)
	}

	after, err := e.repo.GetByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("อ่านผู้ใช้กลับมาไม่สำเร็จ: %v", err)
	}
	if after.MustChangePassword {
		t.Error("เปลี่ยนรหัสผ่านเองแล้ว ธง must_change_password ต้องถูกปลด")
	}
}

func TestCreateStaffRejectsMissingBranch(t *testing.T) {
	e := newEnv(t)

	_, err := e.svc.CreateAdmin(context.Background(), superIdentity(e), account.CreateAdminInput{
		Email:     "nobranch@wisetsuk.test",
		FirstName: "ทดสอบ",
		LastName:  "ไม่มีสาขา",
		Phone:     "0891112222",
	}, "127.0.0.1")

	assertFieldError(t, err, "branch_id")
}

func TestCreateStaffRejectsUnknownBranch(t *testing.T) {
	e := newEnv(t)

	_, err := e.svc.CreateAdmin(context.Background(), superIdentity(e), account.CreateAdminInput{
		Email:     "ghost@wisetsuk.test",
		FirstName: "ทดสอบ",
		LastName:  "สาขาผี",
		Phone:     "0891112222",
		BranchID:  999999,
	}, "127.0.0.1")

	assertFieldError(t, err, "branch_id")
}

// หัวใจของกฎ: สาขาที่มีผู้ดูแลอยู่แล้ว รับคนที่สองไม่ได้
func TestCreateStaffRejectsBranchThatAlreadyHasAdmin(t *testing.T) {
	e := newEnv(t)

	_, err := e.svc.CreateAdmin(context.Background(), superIdentity(e), account.CreateAdminInput{
		Email:     "second@wisetsuk.test",
		FirstName: "ทดสอบ",
		LastName:  "คนที่สอง",
		Phone:     "0891112222",
		BranchID:  e.fixture.BranchAID, // สาขานี้มี admin.a ดูแลอยู่แล้วตั้งแต่ fixture
	}, "127.0.0.1")

	assertFieldError(t, err, "branch_id")
}

func TestUpdateStaffMovesBranch(t *testing.T) {
	e := newEnv(t)
	branch := newBranch(t, e, "branch-move", "สาขาปลายทาง")
	origin := newBranch(t, e, "branch-origin", "สาขาต้นทาง")

	user := createStaff(t, e, "move@wisetsuk.test", origin)

	updated, err := e.svc.UpdateAdmin(context.Background(), superIdentity(e), user.ID, account.UpdateAdminInput{
		FirstName: "ทดสอบ",
		LastName:  "ย้ายสาขา",
		Phone:     "0891112222",
		BranchID:  &branch,
	}, "127.0.0.1")
	if err != nil {
		t.Fatalf("UpdateAdmin คืน error: %v", err)
	}

	if updated.BranchID == nil || *updated.BranchID != branch {
		t.Errorf("สาขาหลังย้าย = %v ต้องเป็น %v", updated.BranchID, branch)
	}
}

func TestUpdateStaffRejectsMovingIntoOccupiedBranch(t *testing.T) {
	e := newEnv(t)
	origin := newBranch(t, e, "branch-free", "สาขาว่าง")

	user := createStaff(t, e, "collide@wisetsuk.test", origin)

	_, err := e.svc.UpdateAdmin(context.Background(), superIdentity(e), user.ID, account.UpdateAdminInput{
		FirstName: "ทดสอบ",
		LastName:  "ชนสาขา",
		Phone:     "0891112222",
		BranchID:  &e.fixture.BranchAID, // มี admin.a ดูแลอยู่แล้ว
	}, "127.0.0.1")

	assertFieldError(t, err, "branch_id")
}

func TestUpdateStaffRejectsBranchForSuperAdmin(t *testing.T) {
	e := newEnv(t)
	f := e.fixture

	_, err := e.svc.UpdateAdmin(context.Background(), superIdentity(e), f.SuperAdminID, account.UpdateAdminInput{
		FirstName: "หัวหน้า",
		LastName:  "ผู้ดูแล",
		Phone:     "0891112222",
		BranchID:  &f.BranchAID,
	}, "127.0.0.1")

	assertFieldError(t, err, "branch_id")
}

// ---------------------------------------------------------------- ชั้นฐานข้อมูล
//
// สองเทสข้างล่างข้ามชั้น service ไปเลย ยิง SQL ตรง ๆ เพื่อพิสูจน์ว่ากฎถูกบังคับที่ DB จริง
// (AGENTS.md ข้อ 4) ถ้าวันหนึ่งมีโค้ดเส้นใหม่ลืมเช็ค ฐานข้อมูลต้องยังปฏิเสธให้อยู่ดี

func TestDatabaseRejectsSecondAdminOnSameBranch(t *testing.T) {
	e := newEnv(t)

	_, err := e.pool.Exec(context.Background(), `
		INSERT INTO users (email, password_hash, first_name, last_name, phone, role, branch_id)
		VALUES ('dup-admin@wisetsuk.test', '$2a$04$abcdefghijklmnopqrstuv', 'ทดสอบ', 'ซ้ำสาขา', '0800000000', 'admin', $1)`,
		e.fixture.BranchAID)

	assertUniqueViolation(t, err, "uq_admin_per_branch")
}

func TestDatabaseRejectsSecondSuperAdmin(t *testing.T) {
	e := newEnv(t)

	_, err := e.pool.Exec(context.Background(), `
		INSERT INTO users (email, password_hash, first_name, last_name, phone, role)
		VALUES ('super2@wisetsuk.test', '$2a$04$abcdefghijklmnopqrstuv', 'หัวหน้า', 'คนที่สอง', '0800000000', 'superadmin')`)

	assertUniqueViolation(t, err, "uq_single_superadmin")
}

// ---------------------------------------------------------------- ตัวช่วย

func assertFieldError(t *testing.T, err error, field string) {
	t.Helper()
	if err == nil {
		t.Fatal("ต้องได้ error แต่สำเร็จ")
	}
	apiErr, ok := err.(*httpx.APIError)
	if !ok {
		t.Fatalf("ต้องได้ APIError แต่ได้ %T: %v", err, err)
	}
	if apiErr.Code != "validation_failed" {
		t.Fatalf("code = %q ต้องเป็น validation_failed (error: %v)", apiErr.Code, err)
	}
	if _, found := apiErr.Fields[field]; !found {
		t.Errorf("ต้องมีข้อความบอกที่ฟิลด์ %q แต่ได้ %v", field, apiErr.Fields)
	}
}

func assertUniqueViolation(t *testing.T, err error, indexName string) {
	t.Helper()
	if err == nil {
		t.Fatal("ฐานข้อมูลต้องปฏิเสธคำสั่งนี้ แต่ผ่านไปได้")
	}
	if !strings.Contains(err.Error(), indexName) {
		t.Errorf("ต้องชนกับ index %q แต่ได้: %v", indexName, err)
	}
}
