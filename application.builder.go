package appbox

import (
	"context"
	"embed"
	"fmt"
	"github.com/Domingor/go-blackbox/appioc"
	"github.com/Domingor/go-blackbox/apputils/apptoken"
	"github.com/Domingor/go-blackbox/seed"
	"github.com/Domingor/go-blackbox/server/cache"
	"github.com/Domingor/go-blackbox/server/cronjobs"
	"github.com/Domingor/go-blackbox/server/datasource"
	"github.com/Domingor/go-blackbox/server/loadconf"
	"github.com/Domingor/go-blackbox/server/mongodb"
	"github.com/Domingor/go-blackbox/server/webiris"
	log "github.com/Domingor/go-blackbox/server/zaplog"
	"net/http"
	"time"
)

const (
	// TimeFormat 日期格式化
	TimeFormat = "2006-01-02 15:04:05"
)

// ApplicationBuilder app builder接口提供系统初始化服务基础功能
type ApplicationBuilder interface {
	EnableWeb(timeFormat, port, logLevel string, components webiris.PartyComponent) *ApplicationBuild // 启动web服务
	EnableDb(dbConfig *datasource.PostgresConfig, models []interface{}) *ApplicationBuild             // 启动数据库
	EnableCache(ctx context.Context, redConfig cache.RedisOptions) *ApplicationBuild                  // 启动缓存
	LoadConfig(configStruct interface{}, loaderFun func(loadconf.Loader)) error                       // 加载配置文件、环境变量等
	InitLog(outDirPath, level string) *ApplicationBuild                                               // 初始化日志打印
	EnableMongoDB(dbConfig *mongodb.MongoDBConfig) *ApplicationBuild                                  // 启动缓存数据库
	InitCronJob() *ApplicationBuild                                                                   // 初始化定时任务
	SetupToken(AMinute, RHour time.Duration, TokenIssuer string) *ApplicationBuild                    // 配置wen-token属性
	EnableStaticSource() *ApplicationBuild                                                            // 加载静态资源
}

type ApplicationBuild struct {
	// 创建Iris实例对象
	irisApp webiris.WebBaseFunc

	// 启动种子list集合
	seeds []seed.SeedFunc

	// 是否启动定时服务，在enableCronjob后为true，会自动start()，即开始调用定时Cron表达式函数
	IsRunningCronJob bool

	isLoadingStaticFs bool
	// 静态服务文件系统
	StaticFs http.FileSystem
	// 是否开启web
	IsEnableWeb bool
	// 是否开启数据库
	IsEnableDB bool
	// 是否开启redis
	IsEnableCache bool
	// 是否开始RabbitMq
	IsEnableRabbitMq bool
	// 是否开始定时任务
	IsEnableCronTask bool
	// 是否开启mongoDB
	IsEnableMongoDB bool
	// 是否开启静态服务文件
	IsEnableStaticFileServe bool
}

// EnableWeb 启动Web服务
func (app *ApplicationBuild) EnableWeb(timeFormat, port, logLevel string, components webiris.PartyComponent) *ApplicationBuild {
	// 初始化iris对象
	app.irisApp = webiris.Init(
		timeFormat, // 日期格式化
		port,       // 监听服务端口
		logLevel,   // 日志级别
		components) // router路由组件

	// 全局上下文对象
	getContext := appioc.GetContext().Ctx

	// 开启协程监听TCP-wen端口服务
	go func() {
		log.SugaredLogger.Info("starting web service...")

		// 判断是否加载静态文件
		if app.isLoadingStaticFs {
			err := app.irisApp.StaticSource(app.StaticFs)
			if err != nil {
				log.SugaredLogger.Debug("app.irisApp.StaticSource fail!")
				return
			}
		}
		// 启动web，此时会阻塞。后面的代码不会被轮到执行
		err := app.irisApp.Run(getContext)

		if err != nil {
			log.SugaredLogger.Infof("run web service error! %s", err)
		}
	}()
	return app
}

// EnableDb 启动数据库操作对象
func (app *ApplicationBuild) EnableDb(dbConfig *datasource.PostgresConfig, models ...interface{}) *ApplicationBuild {
	//	// 初始化数据，注册模型
	datasource.GormInit(dbConfig, models)

	// 放入容器
	appioc.Set(datasource.GetDbInstance())
	return app
}

// EnableCache 启动缓存
func (app *ApplicationBuild) EnableCache(ctx context.Context, redConfig cache.RedisOptions) *ApplicationBuild {
	// 初始化redis，放入容器
	appioc.Set(cache.Init(ctx, redConfig))
	return app
}

// LoadConfig 加载配置文件、环境变量值
func (app *ApplicationBuild) LoadConfig(configStruct interface{}, loaderFun func(loadconf.Loader)) error {
	loader := loadconf.NewLoader()
	if loaderFun == nil {
		return fmt.Errorf("loaderFun is nil")
	}

	// 加载解析配置文件属性
	loaderFun(loader)

	// 读取到的属性值赋值给配置对象
	err := loader.LoadToStruct(configStruct)
	return err
}

// InitLog 初始化自定义日志
func (app *ApplicationBuild) InitLog(outDirPath, level string) *ApplicationBuild {

	if len(outDirPath) > 0 {
		log.CONFIG.Director = outDirPath
	}

	if len(level) > 0 {
		log.CONFIG.Level = level
	}

	// 初始化日志，通过zaplog.日志对象进行调用
	log.Init()
	return app
}

// EnableMongoDB 启动MongoDB客户端
func (app *ApplicationBuild) EnableMongoDB(dbConfig *mongodb.MongoDBConfig) *ApplicationBuild {
	client, err := mongodb.GetClient(dbConfig, appioc.GetContext().Ctx)
	if err != nil {
		log.SugaredLogger.Debugf("init mongoDb fail err %s", err)
	}
	// mongoDb客户端放入容器
	appioc.Set(client)
	return app
}

// SetSeeds 设置启动项目时，要执行的一些钩子函数
func (app *ApplicationBuild) SetSeeds(seedFuncs ...seed.SeedFunc) *ApplicationBuild {
	app.seeds = append(app.seeds, seedFuncs...)
	return app
}

// InitCronJob 初始化定时任务对象，存放入IOC
func (app *ApplicationBuild) InitCronJob() *ApplicationBuild {
	instance := cronjobs.CronInstance()
	// 设置启动定时任务
	app.IsRunningCronJob = true

	// 定时任务客户端放入容器
	appioc.Set(instance)
	return app
}

// SetupToken 设置系统token有效期
func (app *ApplicationBuild) SetupToken(AMinute, RHour time.Duration, TokenIssuer string) *ApplicationBuild {
	apptoken.Init(AMinute, RHour, TokenIssuer)
	return app
}

// EnableStaticSource  加载web服务静态资源文件
func (app *ApplicationBuild) EnableStaticSource(file embed.FS) *ApplicationBuild {
	// 封装为 Https文件系统
	app.isLoadingStaticFs = true
	app.StaticFs = http.FS(file)

	return app
}
