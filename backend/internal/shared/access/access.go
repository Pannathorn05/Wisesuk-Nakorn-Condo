// Package access รวมกฎการจำกัดสิทธิ์รายสาขาที่ทุก module ใช้ร่วมกัน
//
// แยกออกมาเป็น package กลางเพราะกฎนี้ต้องเหมือนกันทุกที่ — ถ้าปล่อยให้แต่ละ module
// เขียนเอง สักวันจะมีที่ใดที่หนึ่งลืมเช็ค แล้วกลายเป็นช่องให้แอดมินข้ามสาขาได้
package access

import (
	"errors"

	"backend/internal/database"
	"backend/internal/httpx"
	"backend/internal/middleware"
	"backend/internal/shared/types"
)

// Branch คืนสาขาที่ผู้เรียกมีสิทธิ์เข้าถึง
//   - super admin: ไม่ระบุ = ทุกสาขา (nil) · ระบุ = เฉพาะสาขานั้น
//   - admin: ถูกบังคับเป็นสาขาของตัวเองเสมอ และถูกปฏิเสธถ้าขอสาขาอื่น
//
// ค่า nil ที่คืนออกไปแปลว่า "ไม่จำกัดสาขา" เท่านั้น ห้ามใช้แทนความหมาย
// "ไม่มีสาขาให้เห็น" เด็ดขาด — กรณีนั้นคืน error ไปแล้วตั้งแต่ในนี้
func Branch(identity middleware.Identity, requested *types.BranchID) (*types.BranchID, error) {
	if requested != nil {
		if !identity.HasBranch(*requested) {
			return nil, httpx.ErrForbidden
		}
		return requested, nil
	}
	if identity.IsSuperAdmin() {
		return nil, nil
	}
	if identity.BranchID == nil {
		return nil, httpx.ErrForbidden
	}
	return identity.BranchID, nil
}

// RequireBranch บังคับว่าต้องระบุสาขาชัดเจน ใช้กับ endpoint ที่แก้ไขข้อมูลของสาขา
//
// หัวหน้าผู้ดูแลที่ไม่ได้ส่ง branch_id มาจะถูกปฏิเสธ เพราะคำสั่งแก้ไขต้องรู้ว่า
// หมายถึงสาขาไหน ไม่ใช่ไปลงทุกสาขาพร้อมกัน
func RequireBranch(identity middleware.Identity, requested *types.BranchID) (types.BranchID, error) {
	scoped, err := Branch(identity, requested)
	if err != nil {
		return 0, err
	}
	if scoped == nil {
		return 0, httpx.BadRequest("กรุณาระบุสาขา (branch_id)")
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
