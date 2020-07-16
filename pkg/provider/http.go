package provider

import (
	"fmt"
	"io"
	"log"
	"net"
	"proxy/pkg/auth"
	"proxy/pkg/checker"
	"proxy/pkg/conn"
	server2 "proxy/pkg/server"
	"proxy/pkg/utils"
	"runtime/debug"
	"strconv"
)

type httpProvider struct {
	name string
	args HTTPArgs
	upConnPool conn.UpstreamConnPool
	checker checker.Checker
	basicAuth auth.BasicAuth
}

func (h httpProvider) GetName() string {
	return h.name
}

func (h httpProvider) Start(args interface{}){
	log.Printf("http provider start")
	h.args = args.(HTTPArgs)
	if h.args.Parent!=""{
		h.checker = checker.NewChecker(h.args.HTTPTimeout,int64(h.args.Interval),h.args.Blocked,h.args.Direct)
	}

	host,port,err:=net.SplitHostPort(h.args.Local)
	if err!=nil{
		panic(err)
	}
	p,_:=strconv.Atoi(port)
	server:=server2.NewServer(host,p)
	if h.args.LocalType == TYPE_TCP{
		err = server.ListenTCP(h.handler)
	}else{
		err = server.ListenTls(h.args.CertBytes,h.args.KeyBytes,h.handler)
	}

	if err == nil{
		log.Printf("%s http(s) proxy on %s",h.args.LocalType,(*server.Listener).Addr())
	}

	utils.WaitSignal()

	h.Stop()
}

func (h httpProvider) Stop() {
	log.Printf("http provider stop")
	if h.upConnPool.Pool !=nil{
		h.upConnPool.Pool.RemoveAll()
	}
}


func NewHttpProvider()Provider{
	return httpProvider{
		upConnPool: conn.UpstreamConnPool{},
		args: HTTPArgs{},
	}
}

func (h httpProvider) handler(inConn net.Conn){
	defer func() {
		if err := recover(); err != nil {
			log.Printf("http(s) conn handler crashed with err : %s \nstack: %s", err, string(debug.Stack()))
		}
	}()
	req, err := NewHTTPRequest(&inConn, 4096, h.IsBasicAuth(), &h.basicAuth)
	if err != nil {
		if err != io.EOF {
			log.Printf("decoder error , form %s, ERR:%s", err, inConn.RemoteAddr())
		}
		conn.CloseConn(&inConn)
		return
	}
	address := req.Host

	useProxy := true
	if h.args.Parent == "" {
		useProxy = false
	} else if h.args.Always {
		useProxy = true
	} else {
		if req.IsHTTPS() {
			h.checker.Add(address, true, req.Method, "", nil)
		} else {
			h.checker.Add(address, false, req.Method, req.URL, req.HeadBuf)
		}
		//var n, m uint
		useProxy, _, _ = h.checker.IsBlocked(req.Host)
	}
	log.Printf("use proxy : %v, %s", useProxy, address)
	//os.Exit(0)
	err = h.OutToTCP(useProxy, address, &inConn, &req)
	if err != nil {
		if h.args.Parent == "" {
			log.Printf("connect to %s fail, ERR:%s", address, err)
		} else {
			log.Printf("connect to %s parent %s fail", h.args.ParentType, h.args.Parent)
		}
		conn.CloseConn(&inConn)
	}
}

func (h *httpProvider) OutToTCP(useProxy bool, address string, inConn *net.Conn, req *HTTPRequest) (err error) {
	inAddr := (*inConn).RemoteAddr().String()
	inLocalAddr := (*inConn).LocalAddr().String()
	//防止死循环
	if h.IsDeadLoop(inLocalAddr, req.Host) {
		conn.CloseConn(inConn)
		err = fmt.Errorf("dead loop detected , %s", req.Host)
		return
	}
	var outConn net.Conn
	var _outConn interface{}
	if useProxy {
		_outConn, err = h.upConnPool.Pool.Get()
		if err == nil {
			outConn = _outConn.(net.Conn)
		}
	} else {
		outConn, err = conn.ConnectHost(address, h.args.Timeout)
	}
	if err != nil {
		log.Printf("connect to %s , err:%s", h.args.Parent, err)
		conn.CloseConn(inConn)
		return
	}

	outAddr := outConn.RemoteAddr().String()
	outLocalAddr := outConn.LocalAddr().String()

	if req.IsHTTPS() && !useProxy {
		req.HTTPSReply()
	} else {
		outConn.Write(req.HeadBuf)
	}
	conn.IoBind((*inConn), outConn, func(isSrcErr bool, err error) {
		log.Printf("conn %s - %s - %s -%s released [%s]", inAddr, inLocalAddr, outLocalAddr, outAddr, req.Host)
		conn.CloseConn(inConn)
		conn.CloseConn(&outConn)
	}, func(n int, d bool) {}, 0)
	log.Printf("conn %s - %s - %s - %s connected [%s]", inAddr, inLocalAddr, outLocalAddr, outAddr, req.Host)
	return
}
func (h *httpProvider) OutToUDP(inConn *net.Conn) (err error) {
	return
}
func (h *httpProvider) InitOutConnPool() {
	if h.args.ParentType == TYPE_TLS || h.args.ParentType == TYPE_TCP {
		//dur int, isTLS bool, certBytes, keyBytes []byte,
		//parent string, timeout int, InitialCap int, MaxCap int
		h.upConnPool = conn.NewUpstreamConnPool(
			h.args.CheckParentInterval,
			h.args.ParentType == TYPE_TLS,
			h.args.CertBytes, h.args.KeyBytes,
			h.args.Parent,
			h.args.Timeout,
			h.args.PoolSize,
			h.args.PoolSize*2,
		)
	}
}
func (h *httpProvider) InitBasicAuth() (err error) {
	h.basicAuth = auth.NewBasicAuth()
	if h.args.AuthFile != "" {
		var n = 0
		n, err = h.basicAuth.AddFromFile(h.args.AuthFile)
		if err != nil {
			err = fmt.Errorf("auth-file ERR:%s", err)
			return
		}
		log.Printf("auth data added from file %d , total:%d", n, h.basicAuth.Total())
	}
	if len(h.args.Auth) > 0 {
		n := h.basicAuth.Add(h.args.Auth)
		log.Printf("auth data added %d, total:%d", n, h.basicAuth.Total())
	}
	return
}
func (h *httpProvider) IsBasicAuth() bool {
	return h.args.AuthFile != "" || len(h.args.Auth) > 0
}
func (h *httpProvider) IsDeadLoop(inLocalAddr string, host string) bool {
	inIP, inPort, err := net.SplitHostPort(inLocalAddr)
	if err != nil {
		return false
	}
	outDomain, outPort, err := net.SplitHostPort(host)
	if err != nil {
		return false
	}
	if inPort == outPort {
		var outIPs []net.IP
		outIPs, err = net.LookupIP(outDomain)
		if err == nil {
			for _, ip := range outIPs {
				if ip.String() == inIP {
					return true
				}
			}
		}
		interfaceIPs, err := utils.GetAllInterfaceAddr()
		if err == nil {
			for _, localIP := range interfaceIPs {
				for _, outIP := range outIPs {
					if localIP.Equal(outIP) {
						return true
					}
				}
			}
		}
	}
	return false
}

