package rancher

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"golang.org/x/crypto/ssh/terminal"

	homedir "github.com/mitchellh/go-homedir"
	v3 "github.com/rancher/types/apis/management.cattle.io/v3"
)

// RancherAPI is the parent struct for interacting with the Rancher API endpoint
type RancherAPI struct {
	Endpoint string
	Insecure bool
	Token    string
}

// ClusterList is the parent struct for parsing response from v3.Clusters
type ClusterList struct {
	CList []ClusterListSpec `json:"data"`
}

type ClusterListSpec struct {
	Name    string            `json:"name"`
	ID      string            `json:"id"`
	Actions map[string]string `json:"actions"`
}

// NewRancherAPI will use the input flags or env variables to init
// the RancherAPI struct for use for interacting with the rancher server
func NewRancherAPI(url string, insecure bool, token string) (r *RancherAPI) {
	r = &RancherAPI{Endpoint: url,
		Insecure: insecure,
		Token:    token,
	}
	return r
}

func (r *RancherAPI) ListClusters() (clusters map[string]string, err error) {
	clusters = make(map[string]string)
	b, err := r.makeCall("/v3/clusters", "GET", nil)
	if err != nil {
		return clusters, err
	}
	c := &ClusterList{}
	err = json.Unmarshal(b, c)
	if err != nil {
		return clusters, err
	}

	for _, cn := range c.CList {
		if cn.ID == "local" {
			clusters["local"] = "local"
		} else {
			clusters[cn.Name] = cn.ID
		}

	}

	return clusters, err
}

func (r *RancherAPI) FetchKubeconfig(clusterID string,
	clusterName string) (filePath string, err error) {
	reqEndpoint := "/v3/clusters/" + clusterID + "?action=generateKubeconfig"
	data, err := r.makeCall(reqEndpoint, "POST", nil)
	if err != nil {
		return "", err
	}

	if len(data) == 0 {
		return "", errors.New("Kubeconfig file looks empty")
	}
	kubeConfig := v3.GenerateKubeConfigOutput{}

	if err := json.Unmarshal(data, &kubeConfig); err != nil {
		return "", err
	}

	// Setup file
	dir, err := homedir.Dir()

	if err != nil {
		return "", err
	}

	if _, err = os.Stat(dir + "/.kube"); os.IsNotExist(err) {
		err = os.Mkdir(dir+"/.kube", 0755)
	}

	if err != nil {
		return "", err
	}

	kubeConfigFile := dir + "/.kube/" + clusterName + ".yaml"
	err = ioutil.WriteFile(kubeConfigFile,
		[]byte(kubeConfig.Config), 0644)

	return kubeConfigFile, err

}

func (r *RancherAPI) makeCall(uri string, method string,
	request io.Reader) (data []byte, err error) {
	req, err := http.NewRequest(method, r.Endpoint+uri, request)
	if err != nil {
		return nil, err
	}
	username, password, err := splitToken(r.Token)
	if err == nil {
		req.SetBasicAuth(username, password)
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	var client *http.Client
	if r.Insecure {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client = &http.Client{Transport: tr}
	} else {
		client = &http.Client{}
	}

	resp, err := client.Do(req)
	if resp.StatusCode > 400 || err != nil {
		return []byte{}, fmt.Errorf("Error during api call: %v, Resp Code: %v \n", err, resp.StatusCode)
	}
	if err != nil {
		return []byte{}, err
	}

	defer resp.Body.Close()
	data, err = ioutil.ReadAll(resp.Body)
	return data, err
}

func splitToken(token string) (username string, password string, err error) {
	result := strings.Split(token, ":")
	if len(result) != 2 {
		return "", "", errors.New("[error] Token looks invalid")
	}
	username = result[0]
	password = result[1]
	return username, password, nil
}

// NewRancherLogin performs the login using api call
func NewRancherLogin(url string, username string, password string, method string, insecure bool) (token string, err error) {

	username, err = checkAndPrompt(username, "RANCHER_USER", false)
	if err != nil {
		return "", err
	}

	password, err = checkAndPrompt(password, "RANCHER_PASSWORD", true)
	if err != nil {
		return "", err
	}

	method, err = checkAndPrompt(method, "RANCHER_LOGIN_METHOD", false)
	if err != nil {
		return "", err
	}

	r := RancherAPI{
		Endpoint: url,
		Insecure: insecure,
		Token:    "",
	}

	reqMap := make(map[string]string)
	reqMap["username"] = username
	reqMap["password"] = password
	reqJson, err := json.Marshal(reqMap)
	if err != nil {
		return "", err
	}

	request := bytes.NewReader(reqJson)
	loginSuffix, err := checkMethod(method)
	if err != nil {
		return "", err
	}

	uri := "/v3-public" + loginSuffix + "?action=login"

	resp, err := r.makeCall(uri, "POST", request)
	if err != nil {
		return "", err
	}

	respMap := make(map[string]interface{})

	if err := json.Unmarshal(resp, &respMap); err != nil {
		return "", err
	}

	return respMap["token"].(string), nil
}

func checkAndPrompt(input string, inputName string, secure bool) (output string, err error) {
	if len(input) != 0 {
		output = input
		return output, nil
	}
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Enter %s:", inputName)
	if secure {
		byteOutput, err := terminal.ReadPassword(0)
		fmt.Println()
		if err == nil {
			output = string(byteOutput)
		}
	} else {
		output, err = reader.ReadString('\n')
	}
	output = strings.TrimSpace(output)
	return output, err
}

func checkMethod(method string) (loginSuffix string, err error) {
	switch method {
	case "local":
		loginSuffix = "/localProviders/local"
	case "ldap":
		loginSuffix = "/openLdapProviders/openldap"
	default:
		err = fmt.Errorf("Invalid login method type")
	}

	return loginSuffix, err
}
