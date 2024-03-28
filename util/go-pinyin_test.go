package util

import (
	"fmt"
	"testing"
)

func TestPinyin(t *testing.T) {
	s := "中国人a"
	fmt.Println(Pinyin(s))
	//s = "中国人1"
	s = "zouyue01"
	fmt.Println(Pinyin(s))
	//unicode.IsLetter(r)

}
