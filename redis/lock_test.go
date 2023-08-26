package redisson

import (
	"github.com/go-redis/redis"
	"sync"
	"testing"
)

// TestLock ...
func TestLock(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "demo",
		DB:       0,
	})
	var wg sync.WaitGroup
	rLock := NewRLock(rdb)
	//defer rLock.UnLock()
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := rLock.Lock("myLock")
		if err != nil {
			t.Errorf("lock fail, err: %v", err)
		}

		err = rLock.Lock("myLock")
		if err != nil {
			t.Errorf("lock fail, err: %v", err)
		}
		defer rLock.UnLock()
		defer rLock.UnLock()
	}()
	wg.Wait()
	t.Log("lock success")

}

// TestTryLock ...
func TestTryLock(t *testing.T) {

	rdb := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "demo",
		DB:       0,
	})

	rLock := NewRLock(rdb)
	err := rLock.TryLock("myTryLock")
	if err != nil {
		t.Errorf("lock fail, err: %v", err)
	}
	t.Log("lock success")
}
