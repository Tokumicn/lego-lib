package postgresql

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	// 引用gorm同时需要引用对应的库 - postgres
	_ "github.com/jinzhu/gorm/dialects/postgres"

	"lego-lib/logs"
)

type (
	// DB 避免业务代码直接引用gorm
	DB = gorm.DB
)

// ErrRecordNotFound 查找结果为空的错误字符串
var ErrRecordNotFound = "record not found"

// Pool mysql连接池再封装
type Pool struct {
	db *gorm.DB
}

// NodeConfig 一个实例配置
type NodeConfig struct {
	Host    string
	Auth    string
	Opts    string
	Name    string
	MaxIdle int `toml:"max_idle"`
	MaxOpen int `toml:"max_open"`
	MaxLife int `toml:"max_life"`
}

// Config mysql对外配置
type Config struct {
	Debug      bool
	NodeConfig *NodeConfig
}

// NewPool 初始化mysql连接池
func NewPool(conf *Config) *Pool {
	pool := &Pool{}

	db, err := connect(conf.Debug, conf.NodeConfig)
	if err != nil {
		logs.Warnf("init postgre connect host:%s err:%s", conf.NodeConfig.Host, err)
		panic(err)
	}
	pool.db = db

	logs.Infof("init mysql finish")
	return pool
}

func connect(debug bool, node *NodeConfig) (*gorm.DB, error) {
	dst := fmt.Sprintf("%s@tcp(%s)/%s", node.Auth, node.Host, node.Name)
	if len(node.Opts) > 0 {
		dst = dst + "?" + node.Opts
	}

	orm, err := gorm.Open("postgres", dst)
	if err != nil {
		return nil, err
	}
	if debug {
		orm.LogMode(true)
	}

	orm.DB().SetMaxIdleConns(node.MaxIdle)
	orm.DB().SetMaxOpenConns(node.MaxOpen)
	orm.DB().SetConnMaxLifetime(time.Duration(node.MaxLife) * time.Second)
	return orm, nil
}
