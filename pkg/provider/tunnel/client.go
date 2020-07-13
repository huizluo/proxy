package tunnel

import (
	"bytes"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"proxy/pkg/conn"
	"proxy/pkg/provider"
	"time"
)

type TunnelClient struct {
	cfg TunnelClientArgs
	name string
}

func NewTunnelClient() provider.Provider {
	return &TunnelClient{
		cfg: TunnelClientArgs{},
	}
}

func (s *TunnelClient) InitService() {
}
func (s *TunnelClient) Check() {
	if s.cfg.Parent != "" {
		log.Printf("use tls parent %s", s.cfg.Parent)
	} else {
		log.Fatalf("parent required")
	}
	if s.cfg.CertBytes == nil || s.cfg.KeyBytes == nil {
		log.Fatalf("cert and key file required")
	}
}
func (s *TunnelClient) GetName()string {
	return s.name
}
func (s *TunnelClient) Stop() {
}
func (s *TunnelClient) Start(args interface{}) (err error) {
	s.cfg = args.(TunnelClientArgs)
	s.Check()
	s.InitService()

	for {
		ctrlConn, err := s.GetInConn(CONN_CONTROL)
		if err != nil {
			log.Printf("control connection err: %s", err)
			time.Sleep(time.Second * 3)
			conn.CloseConn(&ctrlConn)
			continue
		}
		if s.cfg.IsUDP {
			log.Printf("proxy on udp tunnel client mode")
		} else {
			log.Printf("proxy on tcp tunnel client mode")
		}
		for {
			signal := make([]byte, 1)
			if signal[0] == 1 {
				continue
			}
			_, err = ctrlConn.Read(signal)
			if err != nil {
				conn.CloseConn(&ctrlConn)
				log.Printf("read connection signal err: %s", err)
				break
			}
			log.Printf("signal revecived:%s", signal)
			if s.cfg.IsUDP {
				go s.ServeUDP()
			} else {
				go s.ServeConn()
			}
		}
	}
}

func (s *TunnelClient) GetInConn(typ uint8) (outConn net.Conn, err error) {
	outConn, err = s.GetConn()
	if err != nil {
		err = fmt.Errorf("connection err: %s", err)
		return
	}
	keyBytes := []byte(s.cfg.Key)
	keyLength := uint16(len(keyBytes))
	pkg := new(bytes.Buffer)
	binary.Write(pkg, binary.LittleEndian, typ)
	binary.Write(pkg, binary.LittleEndian, keyLength)
	binary.Write(pkg, binary.LittleEndian, keyBytes)
	_, err = outConn.Write(pkg.Bytes())
	if err != nil {
		err = fmt.Errorf("write connection data err: %s ,retrying...", err)
		conn.CloseConn(&outConn)
		return
	}
	return
}
func (s *TunnelClient) GetConn() (conn2 net.Conn, err error) {
	var _conn tls.Conn
	_conn, err = conn.TlsConnectHost(s.cfg.Parent, s.cfg.Timeout, s.cfg.CertBytes, s.cfg.KeyBytes)
	if err == nil {
		conn2 = net.Conn(&_conn)
	}
	return
}
func (s *TunnelClient) ServeUDP() {
	var inConn net.Conn
	var err error
	for {
		for {
			inConn, err = s.GetInConn(CONN_CLIENT)
			if err != nil {
				conn.CloseConn(&inConn)
				log.Printf("connection err: %s, retrying...", err)
				time.Sleep(time.Second * 3)
				continue
			} else {
				break
			}
		}
		log.Printf("conn created , remote : %s ", inConn.RemoteAddr())
		for {
			srcAddr, body, err := conn.ReadUDPPacket(&inConn)
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				log.Printf("connection %s released", srcAddr)
				conn.CloseConn(&inConn)
				break
			}
			//log.Printf("udp packet revecived:%s,%v", srcAddr, body)
			go s.processUDPPacket(&inConn, srcAddr, body)
		}
	}
}
func (s *TunnelClient) processUDPPacket(inConn *net.Conn, srcAddr string, body []byte) {
	dstAddr, err := net.ResolveUDPAddr("udp", s.cfg.Local)
	if err != nil {
		log.Printf("can't resolve address: %s", err)
		conn.CloseConn(inConn)
		return
	}
	clientSrcAddr := &net.UDPAddr{IP: net.IPv4zero, Port: 0}
	udpConn, err := net.DialUDP("udp", clientSrcAddr, dstAddr)
	if err != nil {
		log.Printf("connect to udp %s fail,ERR:%s", dstAddr.String(), err)
		return
	}
	udpConn.SetDeadline(time.Now().Add(time.Millisecond * time.Duration(s.cfg.Timeout)))
	_, err = udpConn.Write(body)
	if err != nil {
		log.Printf("send udp packet to %s fail,ERR:%s", dstAddr.String(), err)
		return
	}
	//log.Printf("send udp packet to %s success", dstAddr.String())
	buf := make([]byte, 512)
	len, _, err := udpConn.ReadFromUDP(buf)
	if err != nil {
		log.Printf("read udp response from %s fail ,ERR:%s", dstAddr.String(), err)
		return
	}
	respBody := buf[0:len]
	//log.Printf("revecived udp packet from %s , %v", dstAddr.String(), respBody)
	_, err = (*inConn).Write(conn.UDPPacket(srcAddr, respBody))
	if err != nil {
		log.Printf("send udp response fail ,ERR:%s", err)
		conn.CloseConn(inConn)
		return
	}
	//log.Printf("send udp response success ,from:%s", dstAddr.String())
}
func (s *TunnelClient) ServeConn() {
	var inConn, outConn net.Conn
	var err error
	for {
		inConn, err = s.GetInConn(CONN_CLIENT)
		if err != nil {
			conn.CloseConn(&inConn)
			log.Printf("connection err: %s, retrying...", err)
			time.Sleep(time.Second * 3)
			continue
		} else {
			break
		}
	}

	i := 0
	for {
		i++
		outConn, err = conn.ConnectHost(s.cfg.Local, s.cfg.Timeout)
		if err == nil || i == 3 {
			break
		} else {
			if i == 3 {
				log.Printf("connect to %s err: %s, retrying...", s.cfg.Local, err)
				time.Sleep(2 * time.Second)
				continue
			}
		}
	}

	if err != nil {
		conn.CloseConn(&inConn)
		conn.CloseConn(&outConn)
		log.Printf("build connection error, err: %s", err)
		return
	}

	conn.IoBind(inConn, outConn, func(isSrcErr bool, err error) {
		log.Printf("%s conn %s - %s - %s - %s released", s.cfg.Key, inConn.RemoteAddr(), inConn.LocalAddr(), outConn.LocalAddr(), outConn.RemoteAddr())
		conn.CloseConn(&inConn)
		conn.CloseConn(&outConn)
	}, func(i int, b bool) {}, 0)
	log.Printf("%s conn %s - %s - %s - %s created", s.cfg.Key, inConn.RemoteAddr(), inConn.LocalAddr(), outConn.LocalAddr(), outConn.RemoteAddr())
}

