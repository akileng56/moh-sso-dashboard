package modulehttp

import (
	"context"
	"errors"
	"fmt"
	"github.com/moh-sso-dashboard/internal/ratelimit"
	"log"
	"net/http"
	"time"
)

// UserLimiter is a shared rate-store capability, independent of any business module.
type UserLimiter interface {
	Available() bool
	Allow(context.Context, string, int, time.Duration) (bool, int, time.Time, error)
}
type throttled struct{ until time.Time }

func (t throttled) Error() string             { return "rate limit exceeded" }
func (t throttled) StatusCode() int           { return 429 }
func (t throttled) RetryAfter() time.Duration { return time.Until(t.until) }

// PerUserLimit never infers identity from headers or fails open on store errors.
func PerUserLimit(limiter UserLimiter) func(*http.Request, string, int, time.Duration) (http.Header, error) {
	return func(r *http.Request, name string, limit int, window time.Duration) (http.Header, error) {
		actor, valid := Actor(r)
		if !valid {
			return nil, denied(401)
		}
		if limiter == nil || !limiter.Available() || name == "" || limit <= 0 || window <= 0 {
			return nil, errors.New("rate enforcement unavailable")
		}
		allowed, remaining, reset, err := limiter.Allow(r.Context(), ratelimit.Key(name, "user", actor.UserID), limit, window)
		if err != nil {
			return nil, errors.New("rate enforcement unavailable")
		}
		if remaining < 0 {
			remaining = 0
		}
		headers := http.Header{}
		headers.Set("X-RateLimit-Policy", name)
		headers.Set("X-RateLimit-Limit", fmt.Sprint(limit))
		headers.Set("X-RateLimit-Remaining", fmt.Sprint(remaining))
		headers.Set("X-RateLimit-Reset", reset.Format(time.RFC3339))
		if !allowed {
			return headers, throttled{reset}
		}
		return headers, nil
	}
}

// StreamError reports a partial-response failure without writing another body.
func StreamError(r *http.Request, err error) {
	log.Printf("module stream failed: request_id=%s method=%s path=%s error=%v", RequestID(r), r.Method, r.URL.Path, err)
}
