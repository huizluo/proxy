package conn

import (
	"crypto/tls"
	"log"
	"net"
	"time"
)

type UpstreamConnPool struct {
	Pool      Pool
	dur       int
	isTLS     bool
	certBytes []byte
	keyBytes  []byte
	address   string
	timeout   int
}

func NewUpstreamConnPool(dur int, isTLS bool, certBytes, keyBytes []byte, address string, timeout int, InitialCap int, MaxCap int) (up UpstreamConnPool) {
	up = UpstreamConnPool{
		dur:       dur,
		isTLS:     isTLS,
		certBytes: certBytes,
		keyBytes:  keyBytes,
		address:   address,
		timeout:   timeout,
	}
	var err error
	up.Pool, err = NewConnPool(poolConfig{
		IsActive: func(conn interface{}) bool { return true },
		Release: func(conn interface{}) {
			if conn != nil {
				conn.(net.Conn).SetDeadline(time.Now().Add(time.Millisecond))
				conn.(net.Conn).Close()
				// log.Println("conn released")
			}
		},
		InitialCap: InitialCap,
		MaxCap:     MaxCap,
		Factory: func() (conn interface{}, err error) {
			conn, err = up.getConn()
			return
		},
	})
	if err != nil {
		log.Fatalf("init conn pool fail ,%s", err)
	} else {
		if InitialCap > 0 {
			log.Printf("init conn pool success")
			up.initPoolDeamon()
		} else {
			log.Printf("conn pool closed")
		}
	}
	return
}
func (up *UpstreamConnPool) getConn() (conn interface{}, err error) {
	if up.isTLS {
		var _conn tls.Conn
		_conn, err = TlsConnectHost(up.address, up.timeout, up.certBytes, up.keyBytes)
		if err == nil {
			conn = net.Conn(&_conn)
		}
	} else {
		conn, err = ConnectHost(up.address, up.timeout)
	}
	return
}

func (up *UpstreamConnPool) initPoolDeamon() {
	go func() {
		if up.dur <= 0 {
			return
		}
		log.Printf("pool deamon started")
		for {
			time.Sleep(time.Second * time.Duration(up.dur))
			conn, err := up.getConn()
			if err != nil {
				log.Printf("pool deamon err %s , release pool", err)
				up.Pool.RemoveAll()
			} else {
				conn.(net.Conn).SetDeadline(time.Now().Add(time.Millisecond))
				conn.(net.Conn).Close()
			}
		}
	}()
}
