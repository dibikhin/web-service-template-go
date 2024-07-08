package middleware

import (
	"time"

	"github.com/go-kit/kit/metrics"

	"ws-dummy-go/internal/dummy"
	"ws-dummy-go/internal/dummy/domain"
)

func NewInstrumentingMiddleware(
	requestCount metrics.Counter, requestLatency metrics.Histogram,
) UserServiceMiddleware {
	return func(next dummy.UserService) dummy.UserService {
		return instrmw{requestCount, requestLatency, next}
	}
}

type instrmw struct {
	requestCount   metrics.Counter
	requestLatency metrics.Histogram

	dummy.UserService
}

func (mw instrmw) Create(s string) (domain.UserID, error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "Create", "error", "false"}
		mw.requestCount.With(lvs...).Add(1)
		mw.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mw.UserService.Create(s)
}
