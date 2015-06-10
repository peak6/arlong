package client

import (
	"strings"
)

func removeDefinitionRef(s string) string {
	return strings.TrimPrefix(s, "#/definitions/")
}
