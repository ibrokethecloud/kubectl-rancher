package cmd

import (
	"errors"
	"fmt"
	"os"

	rancher "github.com/ibrokethecloud/kubectl-rancher/pkg"
	"github.com/sirupsen/logrus"

	"github.com/spf13/viper"

	"github.com/spf13/cobra"
)

func init() {
	viper.AutomaticEnv()
	viper.SetEnvPrefix("RANCHER")
	flags := rootCmd.PersistentFlags()
	flags.String("url", "", "Rancher server url to connect to. Should be of the form http/https..")
	flags.String("token", "", "Rancher server api token to use")
	flags.Bool("insecure", false, "Ignore tls check when connecting to rancher server")
	viper.BindPFlag("url", flags.Lookup("url"))
	viper.BindPFlag("token", flags.Lookup("token"))
	viper.BindPFlag("insecure", flags.Lookup("insecure"))

	rootCmd.AddCommand(listCommand)
	rootCmd.AddCommand(configCommand)
}

var (
	rootCmd = &cobra.Command{
		Use:   "kubectl-rancher",
		Short: "kubectl plugin to interact with Rancher API",
		Long: `The plugin interacts with the Rancher API to list clusters,
generate KUBECONFIG file and set it up for environment.
This allows to use the same rancher api token to quickly
switch between clusters and manipulate objects on them`,
	}

	listCommand = &cobra.Command{
		Use:   "list",
		Short: "List all clusters",
		Long:  `List all the clusters the token has access to via the Rancher server`,
		Run: func(cmd *cobra.Command, args []string) {
			r := rancher.NewRancherAPI(viper.GetString("url"),
				viper.GetBool("insecure"),
				viper.GetString("token"))

			clusters, err := r.ListClusters()
			if err != nil {
				logrus.Error(err)
			}
			fmt.Printf("%7s \t %s\n", "ID", "NAME")
			for name, id := range clusters {
				fmt.Printf("%7s \t %s\n", id, name)
			}
		},
	}

	configCommand = &cobra.Command{
		Use:   "config",
		Short: "Fetch kubeConfig",
		Long: `Fetch the kubeConfig file for the specified cluster, and setup the KUBECONFIG
env variable to point to new file.`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("Command needs a cluster name")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			r := rancher.NewRancherAPI(viper.GetString("url"),
				viper.GetBool("insecure"),
				viper.GetString("token"))

			clusters, err := r.ListClusters()
			if err != nil {
				logrus.Error(err)
			}

			clusterID, ok := clusters[args[0]]
			if !ok {
				return errors.New("Invalid cluster name specified: " + args[0])
			}
			kubeConfigFile, err := r.FetchKubeconfig(clusterID, args[0])
			fmt.Printf("Cluster config stored in %s \n", kubeConfigFile)
			return err
		},
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
