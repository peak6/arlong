Arlong [![godoc badge](http://godoc.org/github.com/plimble/arlong?status.png)](http://godoc.org/github.com/plimble/arlong)
========

Swagger 2.0 Generator

##Install
```
go get -u github.com/plimble/arlong/...
```

##Example
```go
// @DefinitionModel
type Hello struct {
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
// @Contact name="John Doe" url=http://www.company.com email=johndoe@company.com
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
// @GlobalParam userParam name=user required description="sadsadsad" in=body schema.$ref=package.hello
// @GlobalParam userParam2 name=user required description="sadsadsad" in=body schema.$ref=package.Data
//
// @GlobalResponse notFound desc="Entity not found." schema.$ref=package.hello
// @GlobalResponse notFound2 desc="Entity not found." schema.$ref=package.Data
//
// @Path /user/package.Data/{id}
// @Method GET
// @Param name=id required description="sadsadsad" in=path type=string
// @Param name=user required description="sadsadsad" in=body schema.$ref=package.Data
// @Produces json
// @Consumes json
// @Summary this is summary
// @Description this is description
// @Deprecated
// @Schemes http https
// @OperationId GetStart
// @Tags a b c
// @Security petstore_auth=write:pets,read:pets
// @Response 200 desc=123123 schema.$ref=package.NotFound
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

##CLI
```shell
NAME:
   arlong - Genrate Swagger 2.0

USAGE:
   arlong [global options] command [command options] [arguments...]

VERSION:
   1.0.0

AUTHOR(S):

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --path, -p "."   Package path to generate
   --out, -o "."    Output Path
   --file, -f "swagger.json"  Output file name
   --help, -h     show help
   --version, -v    print the version
```

##Todo
 - More format
 - Generate Restful Go Client
 - Unit test
 - Compatible all swagger 2.0 spec
 - Document
 - Validate spec

##Contributing
If you'd like to help out with the project. You can put up a Pull Request.


