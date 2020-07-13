package conn

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"net"
)

func UDPPacket(srcAddr string, packet []byte) []byte {
	addrBytes := []byte(srcAddr)
	addrLength := uint16(len(addrBytes))
	bodyLength := uint16(len(packet))
	pkg := new(bytes.Buffer)
	binary.Write(pkg, binary.LittleEndian, addrLength)
	binary.Write(pkg, binary.LittleEndian, addrBytes)
	binary.Write(pkg, binary.LittleEndian, bodyLength)
	binary.Write(pkg, binary.LittleEndian, packet)
	return pkg.Bytes()
}
func ReadUDPPacket(conn *net.Conn) (srcAddr string, packet []byte, err error) {
	reader := bufio.NewReader(*conn)
	var addrLength uint16
	var bodyLength uint16
	err = binary.Read(reader, binary.LittleEndian, &addrLength)
	if err != nil {
		return
	}
	_srcAddr := make([]byte, addrLength)
	n, err := reader.Read(_srcAddr)
	if err != nil {
		return
	}
	if n != int(addrLength) {
		return
	}
	srcAddr = string(_srcAddr)

	err = binary.Read(reader, binary.LittleEndian, &bodyLength)
	if err != nil {
		return
	}
	packet = make([]byte, bodyLength)
	n, err = reader.Read(packet)
	if err != nil {
		return
	}
	if n != int(bodyLength) {
		return
	}
	return
}
