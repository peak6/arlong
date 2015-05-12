package arlong

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

type Parser struct {
	basePkgPath string
	json        []byte
}

func NewParser(basePkgPath string) *Parser {
	return &Parser{
		basePkgPath: basePkgPath,
	}
}

func (a *Parser) Parse() error {
	newSwagger()
	a.json = nil

	packages, err := parsePackages(a.basePkgPath)
	if err != nil {
		return err
	}

	parseComments(packages)
	parseDefinitionModel(packages)

	return nil
}

func (a *Parser) JSON() ([]byte, error) {
	if a.json == nil {
		if err := a.Parse(); err != nil {
			return nil, err
		}

		result, err := jsonFormat()
		if err != nil {
			return nil, err
		}

		a.json = result
	}

	return a.json, nil
}

func parsePackages(dir string) (map[string]*ast.Package, error) {
	fset := token.NewFileSet() // positions are relative to fset

	return parser.ParseDir(fset, dir, nil, parser.ParseComments)
}

func parseComments(packages map[string]*ast.Package) {
	for _, p := range packages {
		for _, f := range p.Files {
			for i := 0; i < len(f.Comments); i++ {
				readZone(f.Comments[i].List)
			}
		}
	}
}

func parseDefinitionModel(packages map[string]*ast.Package) {
	for _, astPackage := range packages {
		for _, astFile := range astPackage.Files {
			for _, astDeclaration := range astFile.Decls {
				if generalDeclaration, ok := astDeclaration.(*ast.GenDecl); ok && generalDeclaration.Tok == token.TYPE {
					if generalDeclaration.Doc != nil {
						for i := 0; i < len(generalDeclaration.Doc.List); i++ {
							if strings.TrimSpace(strings.TrimPrefix(generalDeclaration.Doc.List[i].Text, "//")) == "@DefinitionModel" {
								for _, astSpec := range generalDeclaration.Specs {
									if typeSpec, ok := astSpec.(*ast.TypeSpec); ok {
										parseDefinition(generalDeclaration.Doc.List, typeSpec)
									}
								}
							}
						}
					}
				}
			}
		}
	}
}

func readZone(comments []*ast.Comment) {
	if len(comments) == 0 {
		return
	}

	for i := 0; i < len(comments); i++ {
		index := findAt(comments[i].Text)

		if index > 0 {
			tag := comments[i].Text[index:]
			header := strings.Split(tag, " ")

			switch header[0] {
			case "@SWAGGER":
				i += parseSwagger(comments[i:])
			case "@GLOBAL_PARAM":
				parseParamGlobal(comments[i : i+1][0])
			case "@DEFINITIONS":
			case "@SECURITY_DEFINITION":
				i += parseSecurityDefinition(comments[i:])
			case "@GLOBAL_RESPONSE":
				parseGlobalResponse(comments[i : i+1][0])
			case "@PATH":
				i += parsePath(comments[i:])
			}
		}
	}
}

func parseSwagger(comments []*ast.Comment) int {
	i := 0
	for ; i < len(comments); i++ {
		if strings.TrimSpace(comments[i].Text) == "//" {
			return i
		}

		index := findAt(comments[i].Text)
		if index > 0 {
			tag, vals := getValues(comments[i].Text[index:])
			switch tag {
			case "@TITLE":
				swagger.Info.Title = vals
			case "@DESCRIPTION":
				swagger.Info.Description = vals
			case "@TERM":
				swagger.Info.TermsOfService = vals
			case "@CONTACT":
				if swagger.Info.Contact == nil {
					swagger.Info.Contact = &Contact{}
				}
				data := getValueByKey(vals)
				swagger.Info.Contact.Name = data["name"]
				swagger.Info.Contact.Email = data["email"]
				swagger.Info.Contact.URL = data["url"]
			case "@LICENSE":
				if swagger.Info.License == nil {
					swagger.Info.License = &License{}
				}
				data := getValueByKey(vals)
				swagger.Info.License.Name = data["name"]
				swagger.Info.License.URL = data["url"]
			case "@VERSION":
				swagger.Info.Version = vals
			case "@SCHEMES":
				swagger.Schemes = getValueStrings(vals)
			case "@CONSUMES":
				valsArray := getValueStrings(vals)
				for i := 0; i < len(valsArray); i++ {
					valsArray[i] = getMime(valsArray[i])
				}
				swagger.Consumes = valsArray
			case "@PRODUCES":
				valsArray := getValueStrings(vals)
				for i := 0; i < len(valsArray); i++ {
					valsArray[i] = getMime(valsArray[i])
				}
				swagger.Produces = valsArray
			case "@SECURITY":
				swagger.Security = append(swagger.Security, getValueMapStrings(vals))
			}
		}
	}

	return i
}

