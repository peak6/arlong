package client

import (
	"bytes"
	"encoding/json"
	"github.com/kr/pretty"
	"github.com/plimble/arlong/schema"
	"io/ioutil"
	"text/template"
)

//gen req/resp obj
//gen model
//gen security method
//gen route method
//gen http client

type Security struct {
	Type             string
	Name             string
	In               string
	Flow             string
	AuthorizationUrl string
	TokenUrl         string
	Scopes           map[string]string
}

type Route struct {
	Request           Request
	ResponseBasicType ResponseBasicType
	ResponseModel     ResponseModel
	Produce           string
	Consume           string
	Security          Security
}

type ResponseBasicType struct {
	Type string
}

type ResponseModel struct {
	Perperties []Property
}

type Request struct {
	Path   []Property
	Query  []Property
	Body   []Property
	Header []KeyVal
}

type KeyVal struct {
	Key string
	Val string
}

type Property struct {
	Name     string
	Type     string
	Required bool
	IsArray  bool
}

type ClientGenerator interface {
	Language() string
	// DefaultContextTpl() []byte
	// DefaultFuncTpl() []byte
	// DefaultClientTpl() []byte
	// ParseContext(swagger schema.Swagger) error
	// ParseFunc(swagger schema.Swagger) error
	// ParseClient(swagger schema.Swagger) error
}

type Config struct {
	Src              string
	Dest             string
	ClientGenerators map[string]ClientGenerator
}

type Generator struct {
	Config
}

func New(config Config) *Generator {
	return &Generator{config}
}

func (g *Generator) Register(cg ClientGenerator) {
	g.ClientGenerators[cg.Language()] = cg
}

func (g *Generator) Generate() {
	swagger := g.getSwagger()

	contexts := g.getContexts(swagger)

	g.parseContext(contexts)

	// contextTpl := config.ClientTemplate.DefaultContextTpl()

	// context := template.New("context")
	// context = context.Parse(string(contextTpl))

	// buf := bytes.NewBuffer(nil)
	// context.Execute(buf)

}

func getPath() {

}

func (g *Generator) getSwagger() *schema.Swagger {
	b, err := ioutil.ReadFile(g.Src)
	if err != nil {
		panic(err)
	}

	swagger := &schema.Swagger{}

	if err := json.Unmarshal(b, swagger); err != nil {
		panic(err)
	}

	return swagger
}

func (g *Generator) getContexts(swagger *schema.Swagger) []Context {
	contexts := make([]Context, len(swagger.Definitions))

	index := 0
	for name, def := range swagger.Definitions {
		context := Context{}
		context.Name = name

		if len(def.Properties) > 0 {
			context.Properties = make([]ContextProperty, len(def.Properties))
			indexProp := 0
			for propName, prop := range def.Properties {
				context.Properties[indexProp] = ContextProperty{}
				context.Properties[indexProp].Name = propName
				if prop.Type != "" {
					context.Properties[indexProp].Type = prop.Type
				} else if prop.Ref != "" {
					context.Properties[indexProp].Type = removeDefinitionRef(prop.Ref)
				}
				indexProp++
			}
		}

		contexts[index] = context
		index++
	}

	return contexts
}

func (g *Generator) parseContext(contexts []Context) {
	var err error
	deftpl := g.ClientTemplate.DefaultContextTpl()
	contextTpl := template.New("context")
	contextTpl, err = contextTpl.Parse(string(deftpl))
	if err != nil {
		panic(err)
	}

	buf := bytes.NewBuffer(nil)
	if err = contextTpl.Execute(buf, contexts); err != nil {
		panic(err)
	}

	pretty.Println(buf.String())
}
