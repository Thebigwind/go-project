package common

import (
	"encoding/json"
	"io/ioutil"
)



type RestServerConfig struct {
	Server      string
	Port        string
	ServerLimit int
}

type PostgresDBConfig struct {
	Server   string
	Port     string
	User     string
	Password string
	Driver   string
	DBName   string
	MaxOpenConns int
	MaxIdleConns int
}

type MysqlDBConfig struct {
	Server   	 string
	Port     	 string
	User     	 string
	Password 	 string
	Driver   	 string
	DBName   	 string
	MaxOpenConns int
	MaxIdleConns int
}


type MongoDBConfig struct {
	Server   string
	Port     string
	User     string
	Password string
	Driver   string
	DBName   string
	MaxOpenConns int
	MaxIdleConns int
}

type RedisDBConfig struct {
	Server   string
	Port     string
	User     string
	Password string
	Driver   string
	DBName   string
}

type ArangoDBConfig struct {
	Server   string
	Port     string
	User     string
	Password string
	Driver   string
	DBName   string
	MaxOpenConns int
	//MaxIdleConns int
}

type RocksDBConfig struct {
	Server   string
	Port     string
	User     string
	Password string
	Driver   string
	DBName   string
	MaxOpenConns int
	//MaxIdleConns int
}



type UserServiceConfig struct {
	EnableLdap bool
}



type RgwServerConfig struct {
	//Enable  bool
	Servers string //`json:"servers"`
}

type KafkaServiceConfig struct {
	Enable  bool   `json:"Enable"`
	Servers string `json:"Servers"`
	Type    string `json:"Type"`
}

type ProjectConfig struct {
	Role           string
	EtcdList       string
	PostgresConfig PostgresDBConfig
	MysqlConfig    MysqlDBConfig
	MongoConfig    MongoDBConfig
	RedisConfig    RedisDBConfig
	ArangoConfig   ArangoDBConfig
	RocksConfig    RocksDBConfig
	RestConfig     RestServerConfig
	UserConfig     UserServiceConfig
	RgwConfig      RgwServerConfig
	KafkaConfig    KafkaServiceConfig
	Test           bool
}

var GlobalProjectConfig *ProjectConfig = nil

func GetProjectConfig() *ProjectConfig {
	return GlobalProjectConfig
}

func ParseProjectConfig(config_file string) *ProjectConfig {
	metaConfig := ProjectConfig{}
	rawData, err := ioutil.ReadFile(config_file)
	if err != nil {
		Logger.Errorf("Can't read config file %s: %s",
			config_file, err.Error())
		return nil
	}
	err = json.Unmarshal(rawData, &metaConfig)
	if err != nil {
		Logger.Errorf("Can't decode config file %s: %s",
			config_file, err.Error())
		return nil
	}
	GlobalProjectConfig = &metaConfig
	return &metaConfig
}
