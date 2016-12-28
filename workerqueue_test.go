package workqueue

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func work(r *WorkRequest) (interface{}, error) {

	duration := r.Data.(time.Duration)
	time.Sleep(duration)

	return nil, nil
}

func TestWorkQueue(t *testing.T) {

	queue := NewDispatcherWithLogger(func(str string, args ...interface{}) {
		fmt.Printf(str, args...)
	})

	queue.Start(10, func(id int) Worker {
		return NewWorker(id, work)
	})

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			queue.RequestAndWait(time.Millisecond * time.Duration(i*100))
			//fmt.Printf("Req %d done\n", i)
		}(i)
	}

	wg.Wait()

}
