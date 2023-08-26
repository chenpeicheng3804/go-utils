package redisson

import (
	"fmt"
	"sync"
	"testing"
)

func TestNextId(t *testing.T) {
	rdb := NewClient(
		"127.0.0.1:6379",
		"demo",
		0,
	)

	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fmt.Println(rdb.NextId("NextId"))
		}()

	}
	wg.Wait()

}
