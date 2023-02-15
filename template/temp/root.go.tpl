package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "{{ .ProjectName }}",
	Short:   "example for gin app",
	Long:    "",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "app.yaml", "config file")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.SetConfigFile(cfgFile)

	viper.AutomaticEnv() // read in environment variables that match

	viper.SetDefault("listenAddr", ":8080")

	fmt.Println("Using config file:", viper.ConfigFileUsed())
    if err := viper.ReadInConfig(); err != nil {
    	panic(err)
    }
    if err := config.InitConfig(viper.GetViper()); err != nil {
    	panic(err)
    }
}

func SetupApp(serviceType string) {
	InitDB()
	InitIDHasher()
	InitRedis()
	InitCeph()
	InitCaller()
	InitWorker()
	InitEs()
}

func InitRedis() {
	logger.Info("init redis")
	if err := cache.InitConnect(config.Cfg.Database.RedisInstanceURL); err != nil {
		panic(err)
	}
}

func InitIDHasher() {
	logger.Info("init id hasher")
	if err := hid.InitHasher(config.Cfg.Hasher); err != nil {
		panic(err)
	}
}

func InitCeph() {
	logger.Info("init ceph")
	if err := ceph.InitAwsClient(config.Cfg.Ceph); err != nil {
		panic(err)
	}
}

func InitDB() {
	logger.Info("init db")
	if err := dbmodels.Connect(config.Cfg.Database.DB); err != nil {
		panic(err)
	}
}

func InitCaller() {
	logger.Info("init caller")
	// TODO
}

func InitTracer(serverType string) {
	tracer.InitTracer(serverType, "marking")
}

func InitWorker() {
	logger.Info("init worker")
	if err := workers.Initialize(config.Cfg.Database.WorkerRedisURL); err != nil {
		panic(err)
	}
}

func InitEs() {
	logger.Info("init es")
	if err := es.Init(config.Cfg.Es.Host, config.Cfg.Es.Username, config.Cfg.Es.Password); err != nil {
		panic(err)
	}
}
