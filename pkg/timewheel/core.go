package timewheel

import (
	"container/list"
	"errors"
	"fmt"
	"sync"
	"time"
)

var timeWheel *TimeWheel
var once sync.Once

// CreateTimeWheel 实现 TimeWheel 单例模式
func CreateTimeWheel(interval time.Duration, slotNums int, job Job) *TimeWheel {
	once.Do(func() {
		timeWheel = New(interval, slotNums, job)
	})
	return timeWheel
}

// New 初始化一个 TimeWheel 对象
func New(interval time.Duration, slotNums int, job Job) *TimeWheel {
	if interval <= 0 || slotNums <= 0 {
		return nil
	}

	timeWheel = &TimeWheel{
		interval:          interval,
		slotNums:          slotNums,
		currentPosition:   0,
		slots:             make([]*list.List, slotNums),
		addTaskChannel:    make(chan *Task),
		removeTaskChannel: make(chan *Task),
		stopChannel:       make(chan bool),
		taskRecords:       &sync.Map{},
		job:               job,
		isRunning:         false,
	}

	timeWheel.initSlots()
	return timeWheel
}

// 初始化时间轮，每个轮上的卡槽用一个双向队列表示，便于插入和删除
func (timeWheel *TimeWheel) initSlots() {
	for i := 0; i < timeWheel.slotNums; i++ {
		timeWheel.slots[i] = list.New()
	}
}

// GetTimeWheel 返回 TimeWheel 对象
func GetTimeWheel() *TimeWheel {
	return timeWheel
}

// Start 启动时间轮
func (timeWheel *TimeWheel) Start() {
	timeWheel.isRunning = true

	timeWheel.ticker = time.NewTicker(timeWheel.interval)
	go timeWheel.start()
}

// 启动时间轮
func (timeWheel *TimeWheel) start() {
	for {
		select {

		case <-timeWheel.ticker.C:
			timeWheel.rollAndRunTask()

		case task := <-timeWheel.addTaskChannel:
			// 如果通过任务的周期定位位置，在服务重启时，任务周期相同的点会被定位到相同的卡槽，造成任务过度集中
			// 利用 Task.createTime 定位任务在时间轮的位置和执行圈数
			timeWheel.addTask(task, false)

		case task := <-timeWheel.removeTaskChannel:
			timeWheel.removeTask(task)

		case <-timeWheel.stopChannel:
			timeWheel.ticker.Stop()

			// 终止流程
			return
		}
	}
}

// 时间检查时间轮点位上的Task，看哪个需要执行
func (timeWheel *TimeWheel) rollAndRunTask() {
	// 获取时间轮位置上的双向链表
	currentList := timeWheel.slots[timeWheel.currentPosition]

	if currentList != nil {
		for item := currentList.Front(); item != nil; {
			task := item.Value.(*Task)

			// 表盘指针每前进一格，该处的任务列表中所有的任务的 circle 都-1, 如果该处的任务 circle = 0，那么执行该任务。

			// 如果执行圈数 > 0，表示没到执行时间
			if task.circle > 0 {
				// 更新执行圈数
				task.circle--

				// 遍历下一个 Task
				item = item.Next()

				continue
			}

			// 执行任务时，Task.job 是第一优先级，然后是 TimeWheel.job
			if task.job != nil {
				go task.job(task.key)
			} else if timeWheel.job != nil {
				go timeWheel.job(task.key)
			} else {
				fmt.Println(fmt.Sprintf("The task %d don't have job to run", task.key))
			}

			// 临时存储
			next := item.Next()

			// 执行完成以后，将该任务从时间轮删除
			timeWheel.taskRecords.Delete(task.key)
			// 移除，注意 e.next = nil
			currentList.Remove(item)

			// 遍历下一个 Task
			item = next

			// 重新添加任务到时间轮
			if task.times != 0 {
				// -1表示一直执行
				if task.times < 0 {
					timeWheel.addTask(task, true)
				} else {
					task.times--
					timeWheel.addTask(task, true)
				}
			} else {
				// 如果 times == 0，说明已经完成执行周期，不需要再添加任务回时间轮
			}
		}
	}

	// 时间轮前进一步
	if timeWheel.currentPosition == timeWheel.slotNums-1 {
		timeWheel.currentPosition = 0
	} else {
		timeWheel.currentPosition++
	}
}

