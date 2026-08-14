package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"backend/internal/config"
	"backend/internal/database"
	"backend/internal/routes"
	"backend/internal/server"
	"backend/internal/storage"
)

func main() {
	if err := run(); err != nil {
		slog.Error("server หยุดทำงาน", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	setupLogger(cfg)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := database.Connect(ctx, cfg.DatabaseURL, cfg.DBMaxConns, cfg.DBMinConns, cfg.DBConnectTimout)
	if err != nil {
		return err
	}
	defer pool.Close()
	slog.Info("เชื่อมต่อฐานข้อมูลสำเร็จ")

	if err := database.Migrate(ctx, pool); err != nil {
		return err
	}

	files, err := storage.NewLocalStore(cfg.UploadDir, cfg.PublicBaseURL, cfg.MaxUploadBytes)
	if err != nil {
		return err
	}

	app := server.New(cfg, pool, files)
	srv := server.NewHTTPServer(":"+cfg.Port, routes.New(app))

	serverErr := make(chan error, 1)
	go func() {
		slog.Info("API พร้อมใช้งาน", "port", cfg.Port, "env", cfg.Env)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		return err
	case <-ctx.Done():
		slog.Info("ได้รับสัญญาณปิดระบบ กำลังปิดอย่างปลอดภัย...")
	}

	// ให้ request ที่ค้างอยู่ทำงานจนจบก่อนปิด
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return err
	}
	slog.Info("ปิดระบบเรียบร้อย")
	return nil
}

func setupLogger(cfg *config.Config) {
	level := slog.LevelDebug
	var h slog.Handler
	if cfg.IsProduction() {
		level = slog.LevelInfo
		h = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	} else {
		h = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	}
	slog.SetDefault(slog.New(h))
}
