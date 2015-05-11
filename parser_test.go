package arlong

import (
	"github.com/kr/pretty"
	"testing"
)

// @DefinitionModel
type Witoo struct {
	// comment on E
	Name string
}

// @DefinitionModel
type Jack struct {
	// description on here
	// sdfsdf
	// sdfsdf
	E *Witoo `json:"jakc,ecom" swagger:"ebola1,required"`
	A Witoo  `swagger:"ebola2,required"`
	*Witoo
	test    int
	private int `swagger:"-"`
	mapping map[string]int
}

// @SWAGGER
// @TITLE Api
// @DESCRIPTION Super api
// @TERM Dont use
// @CONTACT name="witoo harianto" url=http://www.plimble.com email=witooh@gmail.com
// @LICENSE name="Apache 2.0" url=http://google.com
// @VERSION 1.1.1
// @SCHEMES http https ws
// @CONSUMES json xml
// @PRODUCES json xml
// @SECURITY petstore_auth=write:pets,read:pets
//
// @SECURITY_DEFINITION petstore_auth
// @TYPE oauth2
// @FLOW password
// @TOKEN_URL http://swagger.io/api/oauth/token
// @SCOPES write:pets="modify pets in your account" read:pets="read your pets"
//
// @GLOBAL_PARAM userParam name=user required description="sadsadsad" in=body schema.$ref=Witoo
// @GLOBAL_PARAM userParam2 name=user required description="sadsadsad" in=body schema.$ref=Jack
//
// @GLOBAL_RESPONSE notFound desc="Entity not found." schema.$ref=Witoo
// @GLOBAL_RESPONSE notFound2 desc="Entity not found." schema.$ref=Jack
//
// @PATH /user/jack/{id}
// @METHOD GET
// @PARAM name=id required description="sadsadsad" in=path type=string
// @PARAM name=user required description="sadsadsad" in=body schema.$ref=Jack
// @PRODUCES json
// @CONSUMES json
// @SUMMARY this is summary
// @DESCRIPTION this is description
// @DEPRECATED
// @SCHEMES http https
// @OPERATIONID GetStart
// @TAGS a b c
// @SECURITY petstore_auth=write:pets,read:pets
// @RESPONSE 200 desc=123123 schema.$ref=NotFound
func TestAnnotation(t *testing.T) {
	basePath := "/Users/witooh/dev/go/src/github.com/plimble/arlong"
	parser := NewParser(basePath)
	parser.Parse()
	b, _ := parser.JSON()
	pretty.Println(string(b))
}
