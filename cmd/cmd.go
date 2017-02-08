package cmd

import (
	"flag"

	"github.com/spf13/cobra"
)

type Option struct {
	ConfigFile   string
	ProviderType string
}

const (
	CONFIG_FILE_DEFAULT   = "config.yaml"
	PROVIDER_TYPE_DEFAULT = "hdfs"
)

func NewProxyCommand(startFunc func(providerType string, configFile string)) *cobra.Command {
	option := &Option{}
	cmd := &cobra.Command{
		Use:   "acproxy",
		Short: "acproxy is a proxy interacting with active hdfs",
		Long:  "a proxy aims to interact with Hadoop clusters, which supports hdfs temporarily",
		Run: func(cmd *cobra.Command, args []string) {
			startFunc(option.ProviderType, option.ConfigFile)
		},
	}
	cmd.Flags().StringVarP(&option.ConfigFile, "config_file", "c", CONFIG_FILE_DEFAULT, "location of config file")
	cmd.Flags().StringVarP(&option.ProviderType, "type", "t", PROVIDER_TYPE_DEFAULT, "proxy provider type chosen in {hdfs}")
	cmd.Flags().AddGoFlagSet(flag.CommandLine)
	flag.CommandLine.Parse(nil)
	return cmd
}
