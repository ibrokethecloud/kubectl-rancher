# kubectl-rancher
kubectl plugin to interact with rancher api


The plugin is to allow a user to quickly download KUBECONFIG's for the clusters.

### Usage
```cassandraql
The plugin interacts with the Rancher API to list clusters,
generate KUBECONFIG file and set it up for environment.
This allows to use the same rancher api token to quickly
switch between clusters and manipulate objects on them

Usage:
  kubectl-rancher [command]

Available Commands:
  config      Fetch kubeConfig
  help        Help about any command
  list        List all clusters

Flags:
  -h, --help           help for kubectl-rancher
      --insecure       Ignore tls check when connecting to rancher server
      --token string   Rancher server api token to use, or specify as env variable RANCHER_TOKEN
      --url string     Rancher server url to connect to, or specify as env variable RANCHER_URL. Should be of the form http/https..

Use "kubectl-rancher [command] --help" for more information about a command.
```