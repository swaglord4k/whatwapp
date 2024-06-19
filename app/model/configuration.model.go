package model

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type ConfigType string

const (
	PostgresConfig ConfigType = "pg"
	MysqlConfig    ConfigType = "mysql"
)

type ServerConf struct {
	Port int `yaml:"SERVER_PORT"`
}

type DatabaseConf struct {
	Host     string `yaml:"HOST"`
	User     string `yaml:"USER"`
	Password string `yaml:"PASSWORD"`
	Dbname   string `yaml:"DBNAME"`
	Port     int16  `yaml:"PORT"`
}

func GetDatabaseConf(configType ConfigType) *DatabaseConf {
	dbConf := &DatabaseConf{}
	directory, _ := os.Getwd()
	yamlFile, err := os.ReadFile(directory + fmt.Sprintf("/config/conf.%s.yaml", configType))
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, dbConf)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	return dbConf
}

func GetServerConf() *ServerConf {
	serverConf := &ServerConf{}
	directory, _ := os.Getwd()
	yamlFile, err := os.ReadFile(directory + "/config/conf.yaml")
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, serverConf)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	return serverConf
}
