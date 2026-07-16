package server

import "regexp"

var macRegexp = regexp.MustCompile(`^[0-9A-Fa-f]{2}(:|-)?[0-9A-Fa-f]{2}(:|-)?[0-9A-Fa-f]{2}(:|-)?[0-9A-Fa-f]{2}(:|-)?[0-9A-Fa-f]{2}(:|-)?[0-9A-Fa-f]{2}$`)

func isValidMAC(mac string) bool {
	return macRegexp.MatchString(mac)
}
