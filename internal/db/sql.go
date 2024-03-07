package db

import (
	"database/sql"
	"errors"

	log "github.com/ml444/glog"
)

const (
	mysqlCreateTableSQL = `CREATE TABLE IF NOT EXISTS service_init_config  
	(
		id int(10) NOT NULL AUTO_INCREMENT, 
		service_name varchar(50) NOT NULL,
		service_group varchar(50) NOT NULL, 
		start_errcode int unsigned NOT NULL, 
		start_port int unsigned DEFAULT 0, 
		PRIMARY KEY (id),
		UNIQUE KEY IDX_service_name (service_name, service_group) USING BTREE
	) ENGINE=INNODB DEFAULT CHARSET = utf8mb4`

	sqliteCreateTableSQL = `CREATE TABLE IF NOT EXISTS service_init_config  
	(
		id INTEGER PRIMARY KEY AUTOINCREMENT, 
		service_name VARCHAR(50) NOT NULL,
		service_group VARCHAR(50) NOT NULL, 
		start_errcode INTEGER UNSIGNED NOT NULL, 
		start_port INTEGER UNSIGNED DEFAULT 0, 
		UNIQUE (service_name, service_group)
	);`

	postgresCreateTableSQL = `CREATE TABLE IF NOT EXISTS service_init_config  
	(
		id SERIAL PRIMARY KEY, 
		service_name VARCHAR(50) NOT NULL,
		service_group VARCHAR(50) NOT NULL, 
		start_errcode INTEGER NOT NULL, 
		start_port INTEGER DEFAULT 0, 
		UNIQUE (service_name)
	);`
)

func GetCreateTableSQL(dbType string) string {
	switch dbType {
	case "mysql":
		return mysqlCreateTableSQL
	case "postgresql", "postgres":
		return postgresCreateTableSQL
	case "sqlite", "sqlite3":
		return sqliteCreateTableSQL
	}
	return ""
}

func getMaxErrcode(db *sql.DB, svcGroup string, startErrcode int) (int, error) {
	var err error
	var maxErrcode int
	// get max port
	maxErrcodeRow := db.QueryRow(`SELECT start_errcode FROM service_init_config WHERE service_group=? ORDER BY start_errcode DESC LIMIT 1`, svcGroup)
	if Err := maxErrcodeRow.Err(); Err != nil {
		log.Error(Err)
		return 0, Err
	}
	err = maxErrcodeRow.Scan(&maxErrcode)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			maxErrcode = startErrcode
		} else {
			log.Error(err)
			return 0, err
		}
	}
	return maxErrcode, nil
}
func GetModuleId(db *sql.DB, SvcName, SvcGroup string) (int, error) {
	var err error
	if db == nil {
		return 0, err
	}

	row := db.QueryRow(`SELECT id FROM service_init_config WHERE service_name=? AND service_group=?`, SvcName, SvcGroup)
	if err := row.Err(); err != nil {
		log.Error(SvcGroup, SvcName)
		log.Error(err.Error())
		return 0, err
	}
	var moduleId int
	err = row.Scan(&moduleId)
	if err != nil {
		log.Error(SvcGroup, " ", SvcName)
		log.Error(err)
		return 0, err
	}
	return moduleId, nil
}
