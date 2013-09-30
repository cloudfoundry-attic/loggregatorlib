package localip

import "net"
import "strconv"

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
