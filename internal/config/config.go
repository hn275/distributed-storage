package config

import (
	"os"

	"github.com/hn275/distributed-storage/internal/database"
	"gopkg.in/yaml.v3"
)

const DefaultConfigPath = "config/default.yml"

type config struct {
	User         userYaml
	Cluster      clusterYaml
	LoadBalancer loadbalancerYaml `yaml:"load-balancer"`
}

type clusterYaml struct {
	Node uint16
}

type loadbalancerYaml struct {
	Algorithm string `yaml:"algo"`
	LocalPort uint16 `yaml:"local-port"`
}

type userYaml struct {
	Xsmall  int `yaml:"x-small"`
	Small   int `yaml:"small"`
	Medium  int `yaml:"medium"`
	Large   int `yaml:"large"`
	Xlarge  int `yaml:"x-large"`
	XXlarge int `yaml:"xx-large"`
}

func NewConfig(configPath string) (*config, error) {
	conf := &config{}
	err := readConfig(conf, configPath)
	return conf, err
}

func (u *userYaml) GetFiles(db *database.FileIndex) map[string]int {
	files := make(map[string]int)

	files[db.Xsmall] = u.Xsmall
	files[db.Small] = u.Small
	files[db.Medium] = u.Medium
	files[db.Large] = u.Large
	files[db.Xlarge] = u.Xlarge
	files[db.XXlarge] = u.XXlarge

	return files
}

func readConfig(confBuf *config, filePath string) error {
	fd, err := os.OpenFile(filePath, os.O_RDONLY, 0666)
	if err != nil {
		return err
	}
	return yaml.NewDecoder(fd).Decode(confBuf)
}
