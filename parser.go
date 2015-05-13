package arlong

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

type Parser struct {
	packages        []*ast.Package
	usedDefinitions []string
	usedParameters  []string
	usedResponses   []string
	basePkgPath     string
	json            []byte
}

func NewParser(basePkgPath string) *Parser {
	return &Parser{
		packages:    []*ast.Package{},
		basePkgPath: basePkgPath,
	}
}

func (p *Parser) Parse() error {
	newSwagger()
	p.usedDefinitions = []string{}
	p.usedParameters = []string{}
	p.usedResponses = []string{}
	p.json = nil

	if err := p.parsePackages(); err != nil {
		return err
	}

	p.parseComments()
	p.parseDefinitionModel()
	p.mergeCompositeDefinition()
	p.validate()

	return nil
}

func (p *Parser) JSON() ([]byte, error) {
	if p.json == nil {
		if err := p.Parse(); err != nil {
			return nil, err
		}

		result, err := jsonFormat()
		if err != nil {
			return nil, err
		}

		p.json = result
	}

	return p.json, nil
}

func (p *Parser) parsePackages() error {
	return filepath.Walk(p.basePkgPath, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			fset := token.NewFileSet()
			packages, err := parser.ParseDir(fset, path, nil, parser.ParseComments)
			if err != nil {
				return err
			}

			for _, pack := range packages {
				p.packages = append(p.packages, pack)
			}
		}

		return nil
	})
}

func (p *Parser) parseComments() {
	for _, pack := range p.packages {
		for _, f := range pack.Files {
			for i := 0; i < len(f.Comments); i++ {
				p.readZone(f.Comments[i].List)
			}
		}
	}
}

