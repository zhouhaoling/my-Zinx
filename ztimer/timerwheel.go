package ztimer

import (
	"errors"
	"fmt"
	"github.com/dokidokikoi/my-zinx/zlog"
	"sync"
	"time"
)

/*
  tips:
	一个网络服务程序时需要管理大量客户端连接的，
	其中每个客户端连接都需要管理它的 timeout 时间。
	通常连接的超时管理一般设置为30~60秒不等，并不需要太精确的时间控制。
	另外由于服务端管理着多达数万到数十万不等的连接数，
	因此我们没法为每个连接使用一个Timer，那样太消耗资源不现实。

	用时间轮的方式来管理和维护大量的timer调度，会解决上面的问题。
*/

// 时间轮
type TimerWheel struct {
	name string
	// 刻度的时间间隔，单位 ms
	interval int64
	// 每个时间轮上的刻度数
	scales int
	// 当前时间指针的指向
	curIndex int
	// 每个刻度存放的 timer 定时器的最大容量
	maxCap int
	// 当前时间轮上所有 timer
	// 其中int表示当前时间轮的刻度
	timerQueue map[int]map[uint32]*Timer
	// 下一层时间轮
	nextTimerWheel *TimerWheel
	// 互斥锁
	sync.RWMutex
}

func NewTimerWheel(name string, interval int64, scales int, maxCap int) *TimerWheel {
	tw := &TimerWheel{
		name:       name,
		interval:   interval,
		scales:     scales,
		maxCap:     maxCap,
		timerQueue: make(map[int]map[uint32]*Timer, scales),
	}
	// 初始化 map
	for i := 0; i < scales; i++ {
		tw.timerQueue[i] = make(map[uint32]*Timer, maxCap)
	}

	zlog.Info("Init timerWheel name = ", tw.name, " is Done")
	return tw
}

/*
将一个timer定时器加入到分层时间轮中
tID: 每个定时器timer的唯一标识
t: 当前被加入时间轮的定时器
forceNext: 是否强制的将定时器添加到下一层时间轮

我们采用的算法是：
如果当前timer的超时时间间隔 大于一个刻度，那么进行hash计算 找到对应的刻度上添加
如果当前的timer的超时时间间隔 小于一个刻度 :

	如果没有下一轮时间轮
*/
func (tw *TimerWheel) addTimer(tID uint32, t *Timer, forceNext bool) error {
	defer func() error {
		if err := recover(); err != nil {
			errStr := fmt.Sprintf("addTimer function err: %v", err)
			zlog.Error(errStr)
			return errors.New(errStr)
		}
		return nil
	}()

	// 得到当前的超时时间间隔，单位 ms
	delayInterval := t.unixTs - UnixMilli()

	// 如果当前的超时时间大于一个刻度的时间间隔
	if delayInterval >= tw.interval {
		// 得到需要跨越几个刻度
		dn := delayInterval / tw.interval
		// 在对应的刻度上的定时器 Timer 集合 map 加入当前定时器
		tw.timerQueue[(tw.curIndex+int(dn))%tw.scales][tID] = t

		return nil
	}

	// 如果当前的超时时间，小于一个刻度的时间间隔，并且当前时间轮没有下一层
	if delayInterval < tw.interval && tw.nextTimerWheel == nil {
		if forceNext == true {
			/*
				如果设置为强制移至下一个刻度，那么将定时器移至下一个刻度
				这种情况，主要是时间轮自动轮转的情况
				因为这是底层时间轮，该定时器在转动的时候，如果没有被调度者取走的话，该定时器将不会再被发现
				因为时间轮刻度已经过去，如果不强制把该定时器Timer移至下时刻，就永远不会被取走并触发调用
				所以这里强制将timer移至下个刻度的集合中，等待调用者在下次轮转之前取走该定时器
			*/
			tw.timerQueue[(tw.curIndex+1)%tw.scales][tID] = t
		} else {
			// 如果手动添加定时器，那么直接将 timer 添加到对应底层时间轮的当前刻度集合中
			tw.timerQueue[tw.curIndex][tID] = t
		}
		return nil
	}

	// 如果当前超时时间，小于一个刻度的时间间隔，并且有下一轮时间轮
	if delayInterval < tw.interval {
		return tw.nextTimerWheel.AddTimer(tID, t)
	}
	return nil
}

