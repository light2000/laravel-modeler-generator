package meta

import "fmt"

// checkSnakeCode 校验 code 是否为蛇形：小写英文字母开头，仅含小写英文字母、数字与下划线，且不以 '_' 结尾。
// 合法时返回空字符串；否则返回可读错误说明。
func checkSnakeCode(code string) string {
	if code == "" {
		return "code 不能为空"
	}
	if code[0] < 'a' || code[0] > 'z' {
		return "code 须以小写英文字母开头"
	}
	for i := 0; i < len(code); i++ {
		c := code[i]
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_' {
			continue
		}
		return fmt.Sprintf("code 仅允许小写英文字母、数字与下划线（位置 %d 字符 %q 非法）", i, c)
	}
	if code[len(code)-1] == '_' {
		return "code 不能以 '_' 结尾"
	}
	return ""
}

func mustSnakeCode(from string, id string, name string, code string) {
	if msg := checkSnakeCode(code); msg != "" {
		panic(fmt.Sprintf("meta.%s: %s (id=%s, name=%s, code=%q)", from, msg, id, name, code))
	}
}
