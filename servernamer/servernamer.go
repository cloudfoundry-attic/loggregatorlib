package servernamer

import "strings"

func ServerName(hostport, hostname string) string {
	uri := strings.Replace(hostport, ".", "-", -1)
	uri = strings.Replace(uri, ":", "-", -1)
	uri = uri + "-" + hostname

	return uri
}
