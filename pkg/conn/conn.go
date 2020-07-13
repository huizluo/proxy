package conn

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func TlsConnectHost(host string, timeout int, certBytes, keyBytes []byte) (conn tls.Conn, err error) {
	h := strings.Split(host, ":")
	port, _ := strconv.Atoi(h[1])
	return TlsConnect(h[0], port, timeout, certBytes, keyBytes)
}

func TlsConnect(host string, port, timeout int, certBytes, keyBytes []byte) (conn tls.Conn, err error) {
	conf, err := getRequestTlsConfig(certBytes, keyBytes)
	if err != nil {
		return
	}
	_conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), time.Duration(timeout)*time.Millisecond)
	if err != nil {
		return
	}
	return *tls.Client(_conn, conf), err
}

func getRequestTlsConfig(certBytes, keyBytes []byte) (conf *tls.Config, err error) {
	var cert tls.Certificate
	cert, err = tls.X509KeyPair(certBytes, keyBytes)
	if err != nil {
		return
	}
	serverCertPool := x509.NewCertPool()
	ok := serverCertPool.AppendCertsFromPEM(certBytes)
	if !ok {
		err = errors.New("failed to parse root certificate")
	}
	conf = &tls.Config{
		RootCAs:            serverCertPool,
		Certificates:       []tls.Certificate{cert},
		ServerName:         "proxy",
		InsecureSkipVerify: false,
	}
	return
}

func ConnectHost(hostAndPort string, timeout int) (conn net.Conn, err error) {
	conn, err = net.DialTimeout("tcp", hostAndPort, time.Duration(timeout)*time.Millisecond)
	return
}

func ListenTls(ip string, port int, certBytes, keyBytes []byte) (ln *net.Listener, err error) {
	var cert tls.Certificate
	cert, err = tls.X509KeyPair(certBytes, keyBytes)
	if err != nil {
		return
	}
	clientCertPool := x509.NewCertPool()
	ok := clientCertPool.AppendCertsFromPEM(certBytes)
	if !ok {
		err = errors.New("failed to parse root certificate")
	}
	config := &tls.Config{
		ClientCAs:    clientCertPool,
		ServerName:   "proxy",
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
	}
	_ln, err := tls.Listen("tcp", fmt.Sprintf("%s:%d", ip, port), config)
	if err == nil {
		ln = &_ln
	}
	return
}

func HTTPGet(URL string, timeout int) (err error) {
	tr := &http.Transport{}
	var resp *http.Response
	var client *http.Client
	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
		tr.CloseIdleConnections()
	}()
	client = &http.Client{Timeout: time.Millisecond * time.Duration(timeout), Transport: tr}
	resp, err = client.Get(URL)
	if err != nil {
		return
	}
	return
}

func CloseConn(conn *net.Conn) {
	if conn != nil && *conn != nil {
		(*conn).SetDeadline(time.Now().Add(time.Millisecond))
		(*conn).Close()
	}
}
