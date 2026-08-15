// Package server ประกอบ module ทั้งหมดเข้าด้วยกันแล้วสร้าง HTTP server
//
// นี่คือที่เดียวที่รู้ว่ามี module อะไรบ้างและ module ไหนพึ่งพากันอย่างไร
// ตัว module เองไม่รู้จัก server ส่วนการผูก URL ทั้งหมดอยู่ที่ package routes
package server

import (
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"backend/internal/auth"
	"backend/internal/config"
	"backend/internal/database"
	"backend/internal/modules/account"
	"backend/internal/modules/booking"
	"backend/internal/modules/branch"
	"backend/internal/modules/reporting"
	"backend/internal/modules/room"
	"backend/internal/shared/audit"
	"backend/internal/storage"
)

// Server คือ module ทุกตัวที่ต่อสายเรียบร้อยแล้ว พร้อมของกลางที่ชั้น route ต้องใช้
type Server struct {
	Config *config.Config
	Auth   *auth.Manager
	// Assets เก็บรูปสาขา/ห้องไว้ในฐานข้อมูล ชั้น route ใช้เสิร์ฟ GET /files/:assetID
	Assets *storage.DBStore

	Account   *account.Module
	Branch    *branch.Module
	Room      *room.Module
	Booking   *booking.Module
	Reporting *reporting.Module
}

func New(cfg *config.Config, pool *pgxpool.Pool, files *storage.LocalStore) *Server {
	db := database.NewTxManager(pool)
	rec := audit.NewRecorder(db)
	authMgr := auth.NewManager(cfg.JWTSecret, cfg.AccessTokenTTL, cfg.RefreshTokenTTL, cfg.BcryptCost)

	// รูปสาขา/ห้องเก็บในฐานข้อมูล ส่วนสลิปโอนเงินยังเป็นไฟล์บนดิสก์ (files)
	assets := storage.NewDBStore(db, cfg.PublicBaseURL, cfg.MaxUploadBytes)

	// ลำดับการสร้างคือลำดับการพึ่งพา:
	//   branch  -> ไม่พึ่งใคร
	//   account -> ใช้ branch ตรวจว่าสาขาที่จะผูกให้แอดมินมีจริง
	//   room    -> ไม่พึ่งใคร
	//   booking -> ใช้ room (ล็อก/ตรวจว่าง), branch (ค่าธรรมเนียม), account (แจ้งเตือน)
	branchMod := branch.New(db, rec, assets)
	accountMod := account.New(db, authMgr, rec, branchMod.Service)
	roomMod := room.New(db, rec, assets)
	bookingMod := booking.New(db, roomMod.Service, branchMod.Service, accountMod.Service, rec, files)
	reportingMod := reporting.New(db, roomMod.Service, bookingMod.Service)

	return &Server{
		Config: cfg, Auth: authMgr, Assets: assets,
		Account:   accountMod,
		Branch:    branchMod,
		Room:      roomMod,
		Booking:   bookingMod,
		Reporting: reportingMod,
	}
}

// NewHTTPServer ตั้งค่า timeout ที่จำเป็นเพื่อกัน slowloris
func NewHTTPServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}
}
