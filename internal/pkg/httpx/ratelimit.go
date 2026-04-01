package httpx

import (
	"net"
	"net/http"
	"time"

	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
)

// RateLimiterConfig configures the token bucket.
type RateLimiterConfig struct {
	RequestsPerSecond int
	BurstSize         int
}

// TokenBucketRateLimiter implements a token bucket algorithm to rate limit incoming requests by IP using Redis.
func TokenBucketRateLimiter(rdb *redis.Client, cfg RateLimiterConfig) func(http.Handler) http.Handler {
	limiter := redis_rate.NewLimiter(rdb)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract only the IP address, stripping the ephemeral port
			ip := r.RemoteAddr
			if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
				ip = host
			}

			res, err := limiter.Allow(r.Context(), "rate_limit:ip:"+ip, redis_rate.Limit{
				Rate:   cfg.RequestsPerSecond,
				Burst:  cfg.BurstSize,
				Period: time.Second,
			})
			if err != nil {
				// If redis is down, it fails closed or open. We choose fail closed for safety.
				WriteError(w, http.StatusInternalServerError, "internal server error")
				return
			}

			if res.Allowed == 0 {
				WriteError(w, http.StatusTooManyRequests, "too many requests")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

/*
// LEAKY BUCKET IMPLEMENTATION (REDIS):
// The Leaky Bucket algorithm processes requests at a constant, steady rate regardless of burstiness.
// Think of a bucket that drips water at a constant rate. If water is poured in faster than it leaks out, it overflows (HTTP 429).
//
// In Go + Redis, you have a few ways to achieve this:
//
// 1. GCRA (Generic Cell Rate Algorithm) via redis_rate:
//    The package github.com/go-redis/redis_rate actually uses GCRA under the hood. 
//    By setting Burst = Rate (e.g., 10 req/sec and 10 burst), it mathematically enforces a constant 
//    interval between requests (1 request every 100ms), simulating a leaky bucket's "smooth dripping".
//
// 2. Redis Lists as a Queue (True Leaky Bucket):
//    Requests push to a Redis List (RPUSH). A separate worker pops from the list (LPOP) at a constant rate.
//    If the list length (LLEN) exceeds a threshold, the handler immediately returns 429 Too Many Requests.
//
// 3. Custom Lua Script:
//    You can write a Lua script that tracks `water_level` and `last_drip_time`. 
//    On every request, it calculates how much water leaked out since `last_drip_time`, 
//    updates the `water_level`, and adds 1. If `water_level > bucket_capacity`, it rejects the request.
//
// Example GCRA configuration for Leaky Bucket behavior:
// func LeakyBucketRateLimiter(rdb *redis.Client, rps int) func(http.Handler) http.Handler {
// 	limiter := redis_rate.NewLimiter(rdb)
// 	return func(next http.Handler) http.Handler {
// 		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 			ip := r.RemoteAddr
// 			// Setting Rate = Burst ensures a strictly consistent flow (no bursting allows leaky bucket behavior)
// 			res, _ := limiter.Allow(r.Context(), "leaky_limit:ip:"+ip, redis_rate.Limit{
// 				Rate:   rps,
// 				Burst:  rps, 
// 				Period: time.Second,
// 			})
// 			if res.Allowed == 0 {
// 				WriteError(w, http.StatusTooManyRequests, "bucket overflowing")
// 				return
// 			}
// 			next.ServeHTTP(w, r)
// 		})
// 	}
// }
*/
