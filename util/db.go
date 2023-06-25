package util

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	log "github.com/ml444/glog"
)

var createTableSQL = `CREATE TABLE IF NOT EXISTS service_init_config  
				(
					id int(10) NOT NULL AUTO_INCREMENT, 
					service_name varchar(50) NOT NULL,
					service_group varchar(50) NOT NULL, 
					start_errcode int unsigned NOT NULL, 
					start_port int unsigned DEFAULT 0, 
					PRIMARY KEY (id),
					UNIQUE KEY IDX_service_name (service_name) USING BTREE
				) ENGINE=INNODB DEFAULT CHARSET = utf8mb4`

type SvcAssign struct {
	db              *sql.DB
	DbDSN           string
	SvcName         string
	SvcGroup        string
	PortInterval    int
	ErrcodeInterval int
	PortInitMap     map[string]interface{}
	ErrcodeInitMap  map[string]interface{}
}

func NewSvcAssign(
	dbDSN, svcName, svcGroup string,
	portInterval, errcodeInterval int,
	portInitMap, errcodeInitMap map[string]interface{},
) *SvcAssign {
	return &SvcAssign{
		db:              nil,
		DbDSN:           dbDSN,
		SvcName:         svcName,
		SvcGroup:        svcGroup,
		PortInterval:    portInterval,
		ErrcodeInterval: errcodeInterval,
		PortInitMap:     portInitMap,
		ErrcodeInitMap:  errcodeInitMap,
	}
}

func (a *SvcAssign) GetOrAssignPortAndErrcode(port, errCode *int) error {
	err := a.getDb()
	if err != nil {
		return err
	}
	defer a.db.Close()

	// Just tools to use, no performance considerations. Splitting the query in two.
	if port != nil {
		p, err := a.getServerPort()
		if err != nil {
			return err
		}
		*port = p
	}
	if errCode != nil {
		ec, err := a.getErrCode()
		if err != nil {
			return err
		}
		*errCode = ec
	}
	return nil
}

func (a *SvcAssign) getDb() error {
	if a.db != nil {
		return nil
	}
	var err error
	// postgres://user:password@localhost/mydatabase?sslmode=disable
	if a.DbDSN == "" {
		return errors.New(`
		You must setting the env of 'GCTL_DB_DSN':
		MySQL: mysql://username:password@tcp(ip:port)/database
		Postgres: postgres://username:password@ip:port/database
		`)
	}
	sList := strings.Split(a.DbDSN, "://")
	if len(sList) != 2 {
		return errors.New(fmt.Sprintf("database DSN format is error: %s", a.DbDSN))
	}
	driverName := sList[0]
	if !(driverName == "mysql" || driverName == "postgres") {
		return errors.New("you must use one of the following driver names: mysql or postgresql")
	}
	// TODO: postgres unprocessed
	a.db, err = sql.Open(sList[0], sList[1])
	if err != nil {
		log.Error(err)
		return err
	}
	err = a.initTable(a.db, driverName)
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}

func (a *SvcAssign) initTable(db *sql.DB, dbType string) error {
	switch dbType {
	case "mysql":
		_, err := db.Exec(createTableSQL)
		if err != nil {
			log.Error(err)
			return err
		}
	case "postgres", "postgresql":
		// TODO
		return nil
	}
	return nil
}

func (a *SvcAssign) getMaxErrcode() (int, error) {
	var err error
	var maxErrcode int
	// get max port
	maxErrcodeRow := a.db.QueryRow(`SELECT start_errcode FROM service_init_config WHERE service_group=? ORDER BY start_errcode DESC LIMIT 1`, a.SvcGroup)
	if Err := maxErrcodeRow.Err(); Err != nil {
		log.Error(Err)
		return 0, Err
	}
	err = maxErrcodeRow.Scan(&maxErrcode)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			startErrcode, ok := a.ErrcodeInitMap[a.SvcGroup]
			if !ok {
				return 0, errors.New(fmt.Sprintf("the group[%s] of errcode was not found in the environment variable", a.SvcGroup))
			}

			maxErrcode = ToInt(startErrcode)
		} else {
			log.Error(err)
			return 0, err
		}
	}
	return maxErrcode, nil
}

func (a *SvcAssign) GetModuleId() (int, error) {
	var err error
	if a.db == nil {
		err = a.getDb()
		if err != nil {
			return 0, err
		}
		defer a.db.Close()
	}

	row := a.db.QueryRow(`SELECT id FROM service_init_config WHERE service_name=? AND service_group=?`, a.SvcName, a.SvcGroup)
	if err := row.Err(); err != nil {
		log.Error(a.SvcGroup, a.SvcName)
		log.Error(err.Error())
		return 0, err
	}
	var moduleId int
	err = row.Scan(&moduleId)
	if err != nil {
		log.Error(a.SvcGroup, " ", a.SvcName)
		log.Error(err)
		return 0, err
	}
	return moduleId, nil
}

