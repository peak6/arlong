package arlong

import (
	"reflect"
	"strings"
)

func createDifinition(tv reflect.Type) {
	if tv.Kind() == reflect.Ptr {
		tv = tv.Elem()
	}

	if _, ok := swagger.Definitions[tv.Name()]; ok {
		return
	}

	def := &Definition{
		Properties: make(map[string]*Definition),
	}

	swagger.Definitions[tv.Name()] = def

	var fieldType reflect.Type
	var structField reflect.StructField
	for i := 0; i < tv.NumField(); i++ {
		structField = tv.Field(i)

		if tv.Field(i).Type.Kind() == reflect.Ptr {
			fieldType = tv.Field(i).Type.Elem()

		} else {
			fieldType = tv.Field(i).Type
		}

		field := &Definition{}

		name := structField.Tag.Get("field")
		names := strings.Split(name, ",")
		if name == "" {
			name = structField.Name
		}
		if len(names) == 2 && names[1] == "*" {
			name = names[0]
			if def.Required == nil {
				def.Required = []string{}
			}

			def.Required = append(def.Required, name)
		}

		desc := structField.Tag.Get("desc")

		def.Properties[name] = field

		field.Description = desc

		if fieldType.Name() == "Time" {
			field.Type = "string"
			field.Format = "date-time"
			continue
		}

		switch fieldType.Kind() {
		case reflect.Map:
			panic("unsupport map")
		case reflect.Slice:
			field.Type = "array"
			field.Items = &Items{}
			switch fieldType.Elem().Kind() {
			case reflect.Ptr:
				fieldType = fieldType.Elem().Elem()
				field.Description = ""
				field.Items.Ref = "#/definitions/" + fieldType.Name()

				createDifinition(fieldType)
			case reflect.Struct:
				fieldType := fieldType.Elem()
				field.Description = ""
				field.Items.Ref = "#/definitions/" + fieldType.Name()

				createDifinition(fieldType)
			default:
				field.Items.Type, field.Items.Format = setTypeAndFormat(fieldType.Elem())
			}
		case reflect.Struct:
			field.Description = ""
			field.Ref = "#/definitions/" + fieldType.Name()

			createDifinition(fieldType)
		default:
			field.Type, field.Format = setTypeAndFormat(structField.Type)
		}
	}

}

func setTypeAndFormat(tv reflect.Type) (string, string) {
	switch tv.Kind() {
	case reflect.String:
		return "string", ""
	case reflect.Int:
		return "integer", "int32"
	case reflect.Int32:
		return "integer", "int32"
	case reflect.Int64:
		return "integer", "int64"
	case reflect.Float32, reflect.Float64:
		return "number", "float"
	case reflect.Bool:
		return "boolean", ""
	}

	return "unknown", ""
}
