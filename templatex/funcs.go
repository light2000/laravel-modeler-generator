package templatex

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/light2000/laravel-modeler-generator/common"
	"github.com/light2000/laravel-modeler-generator/logger"
)

var (
	funcMap = template.FuncMap{
		"dict": func(values ...any) map[string]any {
			if len(values)%2 != 0 {
				logger.Fatalf("dict expects even number of args")
			}
			m := map[string]any{}
			for i := 0; i < len(values); i += 2 {
				m[values[i].(string)] = values[i+1]
			}
			return m
		},
		"upper": strings.ToUpper,
		"ToVar": func(content string) string {
			return common.ToLowerCamel(content)
		},
		"BigQuot": func(content string) string {
			return "{" + content + "}"
		},
		"DoubleQuot": func(content string) string {
			return "{{" + content + "}}"
		},
		"BigQuotInt8": func(content uint8) string {
			return fmt.Sprintf("{%d}", content)
		},
		"BigLeftQuot": func(content string) string {
			return fmt.Sprintf("{%s", content)
		},
		"DoubleLeftQuot": func(content string) string {
			return fmt.Sprintf("{{%s", content)
		},
		"BigRightQuot": func(content string) string {
			return fmt.Sprintf("%s}", content)
		},
		"DoubleRightQuot": func(content string) string {
			return fmt.Sprintf("%s}}", content)
		},
		"cond": func(b bool, a, c string) string {
			if b {
				return a
			}
			return c
		},
		"sub": func(a, b int) int {
			return a - b
		},
		"PHPEscape": func(s string) string {
			s = strings.ReplaceAll(s, `\`, `\\`)
			s = strings.ReplaceAll(s, `"`, `\"`)
			s = strings.ReplaceAll(s, `$`, `\$`)
			return s
		},
	}
)
