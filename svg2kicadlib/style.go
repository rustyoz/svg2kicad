package svg2kicadlib

import "strings"

func splitStyle(style string) map[string]string {
	var r map[string]string
	r = make(map[string]string)
	props := strings.Split(style, ";")

	for _, keyval := range props {
		kv := strings.Split(keyval, ":")
		r[kv[0]] = kv[1]
	}

	return r
}
