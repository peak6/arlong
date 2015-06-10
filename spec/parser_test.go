package spec

import (
	"github.com/kr/pretty"
	"testing"
)

// @DefinitionModel
type Hello6 map[string]Hello8

// @DefinitionModel
type Hello7 Hello8

// @DefinitionModel
type Hello8 struct {
	// @Description comment e
	// @Required
	Name string
}

// @DefinitionModel
type Hello9 struct {
	c Hello8
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
// @GlobalParam 	userParam		 name=user		required description="sadsadsad"		in=body schema.$ref=arlong.Hello9
// @GlobalParam 	userParam2		 name=user		required description="sadsadsad"		in=body schema.$ref=arlong.Hello9
//
// @GlobalResponse notFound desc="Entity not found." schema.$ref=arlong.Hello9
// @GlobalResponse notFound2 desc="Entity not found." schema.$ref=arlong.Hello9
//
// @Path /attempts
// @Method GET
// @Description Get Array of attempts
// @OperationId GetAttempts
// @Param $ref=limitQuery
// @Param $ref=skipQuery
// @Param $ref=spokenQuery
// @Param $ref=practiseQuery
// @Param $ref=userLangQuery
// @Tags attempts
// @Response 200 schema.type=array schema.items.$ref=arlong.Hello9
//
// @Path /user/jack/{id}
// @Method GET
// @Param name=id required description="sadsadsad" in=path type=string
// @Param name=user required description="sadsadsad" in=body schema.$ref=arlong.Hello9
// @Produces json
// @Consumes json
// @Summary this is summary
// @Description this is description
// @Deprecated
// @Schemes http https
// @OperationId GetStart
// @Tags a b c
// @Security petstore_auth=write:pets,read:pets
// @Response 200 desc=123123 schema.$ref=arlong.Hello9
func TestAnnotation(t *testing.T) {
	basePath := "/Users/witooh/dev/go/src/github.com/plimble/arlong"
	parser := NewParser(basePath)
	b, _ := parser.JSON()
	pretty.Println(string(b))
	pretty.Println(swagger.Definitions)
}
