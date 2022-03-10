package utils

import "strings"

// LastWord 最后一个单词
// @Author: hyc
// @Description:
// @Date: 2021/11/15 16:40
func LastWord(s string) string {
	words := strings.Fields(s)
	if len(words) == 0 {
		return ""
	} else {
		return words[len(words) - 1]
	}
}