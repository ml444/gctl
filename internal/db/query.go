package db

import (
	"database/sql"
	"errors"
	"fmt"
	"path"
	"runtime"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	log "github.com/ml444/glog"

	"github.com/ml444/gctl/config"
)

type SvcAssign struct {
	db              *sql.DB
	DBURI           string
	SvcName         string
	SvcGroup        string
	PortInterval    int
	ErrcodeInterval int
	PortInitMap     map[string]int
	ErrcodeInitMap  map[string]int
}

func NewSvcAssign(svcName, svcGroup string, cfg *config.Config) (*SvcAssign, error) {
	sqlDB, err := getDB(cfg.DbURI)
	if err != nil {
		return nil, err
	}

	return &SvcAssign{
		db:              sqlDB,
		SvcName:         svcName,
		SvcGroup:        svcGroup,
		DBURI:           cfg.DbURI,
		PortInterval:    cfg.SvcPortInterval,
		ErrcodeInterval: cfg.SvcErrcodeInterval,
		PortInitMap:     cfg.SvcGroupInitPortMap,
		ErrcodeInitMap:  cfg.SvcGroupInitErrcodeMap,
	}, nil
}

func (a *SvcAssign) GetOrAssignPortAndErrcode(port, errCode *int) error {
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

func getDB(dbURI string) (db *sql.DB, err error) {
	var driverName, dataSourceName string
	if dbURI == "" {
		_, filename, _, ok := runtime.Caller(0)
		if !ok {
			return nil, fmt.Errorf("failed to get current file path")
		}
		log.Info(path.Dir(filename))
		for _, driver := range sql.Drivers() {
			println(driver)
		}
		driverName = "sqlite3"
		dataSourceName = "./gctl_service.db"

	} else {
		sList := strings.Split(dbURI, "://")
		if len(sList) != 2 {
			return nil, fmt.Errorf("database DSN format is error: %s", dbURI)
		}
		driverName = sList[0]
		dataSourceName = sList[1]

	}

	db, err = sql.Open(driverName, dataSourceName)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	err = initTable(db, driverName)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return db, nil
}

func initTable(db *sql.DB, dbType string) error {
	createTableSQL := GetCreateTableSQL(dbType)
	if createTableSQL == "" {
		return errors.New("not found db type: " + dbType)
	}
	_, err := db.Exec(createTableSQL)
	if err != nil {
		log.Error(err)
		return err
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
				return 0, fmt.Errorf("not found the group: '%s' from errcodeMap", a.SvcGroup)
			}

			maxErrcode = startErrcode
		} else {
			log.Error(err)
			return 0, err
		}
	}
	return maxErrcode, nil
}

func (a *SvcAssign) GetModuleID() (int, error) {
	var err error
	row := a.db.QueryRow(`SELECT id FROM service_init_config WHERE service_name=? AND service_group=?`, a.SvcName, a.SvcGroup)
	if err = row.Err(); err != nil {
		log.Error(a.SvcGroup, a.SvcName)
		log.Error(err.Error())
		return 0, err
	}
	var moduleID int
	err = row.Scan(&moduleID)
	if err != nil {
		log.Error(a.SvcGroup, " ", a.SvcName)
		log.Error(err)
		return 0, err
	}
	return moduleID, nil
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
	var maxErrcode int
	row := a.db.QueryRow(`SELECT start_errcode FROM service_init_config WHERE service_name=? AND service_group=?`, a.SvcName, a.SvcGroup)
	if err = row.Err(); err != nil {
		log.Error(err.Error())
		return 0, err
	}
	var errcode int
	err = row.Scan(&errcode)
	if err != nil {
		// create data
		if errors.Is(err, sql.ErrNoRows) {
			maxErrcode, err = a.getMaxErrcode()
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
		maxErrcode, err = a.getMaxErrcode()
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
					return 0, fmt.Errorf("not found the group '%s' from portMap", a.SvcGroup)
				}
				maxPort = startPort
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