func parseGlobalResponse(comment *ast.Comment) {
	index := findAt(comment.Text)
	if index > 0 {
		data := strings.SplitN(strings.TrimSpace(comment.Text[index:]), " ", 3)
		if len(data) != 3 {
			panic("Invalid @GLOBAL_RESPONSE arguments")
		}

		tag, respName, vals := data[0], data[1], data[2]

		if tag == "@GLOBAL_RESPONSE" {
			resp := &Responses{}
			parseResponse(resp, getValueByKey(vals))
			swagger.Responses[respName] = resp
		}
	}
}

func parseParamGlobal(comment *ast.Comment) {
	index := findAt(comment.Text)
	if index > 0 {
		data := strings.SplitN(strings.TrimSpace(comment.Text[index:]), " ", 3)
		if len(data) != 3 {
			panic("Invalid @GLOBAL_PARAM arguments")
		}

		tag, paramName, vals := data[0], data[1], data[2]

		if tag == "@GLOBAL_PARAM" {
			param := &Parameter{}
			parseParam(param, getValueByKey(vals))
			swagger.Parameters[paramName] = param
		}
	}
}

func parseSecurityDefinition(comments []*ast.Comment) int {
	i := 0
	var def *SecurityDefinitions
	for ; i < len(comments); i++ {
		if strings.TrimSpace(comments[i].Text) == "//" {
			return i
		}

		index := findAt(comments[i].Text)
		if index > 0 {
			tag, vals := getValues(comments[i].Text[index:])
			switch tag {
			case "@SECURITY_DEFINITION":
				def = &SecurityDefinitions{}
				swagger.SecurityDefinitions[vals] = def
			case "@NAME":
				def.Name = vals
			case "@TYPE":
				def.Type = vals
			case "@DESCRIPTION":
				def.Description = vals
			case "@IN":
				def.In = vals
			case "@FLOW":
				def.Flow = vals
			case "@AUTHORIZATION_URL":
				def.AuthorizationUrl = vals
			case "@TOKEN_URL":
				def.TokenUrl = vals
			case "@SCOPES":
				def.Scopes = getValueByKey(vals)
			}
		}
	}

	return i
}

func parsePath(comments []*ast.Comment) int {
	i := 0
	var method *Operation
	path := ""
	for ; i < len(comments); i++ {
		if strings.TrimSpace(comments[i].Text) == "//" {
			return i
		}

		index := findAt(comments[i].Text)
		if index > 0 {
			tag, vals := getValues(comments[i].Text[index:])
			switch tag {
			case "@PATH":
				path = vals
				swagger.Paths[vals] = &Path{}
			case "@METHOD":
				switch vals {
				case "GET":
					method = &Operation{}
					swagger.Paths[path].GET = method
				case "POST":
					method = &Operation{}
					swagger.Paths[path].POST = method
				case "PUT":
					method = &Operation{}
					swagger.Paths[path].PUT = method
				case "DELETE":
					method = &Operation{}
					swagger.Paths[path].DELETE = method
				case "OPTIONS":
					method = &Operation{}
					swagger.Paths[path].OPTIONS = method
				case "HEAD":
					method = &Operation{}
					swagger.Paths[path].HEAD = method
				default:
					panic("Unsupported method")
				}
			case "@CONSUMES":
				valsArray := getValueStrings(vals)
				for i := 0; i < len(valsArray); i++ {
					valsArray[i] = getMime(valsArray[i])
				}
				method.Consumes = valsArray
			case "@PRODUCES":
				valsArray := getValueStrings(vals)
				for i := 0; i < len(valsArray); i++ {
					valsArray[i] = getMime(valsArray[i])
				}
				method.Produces = valsArray
			case "@SUMMARY":
				method.Summary = vals
			case "@DESCRIPTION":
				method.Description = vals
			case "@DEPRECATED":
				method.Deprecated = true
			case "@SCHEMES":
				method.Schemes = getValueStrings(vals)
			case "@OPERATIONID":
				method.OperationId = vals
			case "@SECURITY":
				method.Security = append(method.Security, getValueMapStrings(vals))
			case "@TAGS":
				method.Tags = getValueStrings(vals)
			case "@PARAM":
				if method.Parameters == nil {
					method.Parameters = []*Parameter{}
				}
				param := &Parameter{}
				valArray := getValueByKey(vals)
				parseParam(param, valArray)
				method.Parameters = append(method.Parameters, param)
			case "@RESPONSE":
				if method.Responses == nil {
					method.Responses = make(map[string]*Responses)
				}

				data := strings.SplitN(vals, " ", 2)
				if len(data) != 2 {
					panic("Invalid @RESPONSE arguments")
				}

				code, vals := data[0], data[1]

				resp := &Responses{}
				valArray := getValueByKey(vals)
				parseResponse(resp, valArray)
				method.Responses[code] = resp
			}
		}
	}

	return i
}

