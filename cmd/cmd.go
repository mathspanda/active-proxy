package cmd

import (
	"flag"

	"github.com/spf13/cobra"
)

type Option struct {
	ConfigFile string
}

const (
	CONFIG_FILE_DEFAULT = "config.yaml"
)

func NewProxyCommand(startFunc func(configFile string)) *cobra.Command {
	option := &Option{}
	cmd := &cobra.Command{
		Use:   "acproxy",
		Short: "acproxy is a proxy interacting with active hdfs",
		Long:  "a proxy aims to interact with Hadoop clusters, which supports hdfs temporarily",
		Run: func(cmd *cobra.Command, args []string) {
			startFunc(option.ConfigFile)
		},
	}
	cmd.Flags().StringVarP(&option.ConfigFile, "config_file", "c", CONFIG_FILE_DEFAULT, "set location of config file")
	cmd.Flags().AddGoFlagSet(flag.CommandLine)
	flag.CommandLine.Parse(nil)
	return cmd
}
