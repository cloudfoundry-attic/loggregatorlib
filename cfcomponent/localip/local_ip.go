package localip

import "net"
import "strconv"

func LocalIP() (string, error) {
	addr, err := net.ResolveUDPAddr("udp", "1.2.3.4:1")
	if err != nil {
		return "", err
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return "", err
	}

	defer conn.Close()

	host, _, err := net.SplitHostPort(conn.LocalAddr().String())
	if err != nil {
		return "", err
	}

	return host, nil
}

func LocalIPAndPort() (string, uint32, error) {
	l, err := net.Listen("tcp4", "0.0.0.0:0")
	if err != nil {
		return "", 0, err
	}
	defer l.Close()

	host, port, err := net.SplitHostPort(l.Addr().String())
	if err != nil {
		return "", 0, err
	}

	portValue, err := strconv.ParseUint(port, 10, 32)
	if err != nil {
		return "", 0, err
	}

	return host, uint32(portValue), nil
}
