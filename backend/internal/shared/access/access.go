// Package access รวมกฎการจำกัดสิทธิ์รายสาขาที่ทุก module ใช้ร่วมกัน
//
// แยกออกมาเป็น package กลางเพราะกฎนี้ต้องเหมือนกันทุกที่ — ถ้าปล่อยให้แต่ละ module
// เขียนเอง สักวันจะมีที่ใดที่หนึ่งลืมเช็ค แล้วกลายเป็นช่องให้แอดมินข้ามสาขาได้
package access

import (
	"errors"

	"github.com/google/uuid"

	"backend/internal/database"
	"backend/internal/httpx"
	"backend/internal/middleware"
)

// Branch คืนสาขาที่ผู้เรียกมีสิทธิ์เข้าถึง
//   - super admin: เข้าถึงได้ทุกสาขา (nil = ไม่จำกัด) หรือจะระบุสาขาที่ขอมาก็ได้
//   - admin: ถูกบังคับเป็นสาขาของตัวเองเสมอ และถูกปฏิเสธถ้าขอสาขาอื่น
func Branch(identity middleware.Identity, requested *uuid.UUID) (*uuid.UUID, error) {
	if identity.IsSuperAdmin() {
		return requested, nil
	}
	if identity.BranchID == nil {
		return nil, httpx.ErrForbidden
	}
	if requested != nil && *requested != *identity.BranchID {
		return nil, httpx.ErrForbidden
	}
	return identity.BranchID, nil
}

// RequireBranch บังคับว่าต้องระบุสาขาชัดเจน ใช้กับ endpoint ที่แก้ไขข้อมูลของสาขา
func RequireBranch(identity middleware.Identity, requested *uuid.UUID) (uuid.UUID, error) {
	scoped, err := Branch(identity, requested)
	if err != nil {
		return uuid.Nil, err
	}
	if scoped == nil {
		return uuid.Nil, httpx.BadRequest("กรุณาระบุสาขา (branch_id)")
	}
	return *scoped, nil
}

// MapErr แปลง error จากชั้น repository ให้เป็น error ที่มี HTTP status ถูกต้อง
func MapErr(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, database.ErrNotFound):
		return httpx.ErrNotFound
	case database.IsUniqueViolation(err):
		return httpx.ErrConflict.Wrap(err)
	default:
		var apiErr *httpx.APIError
		if errors.As(err, &apiErr) {
			return err
		}
		return httpx.ErrInternal.Wrap(err)
	}
}
