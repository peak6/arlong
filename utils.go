package arlong

import (
	"path"
	"strconv"
	"strings"
	"unicode/utf8"
)

func findAt(text string) int {
	size := 0
	index := 0
	var char rune
	for size < len(text) {
		char, size = utf8.DecodeRuneInString(text)
		text = text[size:]

		switch char {
		case ' ', '/':
			index += size
			continue
		case '@':
			return index
		default:
			return -1
		}
	}

	return -1
}

func getValues(text string) (string, string) {
	fullText := text
	size := 0
	index := 0
	var char rune
	for size < len(text) {
		char, size = utf8.DecodeRuneInString(text)
		text = text[size:]

		switch char {
		case ' ':
			return strings.TrimSpace(fullText[:index]), strings.TrimSpace(fullText[index:])
		default:
			index += size
		}
	}

	return strings.TrimSpace(fullText), ""
}

func getValueByKey(s string, delim ...rune) map[string]string {
	if len(delim) == 0 {
		delim = []rune{'='}
	}

	result := map[string]string{}
	fullText := s
	size := 0
	index := 0
	key := ""
	inQuote := false
	startVal := 0
	startKey := 0
	var char rune
	for size < len(s) {
		char, size = utf8.DecodeRuneInString(s)
		s = s[size:]

		switch char {
		case ' ', '\t':
			if inQuote {
				index += size
			} else {
				if key != "" {
					result[key] = strings.TrimSpace(fullText[startVal:index])
					key = ""
					index += size
					startKey = index
				} else {
					if fullText[startKey:index] != "" {
						result[strings.TrimSpace(fullText[startKey:index])] = ""
						startKey = index
					}
					index += size
				}
			}
		case '"':
			if !inQuote {
				inQuote = true
				index += size
				startVal = index
			} else {
				result[key] = strings.TrimSpace(fullText[startVal:index])
				inQuote = false
				key = ""
				index += size
				startKey = index
			}
		case delim[0]:
			key = strings.TrimSpace(fullText[startKey:index])
			index += size
			result[key] = ""
			startVal = index
		default:
			index += size
		}
	}

	if key != "" {
		if !inQuote {
			index += size
		}
		result[key] = strings.TrimSpace(fullText[startVal:index])
	}

	return result
}

func checkTypePtr(s string) string {
	char, _ := utf8.DecodeRuneInString(s)
	if char == '&' {
		strs := strings.Split(s, " ")
		s = strings.TrimSuffix(strs[1], "}")
	}
	return s
}

func getTypeFormat(val string) (string, string, bool) {
	switch val {
	case "string":
		return "string", "", true
	case "int":
		return "integer", "int32", true
	case "int32":
		return "integer", "int32", true
	case "int64":
		return "integer", "int64", true
	case "float32", "float64":
		return "number", "float", true
	case "bool":
		return "boolean", "", true
	case "date-time", "time.Time", "Time":
		return "string", "date-time", true
	case "date":
		return "string", "date", true
	case "object":
		return "object", "", true
	case "array":
		return "array", "", true
	}

	return val, "", false
}

func strToInt(val string) int {
	valInt, err := strconv.Atoi(val)
	if err != nil {
		panic(err)
	}

	return valInt
}

func getValueStrings(s string) []string {
	fields := strings.Fields(s)
	for i := 0; i < len(fields); i++ {
		fields[i] = strings.TrimSpace(fields[i])
	}
	return fields
}

func getValueMapStrings(s string) map[string][]string {
	result := map[string][]string{}
	data := getValueByKey(s)
	for key, val := range data {
		result[key] = strings.Split(val, ",")
	}

	return result
}

func getMime(s string) string {
	switch s {
	case "xml":
		return "application/xml"
	case "json":
		return "application/json"
	case "html":
		return "text/html"
	case "text":
		return "text/plain"
	case "form":
		return "application/x-www-form-urlencoded"
	case "multipart":
		return "multipart/form-data"
	}

	return s
}

func pathMatch(pattern, name string) bool {
	exist, err := path.Match(pattern, name)
	if err != nil {
		panic(err)
	}

	return exist
}