// Stop 关闭时间轮
func (timeWheel *TimeWheel) Stop() {
	timeWheel.stopChannel <- true
	timeWheel.isRunning = false
}

// IsRunning 时间轮是否正常运行
func (timeWheel *TimeWheel) IsRunning() bool {
	return timeWheel.isRunning
}

// AddTask 时间轮添加任务
// @param executeCycle    任务周期
// @param key         任务key，必须是唯一，否则添加任务会失败
// @param createTime  任务创建时间
func (timeWheel *TimeWheel) AddTask(interval time.Duration, key interface{}, createdTime time.Time, times int, job Job) error {
	if interval <= 0 || key == nil {
		return errors.New("invalid task params")
	}

	// 检查 Task.Key 是否重复
	_, ok := timeWheel.taskRecords.Load(key)
	if ok {
		return ErrDuplicateTaskKey
	}

	timeWheel.addTaskChannel <- &Task{
		key:          key,
		executeCycle: interval,
		createdTime:  createdTime,
		job:          job,
		times:        times,
	}

	return nil
}

// 添加任务
// @param task       Task  Task对象
// @param byInterval bool  生成Task在时间轮的位置和执行圈数的方式，true表示利用 Task.executeCycle 生成，false表示利用 Task.createTime 生成
func (timeWheel *TimeWheel) addTask(task *Task, byInterval bool) {
	var position, circle int
	if byInterval {
		position, circle = timeWheel.getPosAndCircleByInterval(task.executeCycle)
	} else {
		position, circle = timeWheel.getPosAndCircleByCreatedTime(task.createdTime, task.executeCycle)
	}

	task.position = position
	task.circle = circle

	element := timeWheel.slots[position].PushBack(task)
	timeWheel.taskRecords.Store(task.key, element)
}

// 通过任务的周期计算下次执行的执行圈数和位置
func (timeWheel *TimeWheel) getPosAndCircleByInterval(d time.Duration) (int, int) {
	executeCycleSeconds := int(d.Seconds())
	intervalSeconds := int(timeWheel.interval.Seconds())

	// 执行圈数
	circle := executeCycleSeconds / intervalSeconds / timeWheel.slotNums
	// 位置
	position := (timeWheel.currentPosition + executeCycleSeconds/intervalSeconds) % timeWheel.slotNums

	// 特殊case，当计算的位置和当前位置重叠时，因为当前位置已经走过，所以circle需要减一
	if position == timeWheel.currentPosition && circle != 0 {
		circle--
	}

	return position, circle
}

// 通过任务的创建时间计算下次执行的执行圈数和位置
func (timeWheel *TimeWheel) getPosAndCircleByCreatedTime(createdTime time.Time, d time.Duration) (int, int) {
	passedSeconds := int(time.Since(createdTime).Seconds())

	executeCycleSeconds := int(d.Seconds())
	intervalSeconds := int(timeWheel.interval.Seconds())

	circle := executeCycleSeconds / intervalSeconds / timeWheel.slotNums
	position := (timeWheel.currentPosition + (executeCycleSeconds-(passedSeconds%executeCycleSeconds))/intervalSeconds) % timeWheel.slotNums

	// 特殊case，当计算的位置和当前位置重叠时，因为当前位置已经走过，所以circle需要减一
	if position == timeWheel.currentPosition && circle != 0 {
		circle--
	}

	return position, circle
}

// RemoveTask 时间轮删除任务
func (timeWheel *TimeWheel) RemoveTask(key interface{}) error {
	if key == nil {
		return nil
	}

	// 检查Task是否存在
	val, ok := timeWheel.taskRecords.Load(key)
	if !ok {
		return ErrTaskKeyNotFount
	}

	task := val.(*list.Element).Value.(*Task)
	timeWheel.removeTaskChannel <- task

	return nil
}

// 删除任务
func (timeWheel *TimeWheel) removeTask(task *Task) error {
	val, _ := timeWheel.taskRecords.Load(task.key)

	// 从map结构中删除
	timeWheel.taskRecords.Delete(task.key)

	// 通过 TimeWheel.slots 获取任务
	currentList := timeWheel.slots[task.position]
	currentList.Remove(val.(*list.Element))

	return nil
}
