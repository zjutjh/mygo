package memlock

import "time"

type Options struct {
	//乐观自旋上限
	MaxSpin         int
	MaxSpinDuration time.Duration
	//饥饿阈值
	StarvationThreshold time.Duration
}

var DefaultOpts = Options{
	MaxSpin:             500,
	MaxSpinDuration:     100 * time.Microsecond,
	StarvationThreshold: 1 * time.Millisecond,
}
