package arlong

import (
	"github.com/kr/pretty"
	"testing"
)

// @DefinitionModel
// @Description abcde
type Hello1 string

// @DefinitionModel
// @Description 1234
type Hello2 int

// @DefinitionModel
type Hello3 float64

// @DefinitionModel
type Hello4 []string

// @DefinitionModel
type Hello5 map[string]int

// @DefinitionModel
type Hello6 Hello8

// @DefinitionModel
type Hello7 *Hello8

// @DefinitionModel
type Hello8 struct {
	// @Description comment e
	// @Required
	Name string
}

// @DefinitionModel
type Hello9 struct {
	// @Name ebola1
	// @Description ssssss
	// @Required
	E *Hello8

	// @Required
	A Hello8

	*Hello8
	test int

	// @Name -
	private int

	// @Required
	mapping map[string]int
}

// @Swagger
// @Title Api
// @Description Super api
// @Term Dont use
// @Contact name="witoo harianto" url=http://www.plimble.com email=witooh@gmail.com
// @License name="Apache 2.0" url=http://google.com
// @Version 1.1.1
// @Schemes http https ws
// @Consumes json xml
// @Produces json xml
// @Security petstore_auth=write:pets,read:pets
//
// @SecurityDefinition petstore_auth
// @Type oauth2
// @Flow password
// @TokenUrl http://swagger.io/api/oauth/token
// @Scopes write:pets="modify pets in your account" read:pets="read your pets"
//
// @GlobalParam 	userParam		 name=user		required description="sadsadsad"		in=body schema.$ref=Witoo
// @GlobalParam 	userParam2		 name=user		required description="sadsadsad"		in=body schema.$ref=Jack
//
// @GlobalResponse notFound desc="Entity not found." schema.$ref=Witoo
// @GlobalResponse notFound2 desc="Entity not found." schema.$ref=Jack
//
// @Path /user/jack/{id}
// @Method GET
// @Param name=id required description="sadsadsad" in=path type=string
// @Param name=user required description="sadsadsad" in=body schema.$ref=Jack
// @Produces json
// @Consumes json
// @Summary this is summary
// @Description this is description
// @Deprecated
// @Schemes http https
// @OperationId GetStart
// @Tags a b c
// @Security petstore_auth=write:pets,read:pets
// @Response 200 desc=123123 schema.$ref=NotFound
func TestAnnotation(t *testing.T) {
	basePath := "/Users/witooh/dev/go/src/github.com/plimble/arlong"
	parser := NewParser(basePath)
	b, _ := parser.JSON()
	pretty.Println(string(b))
	pretty.Println(swagger.Parameters)
}
