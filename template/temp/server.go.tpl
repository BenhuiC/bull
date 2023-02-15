package cmd

import (
	"{{ .ProjectName }}/api"
	"{{ .ProjectName }}/pkg/log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var logger = log.NewLogger()

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "server",
	Short: "start server",
	Long:  "start server",
	Run: func(cmd *cobra.Command, args []string) {
		SetupApp("server")
        addr := config.Cfg.Http.ListenAddr
		logger.Infof("Listening and serving HTTP on %s", addr)
		if err := api.Serve(addr); err != nil {
        	panic(err)
        }
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}