func parseParam(param *Parameter, vals map[string]string) {
	for key, val := range vals {
		switch {
		case key == "name":
			param.Name = val
		case key == "$ref":
			param.Ref = "#/parameters/" + val
		case key == "in":
			param.In = val
		case key == "description" || key == "desc":
			param.Description = val
		case key == "required":
			param.Required = true
		case pathMatch("schema.*", key):
			if param.Schema == nil {
				param.Schema = &Schema{}
			}
			parseSchema(param.Schema, strings.TrimPrefix(key, "schema."), val)
		case key == "type":
			param.Type, param.Format = getTypeFormat(val)
		case key == "allowEmptyValue":
			param.AllowEmptyValue = true
		case pathMatch("items.*", key):
			if param.Items == nil {
				param.Items = &Items{}
			}
			parseItem(param.Items, strings.TrimPrefix(key, "items."), val)
		case key == "default":
			param.Default = val
		case key == "maximum":
			param.Maximum = strToInt(val)
		case key == "minimum":
			param.Minimum = strToInt(val)
		case key == "maxLength":
			param.MaxLength = strToInt(val)
		case key == "minLength":
			param.MinLength = strToInt(val)
		case key == "maxItems":
			param.MaxItems = strToInt(val)
		case key == "minItems":
			param.MinItems = strToInt(val)
		}
	}
}

func parseSchema(s *Schema, key, val string) {
	switch {
	case key == "type":
		s.Type, s.Format = getTypeFormat(val)
	case key == "$ref":
		s.Ref = "#/definitions/" + val
	case pathMatch("items.*", key):
		if s.Items == nil {
			s.Items = &Items{}
		}
		parseItem(s.Items, strings.TrimPrefix(key, "items."), val)
	}
}

func parseItem(item *Items, key, val string) {
	switch {
	case key == "$ref":
		item.Ref = "#/definitions/" + val
	case key == "type":
		item.Type, item.Format = getTypeFormat(val)
	case key == "default":
		item.Default = val
	case key == "maximum":
		item.Maximum = strToInt(val)
	case key == "minimum":
		item.Minimum = strToInt(val)
	case key == "maxLength":
		item.MaxLength = strToInt(val)
	case key == "minLength":
		item.MinLength = strToInt(val)
	case key == "maxItems":
		item.MaxItems = strToInt(val)
	case key == "minItems":
		item.MinItems = strToInt(val)
	}
}

func parseResponse(resp *Responses, vals map[string]string) {
	for key, val := range vals {
		switch {
		case key == "$ref":
			resp.Ref = "#/responses/" + val
		case key == "description" || key == "desc":
			resp.Description = val
		case pathMatch("schema.*", key):
			if resp.Schema == nil {
				resp.Schema = &Schema{}
			}
			parseSchema(resp.Schema, strings.TrimPrefix(key, "schema."), val)
		}
	}
}

func parseDefinitionOptions(def *Definition, comments []*ast.Comment) {
	i := 0
	for ; i < len(comments); i++ {
		if strings.TrimSpace(comments[i].Text) == "//" {
			return
		}

		index := findAt(comments[i].Text)
		if index > 0 {
			tag, vals := getValues(comments[i].Text[index:])
			switch tag {
			case "@DESCRIPTION":
				def.Description = vals
			}
		}
	}
}

func parsePropertiesName(comments []*ast.Comment) string {
	i := 0
	for ; i < len(comments); i++ {
		if strings.TrimSpace(comments[i].Text) == "//" {
			return ""
		}

		index := findAt(comments[i].Text)
		if index > 0 {
			tag, vals := getValues(comments[i].Text[index:])
			switch tag {
			case "@NAME":
				return vals
			}
		}
	}

	return ""
}

