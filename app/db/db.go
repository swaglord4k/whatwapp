package db

import (
	"fmt"

	m "de.whatwapp/app/model"
	"gorm.io/gorm"
)

func ConnectToDB(confType m.ConfigType) (*gorm.DB, error) {
	configs := m.GetDatabaseConf(confType)
	switch confType {
	case m.MysqlConfig:
		return ConnectToMySQLDb(configs)
	case m.PostgresConfig:
		return ConnectToPostgresDb(configs)
	}
	return nil, fmt.Errorf("config type %s not supported", confType)
}
