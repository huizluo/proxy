package server

import (
	"fmt"
	"log"
	"net"
	"proxy/pkg/conn"
	"runtime/debug"
)

type Server struct {
	ip          string
	port        int
	Listener    *net.Listener
	UDPListener *net.UDPConn
	errHandler  func(err error)
}

func NewServer(ip string, port int) Server {
	return Server{
		ip:   ip,
		port: port,
		errHandler: func(err error) {
			log.Printf("accept error,ERR: %s", err)
		},
	}
}

func (s *Server) SetErrHandler(f func(err error)) {
	s.errHandler = f
}

func (s *Server) ListenTls(cert, key []byte, f func(c net.Conn)) (err error) {
	s.Listener, err = conn.ListenTls(s.ip, s.port, cert, key)
	if s.Listener != nil && err == nil {
		go func() {
			for {
				var conn net.Conn
				conn, err = (*s.Listener).Accept()
				if err != nil {
					go func() {
						defer func() {
							if e := recover(); e != nil {
								log.Printf("connection handler crashed , ERR: %s , \ntrace:%s", e, string(debug.Stack()))
							}
						}()
						f(conn)
					}()
				} else {
					s.errHandler(err)
					(*s.Listener).Close()
					break
				}
			}
		}()
	}
	return
}

func (s *Server) ListenTCP(fn func(conn net.Conn)) (err error) {
	var l net.Listener
	l, err = net.Listen("tcp", fmt.Sprintf("%s:%d", s.ip, s.port))
	if err == nil {
		s.Listener = &l
		go func() {
			defer func() {
				if e := recover(); e != nil {
					log.Printf("ListenTCP crashed , err : %s , \ntrace:%s", e, string(debug.Stack()))
				}
			}()
			for {
				var conn net.Conn
				conn, err = (*s.Listener).Accept()
				if err == nil {
					go func() {
						defer func() {
							if e := recover(); e != nil {
								log.Printf("connection handler crashed , err : %s , \ntrace:%s", e, string(debug.Stack()))
							}
						}()
						fn(conn)
					}()
				} else {
					s.errHandler(err)
					break
				}
			}
		}()
	}
	return
}


func (s *Server) ListenUDP(fn func(packet []byte, localAddr, srcAddr *net.UDPAddr)) (err error) {
	addr := &net.UDPAddr{IP: net.ParseIP(s.ip), Port: s.port}
	l, err := net.ListenUDP("udp", addr)
	if err == nil {
		s.UDPListener = l
		go func() {
			defer func() {
				if e := recover(); e != nil {
					log.Printf("ListenUDP crashed , err : %s , \ntrace:%s", e, string(debug.Stack()))
				}
			}()
			for {
				var buf = make([]byte, 2048)
				n, srcAddr, err := (*s.UDPListener).ReadFromUDP(buf)
				if err == nil {
					packet := buf[0:n]
					go func() {
						defer func() {
							if e := recover(); e != nil {
								log.Printf("udp data handler crashed , err : %s , \ntrace:%s", e, string(debug.Stack()))
							}
						}()
						fn(packet, addr, srcAddr)
					}()
				} else {
					s.errHandler(err)
					break
				}
			}
		}()
	}
	return
}
