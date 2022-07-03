package time_wheel

import (
	"time"
)

// Task 时间轮上需要执行的任务
type Task struct {
	// 对象标识，唯一值
	key interface{}

	// 任务执行周期
	executeCycle time.Duration
	// 任务创建时间
	createdTime time.Time

	// 任务在时间轮的位置
	position int
	// 任务在时间轮走多少圈才能开始执行
	circle int
	// 任务执行次数，如果一直执行设置成-1
	times int

	// 任务需要执行的 Job，优先级高于 TimeWheel 的 Job
	job Job
}
