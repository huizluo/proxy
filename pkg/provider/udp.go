package provider

import (
	"bufio"
	"fmt"
	"hash/crc32"
	"io"
	"log"
	"net"
	"proxy/pkg/conn"
	"proxy/pkg/server"
	"proxy/pkg/utils"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
)

type UdpProvider struct {
	name string
	args UDPArgs
	upConnPool conn.UpstreamConnPool
	server *server.Server
	p utils.ConcurrentMap
}

func (u *UdpProvider) GetName() string {
	return u.name
}

func (u *UdpProvider) Stop() {
	u.StopService()
}

func NewUdpProvider() Provider {
	return &UdpProvider{
		upConnPool: conn.UpstreamConnPool{},
		p:       utils.NewConcurrentMap(),
	}
}
func (u *UdpProvider) InitService() {
	if u.args.ParentType != TYPE_UDP {
		u.InitOutConnPool()
	}
}
func (u *UdpProvider) StopService() {
	if u.upConnPool.Pool != nil {
		u.upConnPool.Pool.RemoveAll()
	}
}
func (u *UdpProvider) Start(args interface{}) {
	var err error
	u.args = args.(UDPArgs)
	if u.args.Parent != "" {
		log.Printf("use %s parent %s", u.args.ParentType, u.args.Parent)
	} else {
		log.Fatalf("parent required for udp %s", u.args.Local)
	}

	u.InitService()

	host, port, _ := net.SplitHostPort(u.args.Local)
	p, _ := strconv.Atoi(port)
	sc := server.NewServer(host, p)
	u.server = &sc
	err = sc.ListenUDP(u.callback)
	if err != nil {
		log.Fatalf("listen udp on %s:%s error:%s",host,port,err.Error())
	}
	log.Printf("udp proxy on %s", (*sc.UDPListener).LocalAddr())

	utils.WaitSignal()
	u.Stop()
}

