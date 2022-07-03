package token_bucket

import (
	"fmt"
	"github.com/hong0220/fastgo/pkg/token_bucket"
	"testing"
	"time"
)

func Test_TokenBucket(t *testing.T) {
	bucket := token_bucket.NewBucket(5, time.Second)
	for i := 0; i < 1000; i++ {
		time.Sleep(time.Millisecond * 100)
		go DoFunc(bucket, i)
	}
	for {
		fmt.Println("....")
	}
}

func DoFunc(bucket *token_bucket.Bucket, index int) {
	if bucket.GetToken() {
		fmt.Printf("getToken success : %d\n", index)
	} else { // 丢弃
		fmt.Printf("getToken fail : %d\n", index)
	}
}
