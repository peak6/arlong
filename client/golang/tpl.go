package golang

import (
	"github.com/plimble/arlong/schema"
)

type GoClient struct {
}

func New() *GoClient {
	return &GoClient{}
}

func (g *GoClient) DefaultContextTpl() []byte {
	return []byte(`
        {{range .}}
        type {{.Name}} struct { {{range .Properties}}
            {{.Name}} {{.Type}}{{end}}
        }
        {{end}}
    `)
}

func (g *GoClient) Language() string {
	return ""
}

func (g *GoClient) DefaultFuncTpl() []byte {
	return nil
}

func (g *GoClient) DefaultClientTpl() []byte {
	return nil
}

func (g *GoClient) ParseContext(swagger schema.Swagger) error {
	return nil
}

func (g *GoClient) ParseFunc(swagger schema.Swagger) error {
	return nil
}

func (g *GoClient) ParseClient(swagger schema.Swagger) error {
	return nil
}
