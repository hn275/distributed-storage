package config

import (
	"os"

	"github.com/hn275/distributed-storage/internal/database"
	"gopkg.in/yaml.v3"
)

const DefaultConfigPath = "config/default.yml"

type Config struct {
	User         userYaml
	Cluster      clusterYaml
	LoadBalancer loadbalancerYaml `yaml:"load-balancer"`
	Experiment   ExperimentYaml
}

type clusterYaml struct {
	Node     uint16
	Capacity uint16
}

type loadbalancerYaml struct {
	Algorithm string `yaml:"algo"`
	LocalPort uint16 `yaml:"local-port"`
}

type userYaml struct {
	Xsmall   int    `yaml:"x-small"`
	Small    int    `yaml:"small"`
	Medium   int    `yaml:"medium"`
	Large    int    `yaml:"large"`
	Xlarge   int    `yaml:"x-large"`
	XXlarge  int    `yaml:"xx-large"`
	Interval uint32 `yaml:"interval"`
}

type ExperimentYaml struct {
	Name          string `yaml:"name"`
	Latency       uint32 `yaml:"interval"`
	Homogeneous   bool
	OverheadParam int64 `yaml:"overhead-param"`
}

func NewConfig(configPath string) (*Config, error) {
	conf := &Config{}
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

func readConfig(confBuf *Config, filePath string) error {
	fd, err := os.OpenFile(filePath, os.O_RDONLY, 0666)
	if err != nil {
		return err
	}
	return yaml.NewDecoder(fd).Decode(confBuf)
}
