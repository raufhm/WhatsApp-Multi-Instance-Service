package upload

import (
	"errors"
	"math/rand"
	"strings"
	"time"

	"github.com/aws/smithy-go"
)

// Backoff returns an exponential backoff with an optional jitter term, capped
// at max. attempt is 1-based (first retry uses initial). When jitter is 0 the
// returned duration is deterministic for a given attempt, which keeps unit
// tests stable while still converging to the same capped ceiling.
func Backoff(attempt int, initial, max time.Duration, jitter float64) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	shift := uint(attempt - 1)
	if shift > 62 {
		shift = 62
	}
	d := initial * time.Duration(int64(1)<<shift)
	if d < initial || d > max {
		d = max
	}
	if jitter > 0 {
		j := time.Duration(float64(d) * jitter * (rand.Float64()*2 - 1))
		d += j
	}
	if d < 0 {
		return 0
	}
	return d
}

// transientCodes are AWS/S3 error codes that are safe to retry.
var transientCodes = map[string]bool{
	"RequestTimeout":                         true,
	"RequestTimeoutException":                true,
	"RequestThrottled":                       true,
	"RequestThrottledException":              true,
	"SlowDown":                               true,
	"Throttling":                             true,
	"ThrottlingException":                    true,
	"ProvisionedThroughputExceededException": true,
	"ServiceUnavailable":                     true,
	"ServiceUnavailableException":            true,
	"InternalError":                          true,
	"NetworkingError":                        true,
	"PriorRequestNotComplete":                true,
}

// IsTransientError reports whether err is a retryable storage/network failure.
// Missing configuration, invalid object keys, and other permanent errors are
// not transient and should fail immediately.
func IsTransientError(err error) bool {
	if err == nil {
		return false
	}
	var temporary interface{ Temporary() bool }
	if errors.As(err, &temporary) && temporary.Temporary() {
		return true
	}
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		code := apiErr.ErrorCode()
		if strings.HasPrefix(code, "5") || transientCodes[code] {
			return true
		}
	}
	return isNetworkError(err)
}

func isNetworkError(err error) bool {
	msg := strings.ToLower(err.Error())
	for _, frag := range []string{
		"connection refused", "no such host", "i/o timeout", "read timeout",
		"network is unreachable", "broken pipe", "connection reset by peer",
		"tls handshake timeout", "dial tcp", "lookup ", "context deadline exceeded",
	} {
		if strings.Contains(msg, frag) {
			return true
		}
	}
	return false
}
