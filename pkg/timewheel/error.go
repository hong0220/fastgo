package timewheel

import "errors"

// ErrDuplicateTaskKey task key 重复
var ErrDuplicateTaskKey = errors.New("duplicate task key")

// ErrTaskKeyNotFount task key 不存在
var ErrTaskKeyNotFount = errors.New("task key doesn't existed in task list, please check your input")
