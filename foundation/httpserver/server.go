package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/pprof"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/zjutjh/mygo/config"
	"github.com/zjutjh/mygo/feishu"
	"github.com/zjutjh/mygo/foundation/kernel"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/kit"
)

// CommandRegister 启动HTTP Server命令注册
func CommandRegister(routeRegister func(engine *gin.Engine)) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		StartHTTPServer(routeRegister)
		return nil
	}
}

// StartHTTPServer 启动HTTP Server
func StartHTTPServer(routeRegister func(*gin.Engine)) {
	// 获取配置
	conf := DefaultConfig
	config.Pick().UnmarshalKey("http_server", &conf)

	// 初始化gin引擎
	engine, err := initGinEngine(conf)
	if err != nil {
		fmt.Fprintln(os.Stdout, "初始化Gin Engine失败:", err)
		os.Exit(1)
	}

	// 注册路由
	routeRegister(engine)

	// 初始化http server
	server := initHTTPServer(engine, conf)

	// 启动http server
	go listenHTTPServer(server)

	// 监听等待关闭服务
	kernel.ListenStop(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), conf.ShutdownWaitTimeout)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			return fmt.Errorf("HTTP Server等待优雅处理超时, 错误: %w", err)
		} else {
			fmt.Fprintln(os.Stdout, "HTTP Server关闭完成")
		}
		return nil
	})
}

func initGinEngine(conf Config) (*gin.Engine, error) {
	aw, ew, err := initGinLoggerWriter(conf)
	if err != nil {
		return nil, err
	}

	// 创建gin引擎实例
	engine := gin.New()

	// 设置gin参数 有需要时再补充
	// engine.RedirectTrailingSlash = conf.Gin.RedirectTrailingSlash
	// ......

	// 创建gin全局日志记录
	logger := gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: accessLoggerFormatter(),
		Output:    aw,
	})
	// 设置gin崩溃恢复中间件
	recovery := gin.RecoveryWithWriter(ew, recoveryHandler)

	// 设置gin全局中间件
	engine.Use(requestid.New(), logger, recovery)

	// 设置gin pprof
	if conf.Pprof {
		pprof.Register(engine, fmt.Sprintf("%s%s", config.AppName(), pprof.DefaultPrefix))
	}

	return engine, nil
}

func accessLoggerFormatter() gin.LogFormatter {
	return func(param gin.LogFormatterParams) string {
		if param.Latency > time.Minute {
			param.Latency = param.Latency.Truncate(time.Second)
		}

		data := map[string]any{
			"app":         config.AppName(),
			"time":        param.TimeStamp.UnixMilli(),
			"ts":          param.TimeStamp.Format(time.DateTime),
			"api":         param.Path,
			"method":      param.Method,
			"client_ip":   param.ClientIP,
			"query":       param.Request.URL.Query(),
			"header":      param.Request.Header,
			"latency":     param.Latency.String(),
			"status_code": param.StatusCode,
		}
		if param.ErrorMessage != "" {
			data["error"] = param.ErrorMessage
		}
		db, _ := json.Marshal(data)
		return fmt.Sprintf("%s\n", string(db))
	}
}

func recoveryHandler(ctx *gin.Context, err any) {
	reply.Fail(ctx, kit.CodeUnknownError)
	// 发送报警
	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Println("请求飞书Bot发送报警发生了Panic:", err)
			}
		}()
		title := fmt.Sprintf("[%s]HTTP Server Panic!!!", config.AppName())
		message := fmt.Sprintf("请注意: HTTP Server发生了Panic!!!\nPanic: %#v", err)
		feishu.Pick().Send(title, message)
	}()
}

func initHTTPServer(e *gin.Engine, conf Config) *http.Server {
	var handler http.Handler = e

	if conf.H2C.Enable {
		h2s := &http2.Server{}
		handler = h2c.NewHandler(e, h2s)
	}

	return &http.Server{
		Addr:    conf.Addr,
		Handler: handler,
	}
}

func listenHTTPServer(s *http.Server) {
	err := s.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		fmt.Fprintln(os.Stdout, "启动HTTP Server失败:", err)
	}
}
