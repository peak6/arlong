package arlong

import (
	"encoding/json"
)

const (
	QUERY    = "query"
	FORMDATA = "formData"
	HEADER   = "header"
	PATH     = "path"
)

const (
	INT32    = "int32"
	INT64    = "int64"
	STRING   = "string"
	FLOAT    = "float"
	BOOL     = "bool"
	BYTE     = "byte"
	DATETIME = "datetime"
	DATE     = "date"
	PASSWORD = "password"
)

const (
	MIME_XML       = "application/xml"
	MIME_JSON      = "application/json"
	MIME_HTML      = "text/html"
	MIME_TEXT      = "text/plain"
	MIME_FORM      = "application/x-www-form-urlencoded"
	MIME_MULTIPART = "multipart/form-data"
)

var swagger *Swagger

func init() {
	newSwagger()
}

func newSwagger() {
	swagger = &Swagger{
		Swagger:             "2.0",
		Paths:               make(map[string]*Path),
		Definitions:         make(map[string]*Schema),
		Security:            []map[string][]string{},
		SecurityDefinitions: make(map[string]*SecurityDefinitions),
		Parameters:          make(map[string]*Parameter),
		Responses:           make(map[string]*Responses),
	}
}

type Swagger struct {
	Swagger             string                          `json:"swagger,omitempty"`
	Info                Info                            `json:"info"`
	Host                string                          `json:"host,omitempty"`
	BasePath            string                          `json:"basePath,omitempty"`
	Schemes             []string                        `json:"schemes,omitempty"`
	Consumes            []string                        `json:"consumes,omitempty"`
	Produces            []string                        `json:"produces,omitempty"`
	Paths               map[string]*Path                `json:"paths"`
	Definitions         map[string]*Schema              `json:"definitions,omitempty"`
	Security            []map[string][]string           `json:"security,omitempty"`
	SecurityDefinitions map[string]*SecurityDefinitions `json:"securityDefinitions,omitempty"`
	Parameters          map[string]*Parameter           `json:"parameters,omitempty"`
	Responses           map[string]*Responses           `json:"responses,omitempty"`
}

type Info struct {
	Title          string   `json:"title"`
	Description    string   `json:"description,omitempty"`
	TermsOfService string   `json:"termsOfService,omitempty"`
	Contact        *Contact `json:"contact,omitempty"`
	License        *License `json:"license,omitempty"`
	Version        string   `json:"version"`
}

type Contact struct {
	Name  string `json:"name,omitempty"`
	URL   string `json:"url,omitempty"`
	Email string `json:"email,omitempty"`
}

type License struct {
	Name string `json:"name,omitempty"`
	URL  string `json:"url,omitempty"`
}

type Path struct {
	Route      string      `json:"-"`
	Ref        string      `json:"$ref,omitempty"`
	GET        *Operation  `json:"get,omitempty"`
	PUT        *Operation  `json:"put,omitempty"`
	POST       *Operation  `json:"post,omitempty"`
	DELETE     *Operation  `json:"delete,omitempty"`
	OPTIONS    *Operation  `json:"options,omitempty"`
	HEAD       *Operation  `json:"head,omitempty"`
	PATCH      *Operation  `json:"patch,omitempty"`
	Parameters []Parameter `json:"parameters,omitempty"`
}

type Operation struct {
	Tags        []string              `json:"tags,omitempty"`
	Summary     string                `json:"summary,omitempty"`
	Description string                `json:"description,omitempty"`
	OperationId string                `json:"operationId,omitempty"`
	Consumes    []string              `json:"consumes,omitempty"`
	Produces    []string              `json:"produces,omitempty"`
	Parameters  []*Parameter          `json:"parameters,omitempty"`
	Responses   map[string]*Responses `json:"responses,omitempty"`
	Schemes     []string              `json:"schemes,omitempty"`
	Deprecated  bool                  `json:"deprecated,omitempty"`
	Security    []map[string][]string `json:"security,omitempty"`
}

