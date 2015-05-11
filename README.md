Arlong [![godoc badge](http://godoc.org/github.com/plimble/arlong?status.png)](http://godoc.org/github.com/plimble/arlong)
========

Swagger 2.0 Generator

##Example
```go
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
// @NAME abc
// @TYPE oauth2
// @DESCRIPTION oauth2 security
// @IN header
// @FLOW password
// @AUTHORIZATION_URL http://swagger.io/api/oauth/dialog
// @TOKEN_URL http://swagger.io/api/oauth/token
// @SCOPES write:pets="modify pets in your account" read:pets="read your pets"
//
// @GLOBAL_PARAM userParam name=user required description="sadsadsad" in=body schema.$ref=Parameter
// @GLOBAL_PARAM userParam2 name=user required description="sadsadsad" in=body schema.$ref=Parameter
//
// @GLOBAL_RESPONSE notFound desc="Entity not found." schema.$ref=GeneralError
// @GLOBAL_RESPONSE notFound2 desc="Entity not found." schema.$ref=GeneralError
//
// @PATH /user/jack/{id}
// @METHOD GET
// @PARAM name=user required description="sadsadsad" in=body schema.$ref=Parameter
// @PARAM name=id type=array items.$ref=#/parameters/skip
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
func main(){

}
```

##API
```go
func main(){
  a := arlong.NewParser("~/go/src/path/to/package")
  b, err := a.JSON() //generate swagger 2.0 json format
}
```

##Todo
 - CLI
 - More format
 - Generate Restful Go Client
 - Unit test
 - Compatible all swagger 2.0 spec
 - Document
 - Validate spec

##Contributing
If you'd like to help out with the project. You can put up a Pull Request.


