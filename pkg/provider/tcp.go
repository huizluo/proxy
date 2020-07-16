package provider

import (
	"fmt"
	"io"
	"log"
	"net"
	"proxy/pkg/conn"
	server2 "proxy/pkg/server"
	"proxy/pkg/utils"
	"runtime/debug"
	"strconv"
	"time"
)

type TcpProvider struct {
	name string
	args TCPArgs
	upConnPool conn.UpstreamConnPool
}

func NewTcpProvider()Provider{
	return &TcpProvider{
		upConnPool: conn.UpstreamConnPool{},
		args: TCPArgs{},
	}
}

func (t *TcpProvider) GetName() string {
	return t.name
}

func (t *TcpProvider) Stop() {
	t.StopService()
}

func (t *TcpProvider) Start(args interface{}) {
	var err error
	t.args = args.(TCPArgs)
	if t.args.Parent != "" {
		log.Printf("use %s parent %s", t.args.ParentType, t.args.Parent)
	} else {
		log.Fatalf("parent required for %s %s", t.args.Protocol(), t.args.Local)
	}

	t.InitService()

	host, port, _ := net.SplitHostPort(t.args.Local)
	p, _ := strconv.Atoi(port)
	sc := server2.NewServer(host, p)
	if !t.args.IsTLS {
		err = sc.ListenTCP(t.handler)
	} else {
		err = sc.ListenTls(t.args.CertBytes, t.args.KeyBytes, t.handler)
	}
	if err != nil {
		log.Fatalf("listen tcp on addr  %s %s error: %s", host, port,err.Error())
	}

	log.Printf("%s proxy on %s", t.args.Protocol(), (*sc.Listener).Addr())

	utils.WaitSignal()
	t.Stop()
}

func (t *TcpProvider) InitService() {
	t.InitOutConnPool()
}

func (t *TcpProvider) StopService() {
	if t.upConnPool.Pool != nil {
		t.upConnPool.Pool.RemoveAll()
	}
}

func (t *TcpProvider) handler(inConn net.Conn) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("%s conn handler crashed with err : %s \nstack: %s", t.args.Protocol(), err, string(debug.Stack()))
		}
	}()
	var err error
	switch t.args.ParentType {
	case TYPE_TCP:
		fallthrough
	case TYPE_TLS:
		err = t.OutToTCP(&inConn)
	case TYPE_UDP:
		err = t.OutToUDP(&inConn)
	default:
		err = fmt.Errorf("unkown parent type %s", t.args.ParentType)
	}
	if err != nil {
		log.Printf("connect to %s parent %s fail, ERR:%s", t.args.ParentType, t.args.Parent, err)
		conn.CloseConn(&inConn)
	}
}

func (t *TcpProvider) OutToTCP(inConn *net.Conn) (err error) {
	var outConn net.Conn
	var _outConn interface{}
	_outConn, err = t.upConnPool.Pool.Get()
	if err == nil {
		outConn = _outConn.(net.Conn)
	}
	if err != nil {
		log.Printf("connect to %s , err:%s", t.args.Parent, err)
		conn.CloseConn(inConn)
		return
	}
	inAddr := (*inConn).RemoteAddr().String()
	inLocalAddr := (*inConn).LocalAddr().String()
	outAddr := outConn.RemoteAddr().String()
	outLocalAddr := outConn.LocalAddr().String()
	conn.IoBind((*inConn), outConn, func(isSrcErr bool, err error) {
		log.Printf("conn %s - %s - %s -%s released", inAddr, inLocalAddr, outLocalAddr, outAddr)
		conn.CloseConn(inConn)
		conn.CloseConn(&outConn)
	}, func(n int, d bool) {}, 0)
	log.Printf("conn %s - %s - %s -%s connected", inAddr, inLocalAddr, outLocalAddr, outAddr)
	return
}

func (t *TcpProvider) OutToUDP(inConn *net.Conn) (err error) {
	log.Printf("conn created , remote : %s ", (*inConn).RemoteAddr())
	for {
		srcAddr, body, err := conn.ReadUDPPacket(inConn)
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			//log.Printf("connection %s released", srcAddr)
			conn.CloseConn(inConn)
			break
		}
		//log.Debugf("udp packet revecived:%s,%v", srcAddr, body)
		dstAddr, err := net.ResolveUDPAddr("udp", t.args.Parent)
		if err != nil {
			log.Printf("can't resolve address: %s", err)
			conn.CloseConn(inConn)
			break
		}
		clientSrcAddr := &net.UDPAddr{IP: net.IPv4zero, Port: 0}
		udpConn, err := net.DialUDP("udp", clientSrcAddr, dstAddr)
		if err != nil {
			log.Printf("connect to udp %s fail,ERR:%s", dstAddr.String(), err)
			continue
		}
		udpConn.SetDeadline(time.Now().Add(time.Millisecond * time.Duration(t.args.Timeout)))
		_, err = udpConn.Write(body)
		if err != nil {
			log.Printf("send udp packet to %s fail,ERR:%s", dstAddr.String(), err)
			continue
		}
		//log.Debugf("send udp packet to %s success", dstAddr.String())
		buf := make([]byte, 512)
		len, _, err := udpConn.ReadFromUDP(buf)
		if err != nil {
			log.Printf("read udp response from %s fail ,ERR:%s", dstAddr.String(), err)
			continue
		}
		respBody := buf[0:len]
		//log.Debugf("revecived udp packet from %s , %v", dstAddr.String(), respBody)
		_, err = (*inConn).Write(conn.UDPPacket(srcAddr, respBody))
		if err != nil {
			log.Printf("send udp response fail ,ERR:%s", err)
			conn.CloseConn(inConn)
			break
		}
		//log.Printf("send udp response success ,from:%s", dstAddr.String())
	}
	return

}

func (t *TcpProvider) InitOutConnPool() {
	if t.args.ParentType == TYPE_TLS || t.args.ParentType == TYPE_TCP {
		//dur int, isTLS bool, certBytes, keyBytes []byte,
		//parent string, timeout int, InitialCap int, MaxCap int
		t.upConnPool = conn.NewUpstreamConnPool(
			t.args.CheckParentInterval,
			t.args.ParentType == TYPE_TLS,
			t.args.CertBytes, t.args.KeyBytes,
			t.args.Parent,
			t.args.Timeout,
			t.args.PoolSize,
			t.args.PoolSize*2,
		)
	}
}
