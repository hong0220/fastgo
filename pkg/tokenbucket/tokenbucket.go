package tokenbucket

import (
	"time"
)

type Bucket struct {
	capacity int // 容量
	ch       chan bool
	timer    *time.Ticker // 定时填充token
}

func NewBucket(capacity int, interval time.Duration) *Bucket {
	bucket := &Bucket{
		capacity: capacity,
		ch:       make(chan bool, capacity),
		timer:    time.NewTicker(interval),
	}

	// 初始化令牌桶
	for i := 0; i < bucket.capacity; i++ {
		bucket.ch <- true
	}

	go bucket.startTicker()
	return bucket
}

// 定时添加令牌
func (bucket *Bucket) startTicker() {
	for {
		select {
		case <-bucket.timer.C:
			for i := len(bucket.ch); i < bucket.capacity; i++ {
				bucket.AddToken()
			}
		}
	}
}

// 添加令牌
func (bucket *Bucket) AddToken() {
	// 多判断一次
	if len(bucket.ch) < bucket.capacity {
		bucket.ch <- true
	}
}

// 获取令牌
func (bucket *Bucket) GetToken() bool {
	select {
	case <-bucket.ch:
		return true
	default:
		return false
	}
}
