package rancher

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

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
	Clusters v3.ClusterSpec    `json:"appliedSpec"`
	ID       string            `json:"id"`
	Actions  map[string]string `json:"actions"`
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
			clusters[cn.Clusters.DisplayName] = cn.ID
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
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(username, password)
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
