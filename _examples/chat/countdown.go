package main

import (
	"time"

	"github.com/yeeaiclub/fasttui"
)

type CountdownTimer struct {
	timeout   int
	tui       fasttui.Terminal
	onTick    func(int)
	onTimeout func()
	timer     *time.Timer
	ticker    *time.Ticker
	stopChan  chan bool
}

func NewCountdownTimer(
	timeout int,
	tui fasttui.Terminal,
	onTick func(int),
	onTimeout func(),
) *CountdownTimer {
	c := &CountdownTimer{
		timeout:   timeout,
		tui:       tui,
		onTick:    onTick,
		onTimeout: onTimeout,
		stopChan:  make(chan bool),
	}
	c.start()
	return c
}

func (c *CountdownTimer) start() {
	remaining := c.timeout

	c.ticker = time.NewTicker(time.Second)
	c.timer = time.NewTimer(time.Duration(c.timeout) * time.Second)

	go func() {
		for {
			select {
			case <-c.ticker.C:
				remaining--
				if c.onTick != nil {
					c.onTick(remaining)
				}
			case <-c.timer.C:
				if c.onTimeout != nil {
					c.onTimeout()
				}
				return
			case <-c.stopChan:
				return
			}
		}
	}()
}

func (c *CountdownTimer) Dispose() {
	if c.ticker != nil {
		c.ticker.Stop()
	}
	if c.timer != nil {
		c.timer.Stop()
	}
	close(c.stopChan)
}
