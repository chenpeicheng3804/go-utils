package redisson

import (
	"errors"
	"fmt"
	"github.com/chenpeicheng3804/go-utils/util"
	"github.com/go-basic/uuid"
	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

const (
	LockReleaseFlag    = 0
	LockReleaseChannel = "LockReleaseChannel"
	LockExpiration     = 30 * time.Second
	LockTimeout        = 5 * time.Second
)

// rLockMsg ...
var rLockMsg = map[int64]string{
	0:  "锁被占用",
	-1: "加锁成功",
	-2: "可重入锁",
}

// rLockScript ...
var rLockScript = redis.NewScript(`
-- 若锁不存在：则新增锁，并设置锁重入计数为1、设置锁过期时间
if (redis.call('exists', KEYS[1]) == 0) then
    redis.call('HSET', KEYS[1], ARGV[2], 1);
    redis.call('PEXPIRE', KEYS[1], ARGV[1]);
    return -1;
end;

-- 若锁存在，且唯一标识也匹配：则表明当前加锁请求为锁重入请求，故锁重入计数+1，并再次设置锁过期时间
if (redis.call('HEXISTS', KEYS[1], ARGV[2]) == 1) then
    redis.call('HINCRBY', KEYS[1], ARGV[2], 1);
    redis.call('PEXPIRE', KEYS[1], ARGV[1]);
    return -2;
end;

-- 若锁存在，但唯一标识不匹配：表明锁是被其他线程占用，当前线程无权解他人的锁，直接返回锁剩余过期时间
return redis.call('PTTL', KEYS[1]);
`)

// rUnlockMsg ...
var rUnlockMsg = map[int64]string{
	1: "锁不存在",
	2: "不允许解锁其他线程持有的锁",
	3: "可重入锁，已续期",
	4: "解锁成功",
}

// rUnlockScript ...
var rUnlockScript = redis.NewScript(`
-- 若锁不存在：则直接广播解锁消息，并返回1
if (redis.call('exists', KEYS[1]) == 0) then
    redis.call('publish', KEYS[2], ARGV[1]);
    return 1;
end;

-- 若锁存在，但唯一标识不匹配：则表明锁被其他线程占用，当前线程不允许解锁其他线程持有的锁
if (redis.call('hexists', KEYS[1], ARGV[3]) == 0) then
    return 2;
end;

-- 若锁存在，且唯一标识匹配：则先将锁重入计数减1
local counter = redis.call('hincrby', KEYS[1], ARGV[3], -1);
if (counter > 0) then
    -- 锁重入计数减1后还大于0：表明当前线程持有的锁还有重入，不能进行锁删除操作，但可以友好地帮忙设置下过期时期
    redis.call('pexpire', KEYS[1], ARGV[2]);
    return 3;
else
    -- 锁重入计数已为0：间接表明锁已释放了。直接删除掉锁，并广播解锁消息，去唤醒那些争抢过锁但还处于阻塞中的线程
    redis.call('del', KEYS[1]);
    redis.call('publish', KEYS[2], ARGV[1]);
    return 4;
end;
`)

// rLock ...
var rLock = &RLock{
	once: &sync.Once{},
}

// RLock ...
type RLock struct {
	once    *sync.Once
	mutex   *sync.Mutex
	rdb     redis.Cmdable
	name    string
	clintId string
}

// NewRLock ...
func NewRLock(rdb redis.Cmdable) *RLock {
	if _, err := rdb.Ping().Result(); err != nil {
		return nil
	}

	rLock = &RLock{
		mutex:   &sync.Mutex{},
		rdb:     rdb,
		clintId: uuid.New(),
	}

	return rLock
}

// uniqueId ...
func (rLock *RLock) uniqueId() string {
	gid := util.GetGoroutineID()
	return fmt.Sprintf("%s-%d", rLock.clintId, gid)
}

// buildLockArgs ...
func (rLock *RLock) buildLockArgs(args ...time.Duration) time.Duration {
	expiration := LockExpiration

	if len(args) == 1 {
		expiration = args[0]
	}

	expiration = expiration / 1000 / 1000
	return expiration
}

// buildTryLockArgs ...
func (rLock *RLock) buildTryLockArgs(args ...time.Duration) (time.Duration, time.Duration) {
	expiration := LockExpiration
	timeout := LockTimeout

	if len(args) == 1 {
		expiration = args[0]
	}

	if len(args) == 2 {
		expiration = args[0]
		timeout = args[1]
	}

	expiration = expiration / 1000 / 1000
	return expiration, timeout
}

// goTryLock ...
func (rLock *RLock) goTryLock(key, uniqueId string, expiration, timeout time.Duration) error {
	ch := make(chan bool)

	go func() {
		for {
			if _, ok := <-ch; !ok {
				break
			}

			ret, err := rLockScript.Run(rLock.rdb, []string{key}, expiration, uniqueId).Result()
			if err == nil && ret.(int64) <= 0 {
				ch <- true
				break
			}
		}
	}()

	select {
	case <-ch:
		return nil
	case <-time.After(timeout):
		close(ch)
		return errors.New("timeout")
	}
}

// printLog ...
func (rLock *RLock) printLog(name string, ret interface{}) {
	index := ret.(int64)
	if index > 0 {
		index = 0
	}
	logrus.Infof("lock name=%s field=%s code=%v msg=%s", name, rLock.uniqueId(), ret, rLockMsg[index])
}

// Lock ...
func (rLock *RLock) Lock(name string, args ...time.Duration) error {
	if rLock == nil {
		return errors.New("redis conn err")
	}

	expiration := rLock.buildLockArgs(args...)

	rLock.mutex.Lock()
	defer rLock.mutex.Unlock()

	ret, err := rLockScript.Run(rLock.rdb, []string{name}, int64(expiration), rLock.uniqueId()).Result()
	if err != nil {
		return err
	}

	rLock.printLog(name, ret)
	if ret.(int64) > 0 {
		return errors.New(fmt.Sprintf("[%v]lock fail", ret))
	}

	rLock.name = name
	return nil
}

// TryLock ...
func (rLock *RLock) TryLock(name string, args ...time.Duration) error {
	if rLock == nil {
		return errors.New("redis conn err")
	}

	expiration, timeout := rLock.buildTryLockArgs(args...)

	rLock.mutex.Lock()
	defer rLock.mutex.Unlock()

	uniqueId := rLock.uniqueId()
	ret, err := rLockScript.Run(rLock.rdb, []string{name}, int64(expiration), rLock.uniqueId()).Result()
	if err != nil {
		return err
	}

	rLock.printLog(name, ret)
	if ret.(int64) > 0 {
		// 超时时间内，不断尝试加锁
		return rLock.goTryLock(name, uniqueId, expiration, timeout)
	}

	rLock.name = name
	return nil
}

// UnLock ...
func (rLock *RLock) UnLock() error {
	rLock.mutex.Lock()
	defer rLock.mutex.Unlock()

	expiration := int64(LockExpiration) / 1000 / 1000
	ret, err := rUnlockScript.Run(rLock.rdb, []string{rLock.name, LockReleaseChannel}, LockReleaseFlag, expiration, rLock.uniqueId()).Result()
	if err != nil {
		logrus.Error(err)
		return err
	}

	logrus.Infof("unlock name=%s field=%s code=%v msg=%s", rLock.name, rLock.uniqueId(), ret, rUnlockMsg[ret.(int64)])
	return nil
}

// NextId 全局唯一id生成
func (rLock *RLock) NextId(key string) int64 {
	// 1.生成时间戳
	timeUnix := time.Now().Unix()
	// 2.生成序列号
	// 2.1.获取当前日期，精确到天 作为key部分
	//日期格式:yyyy:MM:dd
	date := time.Now().Format(":2006:01:02")

	// 2.2.redis key自增长
	//key拼接 ="icr:"+ "传参key:" + 当前日期
	// Incr value=自增长
	IntCmd := rLock.rdb.Incr(NextidKey + key + date)
	if IntCmd.Err() != nil {
		return 0
	}
	// 3.拼接并返回序列号
	//返回值=（当前时间戳-自定义时间戳）<<COUNT_BITS | 自增长
	return (timeUnix-BeginTimestamp)<<CountBits | IntCmd.Val()

}
