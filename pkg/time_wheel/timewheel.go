package time_wheel

import (
	"container/list"
	"sync"
	"time"
)

// TimeWheel 时间轮
type TimeWheel struct {
	// 时间轮的精度，时间轮每前进一步所需要的时间
	interval time.Duration
	// 时间轮的总齿轮数 interval * slotNums，表示时间轮转一圈走过的时间
	slotNums int
	// 时间轮当前的位置
	currentPosition int

	// 利用数组来实现时间轮，数组中每个元素是个双向链表，用来存储要执行的任务Task
	slots []*list.List

	// 时钟计时器，定时触发
	ticker *time.Ticker

	addTaskChannel    chan *Task
	removeTaskChannel chan *Task
	stopChannel       chan bool

	// 存储任务Task对象，key是任务Task key，value是任务Task对象，value结构是list.Element
	taskRecords *sync.Map

	// 需要执行的任务
	// 如果时间轮上的 Task 执行同一个 Job，可以直接实例化到 TimeWheel 结构体中
	// 此处的优先级低于 Task 中的 Job 参数
	job Job

	// 时间轮是否 running 状态，避免重复启动
	isRunning bool
}
