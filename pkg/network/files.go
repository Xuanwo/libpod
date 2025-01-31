package network

import (
	"encoding/json"
	"io/ioutil"
	"sort"
	"strings"

	"github.com/containernetworking/cni/libcni"
	"github.com/containernetworking/plugins/plugins/ipam/host-local/backend/allocator"
	"github.com/pkg/errors"
)

// LoadCNIConfsFromDir loads all the CNI configurations from a dir
func LoadCNIConfsFromDir(dir string) ([]*libcni.NetworkConfigList, error) {
	var configs []*libcni.NetworkConfigList
	files, err := libcni.ConfFiles(dir, []string{".conflist"})
	if err != nil {
		return nil, err
	}
	sort.Strings(files)

	for _, confFile := range files {
		conf, err := libcni.ConfListFromFile(confFile)
		if err != nil {
			return nil, err
		}
		configs = append(configs, conf)
	}
	return configs, nil
}

// GetCNIConfigPathByName finds a CNI network by name and
// returns its configuration file path
func GetCNIConfigPathByName(name string) (string, error) {
	files, err := libcni.ConfFiles(CNIConfigDir, []string{".conflist"})
	if err != nil {
		return "", err
	}
	for _, confFile := range files {
		conf, err := libcni.ConfListFromFile(confFile)
		if err != nil {
			return "", err
		}
		if conf.Name == name {
			return confFile, nil
		}
	}
	return "", errors.Errorf("unable to find network configuration for %s", name)
}

// ReadRawCNIConfByName reads the raw CNI configuration for a CNI
// network by name
func ReadRawCNIConfByName(name string) ([]byte, error) {
	confFile, err := GetCNIConfigPathByName(name)
	if err != nil {
		return nil, err
	}
	b, err := ioutil.ReadFile(confFile)
	return b, err
}

// GetCNIPlugins returns a list of plugins that a given network
// has in the form of a string
func GetCNIPlugins(list *libcni.NetworkConfigList) string {
	var plugins []string
	for _, plug := range list.Plugins {
		plugins = append(plugins, plug.Network.Type)
	}
	return strings.Join(plugins, ",")
}

// GetNetworksFromFilesystem gets all the networks from the cni configuration
// files
func GetNetworksFromFilesystem() ([]*allocator.Net, error) {
	var cniNetworks []*allocator.Net
	networks, err := LoadCNIConfsFromDir(CNIConfigDir)
	if err != nil {
		return nil, err
	}
	for _, n := range networks {
		for _, cniplugin := range n.Plugins {
			if cniplugin.Network.Type == "bridge" {
				ipamConf := allocator.Net{}
				if err := json.Unmarshal(cniplugin.Bytes, &ipamConf); err != nil {
					return nil, err
				}
				cniNetworks = append(cniNetworks, &ipamConf)
			}
		}
	}
	return cniNetworks, nil
}

// GetNetworkNamesFromFileSystem gets all the names from the cni network
// configuration files
func GetNetworkNamesFromFileSystem() ([]string, error) {
	var networkNames []string
	networks, err := LoadCNIConfsFromDir(CNIConfigDir)
	if err != nil {
		return nil, err
	}
	for _, n := range networks {
		networkNames = append(networkNames, n.Name)
	}
	return networkNames, nil
}
