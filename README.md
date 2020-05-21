# kubectl-rancher
kubectl plugin to interact with rancher api


The plugin is to allow a user to quickly download KUBECONFIG's for the clusters.

After an initial login, the plugin stores the rancher token in $HOME/.kube/rancher.json

This allows the token to be reused across commands to list and get config of clusters once an initial login has been performed.

Only supports local and ldap based logins for now.

### Usage
```cassandraql
kubectl rancher -h
The plugin interacts with the Rancher API to list clusters,
generate KUBECONFIG file and set it up for environment.
This allows to use the same rancher api token to quickly
switch between clusters and manipulate objects on them

Usage:
  kubectl rancher [command]

Available Commands:
  config      Fetch kubeConfig
  help        Help about any command
  list        List all clusters
  login       Login to Rancher

Flags:
  -h, --help           help for kubectl-rancher
      --insecure       Ignore tls check when connecting to rancher server
      --token string   Rancher server api token to use, or specify as env variable RANCHER_TOKEN
      --url string     Rancher server url to connect to, or specify as env variable RANCHER_URL. Should be of the form http/https..


kubectl rancher login -h
Login to Rancher using credentials to generate a token, which can be used for subsequent api calls.

Usage:
  kubectl rancher login [flags]

Flags:
  -h, --help                  help for login
      --login-method string   Method to use to login to Rancher api, or specify as env variable RANCHER_LOGIN_METHOD
      --password string       Password name to use to login to Rancher api, or specify as env variable RANCHER_PASSWORD
      --user string           User name to use to login to Rancher api, or specify as env variable RANCHER_USER

Global Flags:
      --insecure       Ignore tls check when connecting to rancher server
      --token string   Rancher server api token to use, or specify as env variable RANCHER_TOKEN
      --url string     Rancher server url to connect to, or specify as env variable RANCHER_URL. Should be of the form http/https..


kubectl rancher list -h
List all the clusters the token has access to via the Rancher server

Usage:
  kubectl- rancher list [flags]

Flags:
  -h, --help   help for list

Global Flags:
      --insecure       Ignore tls check when connecting to rancher server
      --token string   Rancher server api token to use, or specify as env variable RANCHER_TOKEN
      --url string     Rancher server url to connect to, or specify as env variable RANCHER_URL. Should be of the form http/https..

kubectl rancher config -h
Fetch the kubeConfig file for the specified cluster, and setup the KUBECONFIG
env variable to point to new file.

Usage:
  kubectl rancher config [flags]

Flags:
  -h, --help   help for config

Global Flags:
      --insecure       Ignore tls check when connecting to rancher server
      --token string   Rancher server api token to use, or specify as env variable RANCHER_TOKEN
      --url string     Rancher server url to connect to, or specify as env variable RANCHER_URL. Should be of the form http/https..
```