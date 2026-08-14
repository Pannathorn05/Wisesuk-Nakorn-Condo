package database

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrNotFound ถูกคืนเมื่อ query ไม่พบแถวที่ต้องการ ชั้น service จะแปลงเป็น 404
var ErrNotFound = errors.New("database: not found")

// Executor ครอบทั้ง *pgxpool.Pool และ pgx.Tx
// repository ทุกตัวรับ interface นี้ จึงทำงานได้ทั้งใน/นอก transaction โดยไม่ต้องรู้ว่าอยู่ในบริบทไหน
type Executor interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type txCtxKey struct{}

// TxManager เป็นตัวกลางเดียวที่ตัดสินว่า query ควรวิ่งผ่าน pool หรือ transaction ที่เปิดค้างอยู่
//
// สำคัญสำหรับสถาปัตยกรรมแบบ module: การจองห้องหนึ่งครั้งต้องแตะทั้งตาราง rooms (module room),
// branches (module branch) และ bookings (module booking) ให้อยู่ใน transaction เดียวกัน
// การส่ง transaction ผ่าน context ทำให้ทุก module เข้าร่วม transaction เดียวกันได้
// โดยไม่ต้องผูก repository ของกันและกันไว้ตรง ๆ
type TxManager struct {
	pool *pgxpool.Pool
}

func NewTxManager(pool *pgxpool.Pool) *TxManager {
	return &TxManager{pool: pool}
}

func (m *TxManager) Pool() *pgxpool.Pool { return m.pool }

// Executor คืน transaction ที่ผูกกับ context ถ้ามี ไม่มีก็ใช้ pool ตามปกติ
func (m *TxManager) Executor(ctx context.Context) Executor {
	if tx, ok := ctx.Value(txCtxKey{}).(pgx.Tx); ok {
		return tx
	}
	return m.pool
}

// WithTx รัน fn ภายใน transaction เดียว rollback อัตโนมัติเมื่อ fn คืน error
//
// ถ้า context มี transaction อยู่แล้ว จะเข้าร่วม transaction เดิม (ไม่ซ้อนใหม่)
// เพื่อให้ service ของ module หนึ่งเรียก service ของอีก module ได้อย่างปลอดภัย
func (m *TxManager) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	if _, ok := ctx.Value(txCtxKey{}).(pgx.Tx); ok {
		return fn(ctx)
	}

	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }() // no-op ถ้า commit สำเร็จไปแล้ว

	if err := fn(context.WithValue(ctx, txCtxKey{}, tx)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// ---------------------------------------------------------------- helpers

// NormalizeErr แปลง pgx.ErrNoRows ให้เป็น ErrNotFound ของเรา
func NormalizeErr(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

// IsUniqueViolation บอกว่า error เกิดจากการชนกับ unique constraint หรือไม่
func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

// Itoa แปลง int เป็น string ใช้สร้าง placeholder ($1, $2, ...) ตอนประกอบ query แบบไดนามิก
func Itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}

// Binder ช่วยประกอบ WHERE clause แบบไดนามิกโดยยังใช้ parameterized query เสมอ
type Binder struct {
	Args []any
}

// Bind เพิ่มค่าเข้า args แล้วคืน placeholder ที่ตรงกัน
func (b *Binder) Bind(v any) string {
	b.Args = append(b.Args, v)
	return "$" + Itoa(len(b.Args))
}
