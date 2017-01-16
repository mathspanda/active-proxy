package cmd

import (
	"github.com/spf13/cobra"
)

type Option struct {
	configFile string
	logDir     string
}

const (
	CONFIG_FILE_DEFAULT = "examples/config.yaml"
	LOG_DIR_DEFAULT     = "/var/log/acproxy"
)

func NewProxyCommand(startFunc func(configFile string)) *cobra.Command {
	option := &Option{}
	cmd := &cobra.Command{
		Use:   "acproxy",
		Short: "acproxy is a proxy",
		Long:  "acproxy is a proxy",
		Run: func(cmd *cobra.Command, args []string) {
			startFunc(option.configFile)
		},
	}
	cmd.Flags().StringVarP(&option.logDir, "log_dir", "d", LOG_DIR_DEFAULT, "set location of log dir")
	cmd.Flags().StringVarP(&option.configFile, "config_file", "c", CONFIG_FILE_DEFAULT, "set location of config file")

	return cmd
}
