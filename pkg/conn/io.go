package conn

import (
	"io"
	"log"
	"proxy/pkg/limiter"
	"runtime/debug"
	"sync"
)

func IoBind(dst io.ReadWriter, src io.ReadWriter, fn func(isSrcErr bool, err error), cfn func(count int, isPositive bool), bytesPreSec float64) {
	var one = &sync.Once{}
	go func() {
		defer func() {
			if e := recover(); e != nil {
				log.Printf("IoBind crashed , err : %s , \ntrace:%s", e, string(debug.Stack()))
			}
		}()
		var err error
		var isSrcErr bool
		if bytesPreSec > 0 {
			reader := limiter.NewReadLimiter(src,nil)
			reader.SetRateLimit(bytesPreSec)
			_, isSrcErr, err = ioCopy(dst, reader, func(c int) {
				cfn(c, false)
			})

		} else {
			_, isSrcErr, err = ioCopy(dst, src, func(c int) {
				cfn(c, false)
			})
		}
		if err != nil {
			one.Do(func() {
				fn(isSrcErr, err)
			})
		}
	}()
	go func() {
		defer func() {
			if e := recover(); e != nil {
				log.Printf("IoBind crashed , err : %s , \ntrace:%s", e, string(debug.Stack()))
			}
		}()
		var err error
		var isSrcErr bool
		if bytesPreSec > 0 {
			newReader := limiter.NewReadLimiter(dst,nil)
			newReader.SetRateLimit(bytesPreSec)
			_, isSrcErr, err = ioCopy(src, newReader, func(c int) {
				cfn(c, true)
			})
		} else {
			_, isSrcErr, err = ioCopy(src, dst, func(c int) {
				cfn(c, true)
			})
		}
		if err != nil {
			one.Do(func() {
				fn(isSrcErr, err)
			})
		}
	}()
}
func ioCopy(dst io.Writer, src io.Reader, fn ...func(count int)) (written int64, isSrcErr bool, err error) {
	buf := make([]byte, 32*1024)
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if nw > 0 {
				written += int64(nw)
				if len(fn) == 1 {
					fn[0](nw)
				}
			}
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er != nil {
			err = er
			isSrcErr = true
			break
		}
	}
	return written, isSrcErr, err
}
