package timewheel

// Job 到达时间需要执行的Job
type Job func(interface{})
