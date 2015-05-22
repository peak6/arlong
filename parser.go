package arlong

import (
	"github.com/Sirupsen/logrus"
	"github.com/plimble/utils/parsetype"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

type Parser struct {
	packages        []*ast.Package
	usedDefinitions []*Schema
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
	p.usedDefinitions = []*Schema{}
	p.usedParameters = []string{}
	p.usedResponses = []string{}
	p.json = nil

	if err := p.parsePackages(); err != nil {
		return err
	}

	p.parseComments()
	p.parseDefinitionModels()
	// p.mergeAll()
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

func (p *Parser) parseDefinitionModels() {
	parser := parsetype.NewParser()

	packNames := make(map[string]struct{})
	for _, val := range p.usedDefinitions {
		index := strings.LastIndex(val.rawRefName, ".")
		data := val.rawRefName
		if index > 0 {
			data = val.rawRefName[:index]
		}

		packNames[data] = struct{}{}
	}

	for key := range packNames {
		parser.ParseFromGoPath(key)
	}

	parser.MergeComposite()

	for _, val := range p.usedDefinitions {
		def := &Schema{}

		if _, ok := parser.Types[val.rawRefName]; !ok {
			logrus.Errorf("Could not find %s package", val.rawRefName)
			continue
		}

		pType := parser.Types[val.rawRefName]

		if pType.Doc != nil {
			p.parseDefinitionOptions(def, pType.Doc.List)
		}
		p.parseDefinitionModel(def, pType)

		keyName := val.rawRefName
		swagger.Definitions[keyName] = def
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
			case "@Definition":
				p.parseDefinition(comments[i:])
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
				if !strings.Contains(vals, ":") {
					method.Security = append(method.Security, map[string][]string{vals: []string{}})
				} else {
					method.Security = append(method.Security, getValueMapStrings(vals))
				}
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

func (p *Parser) parseDefinition(comments []*ast.Comment) int {
	i := 0
	var defName string
	for ; i < len(comments); i++ {
		if strings.TrimSpace(comments[i].Text) == "//" {
			return i
		}

		index := findAt(comments[i].Text)
		if index > 0 {
			tag, vals := getValues(comments[i].Text[index:])
			switch tag {
			case "@Definition":
				defName = vals
				if swagger.Definitions[defName] == nil {
					swagger.Definitions[defName] = &Schema{}
				}
			case "@Description":
				swagger.Definitions[defName].Description = vals
			case "@Property":
				if swagger.Definitions[defName].Properties == nil {
					swagger.Definitions[defName].Properties = make(map[string]*Schema)
				}

				propText := strings.Replace(vals, "\t", " ", -1)
				data := strings.SplitN(propText, " ", 2)
				if len(data) != 2 {
					panic("Invalid @Property arguments")
				}

				propName, propVals := strings.TrimSpace(data[0]), strings.TrimSpace(data[1])
				def := &Schema{}
				valArray := getValueByKey(propVals)
				p.parseDefinitionField(def, valArray)
				swagger.Definitions[defName].Properties[propName] = def
			case "@Type":
				swagger.Definitions[defName].Type, swagger.Definitions[defName].Format, _ = getTypeFormat(vals)
			case "@Required":
				swagger.Definitions[defName].Required = getValueStrings(vals)
			case "@Items":
				if swagger.Definitions[defName].Items == nil {
					swagger.Definitions[defName].Items = &Schema{}
				}
				data := getValueByKey(vals)
				for key, val := range data {
					p.parseSchema(swagger.Definitions[defName].Items, key, val)
				}
			}
		}
	}

	return i
}

func (p *Parser) parseDefinitionField(def *Schema, vals map[string]string) {
	for key, val := range vals {
		switch {
		case key == "$ref":
			def.Ref = "#/definitions/" + val
		case key == "type":
			def.Type, def.Format, _ = getTypeFormat(val)
		case key == "description" || key == "desc":
			def.Description = val
		case pathMatch("items.*", key):
			if def.Items == nil {
				def.Items = &Schema{}
			}
			p.parseSchema(def.Items, strings.TrimPrefix(key, "items."), val)
		}
	}
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
		s.rawRefName = val
		p.usedDefinitions = append(p.usedDefinitions, s)
	case pathMatch("items.*", key):
		if s.Items == nil {
			s.Items = &Schema{}
		}
		p.parseSchema(s.Items, strings.TrimPrefix(key, "items."), val)
	}
}

func (p *Parser) parseItem(item *Items, key, val string) {
	switch {
	// case key == "$ref":
	// 	item.Ref = "#/definitions/" + val
	// 	p.usedDefinitions[val] = struct{}{}
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

func (p *Parser) parseDefinitionOptions(def *Schema, comments []*ast.Comment) {
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

func (p *Parser) parsePropertiesOptions(name string, def *Schema, prop *Schema, comments []*ast.Comment) {
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

func (p *Parser) parseDefinitionModel(def *Schema, pType *parsetype.Type) {
	switch pType.Type {
	case "ref":
		if pType.RefType != nil {
			switch pType.RefType.Type {
			case "struct", "ref":
				// p.parseDefinitionModel(def, pType.RefType)
				def.Ref = "#/definitions/" + pType.RefType.Name
				if _, ok := swagger.Definitions[pType.RefType.Name]; !ok {
					swagger.Definitions[pType.RefType.Name] = &Schema{}
					p.parseDefinitionModel(swagger.Definitions[pType.RefType.Name], pType.RefType)
				}
			default:
				def.Type, def.Format, _ = getTypeFormat(pType.RefType.Type)
			}
		}
	case "struct":
		def.Properties = make(map[string]*Schema)
		for key, val := range pType.Properties {
			propDef := &Schema{}

			name := ""
			nameDoc := ""

			if val.Doc != nil {
				nameDoc = p.parsePropertiesName(val.Doc.List)
			}

			if val.Tags != "" && val.Tags.Get(`json`) != "" {
				jsonData := strings.Split(val.Tags.Get(`json`), ",")
				if jsonData[0] == "-" || jsonData[0] == "" {
					continue
				}
				name = jsonData[0]
			} else if nameDoc != "" {
				name = nameDoc
			} else {
				name = key
			}

			if val.Doc != nil {
				p.parsePropertiesOptions(name, def, propDef, val.Doc.List)
			}

			def.Properties[name] = propDef
			p.parseDefinitionModel(propDef, val)
		}
	case "map":
		def.Type = "object"
		def.AdditionalProperties = &Schema{}

		p.parseDefinitionModel(def.AdditionalProperties, pType.MapType)
	case "array":
		def.Type = "array"
		def.Items = &Schema{}
		p.parseDefinitionModel(def.Items, pType.ArrayType)
	default:
		def.Type, def.Format, _ = getTypeFormat(pType.Type)
	}
}

func (p *Parser) validate() {
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

func (p *Parser) mergeAll() {
	for _, val := range p.usedDefinitions {
		cloneSchema := swagger.Definitions[val.rawRefName]
		val.Ref = cloneSchema.Ref
		val.AdditionalProperties = cloneSchema.AdditionalProperties
		val.AllOf = cloneSchema.AllOf
		val.Description = cloneSchema.Description
		val.Format = cloneSchema.Format
		val.Items = cloneSchema.Items
		val.Properties = cloneSchema.Properties
		val.Required = cloneSchema.Required
		val.Type = cloneSchema.Type
	}
	swagger.Definitions = nil
}