func (u *UdpProvider) Clean() {
	u.StopService()
}
func (u *UdpProvider) callback(packet []byte, localAddr, srcAddr *net.UDPAddr) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("udp conn handler crashed with err : %s \nstack: %s", err, string(debug.Stack()))
		}
	}()
	var err error
	switch u.args.ParentType {
	case TYPE_TCP:
		fallthrough
	case TYPE_TLS:
		err = u.OutToTCP(packet, localAddr, srcAddr)
	case TYPE_UDP:
		err = u.OutToUDP(packet, localAddr, srcAddr)
	default:
		err = fmt.Errorf("unkown parent type %s", u.args.ParentType)
	}
	if err != nil {
		log.Printf("connect to %s parent %s fail, ERR:%s", u.args.ParentType, u.args.Parent, err)
	}
}
func (u *UdpProvider) GetConn(connKey string) (conn net.Conn, isNew bool, err error) {
	isNew = !u.p.Has(connKey)
	var _conn interface{}
	if isNew {
		_conn, err = u.upConnPool.Pool.Get()
		if err != nil {
			return nil, false, err
		}
		u.p.Set(connKey, _conn)
	} else {
		_conn, _ = u.p.Get(connKey)
	}
	conn = _conn.(net.Conn)
	return
}
func (u *UdpProvider) OutToTCP(packet []byte, localAddr, srcAddr *net.UDPAddr) (err error) {
	numLocal := crc32.ChecksumIEEE([]byte(localAddr.String()))
	numSrc := crc32.ChecksumIEEE([]byte(srcAddr.String()))
	mod := uint32(u.args.PoolSize)
	if mod == 0 {
		mod = 10
	}
	connKey := uint64((numLocal/10)*10 + numSrc%mod)
	udpConn, isNew, err := u.GetConn(fmt.Sprintf("%d", connKey))
	if err != nil {
		log.Printf("upd get conn to %s parent %s fail, ERR:%s", u.args.ParentType, u.args.Parent, err)
		return
	}
	if isNew {
		go func() {
			defer func() {
				if err := recover(); err != nil {
					log.Printf("udp conn handler out to tcp crashed with err : %s \nstack: %s", err, string(debug.Stack()))
				}
			}()
			log.Printf("conn %d created , local: %s", connKey, srcAddr.String())
			for {
				srcAddrFromConn, body, err := conn.ReadUDPPacket(&udpConn)
				if err == io.EOF || err == io.ErrUnexpectedEOF {
					//log.Printf("connection %d released", connKey)
					u.p.Remove(fmt.Sprintf("%d", connKey))
					break
				}
				if err != nil {
					log.Printf("parse revecived udp packet fail, err: %s", err)
					continue
				}
				//log.Printf("udp packet revecived over parent , local:%s", srcAddrFromConn)
				_srcAddr := strings.Split(srcAddrFromConn, ":")
				if len(_srcAddr) != 2 {
					log.Printf("parse revecived udp packet fail, addr error : %s", srcAddrFromConn)
					continue
				}
				port, _ := strconv.Atoi(_srcAddr[1])
				dstAddr := &net.UDPAddr{IP: net.ParseIP(_srcAddr[0]), Port: port}
				_, err = u.server.UDPListener.WriteToUDP(body, dstAddr)
				if err != nil {
					log.Printf("udp response to local %s fail,ERR:%s", srcAddr, err)
					continue
				}
				//log.Printf("udp response to local %s success", srcAddr)
			}
		}()
	}
	//log.Printf("select conn %d , local: %s", connKey, srcAddr.String())
	writer := bufio.NewWriter(udpConn)
	//fmt.Println(conn, writer)
	writer.Write(conn.UDPPacket(srcAddr.String(), packet))
	err = writer.Flush()
	if err != nil {
		log.Printf("write udp packet to %s fail ,flush err:%s", u.args.Parent, err)
		return
	}
	//log.Printf("write packet %v", packet)
	return
}
func (u *UdpProvider) OutToUDP(packet []byte, localAddr, srcAddr *net.UDPAddr) (err error) {
	//log.Printf("udp packet revecived:%s,%v", srcAddr, packet)
	dstAddr, err := net.ResolveUDPAddr("udp", u.args.Parent)
	if err != nil {
		log.Printf("resolve udp addr %s fail  fail,ERR:%s", dstAddr.String(), err)
		return
	}
	clientSrcAddr := &net.UDPAddr{IP: net.IPv4zero, Port: 0}
	conn, err := net.DialUDP("udp", clientSrcAddr, dstAddr)
	if err != nil {
		log.Printf("connect to udp %s fail,ERR:%s", dstAddr.String(), err)
		return
	}
	conn.SetDeadline(time.Now().Add(time.Millisecond * time.Duration(u.args.Timeout)))
	_, err = conn.Write(packet)
	if err != nil {
		log.Printf("send udp packet to %s fail,ERR:%s", dstAddr.String(), err)
		return
	}
	//log.Printf("send udp packet to %s success", dstAddr.String())
	buf := make([]byte, 512)
	len, _, err := conn.ReadFromUDP(buf)
	if err != nil {
		log.Printf("read udp response from %s fail ,ERR:%s", dstAddr.String(), err)
		return
	}
	//log.Printf("revecived udp packet from %s , %v", dstAddr.String(), respBody)
	_, err = u.server.UDPListener.WriteToUDP(buf[0:len], srcAddr)
	if err != nil {
		log.Printf("send udp response to cluster fail ,ERR:%s", err)
		return
	}
	//log.Printf("send udp response to cluster success ,from:%s", dstAddr.String())
	return
}
func (u *UdpProvider) InitOutConnPool() {
	if u.args.ParentType == TYPE_TLS || u.args.ParentType == TYPE_TCP {
		//dur int, isTLS bool, certBytes, keyBytes []byte,
		//parent string, timeout int, InitialCap int, MaxCap int
		u.upConnPool = conn.NewUpstreamConnPool(
			u.args.CheckParentInterval,
			u.args.ParentType == TYPE_TLS,
			u.args.CertBytes, u.args.KeyBytes,
			u.args.Parent,
			u.args.Timeout,
			u.args.PoolSize,
			u.args.PoolSize*2,
		)
	}
}

