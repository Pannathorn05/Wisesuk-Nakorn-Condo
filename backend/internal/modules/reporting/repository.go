package reporting

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"backend/internal/database"
	"backend/internal/shared/audit"
	"backend/internal/shared/types"
)

// Repository อ่าน activity log สำหรับหน้า "ประวัติการใช้งาน"
//
// ฝั่งเขียน log อยู่ที่ shared/audit เพราะทุก module ต้องเขียนได้
// ถ้าเอา recorder มาไว้ที่นี่ ทุก module จะต้อง import reporting แล้วเกิด import cycle
type Repository struct{ db *database.TxManager }

func NewRepository(db *database.TxManager) *Repository { return &Repository{db: db} }

type ListParams struct {
	ActorID   *uuid.UUID
	ActorRole *types.Role
	BranchID  *uuid.UUID
	Action    string
	Limit     int
	Offset    int
}

func (r *Repository) ListLogs(ctx context.Context, p ListParams) ([]audit.Log, int, error) {
	where := []string{"TRUE"}
	b := &database.Binder{}

	if p.ActorID != nil {
		where = append(where, "al.actor_id = "+b.Bind(*p.ActorID))
	}
	if p.ActorRole != nil {
		where = append(where, "al.actor_role = "+b.Bind(*p.ActorRole))
	}
	if p.BranchID != nil {
		where = append(where, "al.branch_id = "+b.Bind(*p.BranchID))
	}
	if a := strings.TrimSpace(p.Action); a != "" {
		where = append(where, "al.action = "+b.Bind(a))
	}

	whereSQL := "WHERE " + strings.Join(where, " AND ")
	const joins = ` FROM activity_logs al LEFT JOIN branches b ON b.id = al.branch_id `

	exec := r.db.Executor(ctx)

	var total int
	if err := exec.QueryRow(ctx, `SELECT COUNT(*)`+joins+whereSQL, b.Args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	limit := p.Limit
	if limit <= 0 {
		limit = 50
	}
	q := `SELECT al.id, al.actor_id, al.actor_role, al.actor_name, al.branch_id,
	             COALESCE(b.name, ''), al.action, al.entity_type, al.entity_id,
	             al.detail, al.ip_address, al.created_at` +
		joins + whereSQL +
		` ORDER BY al.created_at DESC LIMIT ` + b.Bind(limit) + ` OFFSET ` + b.Bind(p.Offset)

	rows, err := exec.Query(ctx, q, b.Args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := []audit.Log{}
	for rows.Next() {
		var l audit.Log
		if err := rows.Scan(&l.ID, &l.ActorID, &l.ActorRole, &l.ActorName, &l.BranchID,
			&l.BranchName, &l.Action, &l.EntityType, &l.EntityID,
			&l.Detail, &l.IPAddress, &l.CreatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, l)
	}
	return out, total, rows.Err()
}
