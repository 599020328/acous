package conf

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"sync"
)

type Conf struct {
	ServerPort string `yaml:"port"`

	Number string `yaml:"number"`

	Rho         float64 `yaml:"rho"`
	Q           int     `yaml:"q"`
	Round       int     `yaml:"round"`
	Ant         int     `yaml:"ant"`
	UpdateRound int     `yaml:"update_round"`
	Alpha       float64 `yaml:"alpha"`
	Gamma       float64 `yaml:"gamma"` //todo data size   ask neighbor for delay

	PathThreshold int `yaml:"path_threshold"`

	GetNeighborDelayRound int `yaml:"get_neighbor_delay_round"`
	DataSize              int `yaml:"data_size"`

	RequestRelativePath     string `yaml:"request_relative_path"`
	UpdateDelayRelativePath string `yaml:"update_delay_relative_path"`
	ResponseRelativePath    string `yaml:"response_relative_path"`
	UpdateRelativePath      string `yaml:"update_relative_path"`
	UpdateDataPath			string `yaml:"update_data_path"`

	NeighborList map[string]Neighbor `yaml:"neighbor"`

	UpdateList     []UpdateInfo
	PathLog        map[string]int
	PathTotalDelay map[string]float64

	sync.RWMutex
}

type Neighbor struct {
	Url       string  `yaml:"url"`       //编号
	Tau       float64 `yaml:"tau"`       //信息素浓度
	Bandwidth int     `yaml:"bandwidth"` //带宽
	Delay     float64 `yaml:"delay"`     //延时
}

type UpdateInfo struct {
	Path     string `json:"path"`
	PathCost string `json:"path_cost"`
}

// GetConfOrDie Loading config from local file
func (c *Conf) GetConfOrDie(path string) *Conf {

	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("Open Configure file failed, %v ", err)
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}
	return c
}

func (c *Conf) UpdateTau(nodeNum string, newTau float64) {
	tmp := c.NeighborList[nodeNum]
	tmp.Tau = newTau
	c.NeighborList[nodeNum] = tmp
}
