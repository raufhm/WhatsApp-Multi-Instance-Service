package totp

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base32"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
	"time"

	qrcode "github.com/skip2/go-qrcode"
	"golang.org/x/crypto/bcrypt"
)

const (
	defaultKeyString = "whatsapp-multi-instance-totp-key-32b!"
	Period           = 30
	Digits           = 6
	Issuer           = "whops"
	AccountPrefix    = Issuer
)

var (
	ErrInvalidSecret     = errors.New("invalid TOTP secret")
	ErrInvalidCiphertext = errors.New("invalid ciphertext")
)

// getEncryptionKey returns a 32-byte AES key from env TOTP_ENCRYPTION_KEY or a default.
func getEncryptionKey(overrideKey ...[]byte) []byte {
	if len(overrideKey) > 0 && len(overrideKey[0]) == 32 {
		return overrideKey[0]
	}
	keyStr := os.Getenv("TOTP_ENCRYPTION_KEY")
	if keyStr == "" {
		keyStr = defaultKeyString
	}
	hash := sha256.Sum256([]byte(keyStr))
	return hash[:]
}

// EncryptSecret encrypts a plaintext secret using AES-256-GCM.
func EncryptSecret(plaintext string, overrideKey ...[]byte) (string, error) {
	key := getEncryptionKey(overrideKey...)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptSecret decrypts an AES-256-GCM encrypted secret.
func DecryptSecret(encrypted string, overrideKey ...[]byte) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}
	key := getEncryptionKey(overrideKey...)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", ErrInvalidCiphertext
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

// GenerateSecret generates a 20-byte random Base32 encoded string without padding.
func GenerateSecret() (string, error) {
	b := make([]byte, 20)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b), nil
}

// GenerateCode calculates the 6-digit TOTP code for a given timestamp.
func GenerateCode(secretBase32 string, t time.Time) (string, error) {
	cleaned := strings.ToUpper(strings.TrimSpace(secretBase32))
	secret, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(cleaned)
	if err != nil {
		// try with standard padding
		secret, err = base32.StdEncoding.DecodeString(cleaned)
		if err != nil {
			return "", ErrInvalidSecret
		}
	}
	counter := uint64(t.Unix() / Period)
	return generateCodeAtCounter(secret, counter), nil
}

func generateCodeAtCounter(secret []byte, counter uint64) string {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, counter)

	mac := hmac.New(sha1.New, secret)
	mac.Write(buf)
	h := mac.Sum(nil)

	offset := h[len(h)-1] & 0x0f
	codeInt := (binary.BigEndian.Uint32(h[offset:offset+4]) & 0x7fffffff) % 1000000
	return fmt.Sprintf("%06d", codeInt)
}

// VerifyCode verifies a 6-digit TOTP code with +/- 1 time step tolerance (window=1).
func VerifyCode(secretBase32, code string, t time.Time) bool {
	code = strings.TrimSpace(code)
	if len(code) != Digits {
		return false
	}
	cleaned := strings.ToUpper(strings.TrimSpace(secretBase32))
	secret, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(cleaned)
	if err != nil {
		secret, err = base32.StdEncoding.DecodeString(cleaned)
		if err != nil {
			return false
		}
	}

	currentStep := uint64(t.Unix() / Period)
	for i := int64(-1); i <= 1; i++ {
		step := uint64(int64(currentStep) + i)
		expected := generateCodeAtCounter(secret, step)
		if subtle.ConstantTimeCompare([]byte(expected), []byte(code)) == 1 {
			return true
		}
	}
	return false
}

// VerifyTOTP verifies a candidate 6-digit code against the current time.
func VerifyTOTP(secretBase32, code string) bool {
	return VerifyCode(secretBase32, code, time.Now())
}

// GenerateOtpauthURL formats an RFC 6238 otpauth URI.
func GenerateOtpauthURL(account, secretBase32 string) string {
	cleaned := strings.ToUpper(strings.TrimSpace(secretBase32))
	label := fmt.Sprintf("%s:%s", AccountPrefix, account)
	escapedLabel := url.PathEscape(label)

	v := url.Values{}
	v.Set("secret", cleaned)
	v.Set("issuer", Issuer)
	v.Set("algorithm", "SHA1")
	v.Set("digits", "6")
	v.Set("period", "30")

	return fmt.Sprintf("otpauth://totp/%s?%s", escapedLabel, v.Encode())
}

// GenerateQRCodeDataURL generates a base64 encoded PNG data URL from an otpauth URI.
func GenerateQRCodeDataURL(otpauthURL string) (string, error) {
	png, err := qrcode.Encode(otpauthURL, qrcode.Medium, 256)
	if err != nil {
		return "", err
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(png), nil
}

const backupCodeChars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // omit confusing 0,O,1,I

// GenerateBackupCodes generates `count` random 8-character codes formatted as XXXX-XXXX.
func GenerateBackupCodes(count int) ([]string, error) {
	codes := make([]string, count)
	for i := 0; i < count; i++ {
		raw := make([]byte, 8)
		if _, err := rand.Read(raw); err != nil {
			return nil, err
		}
		var sb strings.Builder
		for j, b := range raw {
			if j == 4 {
				sb.WriteByte('-')
			}
			sb.WriteByte(backupCodeChars[int(b)%len(backupCodeChars)])
		}
		codes[i] = sb.String()
	}
	return codes, nil
}

// NormalizeBackupCode normalizes backup code by removing whitespace and hyphens and converting to uppercase.
func NormalizeBackupCode(code string) string {
	cleaned := strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(code), "-", ""))
	if len(cleaned) == 8 {
		return cleaned[:4] + "-" + cleaned[4:]
	}
	return cleaned
}

// HashBackupCode computes a bcrypt hash of a backup code.
func HashBackupCode(code string) (string, error) {
	normalized := NormalizeBackupCode(code)
	hash, err := bcrypt.GenerateFromPassword([]byte(normalized), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// VerifyBackupCode verifies a candidate backup code against a bcrypt hash.
func VerifyBackupCode(code, codeHash string) bool {
	normalized := NormalizeBackupCode(code)
	return bcrypt.CompareHashAndPassword([]byte(codeHash), []byte(normalized)) == nil
}

// HashToken computes a SHA-256 hash of a token string.
func HashToken(rawToken string) string {
	sum := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(sum[:])
}
