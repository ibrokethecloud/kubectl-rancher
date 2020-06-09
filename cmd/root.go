package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/mitchellh/go-homedir"

	rancher "github.com/ibrokethecloud/kubectl-rancher/pkg"
	"github.com/sirupsen/logrus"

	"github.com/spf13/viper"

	"github.com/spf13/cobra"
)

func init() {
	viper.AutomaticEnv()
	err := checkandcreate()
	if err != nil {
		logrus.Fatal(err)
	}
	flags := rootCmd.PersistentFlags()
	flags.String("url", "", "Rancher server url to connect to, or specify as env variable RANCHER_URL. Should be of the form http/https..")
	flags.String("token", "", "Rancher server api token to use, or specify as env variable RANCHER_TOKEN")
	flags.String("ca", "", "Rancher server ca cert location, or specify as env variable RANCHER_CA")
	flags.Bool("insecure", false, "Ignore tls check when connecting to rancher server")
	viper.BindPFlag("url", flags.Lookup("url"))
	viper.BindPFlag("token", flags.Lookup("token"))
	viper.BindPFlag("insecure", flags.Lookup("insecure"))
	rootCmd.AddCommand(listCommand)
	rootCmd.AddCommand(configCommand)
	rootCmd.AddCommand(loginCommand)
	loginCommand.Flags().String("user", "", "User name to use to login to Rancher api, or specify as env variable RANCHER_USER")
	loginCommand.Flags().String("password", "", "Password name to use to login to Rancher api, or specify as env variable RANCHER_PASSWORD")
	loginCommand.Flags().String("login-method", "", "Method to use to login to Rancher api, or specify as env variable RANCHER_LOGIN_METHOD")
	viper.BindPFlag("user", loginCommand.Flags().Lookup("user"))
	viper.BindPFlag("password", loginCommand.Flags().Lookup("password"))
	viper.BindPFlag("login-method", loginCommand.Flags().Lookup("login-method"))
	viper.SetEnvPrefix("RANCHER")
	viper.SetConfigName("rancher")
	viper.SetConfigType("json")
	viper.AddConfigPath("$HOME/.kube/")
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
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.ReadInConfig()
		},
		Run: func(cmd *cobra.Command, args []string) {
			r := rancher.NewRancherAPI(viper.GetString("url"),
				viper.GetBool("insecure"),
				viper.GetString("token"),
				viper.GetString("ca"))

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
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.ReadInConfig()
		},
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("Command needs a cluster name")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			r := rancher.NewRancherAPI(viper.GetString("url"),
				viper.GetBool("insecure"),
				viper.GetString("token"),
				viper.GetString("ca"))

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

	loginCommand = &cobra.Command{
		Use:   "login",
		Short: "Login to Rancher",
		Long:  `Login to Rancher using credentials to generate a token, which can be used for subsequent api calls.`,
		Run: func(cmd *cobra.Command, args []string) {
			token, err := rancher.NewRancherLogin(viper.GetString("url"),
				viper.GetString("user"),
				viper.GetString("password"),
				viper.GetString("login-method"),
				viper.GetBool("insecure"),
				viper.GetString("ca"))

			if err != nil {
				logrus.Error(err)
			}
			viper.Set("password", "")
			viper.Set("token", token)
			if err := viper.WriteConfig(); err != nil {
				logrus.Error(err)
			}
		},
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func checkandcreate() (err error) {
	dir, err := homedir.Dir()
	if err != nil {
		return err
	}

	if _, err = os.Stat(dir + "/.kube"); os.IsNotExist(err) {
		err = os.Mkdir(dir+"/.kube", 0755)
	}

	if err != nil {
		return err
	}

	stateFileName := dir + "/.kube/" + "rancher.json"
	_, err = os.Stat(stateFileName)
	if os.IsNotExist(err) {
		stateFile, err := os.Create(stateFileName)
		if err != nil {
			return err
		}

		defer stateFile.Close()
	}
	return nil
}
