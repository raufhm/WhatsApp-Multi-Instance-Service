package totp

import (
	"bytes"
	"encoding/base64"
	"image"
	_ "image/png"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/makiuchi-d/gozxing"
	zxingqr "github.com/makiuchi-d/gozxing/qrcode"
)

func TestSecretEncryptionDecryption(t *testing.T) {
	secret, err := GenerateSecret()
	if err != nil {
		t.Fatalf("GenerateSecret failed: %v", err)
	}
	if len(secret) == 0 {
		t.Fatal("expected non-empty secret")
	}

	encrypted, err := EncryptSecret(secret)
	if err != nil {
		t.Fatalf("EncryptSecret failed: %v", err)
	}
	if encrypted == secret {
		t.Fatal("encrypted text should not match plaintext")
	}

	decrypted, err := DecryptSecret(encrypted)
	if err != nil {
		t.Fatalf("DecryptSecret failed: %v", err)
	}
	if decrypted != secret {
		t.Fatalf("expected decrypted %q, got %q", secret, decrypted)
	}
}

func TestTOTPGenerationAndVerification(t *testing.T) {
	secret, err := GenerateSecret()
	if err != nil {
		t.Fatalf("GenerateSecret failed: %v", err)
	}

	now := time.Now()
	code, err := GenerateCode(secret, now)
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}
	if len(code) != 6 {
		t.Fatalf("expected 6-digit code, got %q", code)
	}

	// Verify at exact same time
	if !VerifyCode(secret, code, now) {
		t.Fatalf("VerifyCode failed for valid code %q at time %v", code, now)
	}

	// Verify with +25s drift (same or +1 step)
	if !VerifyCode(secret, code, now.Add(25*time.Second)) {
		t.Fatalf("VerifyCode failed for valid code with 25s forward drift")
	}

	// Verify with -25s drift (same or -1 step)
	if !VerifyCode(secret, code, now.Add(-25*time.Second)) {
		t.Fatalf("VerifyCode failed for valid code with 25s backward drift")
	}

	// Verify with +1 step (30s)
	if !VerifyCode(secret, code, now.Add(30*time.Second)) {
		t.Fatalf("VerifyCode should accept +1 step (30s)")
	}

	// Verify with -1 step (-30s)
	if !VerifyCode(secret, code, now.Add(-30*time.Second)) {
		t.Fatalf("VerifyCode should accept -1 step (-30s)")
	}

	// Verify with +2 steps (+65s) should fail
	if VerifyCode(secret, code, now.Add(65*time.Second)) {
		t.Fatalf("VerifyCode should reject +2 steps drift")
	}

	// Verify with wrong code should fail
	if VerifyCode(secret, "000000", now) && code != "000000" {
		t.Fatalf("VerifyCode should reject invalid code")
	}
}

func TestRFC6238KnownVectors(t *testing.T) {
	// RFC 6238 Appendix B uses Base32 secret "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ" (ASCII "12345678901234567890")
	// Time = 59s (T=1): 8 digits 94287082 -> 6 digits 287082
	// Time = 1111111109s (T=37037036): 8 digits 07081804 -> 6 digits 081804
	// Time = 1111111111s (T=37037037): 8 digits 14050471 -> 6 digits 050471
	// Time = 1234567890s (T=41152263): 8 digits 89005924 -> 6 digits 005924
	secret := "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ"

	t59 := time.Unix(59, 0)
	code59, err := GenerateCode(secret, t59)
	if err != nil {
		t.Fatal(err)
	}
	if code59 != "287082" {
		t.Errorf("at t=59s expected 287082, got %s", code59)
	}

	t1109 := time.Unix(1111111109, 0)
	code1109, err := GenerateCode(secret, t1109)
	if err != nil {
		t.Fatal(err)
	}
	if code1109 != "081804" {
		t.Errorf("at t=1111111109s expected 081804, got %s", code1109)
	}

	t1234 := time.Unix(1234567890, 0)
	code1234, err := GenerateCode(secret, t1234)
	if err != nil {
		t.Fatal(err)
	}
	if code1234 != "005924" {
		t.Errorf("at t=1234567890s expected 005924, got %s", code1234)
	}
}

