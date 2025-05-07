package util

import (
	"fmt"
	"testing"
)

func TestPinyin(t *testing.T) {
	s := "⼭"
	fmt.Println(Pinyin(s))
	////s = "中国人1"
	s = "山"
	fmt.Println(Pinyin(s))
	//unicode.IsLetter(r)

}
