package client

import (
	"github.com/plimble/arlong/client/golang"
	"testing"
)

func TestGenerate(t *testing.T) {
	g := New(Config{
		Src:            "./swagger.json",
		Dest:           "./out",
		ClientTemplate: golang.New(),
		TemplatePath:   "",
	})

	g.Generate()

}
