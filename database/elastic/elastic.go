package elastic

import (
	"github.com/Tokumicn/go-frame/lego-lib/logs"
	"gopkg.in/olivere/elastic.v5"
)

const (
	defaultHost    = "http://127.0.0.1:9200"
	defaultTimeout = 5 // 秒
)

type Config struct {
	Hosts          string // 多个用逗号隔开: http://node01:9200,http://node02:9200,http://node03:9200
	RequestTimeout int    // 秒
}

type Elastic struct {
	config *Config
	Client *elastic.Client
}

// default Host is http://127.0.0.1:9200
func NewElastic(conf *Config) *Elastic {
	newCli, err := connect(conf)
	if err != nil {
		logs.Warnf("init elastic connect host:%s err:%s", conf.Hosts, err)
		return nil
	}

	return &Elastic{
		config: conf,
		Client: newCli,
	}
}

func connect(conf *Config) (*elastic.Client, error) {

	client, err := elastic.NewClient(elastic.SetURL(conf.Hosts))
	if err != nil {
		return nil, err
	}

	esVersion, err := client.ElasticsearchVersion(conf.Hosts)
	if err != nil {
		// Handle error
		return nil, err
	}
	logs.Infof("Elasticsearch version %s\n", esVersion)

	return client, nil
}
