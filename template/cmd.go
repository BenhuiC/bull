package template

import "text/template"

var CmdMap = map[string]*template.Template{
	"root":   template.Must(template.New("root").Parse(RootCmd())),
	"server": template.Must(template.New("server").Parse(ServerCmd())),
}

// RootCmd projectDir/cmd/root.go
func RootCmd() string {
	return `
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

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
`
}

// ServerCmd projectDir/cmd/server.go
func ServerCmd() string {
	return `
package cmd

import (
	"{{ .ProjectName }}/apis"
	"{{ .ProjectName }}/models"
	"{{ .ProjectName }}/pkg/log"
	"{{ .ProjectName }}/workers"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var logger = log.NewLogger()

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "start server",
	Long:  "start server",
	Run: func(cmd *cobra.Command, args []string) {
		InitDB()
		InitWorker()
		addr := viper.GetString("listenAddr")
		logger.Infof("Listening and serving HTTP on %s", addr)
		apis.Serve(addr)
	},
}

func InitDB() {
	logger.Info("Connect to db")
	if err := models.Connect(viper.GetString("db")); err != nil {
		panic(err)
	}
}

func InitWorker() {
	logger.Info("init worker")
	if err := workers.Initialize(viper.GetString("workerRedisURL")); err != nil {
		panic(err)
	}
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
`
}
