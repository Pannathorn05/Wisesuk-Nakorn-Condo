// Package audit เขียน activity log
//
// อยู่ใน shared เพราะเป็นเรื่องที่ตัดขวางทุก module — ทุก module ต้องบันทึก log ได้
// ถ้าให้ recorder ไปอยู่ใน module reporting จะเกิด import cycle ทันที
// เพราะ reporting ต้อง import room กับ booking เพื่อสร้างแดชบอร์ด
//
// ฝั่ง "อ่าน" log อยู่ที่ modules/reporting/repository.go ซึ่งเป็นเรื่องของการแสดงผล
package audit

import (
	"context"
	"time"

	"github.com/google/uuid"

	"backend/internal/database"
	"backend/internal/middleware"
	"backend/internal/shared/types"
)

// Log คือหนึ่งบรรทัดในหน้า "ประวัติการใช้งาน"
type Log struct {
	ID         uuid.UUID      `json:"id"`
	ActorID    *uuid.UUID     `json:"actor_id,omitempty"`
	ActorRole  types.Role     `json:"actor_role"`
	ActorName  string         `json:"actor_name"`
	BranchID   *uuid.UUID     `json:"branch_id,omitempty"`
	BranchName string         `json:"branch_name,omitempty"`
	Action     string         `json:"action"`
	EntityType string         `json:"entity_type"`
	EntityID   string         `json:"entity_id"`
	Detail     map[string]any `json:"detail"`
	IPAddress  string         `json:"ip_address"`
	CreatedAt  time.Time      `json:"created_at"`
}

// Entry คือรายการที่จะบันทึก
type Entry struct {
	ActorID    *uuid.UUID
	ActorRole  types.Role
	ActorName  string
	BranchID   *uuid.UUID
	Action     string
	EntityType string
	EntityID   string
	Detail     map[string]any
	IPAddress  string
}

type Recorder struct {
	db *database.TxManager
}

func NewRecorder(db *database.TxManager) *Recorder { return &Recorder{db: db} }

// Record บันทึกแบบ best-effort — ถ้าเขียน log ไม่สำเร็จ ไม่ควรทำให้ธุรกรรมหลัก
// ที่สำเร็จไปแล้วล้มตาม จึงกลืน error ไว้ที่นี่
func (r *Recorder) Record(ctx context.Context, e Entry) {
	if e.Detail == nil {
		e.Detail = map[string]any{}
	}
	const q = `
		INSERT INTO activity_logs
			(actor_id, actor_role, actor_name, branch_id, action, entity_type, entity_id, detail, ip_address)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, _ = r.db.Executor(ctx).Exec(ctx, q,
		e.ActorID, e.ActorRole, e.ActorName, e.BranchID,
		e.Action, e.EntityType, e.EntityID, e.Detail, e.IPAddress)
}

// RecordFor เป็นทางลัดที่ใช้บ่อยที่สุด — บันทึกจาก identity ของผู้เรียก API
func (r *Recorder) RecordFor(ctx context.Context, identity middleware.Identity, action, entityType, entityID string, detail map[string]any, ip string) {
	actorID := identity.UserID
	r.Record(ctx, Entry{
		ActorID: &actorID, ActorRole: identity.Role, ActorName: identity.Name,
		BranchID: identity.BranchID, Action: action,
		EntityType: entityType, EntityID: entityID, Detail: detail, IPAddress: ip,
	})
}