func parsePropertiesOptions(name string, def *Definition, prop *Definition, comments []*ast.Comment) {
	i := 0
	for ; i < len(comments); i++ {
		if strings.TrimSpace(comments[i].Text) == "//" {
			return
		}

		index := findAt(comments[i].Text)
		if index > 0 {
			tag, vals := getValues(comments[i].Text[index:])
			switch tag {
			case "@DESCRIPTION":
				prop.Description = vals
			case "@REQUIRED":
				def.Required = append(def.Required, name)
			}
		}
	}
}

func parseDefinition(comments []*ast.Comment, astTypeSpec *ast.TypeSpec) {
	name := astTypeSpec.Name.String()
	if _, ok := swagger.Definitions[name]; ok {
		return
	}

	def := &Definition{}
	parseDefinitionOptions(def, comments)
	swagger.Definitions[name] = def

	switch astType := astTypeSpec.Type.(type) {
	default:
		parseProperties(def, astType)
	case *ast.StructType:
		def.Type = "object"
		def.Properties = make(map[string]*Definition)

		for i := 0; i < astType.Fields.NumFields(); i++ {
			propName := ""
			astField := astType.Fields.List[i]

			if astField.Doc != nil {
				propName = parsePropertiesName(astField.Doc.List)
			}

			if propName == "-" {
				continue
			}

			if propName == "" {
				if len(astField.Names) == 0 {
					if astSelectorExpr, ok := astField.Type.(*ast.SelectorExpr); ok {
						propName = strings.TrimPrefix(astSelectorExpr.Sel.Name, "*")
					} else if astTypeIdent, ok := astField.Type.(*ast.Ident); ok {
						propName = astTypeIdent.Name
					} else if astStarExpr, ok := astField.Type.(*ast.StarExpr); ok {
						if astIdent, ok := astStarExpr.X.(*ast.Ident); ok {
							propName = astIdent.Name
						}
					} else {
						panic(fmt.Errorf("Something goes wrong: %#v", astField.Type))
					}
				} else {
					propName = astField.Names[0].String()
				}
			}

			field := &Definition{}

			if astField.Doc != nil {
				parsePropertiesOptions(propName, def, field, astField.Doc.List)
			}

			parseProperties(field, astField.Type)
			def.Properties[propName] = field
		}
	}
}

func parseProperties(def *Definition, astType ast.Expr) {
	switch fieldType := astType.(type) {
	case *ast.MapType:
		def.Type = "object"
		if def.AdditionalProperties == nil {
			def.AdditionalProperties = &Schema{}
		}
		switch mapType := fieldType.Value.(type) {
		case *ast.InterfaceType:
			def.AdditionalProperties.Type = "any"
		case *ast.Ident:
			def.AdditionalProperties.Type, def.AdditionalProperties.Format = getTypeFormat(mapType.String())
			if def.AdditionalProperties.Type == "unknown" {
				def.AdditionalProperties.Type = ""
				def.AdditionalProperties.Format = ""
				def.AdditionalProperties.Ref = "#/definitions/" + mapType.String()
			}
		}
	case *ast.ArrayType:
		def.Type = "array"
		def.Items = &Items{}
		switch arrayType := fieldType.Elt.(type) {
		case *ast.InterfaceType:
			def.Items.Type = "any"
		case *ast.Ident:
			def.Items.Type, def.Items.Format = getTypeFormat(arrayType.String())
		}
	case *ast.StarExpr:
		def.Type, def.Format = getTypeFormat(fmt.Sprint(fieldType.X))
		if def.Type == "unknown" {
			def.Type = ""
			def.Format = ""
			def.Ref = "#/definitions/" + fmt.Sprint(fieldType.X)
		}
	case *ast.SelectorExpr:
		def.Type, def.Format = getTypeFormat(fieldType.Sel.Name)
		if def.Type == "unknown" {
			def.Type = ""
			def.Format = ""
			def.Ref = "#/definitions/" + fieldType.Sel.Name
		}
	case *ast.Ident:
		def.Type, def.Format = getTypeFormat(fieldType.String())
		if def.Type == "unknown" {
			def.Type = ""
			def.Format = ""
			def.Ref = "#/definitions/" + fieldType.String()
		}
	}
}
