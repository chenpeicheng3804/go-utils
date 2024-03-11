package util

import (
	"github.com/mozillazg/go-pinyin"
)

func Pinyin(s string) (pinyinName string) {
	pinyin := pinyin.LazyConvert(s, nil)

	for _, dd := range pinyin {
		pinyinName = pinyinName + dd
	}

	return pinyinName
}
