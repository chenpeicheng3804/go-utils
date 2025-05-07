package util

import (
	"math/rand"
	"strings"
	"time"
)

// 随机字符串 key
const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890!@#$%^&*()"

// 生成随机密码
func RandomString(n int) string {
	seed := time.Now().UnixNano()       // 使用当前时间的纳秒级别作为种子
	r := rand.New(rand.NewSource(seed)) // 创建一个独立的随机数生成器
	sb := strings.Builder{}
	sb.Grow(n)
	for i := 0; i < n; i++ {
		sb.WriteByte(charset[r.Intn(len(charset))])
	}
	return sb.String()
}
