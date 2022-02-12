package main

import (
	"emperror.dev/emperror"
	"emperror.dev/errors"
	logurhandler "emperror.dev/handler/logur"
	"fmt"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/core/router"
	"github.com/kataras/iris/v12/mvc"
	"os"

	"github.com/sagikazarmark/appkit/buildinfo"
	"github.com/seanb4t/mu-portal-server/internal/platform/log"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
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

	app := iris.Default()
	rootAPI := rootAPI(app)
	telemetryAPI := telemetryAPI(app)
	wsAPI := wsAPI(app)

	mvcApp := mvc.New(rootAPI)
	mvcApp.Party("/telemetry").Register(telemetryAPI)
	mvcApp.Party("/ws").Register(wsAPI)
}

func rootAPI(app *iris.Application) router.Party {
	return app.Party("/")
}

func telemetryAPI(app *iris.Application) router.Party {
	return app.PartyFunc("/telemetry", func(p router.Party) {
		//p.Get("/", server.GetTelemetry)
	})
}

func wsAPI(app *iris.Application) router.Party {
	return app.Party("/ws")
}
