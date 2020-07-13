package limiter

import (
	"golang.org/x/net/context"
	"golang.org/x/time/rate"
	"io"
	"time"
)

const Limit = 1000 * 1000 * 1000

type ReadLimiter struct {
	r io.Reader
	ctx context.Context
	l *rate.Limiter
}

type WriteLimiter struct {
	w io.Writer
	ctx context.Context
	l *rate.Limiter
}

func NewReadLimiter(r io.Reader,ctx context.Context)*ReadLimiter{
	return &ReadLimiter{
		r:r,
		ctx: ctx,
	}
}

func NewWriteLimiter(w io.Writer,ctx context.Context)*WriteLimiter{
	return &WriteLimiter{
		w :w,
		ctx: ctx,
	}
}
//SetLimit set rate limit (bytes/sec) to reader
func (r *ReadLimiter)SetRateLimit(bytePerSec float64){
	r.l = rate.NewLimiter(rate.Limit(bytePerSec),Limit)
	r.l.AllowN(time.Now(),Limit)
}

// Read reads bytes into p.
func (r *ReadLimiter) Read(p []byte) (int, error) {
	if r.l == nil {
		return r.r.Read(p)
	}
	n, err := r.r.Read(p)
	if err != nil {
		return n, err
	}
	if err := r.l.WaitN(r.ctx, n); err != nil {
		return n, err
	}
	return n, nil
}

// SetRateLimit sets rate limit (bytes/sec) to the writer.
func (w *WriteLimiter) SetRateLimit(bytesPerSec float64) {
	w.l = rate.NewLimiter(rate.Limit(bytesPerSec), Limit)
	w.l.AllowN(time.Now(), Limit) // spend initial burst
}

// Write writes bytes from p.
func (w *WriteLimiter) Write(p []byte) (int, error) {
	if w.l == nil {
		return w.w.Write(p)
	}
	n, err := w.w.Write(p)
	if err != nil {
		return n, err
	}
	if err := w.l.WaitN(w.ctx, n); err != nil {
		return n, err
	}
	return n, err
}

