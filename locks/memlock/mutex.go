package memlock

import (
	"runtime"
	"sync/atomic"
	"time"
)

const (
	mutexLocked = 1 << iota
	mutexWoken
	mutexStarving
	mutexWaiterShift = iota

	starvationThresholdNs = 1e6
	maxSpins              = 4
	activeSpinCount       = 30
)

type Mutex struct {
	state int32
	ch    chan struct{} // 用于等待通知的channel
}

func NewMutex() *Mutex {
	return &Mutex{
		ch: make(chan struct{}, 1),
	}
}

// 获取锁
func (m *Mutex) Lock() {
	// 快速路径
	if atomic.CompareAndSwapInt32(&m.state, 0, mutexLocked) {
		return
	}
	m.lockSlow()
}

// 尝试获取锁
func (m *Mutex) TryLock() bool {
	return atomic.CompareAndSwapInt32(&m.state, 0, mutexLocked)
}

// 释放锁
func (m *Mutex) Unlock() {
	// 快速路径
	new := atomic.AddInt32(&m.state, -mutexLocked)
	if new != 0 {
		m.unlockSlow(new)
	}
}

// 慢路径加锁
func (m *Mutex) lockSlow() {
	var waitStartTime int64
	starving := false
	awoke := false
	iter := 0

	for {
		old := atomic.LoadInt32(&m.state)

		// 自旋阶段
		if old&(mutexLocked|mutexStarving) == mutexLocked && shouldSpin(iter, old) {
			if !awoke && old&mutexWoken == 0 && old>>mutexWaiterShift != 0 &&
				atomic.CompareAndSwapInt32(&m.state, old, old|mutexWoken) {
				awoke = true
			}
			m.doSpin()
			iter++
			continue
		}

		new := old
		// 非饥饿模式尝试获取锁
		if old&mutexStarving == 0 {
			new |= mutexLocked
		}
		// 增加等待者计数
		if old&(mutexLocked|mutexStarving) != 0 {
			new += 1 << mutexWaiterShift
		}
		// 设置唤醒标志
		if starving && old&mutexLocked != 0 {
			new |= mutexWoken
		}
		if awoke {
			new &^= mutexWoken
		}

		if atomic.CompareAndSwapInt32(&m.state, old, new) {
			if old&(mutexLocked|mutexStarving) == 0 {
				// 成功获取锁
				break
			}

			// 记录等待开始时间
			queueLifo := waitStartTime != 0
			if waitStartTime == 0 {
				waitStartTime = time.Now().UnixNano()
			}

			// 进入等待
			m.semacquire(queueLifo)

			// 被唤醒后
			starving = starving || time.Now().UnixNano()-waitStartTime > starvationThresholdNs
			old = atomic.LoadInt32(&m.state)
			if old&mutexStarving != 0 {
				// 饥饿模式处理
				delta := int32(mutexLocked - 1<<mutexWaiterShift)
				if !starving || old>>mutexWaiterShift == 1 {
					delta -= mutexStarving
				}
				atomic.AddInt32(&m.state, delta)
				break
			}
			awoke = true
			iter = 0
		} else {
			old = atomic.LoadInt32(&m.state)
		}
	}
}

// 慢路径解锁
func (m *Mutex) unlockSlow(new int32) {
	if (new+mutexLocked)&mutexLocked == 0 {
		panic("sync: unlock of unlocked mutex")
	}

	old := new
	for {
		// 如果没有等待者，或者有其他状态，直接返回
		if old>>mutexWaiterShift == 0 || old&(mutexLocked|mutexWoken|mutexStarving) != 0 {
			return
		}

		// 尝试唤醒一个等待者
		new = (old - 1<<mutexWaiterShift) | mutexWoken
		if atomic.CompareAndSwapInt32(&m.state, old, new) {
			m.semrelease()
			return
		}
		old = atomic.LoadInt32(&m.state)
	}
}

// 信号量获取
func (m *Mutex) semacquire(lifo bool) {
	// 非阻塞检查一次
	old := atomic.LoadInt32(&m.state)
	if old&mutexLocked == 0 {
		if atomic.CompareAndSwapInt32(&m.state, old, old|mutexLocked) {
			return
		}
	}

	// 使用channel等待通知
	select {
	case <-m.ch:
		// 被唤醒，再次尝试获取锁
		return
	default:
		// channel空，等待通知
		<-m.ch
	}
}

// 信号量释放 - 发送通知
func (m *Mutex) semrelease() {
	// 非阻塞发送通知
	select {
	case m.ch <- struct{}{}:
	default:
		// channel已满，说明已经有goroutine被唤醒
	}
}

// 判断是否应该自旋
func shouldSpin(iter int, state int32) bool {
	if iter >= maxSpins {
		return false
	}
	if runtime.GOMAXPROCS(0) == 1 {
		return false
	}
	if runtime.NumGoroutine() > runtime.GOMAXPROCS(0)*2 {
		return false
	}
	return state&mutexLocked != 0 && state&mutexStarving == 0
}

// 执行自旋
func (m *Mutex) doSpin() {
	for i := 0; i < activeSpinCount; i++ {
		runtime.Gosched()
	}
}
