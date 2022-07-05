package timewheel

import (
	"errors"
	"fmt"
	"github.com/hong0220/fastgo/pkg/timewheel"
	"testing"
	"time"
)

var fiveSecond int64 = 5
var twoSecond int64 = 2

func Test_TimeWheel(t *testing.T) {
	// 初始化一个时间间隔是1s，一共有60个齿轮的时间轮，默认时间轮转动一圈的时间是60s
	timeWheel := timewheel.CreateTimeWheel(1*time.Second, 60, TimeWheelDefaultJob)

	// 启动时间轮
	timeWheel.Start()

	if timeWheel.IsRunning() {
		// 添加一个task，每隔5s执行一次，task名字叫task1，task的创建时间是time.Now()
		// task执行的任务设置为nil，所以默认执行 timewheel 的Job，也就是 TimeWheelDefaultJob
		fmt.Println(fmt.Sprintf("%v Add task task-5s", time.Now().Format(time.RFC3339)))
		err := timeWheel.AddTask(time.Duration(fiveSecond)*time.Second, "task-5s", time.Now(), -1, nil)
		if err != nil {
			panic(err)
		}

		// Task执行 TaskJob
		fmt.Println(fmt.Sprintf("%v Add task task-2s", time.Now().Format(time.RFC3339)))
		err = timeWheel.AddTask(time.Duration(twoSecond)*time.Second, "task-2s", time.Now(), -1, TaskJob)
		if err != nil {
			panic(err)
		}
	} else {
		panic("TimeWheel is not running")
	}

	time.Sleep(60 * time.Second)

	// 删除task
	fmt.Println("Remove task task-5s")
	err := timeWheel.RemoveTask("task-5s")
	if err != nil {
		panic(err)
	}

	time.Sleep(60 * time.Second)

	// 删除task
	fmt.Println("Remove task task-2s")
	err = timeWheel.RemoveTask("task-2s")
	if err != nil {
		panic(err)
	}

	// 关闭时间轮
	timeWheel.Stop()
}

var lastTimeDefault time.Time

func TimeWheelDefaultJob(key interface{}) {
	currentTime := time.Now()
	if !lastTimeDefault.IsZero() {
		//fmt.Printf("key=%s, current=%v, last=%v, diff=%v\n", key,
		//	currentTime.Unix(), lastTimeDefault.Unix(), currentTime.Unix()-lastTimeDefault.Unix())
		if currentTime.Unix()-lastTimeDefault.Unix() != fiveSecond {
			panic(errors.New("TimeWheelDefaultJob 时间轮异常"))
		}
	}
	lastTimeDefault = currentTime
	fmt.Printf("%v This is a timewheel job with key: %v\n", currentTime.Format(time.RFC3339), key)
}

var lastTime time.Time

func TaskJob(key interface{}) {
	currentTime := time.Now()
	if !lastTime.IsZero() {
		//fmt.Printf("key=%s, current=%v, last=%v, diff=%v\n", key, currentTime.Unix(), lastTime.Unix(),
		//	currentTime.Unix()-lastTime.Unix())
		if currentTime.Unix()-lastTime.Unix() != twoSecond {
			panic(errors.New("TaskJob 时间轮异常"))
		}
	}
	lastTime = currentTime
	fmt.Printf("%v This is a timewheel job with key: %v\n", currentTime.Format(time.RFC3339), key)
}
