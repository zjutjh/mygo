package crontab

import (
	"errors"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/robfig/cron/v3"
	"github.com/spf13/cobra"

	"github.com/zjutjh/mygo/config"
	"github.com/zjutjh/mygo/feishu"
	"github.com/zjutjh/mygo/foundation/kernel"
)

// CommandRegister 启动定时任务命令注册
func CommandRegister(jobRegister func(c *cron.Cron)) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		Run(jobRegister)
		return nil
	}
}

// Run 启动定时任务
func Run(jobRegister func(c *cron.Cron)) {
	// 获取配置
	conf := DefaultConfig
	config.Pick().UnmarshalKey("cron", &conf)

	_, err := os.OpenFile(conf.Log.ErrorFilename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Fprintln(os.Stdout, "初始化cron引擎日志错误:", err)
		os.Exit(1)
	}
	ew := &lumberjack.Logger{
		Filename:   conf.Log.ErrorFilename,
		MaxSize:    conf.Log.MaxSize,
		MaxAge:     conf.Log.MaxAge,
		MaxBackups: conf.Log.MaxBackups,
		LocalTime:  conf.Log.LocalTime,
		Compress:   conf.Log.Compress,
	}
	logger := cron.PrintfLogger(log.New(ew, "\ncron: ", log.LstdFlags))

	// 初始化cron实例
	c := cron.New(cron.WithSeconds(), cron.WithChain(Recover(logger)))

	// 启动cron
	c.Start()

	// 注册任务
	jobRegister(c)

	// 监听并等待关闭服务
	kernel.ListenStop(func() error {
		ctx := c.Stop()
		timer := time.NewTimer(conf.ShutdownWaitTimeout)
		select {
		case <-ctx.Done():
			timer.Stop()
			fmt.Fprintln(os.Stdout, "Cron关闭完成")
			return nil
		case <-timer.C:
			return errors.New("cron等待优雅处理超时, 强制关闭")
		}
	})
}

func Recover(logger cron.Logger) cron.JobWrapper {
	return func(j cron.Job) cron.Job {
		return cron.FuncJob(func() {
			defer func() {
				if r := recover(); r != nil {
					const size = 64 << 10
					buf := make([]byte, size)
					buf = buf[:runtime.Stack(buf, false)]
					err, ok := r.(error)
					if !ok {
						err = fmt.Errorf("%v", r)
					}
					logger.Error(err, "panic", "stack", "...\n"+string(buf))
					// 发送报警
					if !feishu.Exist() {
						return
					}
					go func() {
						defer func() {
							if err2 := recover(); err2 != nil {
								log.Println("请求飞书Bot发送报警发生了Panic", err2)
							}
						}()
						title := fmt.Sprintf("[%s]CronJob Panic!!!", config.AppName())
						message := fmt.Sprintf("请注意: CronJob[%#v]发生了Panic!!!\nPanic: %#v", j, r)
						feishu.Pick().Send(title, message)
					}()
				}
			}()
			j.Run()
		})
	}
}
