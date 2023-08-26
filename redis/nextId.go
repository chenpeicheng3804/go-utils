package redisson

import (
	"time"
)

const (
	// BeginTimestamp 开始时间戳
	BeginTimestamp = 951969723 //2000-03-02 12:02:03 +0800 CST
	// CountBits 序列号的位数
	CountBits = 32
	// NextidKey 唯一ID前缀
	NextidKey = "icr:"
)

// NextId
// 基于redis生成唯一id
func (rdb *Rdb) NextId(key string) int64 {
	// 1.生成时间戳
	timeUnix := time.Now().Unix()
	// 2.生成序列号
	// 2.1.获取当前日期，精确到天 作为key部分
	//日期格式:yyyy:MM:dd
	date := time.Now().Format(":2006:01:02")

	// 2.2.redis key自增长
	//key拼接 ="icr:"+ "传参key:" + 当前日期
	// Incr value=自增长
	IntCmd := rdb.Client.Incr(NextidKey + key + date)
	if IntCmd.Err() != nil {
		return 0
	}
	// 3.拼接并返回序列号
	//返回值=（当前时间戳-自定义时间戳）<<COUNT_BITS | 自增长
	return (timeUnix-BeginTimestamp)<<CountBits | IntCmd.Val()

}
