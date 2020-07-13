package provider

const (
	TYPE_TCP     = "tcp"
	TYPE_UDP     = "udp"
	TYPE_HTTP    = "http"
	TYPE_TLS     = "tls"

)

type Args struct {
	Local     string
	Parent    string
	CertBytes []byte
	KeyBytes  []byte
}

type TCPArgs struct {
	Args
	ParentType          string
	IsTLS               bool
	Timeout             int
	PoolSize            int
	CheckParentInterval int
}

type HTTPArgs struct {
	Args
	Always              bool
	HTTPTimeout         int
	Interval            int
	Blocked             string
	Direct              string
	AuthFile            string
	Auth                []string
	ParentType          string
	LocalType           string
	Timeout             int
	PoolSize            int
	CheckParentInterval int
}
type UDPArgs struct {
	Args
	ParentType          string
	Timeout             int
	PoolSize            int
	CheckParentInterval int
}

func (a TCPArgs) Protocol() string {
	if a.IsTLS {
		return "tls"
	}
	return "tcp"
}
