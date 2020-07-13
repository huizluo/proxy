package utils

import (
	"log"
	"runtime/debug"
)

func GoWithRecover(handler func(),recoverHandler func()){
	go func() {
		defer func() {
			if e:=recover();e!=nil{
				log.Printf("%s goroutine panic: %v\n%s\n", CacheTime(), e, string(debug.Stack()))
			}
			if recoverHandler!=nil{
				recoverHandler()
			}
		}()
		handler()
	}()
}

func WithRecover(handler func(),recoverHandler func()){
	defer func() {
		if e:=recover();e!=nil{
			log.Printf("%s goroutine panic: %v\n%s\n", CacheTime(), e, string(debug.Stack()))
		}
		if recoverHandler!=nil{
			recoverHandler()
		}
	}()
	handler()
}
