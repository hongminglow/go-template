package httpx

import (
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiterConfig configures the token bucket.
type RateLimiterConfig struct {
	RequestsPerSecond float64
	BurstSize         int
}

// TokenBucketRateLimiter implements a token bucket algorithm to rate limit incoming requests by IP.
func TokenBucketRateLimiter(cfg RateLimiterConfig) func(http.Handler) http.Handler {
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	// Background job to clean up old IP entries from memory every minute
	go func() {
		for {
			time.Sleep(time.Minute)
			mu.Lock()
			for ip, c := range clients {
				if time.Since(c.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Basic way to get client IP. Use real IP from proxy if behind load balancer.
			ip := r.RemoteAddr

			mu.Lock()
			if _, found := clients[ip]; !found {
				// Initialize a new rate limiter for that specific IP
				clients[ip] = &client{
					limiter: rate.NewLimiter(rate.Limit(cfg.RequestsPerSecond), cfg.BurstSize),
				}
			}

			clients[ip].lastSeen = time.Now()

			if !clients[ip].limiter.Allow() {
				mu.Unlock()
				WriteError(w, http.StatusTooManyRequests, "too many requests")
				return
			}
			mu.Unlock()

			next.ServeHTTP(w, r)
		})
	}
}

/*
// LEAKY BUCKET IMPLEMENTATION:
// The Leaky Bucket algorithm processes requests at a constant, steady rate regardless of burstiness.
// Think of a bucket that drips water at a constant rate. If water is poured in faster than it leaks out, the bucket overflows.
//
// In Go, leaky buckets are often implemented using channels, wait groups, or custom goroutine schedulers.
// Wait, the golang.org/x/time/rate package acts as a token bucket, but you can also configure it to behave similarly to a leaky bucket
// by keeping burst size small and requests consistent.
// A pure leaky bucket explicitly queues requests or immediately shapes them into a constant drain rate.

// Commented Example of Leaky Bucket Middleware logic:
//
// import "go.uber.org/ratelimit" // Popular leaky bucket package
//
// func LeakyBucketRateLimiter(rps int) func(http.Handler) http.Handler {
// 	// Creates a leaky bucket that allows X operations per second (constant processing rate)
//  // ratelimit.New(rps) creates it.
//  lim := ratelimit.New(rps)
//
// 	return func(next http.Handler) http.Handler {
// 		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 			// This will block until the bucket is ready to process the "drip".
//          // If you don't want to block connection completely, you'd check a queue bound.
// 			lim.Take()
// 			next.ServeHTTP(w, r)
// 		})
// 	}
// }
*/
