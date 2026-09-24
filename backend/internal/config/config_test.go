package config

import (
	"reflect"
	"testing"
)

// setRequired ตั้งค่าขั้นต่ำที่ Load บังคับ เพื่อให้เทสแต่ละตัวโฟกัสเฉพาะตัวแปรที่สนใจ
func setRequired(t *testing.T) {
	t.Helper()
	t.Setenv("DATABASE_URL", "postgres://localhost/test")
	t.Setenv("JWT_SECRET", "config-test-secret-at-least-32-characters")
}

func TestTrustedProxiesDefaultsToNone(t *testing.T) {
	setRequired(t)
	t.Setenv("TRUSTED_PROXIES", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(cfg.TrustedProxies) != 0 {
		t.Errorf("TrustedProxies = %v ต้องว่าง", cfg.TrustedProxies)
	}
}

func TestTrustedProxiesAcceptsIPAndCIDR(t *testing.T) {
	setRequired(t)
	t.Setenv("TRUSTED_PROXIES", " 10.0.0.1 , 172.16.0.0/12,::1")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	want := []string{"10.0.0.1", "172.16.0.0/12", "::1"}
	if !reflect.DeepEqual(cfg.TrustedProxies, want) {
		t.Errorf("TrustedProxies = %v ต้องการ %v", cfg.TrustedProxies, want)
	}
}

// ค่าที่พิมพ์ผิดต้องทำให้ server ไม่ start — ถ้าปล่อยผ่านเงียบ ๆ ระบบจะไม่เชื่อ proxy ตัวจริง
// แล้ว IP ใน log กลายเป็น IP ของ proxy ทั้งหมดโดยไม่มีใครรู้
func TestTrustedProxiesRejectsInvalidValue(t *testing.T) {
	setRequired(t)
	t.Setenv("TRUSTED_PROXIES", "10.0.0.1,nginx")

	if _, err := Load(); err == nil {
		t.Fatal("ต้องได้ error เมื่อ TRUSTED_PROXIES มีค่าที่ไม่ใช่ IP หรือ CIDR")
	}
}
