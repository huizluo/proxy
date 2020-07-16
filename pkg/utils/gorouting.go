package utils

import (
	"log"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
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

func WaitSignal(){
	sig:=make(chan os.Signal,1)
	signal.Notify(sig,
		os.Interrupt,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		)

	res,_:=<-sig
	log.Printf("proxy received [%s] ,service will closed",res.String())
}
