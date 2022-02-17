package main

import (
	"context"
	"contrib.go.opencensus.io/exporter/ocagent"
	"contrib.go.opencensus.io/exporter/prometheus"
	"emperror.dev/emperror"
	"emperror.dev/errors"
	"emperror.dev/errors/match"
	logurhandler "emperror.dev/handler/logur"
	"fmt"
	"github.com/AppsFlyer/go-sundheit/checks"
	"github.com/kataras/iris/v12"
	iriscontext "github.com/kataras/iris/v12/context"
	"github.com/kataras/iris/v12/core/router"
	"github.com/kataras/iris/v12/mvc"
	"github.com/seanb4t/mu-portal-server/internal/platform/database"
	gosundheit_logger "github.com/seanb4t/mu-portal-server/internal/platform/gosundheit"
	"github.com/AppsFlyer/go-sundheit"
	_ "go.mongodb.org/mongo-driver/mongo/readpref"
	"logur.dev/logur"
	"time"

	//"logur.dev/logur"
	"os"
	"os/signal"
	"syscall"

	health "github.com/AppsFlyer/go-sundheit"
	"github.com/cloudflare/tableflip"
	"github.com/oklog/run"
	"github.com/sagikazarmark/appkit/buildinfo"
	"github.com/seanb4t/mu-portal-server/internal/platform/log"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	appkitrun "github.com/sagikazarmark/appkit/run"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
)

// Provisioned by ldflags
// nolint: gochecknoglobals
//goland:noinspection GoUnusedGlobalVariable
var (
	version    string
	commitHash string
	buildDate  string
)

const (
	// appName is an identifier-like name used anywhere this app needs to be identified.
	//
	// It identifies the application itself, the actual instance needs to be identified via environment
	// and other details.
	appName = "mu-portal-server"

	// friendlyAppName is the visible name of the application.
	friendlyAppName = "MuSH Portal Server"
)

func main() {

	v, f := viper.New(), pflag.NewFlagSet(friendlyAppName, pflag.ExitOnError)

	configure(v, f)

	f.String("config", "", "Configuration file")
	f.Bool("version", false, "Show version information")

	_ = f.Parse(os.Args[1:])

	if v, _ := f.GetBool("version"); v {
		fmt.Printf("%s version %s (%s) built on %s\n", friendlyAppName, version, commitHash, buildDate)

		os.Exit(0)
	}

	if c, _ := f.GetString("config"); c != "" {
		v.SetConfigFile(c)
	}

	err := v.ReadInConfig()
	_, configFileNotFound := err.(viper.ConfigFileNotFoundError)
	if !configFileNotFound {
		emperror.Panic(errors.Wrap(err, "failed to read configuration"))
	}

	var config configuration
	err = v.Unmarshal(&config)
	emperror.Panic(errors.Wrap(err, "failed to unmarshal configuration"))

	err = config.Process()
	emperror.Panic(errors.WithMessage(err, "failed to process configuration"))

	// Create logger (first thing after configuration loading)
	logger := log.NewLogger(config.Log)

	// Override the global standard library logger to make sure everything uses our logger
	log.SetStandardLogger(logger)

	if configFileNotFound {
		logger.Warn("configuration file not found")
	}

	err = config.Validate()
	if err != nil {
		logger.Error(err.Error())

		os.Exit(3)
	}

	// Configure error handler
	errorHandler := logurhandler.New(logger)
	defer emperror.HandleRecover(errorHandler)

	buildInfo := buildinfo.New(version, commitHash, buildDate)

	logger.Info("starting application", buildInfo.Fields())

	healthChecker := health.New()

	app := iris.Default()

	rootAPI := rootAPI(app, logger)
	telemetryAPI(app, healthChecker, logger)
	wsAPI(app, logger)
	metricsAPI(app, config, logger, errorHandler)
	mvc.New(rootAPI)

	upg, _ := tableflip.New(tableflip.Options{})
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGHUP)
		for range ch {
			logger.Info("gracefully reloading")
			_ = upg.Upgrade()
		}
	}()

	// Connect to the database
	logger.Info("connecting to database")
	dbConnector, err := database.NewConnector(config.Database)
	emperror.Panic(errors.Wrap(err, "failed to create database connector"))

	err = dbConnector.Connect(context.Background())
	emperror.Panic(errors.Wrap(err, "failed to connect to database"))
	defer dbConnector.Disconnect(context.Background())

	_ = healthChecker.RegisterCheck(
		&checks.CustomCheck{
			CheckName: "database.check",
			CheckFunc: database.NewDbPingCheck(dbConnector),
		},
		gosundheit.InitialDelay(0),
		gosundheit.ExecutionPeriod(2*time.Minute),
		gosundheit.ExecutionTimeout(5*time.Second),
	)

	var group run.Group

	{
		const name = "app"
		//logger := logur.WithField(logger, "server", name)

		group.Add(
			func() error { return app.Run(iris.Addr(config.Server.Addr)) },
			func(err error) {
				app.Shutdown(context.Background())
			},
		)
	}

	// Setup signal handler
	group.Add(run.SignalHandler(context.Background(), syscall.SIGINT, syscall.SIGTERM))

	// Setup graceful restart
	group.Add(appkitrun.GracefulRestart(context.Background(), upg))

	err = group.Run()
	emperror.WithFilter(errorHandler, match.As(&run.SignalError{}).MatchError).Handle(err)

}

