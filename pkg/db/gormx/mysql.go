package gormx

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// MySQL mysql连接配置
type MysqlCfg struct {
	Host     string //	主机
	Port     string //	端口号
	Username string //	用户名
	Password string //	密码
	Database string //	数据库名
}

type DBC struct {
	*gorm.DB
	cfg MysqlCfg
}

//	获取mysql环境变量
func getMysqlCfg() *MysqlCfg {
	var cfg = &MysqlCfg{
		Host:     "127.0.0.1",
		Port:     "33061",
		Username: "root",
		Password: "123456",
		Database: "servicesMatch",
	}

	//	获取环境变量
	host, present := os.LookupEnv("MYSQL_HOST")
	if present {
		cfg.Host = host
	}
	port, present := os.LookupEnv("MYSQL_PORT")
	if present {
		cfg.Port = port
	}
	username, present := os.LookupEnv("MYSQL_USERNAME")
	if present {
		cfg.Username = username
	}
	password, present := os.LookupEnv("MYSQL_PASSWORD")
	if present {
		cfg.Password = password
	}
	database, present := os.LookupEnv("MYSQL_DATABASE")
	if present {
		cfg.Database = database
	}

	return cfg
}

// 获取数据库连接地址
func (cfg *MysqlCfg) GetURL() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database)
}

// 建立mysql连接
func Open(cfg MysqlCfg) (*DBC, error) {
	var (
		dbProxy *gorm.DB
		err     error
		db      DBC
	)

	url := cfg.GetURL()
	dbProxy, err = gorm.Open(mysql.Open(url), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("get mysql url failed: %v", err)
	}

	//	允许数据库执行期间输出一些log
	dbProxy.Logger.LogMode(logger.Info)

	err = dbProxy.AutoMigrate(&MatchRecord{}, &Match{})
	if err != nil {
		return nil, fmt.Errorf("auto migrate failed: %v", err)
	}

	log.Info("open mysql successfully!")

	db = DBC{DB: dbProxy, cfg: cfg}
	return &db, nil
}

//	建立mysql连接
func OpenMysql() *DBC {
	cfg := getMysqlCfg()
	dbProxy, err := Open(*cfg)
	if err != nil {
		log.Fatalf("open mysql failed: %v", err)
	}
	return dbProxy
}
