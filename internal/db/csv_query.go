package db

import (
	"encoding/csv"
	"os"
	"strconv"
	"time"

	log "github.com/ml444/glog"

	"github.com/ml444/gctl/config"
)

type CSVAssign struct {
	file            *os.File
	cfg             *config.Config
	SvcName         string
	SvcGroup        string
	PortInterval    int
	ErrcodeInterval int
	PortInitMap     map[string]int
	ErrcodeInitMap  map[string]int

	records []*ModelServiceConfig
}

func NewCSVAssign(svcName, svcGroup string, cfg *config.Config) (IDispatcher, error) {
	file, err := os.OpenFile("gctl_service_cfg.cc", os.O_RDWR|os.O_CREATE, 0664)
	if err != nil {
		return nil, err
	}

	return &CSVAssign{
		file:            file,
		cfg:             cfg,
		SvcName:         svcName,
		SvcGroup:        svcGroup,
		PortInterval:    cfg.SvcPortInterval,
		ErrcodeInterval: cfg.SvcErrcodeInterval,
		PortInitMap:     cfg.SvcGroupInitPortMap,
		ErrcodeInitMap:  cfg.SvcGroupInitErrcodeMap,
	}, nil
}
func (a *CSVAssign) Close() {
	_ = a.file.Close()
}
func (a *CSVAssign) GetOrAssignPortAndErrcode(port, errCode *int) error {
	reader := csv.NewReader(a.file)
	records, err := reader.ReadAll()
	if err != nil {
		return err
	}
	foundRecord := a.findRecord(records)
	if port != nil {
		if foundRecord == nil {
			// create record
			foundRecord = a.createRecord()
			log.Info("create new service record: %v", foundRecord)
		} else {
			*port = int(foundRecord.StartPort)
		}
	}
	if errCode != nil {
		if foundRecord == nil {
			// create record
			foundRecord = a.createRecord()
			log.Info("create new service record: %v", foundRecord)
		} else {
			*errCode = int(foundRecord.StartErrCode)

		}
	}
	return nil
}

func (a *CSVAssign) GetModuleID() (int, error) {
	reader := csv.NewReader(a.file)
	records, err := reader.ReadAll()
	if err != nil {
		return 0, err
	}
	foundRecord := a.findRecord(records)
	return int(foundRecord.Id), nil
}

func (a *CSVAssign) findRecord(records [][]string) *ModelServiceConfig {
	var result *ModelServiceConfig
	var ModelList []*ModelServiceConfig
	for _, record := range records {
		m := &ModelServiceConfig{
			Id:           toUint32(record[0]),
			CreatedAt:    toUint32(record[1]),
			UpdatedAt:    toUint32(record[2]),
			DeletedAt:    toUint32(record[3]),
			ServiceName:  record[4],
			ServiceGroup: record[5],
			StartPort:    toUint32(record[6]),
			StartErrCode: toUint32(record[7]),
		}
		if record[5] != a.SvcGroup {
			continue
		}
		if record[4] == a.SvcName {
			result = m
		}
		ModelList = append(ModelList, m)
	}
	a.records = ModelList
	return result
}

func toUint32(s string) uint32 {
	u64, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		println(err.Error())
		return 0
	}
	return uint32(u64)
}

func (a *CSVAssign) createRecord() (record *ModelServiceConfig) {
	var maxID, maxPort, maxErrCode uint32
	for _, r := range a.records {
		if r.Id > maxID {
			maxID = r.Id
		}
		if r.ServiceGroup != a.SvcGroup {
			continue
		}
		if r.StartPort > maxPort {
			maxPort = r.StartPort
		}
		if r.StartErrCode > maxErrCode {
			maxErrCode = r.StartErrCode
		}
	}
	if maxPort == 0 && maxErrCode == 0 {
		// this gourp serice is empty
		initPort, ok := a.PortInitMap[a.SvcGroup]
		if !ok {
			maxPort = uint32(a.cfg.DefaultStartingPort)
		} else {
			maxPort = uint32(initPort)
		}
		initErrCode, ok := a.ErrcodeInitMap[a.SvcGroup]
		if !ok {
			maxErrCode = uint32(a.cfg.DefaultStartingErrcode)
		} else {
			maxErrCode = uint32(initErrCode)
		}
	}
	timeAt := time.Now().Unix()
	record = &ModelServiceConfig{
		Id:           maxID + 1,
		CreatedAt:    uint32(timeAt),
		UpdatedAt:    uint32(timeAt),
		DeletedAt:    0,
		ServiceName:  a.SvcName,
		ServiceGroup: a.SvcGroup,
		StartPort:    maxPort + uint32(a.PortInterval),
		StartErrCode: maxErrCode + uint32(a.ErrcodeInterval),
	}
	a.records = append(a.records, record)

	// insert
	writer := csv.NewWriter(a.file)
	defer writer.Flush()
	writer.Write(toStrSlice(record))
	return record
}

func toStrSlice(m *ModelServiceConfig) []string {
	return []string{
		strconv.FormatUint(uint64(m.Id), 10),
		strconv.FormatUint(uint64(m.CreatedAt), 10),
		strconv.FormatUint(uint64(m.UpdatedAt), 10),
		strconv.FormatUint(uint64(m.DeletedAt), 10),
		m.ServiceName,
		m.ServiceGroup,
		strconv.FormatUint(uint64(m.StartPort), 10),
		strconv.FormatUint(uint64(m.StartErrCode), 10),
	}
}