func TestBackupCodes(t *testing.T) {
	codes, err := GenerateBackupCodes(10)
	if err != nil {
		t.Fatalf("GenerateBackupCodes: %v", err)
	}
	if len(codes) != 10 {
		t.Fatalf("expected 10 backup codes, got %d", len(codes))
	}
	for _, c := range codes {
		if len(c) != 9 || c[4] != '-' {
			t.Fatalf("invalid backup code format: %q", c)
		}
		hash, err := HashBackupCode(c)
		if err != nil {
			t.Fatalf("HashBackupCode failed: %v", err)
		}
		if !VerifyBackupCode(c, hash) {
			t.Fatalf("VerifyBackupCode failed for %q", c)
		}
		// Also verify case insensitivity and hyphen insensitivity
		withoutHyphen := strings.ReplaceAll(c, "-", "")
		if !VerifyBackupCode(withoutHyphen, hash) {
			t.Fatalf("VerifyBackupCode without hyphen failed for %q", withoutHyphen)
		}
		lower := strings.ToLower(c)
		if !VerifyBackupCode(lower, hash) {
			t.Fatalf("VerifyBackupCode lower failed for %q", lower)
		}
		if VerifyBackupCode("WRON-CODE", hash) {
			t.Fatalf("VerifyBackupCode should fail for incorrect code")
		}
	}
}

func TestQRCodeDataURLAndOtpauthURL(t *testing.T) {
	secret := "JBSWY3DPEHPK3PXP"
	account := "user@example.com"
	expectedURL := "otpauth://totp/whops:user@example.com?algorithm=SHA1&digits=6&issuer=whops&period=30&secret=JBSWY3DPEHPK3PXP"
	url := GenerateOtpauthURL(account, secret)
	if url != expectedURL {
		t.Fatalf("unexpected otpauth url:\nexpected: %s\ngot:      %s", expectedURL, url)
	}

	dataURL, err := GenerateQRCodeDataURL(url)
	if err != nil {
		t.Fatalf("GenerateQRCodeDataURL failed: %v", err)
	}
	if !strings.HasPrefix(dataURL, "data:image/png;base64,") {
		t.Fatalf("expected data:image/png;base64, prefix, got %s", dataURL[:30])
	}
}

func TestQRCodeDataURLRoundTrip(t *testing.T) {
	secret := "JBSWY3DPEHPK3PXP"
	account := "user@example.com"
	otpauthURL := GenerateOtpauthURL(account, secret)

	dataURL, err := GenerateQRCodeDataURL(otpauthURL)
	if err != nil {
		t.Fatalf("GenerateQRCodeDataURL failed: %v", err)
	}
	const prefix = "data:image/png;base64,"
	if !strings.HasPrefix(dataURL, prefix) {
		t.Fatalf("expected data URL to start with %q", prefix)
	}

	rawPNG, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(dataURL, prefix))
	if err != nil {
		t.Fatalf("failed to decode base64 png: %v", err)
	}

	img, _, err := image.Decode(bytes.NewReader(rawPNG))
	if err != nil {
		t.Fatalf("failed to decode PNG image: %v", err)
	}

	bmp, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		t.Fatalf("failed to create binary bitmap: %v", err)
	}

	res, err := zxingqr.NewQRCodeReader().Decode(bmp, nil)
	if err != nil {
		t.Fatalf("failed to decode QR matrix: %v", err)
	}

	decodedURL := res.GetText()
	if decodedURL != otpauthURL {
		t.Fatalf("decoded QR payload does not match original.\nexpected: %s\ngot:      %s", otpauthURL, decodedURL)
	}

	u, err := url.Parse(decodedURL)
	if err != nil {
		t.Fatalf("failed to parse decoded otpauth URL: %v", err)
	}
	if u.Scheme != "otpauth" {
		t.Errorf("expected scheme otpauth, got %s", u.Scheme)
	}
	if u.Host != "totp" {
		t.Errorf("expected host totp, got %s", u.Host)
	}
	if u.Path != "/whops:user@example.com" {
		t.Errorf("expected path /whops:user@example.com, got %s", u.Path)
	}
	q := u.Query()
	if q.Get("issuer") != Issuer {
		t.Errorf("expected issuer %s, got %s", Issuer, q.Get("issuer"))
	}
	if q.Get("secret") != secret {
		t.Errorf("expected secret %s, got %s", secret, q.Get("secret"))
	}
	if q.Get("algorithm") != "SHA1" {
		t.Errorf("expected algorithm SHA1, got %s", q.Get("algorithm"))
	}
	if q.Get("digits") != "6" {
		t.Errorf("expected digits 6, got %s", q.Get("digits"))
	}
	if q.Get("period") != "30" {
		t.Errorf("expected period 30, got %s", q.Get("period"))
	}
}