func (p *Parser) parseDefinitionModel() {
	for _, astPackage := range p.packages {
		for _, astFile := range astPackage.Files {
			for _, astDeclaration := range astFile.Decls {
				if generalDeclaration, ok := astDeclaration.(*ast.GenDecl); ok && generalDeclaration.Tok == token.TYPE {
					if generalDeclaration.Doc != nil {
						for i := 0; i < len(generalDeclaration.Doc.List); i++ {
							if strings.TrimSpace(strings.TrimPrefix(generalDeclaration.Doc.List[i].Text, "//")) == "@DefinitionModel" {
								for _, astSpec := range generalDeclaration.Specs {
									if typeSpec, ok := astSpec.(*ast.TypeSpec); ok {
										p.parseDefinition(astPackage.Name, generalDeclaration.Doc.List, typeSpec)
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

func (p *Parser) readZone(comments []*ast.Comment) {
	if len(comments) == 0 {
		return
	}

	for i := 0; i < len(comments); i++ {
		index := findAt(comments[i].Text)

		if index > 0 {
			tag := comments[i].Text[index:]
			header := strings.Split(tag, " ")

			switch header[0] {
			case "@Swagger":
				i += p.parseSwagger(comments[i:])
			case "@GlobalParam":
				p.parseParamGlobal(comments[i : i+1][0])
			case "@SecurityDefinition":
				i += p.parseSecurityDefinition(comments[i:])
			case "@GlobalResponse":
				p.parseGlobalResponse(comments[i : i+1][0])
			case "@Path":
				i += p.parsePath(comments[i:])
			}
		}
	}
}

func (p *Parser) parseSwagger(comments []*ast.Comment) int {
	i := 0
	for ; i < len(comments); i++ {
		if strings.TrimSpace(comments[i].Text) == "//" {
			return i
		}

		index := findAt(comments[i].Text)
		if index > 0 {
			tag, vals := getValues(comments[i].Text[index:])
			switch tag {
			case "@Title":
				swagger.Info.Title = vals
			case "@Description":
				swagger.Info.Description = vals
			case "@BasePath":
				swagger.BasePath = vals
			case "@Term":
				swagger.Info.TermsOfService = vals
			case "@Contact":
				if swagger.Info.Contact == nil {
					swagger.Info.Contact = &Contact{}
				}
				data := getValueByKey(vals)
				swagger.Info.Contact.Name = data["name"]
				swagger.Info.Contact.Email = data["email"]
				swagger.Info.Contact.URL = data["url"]
			case "@License":
				if swagger.Info.License == nil {
					swagger.Info.License = &License{}
				}
				data := getValueByKey(vals)
				swagger.Info.License.Name = data["name"]
				swagger.Info.License.URL = data["url"]
			case "@Version":
				swagger.Info.Version = vals
			case "@Schemes":
				swagger.Schemes = getValueStrings(vals)
			case "@Consumes":
				valsArray := getValueStrings(vals)
				for i := 0; i < len(valsArray); i++ {
					valsArray[i] = getMime(valsArray[i])
				}
				swagger.Consumes = valsArray
			case "@Produces":
				valsArray := getValueStrings(vals)
				for i := 0; i < len(valsArray); i++ {
					valsArray[i] = getMime(valsArray[i])
				}
				swagger.Produces = valsArray
			case "@Security":
				swagger.Security = append(swagger.Security, getValueMapStrings(vals))
			}
		}
	}

	return i
}

func (p *Parser) parseGlobalResponse(comment *ast.Comment) {
	index := findAt(comment.Text)
	if index > 0 {
		commentText := strings.Replace(comment.Text[index:], "\t", " ", -1)
		data := strings.SplitN(commentText, " ", 3)
		if len(data) != 3 {
			panic("Invalid @GlobalResponse arguments")
		}

		tag, respName, vals := strings.TrimSpace(data[0]), strings.TrimSpace(data[1]), strings.TrimSpace(data[2])

		if tag == "@GlobalResponse" {
			resp := &Responses{}
			p.parseResponse(resp, getValueByKey(vals))
			swagger.Responses[respName] = resp
		}
	}
}

func (p *Parser) parseParamGlobal(comment *ast.Comment) {
	index := findAt(comment.Text)
	if index > 0 {
		commentText := strings.Replace(comment.Text[index:], "\t", " ", -1)
		data := strings.SplitN(commentText, " ", 3)
		if len(data) != 3 {
			panic("Invalid @GlobalParam arguments")
		}

		tag, paramName, vals := strings.TrimSpace(data[0]), strings.TrimSpace(data[1]), strings.TrimSpace(data[2])

		if tag == "@GlobalParam" {
			param := &Parameter{}
			p.parseParam(param, getValueByKey(vals))
			swagger.Parameters[paramName] = param
		}
	}
}

func (p *Parser) parseSecurityDefinition(comments []*ast.Comment) int {
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
			case "@SecurityDefinition":
				def = &SecurityDefinitions{}
				swagger.SecurityDefinitions[vals] = def
			case "@Name":
				def.Name = vals
			case "@Type":
				def.Type = vals
			case "@Description":
				def.Description = vals
			case "@In":
				def.In = vals
			case "@Flow":
				def.Flow = vals
			case "@AuthorizationUrl":
				def.AuthorizationUrl = vals
			case "@TokenUrl":
				def.TokenUrl = vals
			case "@Scopes":
				def.Scopes = getValueByKey(vals)
			}
		}
	}

	return i
}

func (p *Parser) parsePath(comments []*ast.Comment) int {
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
			case "@Path":
				path = vals
				if swagger.Paths[vals] == nil {
					swagger.Paths[vals] = &Path{}
				}
			case "@Method":
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
			case "@Consumes":
				valsArray := getValueStrings(vals)
				for i := 0; i < len(valsArray); i++ {
					valsArray[i] = getMime(valsArray[i])
				}
				method.Consumes = valsArray
			case "@Produces":
				valsArray := getValueStrings(vals)
				for i := 0; i < len(valsArray); i++ {
					valsArray[i] = getMime(valsArray[i])
				}
				method.Produces = valsArray
			case "@Summary":
				method.Summary = vals
			case "@Description":
				method.Description = vals
			case "@Deprecated":
				method.Deprecated = true
			case "@Schemes":
				method.Schemes = getValueStrings(vals)
			case "@OperationId":
				method.OperationId = vals
			case "@Security":
				method.Security = append(method.Security, getValueMapStrings(vals))
			case "@Tags":
				method.Tags = getValueStrings(vals)
			case "@Param":
				if method.Parameters == nil {
					method.Parameters = []*Parameter{}
				}
				param := &Parameter{}
				valArray := getValueByKey(vals)
				p.parseParam(param, valArray)
				method.Parameters = append(method.Parameters, param)
			case "@Response":
				if method.Responses == nil {
					method.Responses = make(map[string]*Responses)
				}

				data := strings.SplitN(vals, " ", 2)
				if len(data) < 1 {
					panic("Invalid @Response arguments")
				}

				vals = ""
				code := data[0]
				if len(data) == 2 {
					vals = data[1]
				}

				resp := &Responses{}
				valArray := getValueByKey(vals)
				p.parseResponse(resp, valArray)
				method.Responses[code] = resp
			}
		}
	}

	return i
}

func (p *Parser) parseParam(param *Parameter, vals map[string]string) {
	for key, val := range vals {
		switch {
		case key == "name":
			param.Name = val
		case key == "$ref":
			param.Ref = "#/parameters/" + val
			p.usedParameters = append(p.usedParameters, val)
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
			p.parseSchema(param.Schema, strings.TrimPrefix(key, "schema."), val)
		case key == "type":
			param.Type, param.Format, _ = getTypeFormat(val)
		case key == "allowEmptyValue":
			param.AllowEmptyValue = true
		case pathMatch("items.*", key):
			if param.Items == nil {
				param.Items = &Items{}
			}
			p.parseItem(param.Items, strings.TrimPrefix(key, "items."), val)
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

func (p *Parser) parseSchema(s *Schema, key, val string) {
	switch {
	case key == "type":
		s.Type, s.Format, _ = getTypeFormat(val)
	case key == "$ref":
		s.Ref = "#/definitions/" + val
		p.usedDefinitions = append(p.usedDefinitions, val)
	case pathMatch("items.*", key):
		if s.Items == nil {
			s.Items = &Items{}
		}
		p.parseItem(s.Items, strings.TrimPrefix(key, "items."), val)
	}
}

func (p *Parser) parseItem(item *Items, key, val string) {
	switch {
	case key == "$ref":
		item.Ref = "#/definitions/" + val
		p.usedDefinitions = append(p.usedDefinitions, val)
	case key == "type":
		item.Type, item.Format, _ = getTypeFormat(val)
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

func (p *Parser) parseResponse(resp *Responses, vals map[string]string) {
	for key, val := range vals {
		switch {
		case key == "$ref":
			resp.Ref = "#/responses/" + val
			p.usedResponses = append(p.usedResponses, val)
		case key == "description" || key == "desc":
			resp.Description = val
		case pathMatch("schema.*", key):
			if resp.Schema == nil {
				resp.Schema = &Schema{}
			}
			p.parseSchema(resp.Schema, strings.TrimPrefix(key, "schema."), val)
		}
	}
}

func (p *Parser) parseDefinitionOptions(def *Definition, comments []*ast.Comment) {
	i := 0
	for ; i < len(comments); i++ {
		if strings.TrimSpace(comments[i].Text) == "//" {
			return
		}

		index := findAt(comments[i].Text)
		if index > 0 {
			tag, vals := getValues(comments[i].Text[index:])
			switch tag {
			case "@Description":
				def.Description = vals
			}
		}
	}
}

func (p *Parser) parsePropertiesName(comments []*ast.Comment) string {
	i := 0
	for ; i < len(comments); i++ {
		if strings.TrimSpace(comments[i].Text) == "//" {
			return ""
		}

		index := findAt(comments[i].Text)
		if index > 0 {
			tag, vals := getValues(comments[i].Text[index:])
			switch tag {
			case "@Name":
				return vals
			}
		}
	}

	return ""
}

func (p *Parser) parsePropertiesOptions(name string, def *Definition, prop *Definition, comments []*ast.Comment) {
	i := 0
	for ; i < len(comments); i++ {
		if strings.TrimSpace(comments[i].Text) == "//" {
			return
		}

		index := findAt(comments[i].Text)
		if index > 0 {
			tag, vals := getValues(comments[i].Text[index:])
			switch tag {
			case "@Description":
				prop.Description = vals
			case "@Required":
				def.Required = append(def.Required, name)
			}
		}
	}
}

func (p *Parser) parseDefinition(packName string, comments []*ast.Comment, astTypeSpec *ast.TypeSpec) {
	name := packName + "." + astTypeSpec.Name.String()
	if _, ok := swagger.Definitions[name]; ok {
		return
	}

	compositeDef := []*Definition{}
	def := &Definition{}
	p.parseDefinitionOptions(def, comments)

	switch astType := astTypeSpec.Type.(type) {
	default:
		p.parseProperties(packName, def, astType)
	case *ast.StructType:
		def.Type = "object"
		def.Properties = make(map[string]*Definition)

		for i := 0; i < astType.Fields.NumFields(); i++ {
			propName := ""
			astField := astType.Fields.List[i]

			if astField.Doc != nil {
				propName = p.parsePropertiesName(astField.Doc.List)
			}

			if propName == "-" {
				continue
			}

			if propName == "" {
				if len(astField.Names) == 0 {
					if astSelectorExpr, ok := astField.Type.(*ast.SelectorExpr); ok {
						fieldPackName := astSelectorExpr.X.(*ast.Ident).Name
						propName = fieldPackName + "." + strings.TrimPrefix(astSelectorExpr.Sel.Name, "*")
					} else if astTypeIdent, ok := astField.Type.(*ast.Ident); ok {
						propName = astTypeIdent.Name
					} else if astStarExpr, ok := astField.Type.(*ast.StarExpr); ok {
						if astSelectorExpr, ok := astStarExpr.X.(*ast.SelectorExpr); ok {
							fieldPackName := astSelectorExpr.X.(*ast.Ident).Name
							compositeDef = append(compositeDef, &Definition{
								Ref: "#/definitions/" + fieldPackName + "." + strings.TrimPrefix(astSelectorExpr.Sel.Name, "*"),
							})
							continue
						}

						if astIdent, ok := astStarExpr.X.(*ast.Ident); ok {
							compositeDef = append(compositeDef, &Definition{
								Ref: "#/definitions/" + packName + "." + astIdent.Name,
							})
							// propName = astIdent.Name
							continue
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
				p.parsePropertiesOptions(propName, def, field, astField.Doc.List)
			}

			p.parseProperties(packName, field, astField.Type)
			def.Properties[propName] = field
		}
	}

	if len(compositeDef) > 0 {
		compositeDef = append(compositeDef, def)
		swagger.Definitions[name] = &Definition{
			AllOf: compositeDef,
		}
	} else {
		swagger.Definitions[name] = def
	}
}

func (p *Parser) parseProperties(packName string, def *Definition, astType ast.Expr) {
	var ok bool
	switch fieldType := astType.(type) {
	case *ast.MapType:
		def.Type = "object"
		if def.AdditionalProperties == nil {
			def.AdditionalProperties = &Definition{}
		}
		p.parseProperties(packName, def.AdditionalProperties, fieldType.Value)
	case *ast.ArrayType:
		def.Type = "array"
		def.Items = &Items{}
		switch arrayType := fieldType.Elt.(type) {
		case *ast.InterfaceType:
			def.Items.Type = "any"
		case *ast.Ident:
			def.Items.Type, def.Items.Format, _ = getTypeFormat(arrayType.String())
		}
	case *ast.StarExpr:
		def.Type, def.Format, ok = getTypeFormat(checkTypePtr(fmt.Sprint(fieldType.X)))
		if !ok {
			if strings.Contains(def.Type, ".") {
				def.Ref = "#/definitions/" + def.Type
			} else {
				def.Ref = "#/definitions/" + packName + "." + def.Type
			}
			def.Type = ""
			def.Format = ""
		}
	case *ast.SelectorExpr:
		packName = fieldType.X.(*ast.Ident).Name
		def.Type, def.Format, ok = getTypeFormat(fieldType.Sel.Name)
		if !ok {
			def.Ref = "#/definitions/" + packName + "." + def.Type
			def.Type = ""
			def.Format = ""
		}
	case *ast.Ident:
		def.Type, def.Format, ok = getTypeFormat(fieldType.String())
		if !ok {
			def.Ref = "#/definitions/" + packName + "." + def.Type
			def.Type = ""
			def.Format = ""
		}
	}
}

func (p *Parser) isUsedDefinition(name string, defs []string) bool {
	for i := 0; i < len(defs); i++ {
		if name == defs[i] {
			return true
		}
	}

	return false
}

func (p *Parser) mergeCompositeDefinition() {
	usedMerge := []string{}
	for defName, def := range swagger.Definitions {
		if len(def.AllOf) > 0 {
			newDef := &Definition{
				Properties: make(map[string]*Definition),
			}

			for i := 0; i < len(def.AllOf); i++ {
				if len(def.AllOf[i].Properties) != 0 {
					newDef.Description = def.Description
					for key, val := range def.AllOf[i].Properties {
						newDef.Properties[key] = val
					}
				}

				if def.AllOf[i].Ref != "" {
					refDefName := strings.TrimPrefix(def.AllOf[i].Ref, "#/definitions/")
					if _, ok := swagger.Definitions[refDefName]; !ok {
						panic("cannot merged composite struct, cannot found " + refDefName + " definition")
					}

					usedMerge = append(usedMerge, refDefName)

					for key, val := range swagger.Definitions[refDefName].Properties {
						newDef.Properties[key] = val
					}
				}
			}
			swagger.Definitions[defName] = newDef
		}
	}

	for defName, _ := range swagger.Definitions {
		if p.isUsedDefinition(defName, usedMerge) {
			delete(swagger.Definitions, defName)
		}
	}
}

func (p *Parser) validate() {
	for _, defName := range p.usedDefinitions {
		if _, ok := swagger.Definitions[defName]; !ok {
			panic("cannot found " + defName + " in definitions")
		}
	}

	for _, defName := range p.usedParameters {
		if _, ok := swagger.Parameters[defName]; !ok {
			panic("cannot found " + defName + " in parameters")
		}
	}

	for _, defName := range p.usedResponses {
		if _, ok := swagger.Responses[defName]; !ok {
			panic("cannot found " + defName + " in responses")
		}
	}
}