// AddTimer 添加一个timer到一个时间轮中(非时间轮自转情况)
func (tw *TimerWheel) AddTimer(tID uint32, t *Timer) error {
	tw.Lock()
	defer tw.Unlock()

	return tw.addTimer(tID, t, false)
}

// 根据 tID 删除定时器
func (tw *TimerWheel) RemoveTimer(tID uint32) {
	tw.Lock()
	defer tw.Unlock()

	for i := 0; i < tw.scales; i++ {
		if _, ok := tw.timerQueue[i][tID]; ok {
			delete(tw.timerQueue[i], tID)
		}
	}
}

// 给一个时间轮添加下层时间轮，比如给小时时间轮添加分钟时间轮，
// 给分钟时间轮添加秒时间轮
func (tw *TimerWheel) AddTimerWheel(next *TimerWheel) {
	tw.nextTimerWheel = next
	zlog.Infof("Add timerWheel[%s]'s next [%s] is succ!", tw.name, next.name)
}

// 启动时间轮
func (tw *TimerWheel) run() {
	for {
		// 时间轮每隔 interval 一刻度时间，触发转动一次
		time.Sleep(time.Duration(tw.interval) * time.Millisecond)

		tw.Lock()
		// 取出挂载在当前刻度的全部定时器
		curTimer := tw.timerQueue[tw.curIndex]
		// 当前定时器要重新添加，所以给当前刻度再重新开辟一个 map timer 容器
		tw.timerQueue[tw.curIndex] = make(map[uint32]*Timer, tw.maxCap)
		for tID, timer := range curTimer {
			// 这里属于时间轮自动转动，forceNext 设为 true
			// 将当前刻度的定时器移动到下一级时间轮
			tw.addTimer(tID, timer, true)
		}

		// 取出下一个刻度挂载的全部定时器，进行重新添加
		nextTimers := tw.timerQueue[(tw.curIndex+1)%tw.scales]
		tw.timerQueue[(tw.curIndex+1)%tw.scales] = make(map[uint32]*Timer, tw.maxCap)
		for tID, timer := range nextTimers {
			tw.addTimer(tID, timer, true)
		}

		// 当前刻度指针走一格
		tw.curIndex = (tw.curIndex + 1) % tw.scales

		tw.Unlock()
	}
}

// Run 非阻塞的方式让时间轮转起来
func (tw *TimerWheel) Run() {
	go tw.run()
	zlog.Info("timerWheel name =", tw.name, "is running...")
}

// 获取定时器在一段时间间隔内的Timer
func (tw *TimerWheel) GetTimerWithIn(duration time.Duration) map[uint32]*Timer {
	// 最终触发定时器的一定是挂载最底层时间轮上的定时器
	// 1.找到最底层时间轮
	leafTw := tw
	for leafTw.nextTimerWheel != nil {
		leafTw = leafTw.nextTimerWheel
	}

	leafTw.Lock()
	defer leafTw.Unlock()
	// 返回 timer 的集合
	timerList := make(map[uint32]*Timer)

	now := UnixMilli()

	// 取出当前时间轮刻度内全部 timer
	for tID, timer := range leafTw.timerQueue[leafTw.curIndex] {
		if timer.unixTs-now < int64(duration/1e6) {
			// 当前定时器已经超时
			timerList[tID] = timer
			// 定时器已经超时被取走，从当前时间轮上摘除该定时器
			delete(leafTw.timerQueue[leafTw.curIndex], tID)
		}
	}

	return timerList
}
