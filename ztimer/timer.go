package ztimer

import "time"

const (
	HourName = "HOUR"
	// 每小时间隔，精度为 ms
	HourInterval = 60 * 60 * 1e3
	// 12 小时制
	HourScales = 12

	MinuteName = "MINUTE"
	// 每分钟间隔，精度为 ms
	MinuteInterval = 60 * 1e3
	// 60 分钟制
	MinuteScales = 60

	SecondName = "SECOND"
	// 每秒间隔，精度为 ms
	SecondInterval = 1e3
	// 60 秒制
	SecondScales = 12
	// 每个时间轮刻度挂载定时器的最大个数
	TimersMaxCap = 2048
)

/*
   注意：
    有关时间的几个换算
   	time.Second(秒) = time.Millisecond * 1e3
	time.Millisecond(毫秒) = time.Microsecond * 1e3
	time.Microsecond(微秒) = time.Nanosecond * 1e3

	time.Now().UnixNano() ==> time.Nanosecond (纳秒)
*/

// 定时器
type Timer struct {
	// 延迟调用函数
	delayFunc *DelayFunc
	// 调用时间戳(uinx 时间，单位 ms)
	unixTs int64
}

// 返回 1970-1-1 至今经历的毫秒数
func UnixMilli() int64 {
	return time.Now().UnixNano() / 1e6
}

// 创建一个定时器，在指定时间触发定时器方法
func NewTimerAt(df *DelayFunc, unixNano int64) *Timer {
	return &Timer{
		delayFunc: df,
		unixTs:    unixNano / 1e6,
	}
}

// 创建一个定时器，在当前时间延迟 duration 之后触发定时器方法
func NewTimerAfter(df *DelayFunc, duration time.Duration) *Timer {
	return NewTimerAt(df, time.Now().UnixNano()+int64(duration))
}

func (t *Timer) Run() {
	go func() {
		now := UnixMilli()
		// 设置的定时器是否在当前调用时间之后
		if t.unixTs > now {
			// 睡眠，直到时间超时，以微秒为单位进行睡眠
			time.Sleep(time.Duration(t.unixTs-now) * time.Millisecond)
		}

		// 触发定时器方法
		t.delayFunc.Call()
	}()
}
