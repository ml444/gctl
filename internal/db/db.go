package db

import (
	"errors"
)

type ModelServiceConfig struct {
	Id           uint64 `gorm:"comment:主键;primarykey"`
	CreatedAt    uint32 `gorm:"comment:创建时间"`
	UpdatedAt    uint32 `gorm:"comment:更新时间"`
	DeletedAt    uint32 `gorm:"comment:删除时间"`
	ServiceName  string `gorm:"comment:服务名称;type:varchar(50)"`
	ServiceGroup string `gorm:"comment:服务分组;type:varchar(50)"`
	StartPort    uint32 `gorm:"comment:起始端口;type:int;default:0"`
	StartErrCode uint32 `gorm:"comment:起始错误码;type:int;default:0"`
}

// func InitDB(dbUri string, debug bool) (db *gorm.DB, err error) {
// 	//uri := "user:pass@tcp(127.0.0.1:3306)/sms?charset=utf8mb4&parseTime=True&loc=Local"
// 	if dbUri == "" {
// 		_, filename, _, ok := runtime.Caller(0)
// 		if !ok {
// 			return nil, fmt.Errorf("failed to get current file path")
// 		}
// 		log.Info(filepath.Join(path.Dir(filename), "config.yml"))
// 		db, err = gorm.Open(sqlite.Open("gctl-service-cfg.db"), &gorm.Config{Logger: logger.Default.LogMode(logger.Info)})
// 	} else if debug {
// 		db, err = gorm.Open(mysql.Open(dbUri), &gorm.Config{Logger: logger.Default.LogMode(logger.Info)})
// 	} else {
// 		db, err = gorm.Open(mysql.Open(dbUri), &gorm.Config{})
// 	}
// 	if err != nil {
// 		log.Error(err)
// 		return nil, err
// 	}
// 	err = db.AutoMigrate(&ModelServiceConfig{})
// 	if err != nil {
// 		log.Error(err)
// 		return nil, err
// 	}
// 	return db, nil
// }

var initErr = errors.New(`
		You must setting the env of 'GCTL_DB_DSN':
		MySQL: mysql://username:password@tcp(ip:port)/database
		Postgres: postgres://username:password@ip:port/database	or "user=astaxie password=astaxie dbname=test sslmode=disable"
		SQLite: sqlite://username:password@ip:port/database

		`)
