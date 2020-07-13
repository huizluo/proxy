package tunnel

import "proxy/pkg/provider"

const (
	CONN_CONTROL = uint8(1)
	CONN_SERVER  = uint8(2)
	CONN_CLIENT  = uint8(3)
)

type TunnelServerArgs struct {
	provider.Args
	IsUDP   bool
	Key     string
	Timeout int
}
type TunnelClientArgs struct {
	provider.Args
	IsUDP   bool
	Key     string
	Timeout int
}
type TunnelBridgeArgs struct {
	provider.Args
	Timeout int
}
