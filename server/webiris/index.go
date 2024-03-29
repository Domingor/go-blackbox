package webiris

import (
	"context"
	"fmt"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/recover"
	"net/http"
)

/**
* @Author: Connor
* @Date:   23.3.22 15:53
* @Description:
 */

// PartyComponent 路由组件
type PartyComponent func(app *iris.Application)

type WebBaseFunc interface {
	Run(ctx context.Context) error
	StaticSource(fs http.FileSystem) error
}

type WebIris struct {
	app        *iris.Application
	port       string // 监听端口地址
	timeFormat string // 时间格式化
}

// Init 初始化iris配置
func Init(timeFormat, port, logLevel string, components PartyComponent) *WebIris {
	// 创建iris实例
	application := iris.New()

	// 一个可以让程序从任意的 http-relative panics 中恢复过来，
	application.Use(recover.New())

	// 日志级别
	application.Logger().SetLevel(logLevel)

	if components != nil {
		// 注册路路由
		components(application)
	}

	// 返回WebIris实例
	return &WebIris{
		app:        application,
		port:       port,
		timeFormat: timeFormat,
	}
}

//func (w *WebIris) shutdownFuture(ctx context.Context) {
//	if ctx == nil {
//		return
//	}
//	var c context.Context
//	var cancel context.CancelFunc
//	defer func() {
//		if cancel != nil {
//			cancel()
//		}
//	}()
//	for {
//		select {
//		case <-ctx.Done():
//			c = context.TODO()
//			if err := w.app.Shutdown(c); nil != err {
//			}
//			return
//		default:
//			time.Sleep(time.Millisecond * 500)
//		}
//	}
//}

// Run 启动iris服务并监听端口
func (w *WebIris) Run(ctx context.Context) (err error) {
	// 启动web服务，监听端口（阻塞）
	err = w.app.Listen(w.port,
		iris.WithoutInterruptHandler,
		iris.WithoutServerError(iris.ErrServerClosed),
		iris.WithOptimizations,
		iris.WithTimeFormat(w.timeFormat))
	fmt.Println(err)
	return
}

// StaticSource 配置静态文件访问路径
func (w *WebIris) StaticSource(fs http.FileSystem) (err error) {
	// 添加静态资源，如vue打包后的asset/*,index.html 可直接通过服务 / 访问静态网站
	w.app.HandleDir("/", fs)
	return
}
