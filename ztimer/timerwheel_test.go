package ztimer

import (
	"fmt"
	"testing"
	"time"
)

func TestTimerWheel(t *testing.T) {
	// 创建秒级时间轮
	secondTw := NewTimerWheel(SecondName, SecondInterval, SecondScales, TimersMaxCap)
	// 创建分钟级时间轮
	minuteTw := NewTimerWheel(MinuteName, MinuteInterval, MinuteScales, TimersMaxCap)
	// 创建小时级时间轮
	hourTw := NewTimerWheel(HourName, HourInterval, HourScales, TimersMaxCap)

	// 将分层时间轮做关联
	hourTw.AddTimerWheel(minuteTw)
	minuteTw.AddTimerWheel(secondTw)

	fmt.Println("init timerWheels done!")

	// 给时间轮添加定时器
	timer1 := NewTimerAfter(NewDelayFunc(myFunc, []interface{}{1, 10}), 10*time.Second)
	hourTw.AddTimer(1, timer1)
	fmt.Println("add timer 1 done!")

	//给时间轮添加定时器
	timer2 := NewTimerAfter(NewDelayFunc(myFunc, []interface{}{2, 20}), 20*time.Second)
	_ = hourTw.AddTimer(2, timer2)
	fmt.Println("add timer 2 done!")

	//给时间轮添加定时器
	timer3 := NewTimerAfter(NewDelayFunc(myFunc, []interface{}{3, 30}), 30*time.Second)
	_ = hourTw.AddTimer(3, timer3)
	fmt.Println("add timer 3 done!")

	//给时间轮添加定时器
	timer4 := NewTimerAfter(NewDelayFunc(myFunc, []interface{}{4, 40}), 40*time.Second)
	_ = hourTw.AddTimer(4, timer4)
	fmt.Println("add timer 4 done!")

	//给时间轮添加定时器
	timer5 := NewTimerAfter(NewDelayFunc(myFunc, []interface{}{5, 50}), 50*time.Second)
	_ = hourTw.AddTimer(5, timer5)
	fmt.Println("add timer 5 done!")

	//时间轮运行
	secondTw.Run()
	minuteTw.Run()
	hourTw.Run()

	fmt.Println("timerWheels are running!")

	go func() {
		n := 0.0
		for {
			fmt.Println("tick...", n)

			//取出近1ms的超时定时器有哪些
			timers := hourTw.GetTimerWithIn(1000 * time.Millisecond)
			for _, timer := range timers {
				//调用定时器方法
				timer.delayFunc.Call()
			}

			time.Sleep(500 * time.Millisecond)
			n += 0.5
		}
	}()

	//主进程等待其他go，由于Run()方法是用一个新的go承载延迟方法，这里不能用waitGroup
	time.Sleep(10 * time.Minute)
}