func (a *SvcAssign) getErrCode() (int, error) {
	insertNewErrcode := func(maxErrcode int) (int, error) {
		newErrcode := maxErrcode + a.ErrcodeInterval
		result, err := a.db.Exec(`INSERT INTO service_init_config (service_name, service_group, start_errcode) VALUES (?, ?, ?)`, a.SvcName, a.SvcGroup, newErrcode)
		if err != nil {
			log.Error(err)
			return 0, err
		}
		lastID, err := result.LastInsertId()
		if err != nil {
			log.Error(err)
			return 0, err
		}
		rowsCount, err := result.RowsAffected()
		if err != nil {
			log.Error(err)
			return 0, err
		}
		if lastID > 0 && rowsCount == 1 {
			log.Infof("success insert data: %d, %s, %d, %d", lastID, a.SvcName, a.SvcGroup, newErrcode)
		} else {
			log.Warn("there may be some unknow errors")
		}
		return newErrcode, nil
	}
	updateNewErrcode := func(maxErrcode int) (int, error) {
		newErrcode := maxErrcode + a.ErrcodeInterval
		result, err := a.db.Exec(`UPDATE service_init_config SET start_errcode=? WHERE service_name=? AND service_group=?;`, newErrcode, a.SvcName, a.SvcGroup)
		if err != nil {
			log.Error(err)
			return 0, err
		}
		lastID, err := result.LastInsertId()
		if err != nil {
			log.Error(err)
			return 0, err
		}
		rowsCount, err := result.RowsAffected()
		if err != nil {
			log.Error(err)
			return 0, err
		}
		if lastID > 0 && rowsCount == 1 {
			log.Infof("success insert data: %d, %s, %d, %d", lastID, a.SvcName, a.SvcGroup, newErrcode)
		} else {
			log.Warn("there may be some unknow errors")
		}
		return newErrcode, nil
	}

	var err error
	row := a.db.QueryRow(`SELECT start_errcode FROM service_init_config WHERE service_name=? AND service_group=?`, a.SvcName, a.SvcGroup)
	if err := row.Err(); err != nil {
		log.Error(err.Error())
		return 0, err
	}
	var errcode int
	err = row.Scan(&errcode)
	if err != nil {
		// create data
		if errors.Is(err, sql.ErrNoRows) {
			maxErrcode, err := a.getMaxErrcode()
			if err != nil {
				log.Error(err)
				return 0, err
			}
			return insertNewErrcode(maxErrcode)
		}
		log.Error(err)
		return 0, err
	}
	if errcode == 0 {
		maxErrcode, err := a.getMaxErrcode()
		if err != nil {
			log.Error(err)
			return 0, err
		}
		return updateNewErrcode(maxErrcode)
	}
	return errcode, nil
}
func (a *SvcAssign) getServerPort() (int, error) {
	getMaxPort := func() (int, error) {
		var err error
		var maxPort int
		// get max port
		maxPortRow := a.db.QueryRow(`SELECT start_port FROM service_init_config WHERE service_group=? ORDER BY start_port DESC LIMIT 1`, a.SvcGroup)
		if Err := maxPortRow.Err(); Err != nil {
			log.Error(Err)
			return 0, Err
		}
		err = maxPortRow.Scan(&maxPort)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				startPort, ok := a.PortInitMap[a.SvcGroup]
				if !ok {
					return 0, errors.New(fmt.Sprintf("the group[%s] of port was not found in the environment variable", a.SvcGroup))
				}
				maxPort = ToInt(startPort)
			} else {
				log.Error(err)
				return 0, err
			}
		}
		return maxPort, nil
	}
	insertNewPort := func(maxPort int) (int, error) {
		newPort := maxPort + a.PortInterval
		result, err := a.db.Exec(`INSERT INTO service_init_config (service_name, service_group, start_port) VALUES (?, ?, ?)`, a.SvcName, a.SvcGroup, newPort)
		if err != nil {
			log.Error(err)
			return 0, err
		}
		lastID, err := result.LastInsertId()
		if err != nil {
			log.Error(err)
			return 0, err
		}
		rowsCount, err := result.RowsAffected()
		if err != nil {
			log.Error(err)
			return 0, err
		}
		if lastID > 0 && rowsCount == 1 {
			log.Infof("success insert data: %d, %s, %d, %d", lastID, a.SvcName, a.SvcGroup, newPort)
		} else {
			log.Warn("there may be some unknow errors")
		}
		return newPort, nil
	}
	updateNewPort := func(maxPort int) (int, error) {
		newPort := maxPort + a.PortInterval
		result, err := a.db.Exec(`UPDATE service_init_config SET start_port=? WHERE service_name=? AND service_group=?;`, newPort, a.SvcName, a.SvcGroup)
		if err != nil {
			log.Error(err)
			return 0, err
		}
		lastID, err := result.LastInsertId()
		if err != nil {
			log.Error(err)
			return 0, err
		}
		rowsCount, err := result.RowsAffected()
		if err != nil {
			log.Error(err)
			return 0, err
		}
		if lastID > 0 && rowsCount == 1 {
			log.Infof("success insert data: %d, %s, %d, %d", lastID, a.SvcName, a.SvcGroup, newPort)
		} else {
			log.Warn("there may be some unknow errors")
		}
		return newPort, nil
	}

	var err error
	row := a.db.QueryRow(`SELECT start_port FROM service_init_config WHERE service_name=? AND service_group=?`, a.SvcName, a.SvcGroup)
	if err := row.Err(); err != nil {
		log.Error(err.Error())
		return 0, err
	}
	var port int
	err = row.Scan(&port)
	if err != nil {
		// create data
		if errors.Is(err, sql.ErrNoRows) {
			maxPort, err := getMaxPort()
			if err != nil {
				log.Error(err)
				return 0, err
			}
			return insertNewPort(maxPort)
		}
		log.Error(err)
		return 0, err
	}
	if port == 0 {
		maxPort, err := getMaxPort()
		if err != nil {
			log.Error(err)
			return 0, err
		}
		return updateNewPort(maxPort)
	}
	return port, nil
}