func rootAPI(app *iris.Application, logger logur.LoggerFacade) router.Party {
	return app.Party("/")
}

func telemetryAPI(app *iris.Application, healthChecker health.Health, logger logur.LoggerFacade) {
	app.PartyFunc("/telemetry", func(p router.Party) {
		// build info
		buildInfo := buildinfo.New(version, commitHash, buildDate)
		p.Get("/buildinfo", func(ctx iriscontext.Context) {
			ctx.JSON(buildInfo)
		})

		// health check

		health.WithCheckListeners(gosundheit_logger.NewLogger(logur.WithField(logger, "comonent", "healthcheck")))
		healthCheckHandler := func(ctx iriscontext.Context) {
			results, healthy := healthChecker.Results()
			ctx.Negotiation().JSON()
			if healthy {
				ctx.StatusCode(iris.StatusOK)
			} else {
				ctx.StatusCode(iris.StatusServiceUnavailable)
			}
			ctx.JSON(results)
		}
		p.Get("/health", healthCheckHandler)
		p.Get("/health/ready", healthCheckHandler)
		p.Get("/health/live", func(ctx iriscontext.Context) {
			_, healthy := healthChecker.Results()
			ctx.Negotiation().JSON()
			if healthy {
				ctx.StatusCode(iris.StatusOK)
			} else {
				ctx.StatusCode(iris.StatusServiceUnavailable)
			}
			ctx.Text("ok")
		})
	})
}

func metricsAPI(app *iris.Application, config configuration, logger logur.LoggerFacade, errorHandler *logurhandler.Handler) {
	trace.ApplyConfig(config.Opencensus.Trace.Config())
	if config.Opencensus.Exporter.Enabled {
		exporter, err := ocagent.NewExporter(append(config.Opencensus.Exporter.Options(),
			ocagent.WithServiceName(appName),
		)...)
		emperror.Panic(err)

		trace.RegisterExporter(exporter)
		view.RegisterExporter(exporter)
	}
	exporter, err := prometheus.NewExporter(prometheus.Options{
		OnError: emperror.WithDetails(
			errorHandler,
			"component", "opencensus",
			"exporter", "prometheus",
		).Handle,
	})
	emperror.Panic(err)
	view.RegisterExporter(exporter)

	app.PartyFunc("/metrics", func(p router.Party) {
		p.Get("/", func(ctx iriscontext.Context) {
			exporter.ServeHTTP(ctx.ResponseWriter(), ctx.Request())
		})
	})
}

func wsAPI(app *iris.Application, logger logur.LoggerFacade) router.Party {
	return app.Party("/ws")
}
