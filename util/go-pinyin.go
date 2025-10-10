package util

import (
	"golang.org/x/text/language"
	"strings"
	"unicode"

	"github.com/mozillazg/go-pinyin"
	"golang.org/x/text/cases"
)

// Pinyin 中文转拼音
// 定义一个函数Pinyin，接收一个字符串参数s，返回一个拼音字符串pinyinName
func Pinyin(s string) (pinyinName string) {
	// 如果输入字符串为空，则直接返回空字符串
	if s == "" {
		return ""
	}

	// 定义两个字符串构建器，用于存储汉字和非汉字字符
	var hanChars, serialNumber strings.Builder
	// 遍历输入字符串中的每个字符
	for _, char := range s {
		// 判断字符是否为汉字，将汉字字符追加到hanChars中，非汉字字符追加到serialNumber中
		if unicode.Is(unicode.Scripts["Han"], char) {
			hanChars.WriteRune(char)
		} else {
			serialNumber.WriteRune(char)
		}
	}

	// 调用pinyin.LazyConvert方法将hanChars中的汉字转换为拼音数组
	pinyin := pinyin.LazyConvert(hanChars.String(), nil)

	// 将拼音数组转换为字符串，并与非汉字字符拼接后赋值给pinyinName
	pinyinName = strings.Join(pinyin, "")

	// 返回拼音字符串和非汉字字符拼接后的结果
	return pinyinName + serialNumber.String()
}

// PinyinHump
// 中文转拼音驼峰
func PinyinHump(s string) (pinyinName string) {
	// 如果输入字符串为空，则直接返回空字符串
	if s == "" {
		return ""
	}

	// 定义两个字符串构建器，用于存储汉字和非汉字字符
	var hanChars, serialNumber strings.Builder
	// 遍历输入字符串中的每个字符
	for _, char := range s {
		// 判断字符是否为汉字，将汉字字符追加到hanChars中，非汉字字符追加到serialNumber中
		if unicode.Is(unicode.Scripts["Han"], char) {
			hanChars.WriteRune(char)
		} else {
			serialNumber.WriteRune(char)
		}
	}

	// 调用pinyin.LazyConvert方法将hanChars中的汉字转换为拼音数组
	pinyin := pinyin.LazyConvert(hanChars.String(), nil)

	// 将拼音数组转换为字符串，并与非汉字字符拼接后赋值给pinyinName
	// 首字母转大写
	// pinyinName =
	c := cases.Title(language.English)
	for _, v := range pinyin {
		pinyinName = pinyinName + c.String(v)
	}

	// 返回拼音字符串和非汉字字符拼接后的结果
	return pinyinName + serialNumber.String()
}