type Parameter struct {
	Ref             string      `json:"$ref,omitempty"`
	Name            string      `json:"name,omitempty"`
	In              string      `json:"in,omitempty"`
	Description     string      `json:"description,omitempty"`
	Required        bool        `json:"required,omitempty"`
	Schema          *Schema     `json:"schema,omitempty"`
	Type            string      `json:"type,omitempty"`
	Format          string      `json:"format,omitempty"`
	AllowEmptyValue bool        `json:"allowEmptyValue,omitempty"`
	Items           *Items      `json:"items,omitempty"`
	Default         interface{} `json:"default,omitempty"`
	Maximum         int         `json:"maximum,omitempty"`
	Minimum         int         `json:"minimum,omitempty"`
	MaxLength       int         `json:"maxLength,omitempty"`
	MinLength       int         `json:"minLength,omitempty"`
	MaxItems        int         `json:"maxItems,omitempty"`
	MinItems        int         `json:"minItems,omitempty"`
}

type Schema struct {
	AllOf                []*Schema          `json:"allOf,omitempty"`
	Properties           map[string]*Schema `json:"properties,omitempty"`
	Required             []string           `json:"required,omitempty"`
	Type                 string             `json:"type,omitempty"`
	Format               string             `json:"format,omitempty"`
	Description          string             `json:"description,omitempty"`
	Items                *Schema            `json:"items,omitempty"`
	Ref                  string             `json:"$ref,omitempty"`
	AdditionalProperties *Schema            `json:"additionalProperties,omitempty"`
	rawRefName           string             `json:"-"`
}

type Items struct {
	Type      string      `json:"type,omitempty"`
	Format    string      `json:"format,omitempty"`
	Default   interface{} `json:"default,omitempty"`
	Maximum   int         `json:"maximum,omitempty"`
	Minimum   int         `json:"minimum,omitempty"`
	MaxLength int         `json:"maxLength,omitempty"`
	MinLength int         `json:"minLength,omitempty"`
	MaxItems  int         `json:"maxItems,omitempty"`
	MinItems  int         `json:"minItems,omitempty"`
}

type Responses struct {
	Ref         string             `json:"$ref,omitempty"`
	Description string             `json:"description"`
	Schema      *Schema            `json:"schema,omitempty"`
	Headers     map[string]*Header `json:"headers,omitempty"`
}

type Field struct {
	Type                 string  `json:"type,omitempty"`
	Description          string  `json:"description,omitempty"`
	Format               string  `json:"format,omitempty"`
	Items                *Items  `json:"items,omitempty"`
	Ref                  string  `json:"$ref,omitempty"`
	AdditionalProperties *Schema `json:"additionalProperties,omitempty"`
}

type SecurityDefinitions struct {
	Type             string            `json:"type,omitempty"`
	Description      string            `json:"description,omitempty"`
	Name             string            `json:"name,omitempty"`
	In               string            `json:"in,omitempty"`
	Flow             string            `json:"flow,omitempty"`
	AuthorizationUrl string            `json:"authorizationUrl,omitempty"`
	TokenUrl         string            `json:"tokenUrl,omitempty"`
	Scopes           map[string]string `json:"scopes,omitempty"`
}

type Header struct {
	Description string `json:"description,omitempty"`
	Type        string `json:"type,omitempty"`
	Format      string `json:"format,omitempty"`
	Items       *Items `json:"items,omitempty"`
	Default     string `json:"default,omitempty"`
	Maximum     int    `json:"maximum,omitempty"`
	Minimum     int    `json:"minimum,omitempty"`
	MaxLength   int    `json:"maxLength,omitempty"`
	MinLength   int    `json:"minLength,omitempty"`
	MaxItems    int    `json:"maxItems,omitempty"`
	MinItems    int    `json:"minItems,omitempty"`
}

func jsonFormat() ([]byte, error) {
	return json.Marshal(swagger)
}
