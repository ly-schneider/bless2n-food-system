package utils

import "testing"

func TestHashAndVerifyOTPArgon2(t *testing.T) {
    code := "123456"
    phc, err := HashOTPArgon2(code)
    if err != nil { t.Fatalf("hash error: %v", err) }
    if phc == "" { t.Fatalf("empty phc string") }
    ok, err := VerifyOTPArgon2(code, phc)
    if err != nil { t.Fatalf("verify error: %v", err) }
    if !ok { t.Fatalf("expected verification ok") }
    ok2, _ := VerifyOTPArgon2("000000", phc)
    if ok2 { t.Fatalf("expected verification to fail for wrong code") }
}

