// Package format contains common declarations for openapi3 supported formats
package format

// TODO: https://spec.openapis.org/registry/format/

const (
	Hostname = "hostname"
	IPv4     = "ipv4"
	IPv6     = "ipv6"
	URI      = "uri"
	Email    = "email"
	UUID     = "uuid"
	Binary   = "binary"
	Date     = "date"
	DateTime = "date-time"
)

var known = map[string]struct{}{
	Hostname: {},
	IPv4:     {},
	IPv6:     {},
	Email:    {},
	UUID:     {},
	Binary:   {},
	Date:     {},
	DateTime: {},
}

func Register(format string) bool {
	if _, ok := known[format]; !ok {
		known[format] = struct{}{}
		return true
	}
	return false
}

// IsKnown returns true if provided format is known
func IsKnown(format string) bool {
	_, ok := known[format]
	return ok
}
