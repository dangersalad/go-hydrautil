package hydrautil

import (
	"net/http"
	"regexp"
)

// Bypass configures auth bypasses for methods on paths that match the
// given pattern
type Bypass struct {
	Pattern *regexp.Regexp
	Methods []string
}

func (b *Bypass) canBypass(r *http.Request) bool {
	for _, m := range b.Methods {
		if r.Method == m {
			return b.Pattern.MatchString(r.URL.Path)
		}
	}
	return false
}
