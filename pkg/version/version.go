package version

import (
	"fmt"
	"regexp"
	"runtime"
	"strings"
)

// Version Properties
var (
	Number    = "0.0.1"
	BuildDate string
	BuildHash string
)

var uaReplacer = regexp.MustCompile(`[^a-z0-9.\-/]`)

// HTTPUserAgent returns a HTTP User-Agent for the product using
// the Number, BuildDate and BuildHash.
func HTTPUserAgent(product string) string {
	s := new(strings.Builder)
	fmt.Fprintf(s, "%s/v%v (", uaReplacer.ReplaceAllString(product, "-"), Number)

	if BuildDate != "" {
		s.WriteString(BuildDate)
		s.WriteString("; ")
	}

	if BuildHash != "" {
		s.WriteString(BuildHash)
		s.WriteString("; ")
	}

	fmt.Fprintf(s, "%v; %v-%v)", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	return s.String()
}
