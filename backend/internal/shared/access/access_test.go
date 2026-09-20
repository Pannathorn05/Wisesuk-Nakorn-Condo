package access_test

import (
	"errors"
	"testing"

	"backend/internal/httpx"
	"backend/internal/middleware"
	"backend/internal/shared/access"
	"backend/internal/shared/types"
)

const (
	branchA = types.BranchID(1)
	branchB = types.BranchID(2)
)

func superAdmin() middleware.Identity {
	return middleware.Identity{UserID: 1, Role: types.RoleSuperAdmin, Name: "หัวหน้าผู้ดูแล"}
}

func admin(branch types.BranchID) middleware.Identity {
	return middleware.Identity{UserID: 2, Role: types.RoleAdmin, Name: "ผู้ดูแล", BranchID: &branch}
}

func member() middleware.Identity {
	return middleware.Identity{UserID: 3, Role: types.RoleMember, Name: "สมาชิก"}
}

func ptr(id types.BranchID) *types.BranchID { return &id }

func TestBranchScope(t *testing.T) {
	tests := []struct {
		name      string
		identity  middleware.Identity
		requested *types.BranchID
		wantAll   bool // true = คาดหวัง nil ซึ่งแปลว่าไม่จำกัดสาขา
		want      types.BranchID
		wantErr   error
	}{
		{
			name:     "หัวหน้าผู้ดูแลไม่ระบุสาขา เห็นทุกสาขา",
			identity: superAdmin(),
			wantAll:  true,
		},
		{
			name:      "หัวหน้าผู้ดูแลระบุสาขา เห็นเฉพาะสาขานั้น",
			identity:  superAdmin(),
			requested: ptr(branchB),
			want:      branchB,
		},
		{
			name:     "ผู้ดูแลไม่ระบุ ถูกบังคับเป็นสาขาตัวเอง",
			identity: admin(branchA),
			want:     branchA,
		},
		{
			name:      "ผู้ดูแลเจาะดูสาขาของตัวเองได้",
			identity:  admin(branchA),
			requested: ptr(branchA),
			want:      branchA,
		},
		{
			name:      "ผู้ดูแลขอสาขาอื่น ถูกปฏิเสธ",
			identity:  admin(branchA),
			requested: ptr(branchB),
			wantErr:   httpx.ErrForbidden,
		},
		{
			name:     "สมาชิกไม่มีสาขา จึงไม่เห็นอะไรเลย ไม่ใช่เห็นทุกสาขา",
			identity: member(),
			wantErr:  httpx.ErrForbidden,
		},
		{
			name:      "สมาชิกระบุสาขาก็ยังถูกปฏิเสธ",
			identity:  member(),
			requested: ptr(branchA),
			wantErr:   httpx.ErrForbidden,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := access.Branch(tc.identity, tc.requested)

			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("err = %v ต้องเป็น %v", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("ไม่ควรมี error แต่ได้ %v", err)
			}
			if tc.wantAll {
				if got != nil {
					t.Errorf("ต้องได้ nil (ทุกสาขา) แต่ได้ %v", *got)
				}
				return
			}
			if got == nil {
				t.Fatalf("ต้องได้สาขา %v แต่ได้ nil ซึ่งแปลว่าไม่จำกัดสาขา", tc.want)
			}
			if *got != tc.want {
				t.Errorf("สาขา = %v ต้องเป็น %v", *got, tc.want)
			}
		})
	}
}

func TestRequireBranch(t *testing.T) {
	tests := []struct {
		name      string
		identity  middleware.Identity
		requested *types.BranchID
		want      types.BranchID
		wantCode  string
	}{
		{
			name:     "ผู้ดูแลไม่ต้องระบุ ระบบรู้อยู่แล้วว่าสาขาไหน",
			identity: admin(branchA),
			want:     branchA,
		},
		{
			name:      "หัวหน้าผู้ดูแลระบุสาขาไหนก็ได้",
			identity:  superAdmin(),
			requested: ptr(branchB),
			want:      branchB,
		},
		{
			name:     "หัวหน้าผู้ดูแลไม่ระบุสาขา ต้องไม่ไปแก้ทุกสาขาพร้อมกัน",
			identity: superAdmin(),
			wantCode: "bad_request",
		},
		{
			name:      "ผู้ดูแลระบุสาขาอื่น ถูกปฏิเสธ",
			identity:  admin(branchA),
			requested: ptr(branchB),
			wantCode:  "forbidden",
		},
		{
			name:     "สมาชิกแก้ข้อมูลสาขาไม่ได้",
			identity: member(),
			wantCode: "forbidden",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := access.RequireBranch(tc.identity, tc.requested)

			if tc.wantCode != "" {
				var apiErr *httpx.APIError
				if !errors.As(err, &apiErr) {
					t.Fatalf("ต้องได้ APIError แต่ได้ %v", err)
				}
				if apiErr.Code != tc.wantCode {
					t.Errorf("code = %q ต้องเป็น %q", apiErr.Code, tc.wantCode)
				}
				return
			}
			if err != nil {
				t.Fatalf("ไม่ควรมี error แต่ได้ %v", err)
			}
			if got != tc.want {
				t.Errorf("สาขา = %v ต้องเป็น %v", got, tc.want)
			}
		})
	}
}

func TestIdentityHasBranch(t *testing.T) {
	if !superAdmin().HasBranch(branchB) {
		t.Error("หัวหน้าผู้ดูแลต้องแตะได้ทุกสาขา")
	}
	if !admin(branchA).HasBranch(branchA) {
		t.Error("ผู้ดูแลต้องแตะสาขาของตัวเองได้")
	}
	if admin(branchA).HasBranch(branchB) {
		t.Error("ผู้ดูแลต้องแตะสาขาอื่นไม่ได้")
	}
	if member().HasBranch(branchA) {
		t.Error("สมาชิกต้องไม่มีสิทธิ์กับสาขาใด")
	}
}
