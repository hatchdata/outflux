package ts

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/timescale/outflux/internal/idrf"
)

const (
	createTableQueryTemplate      = "CREATE TABLE %s(%s)"
	columnDefTemplate             = "%s %s"
	createHypertableQueryTemplate = "SELECT create_hypertable('%s', '%s');"
	createTimescaleExtensionQuery = "CREATE EXTENSION IF NOT EXISTS timescaledb"
)

type tableCreator interface {
	Create(dbConn *sql.DB, info *idrf.DataSetInfo) error
}

func newTableCreator() tableCreator {
	return &defaultTableCreator{}
}

type defaultTableCreator struct{}

func (d *defaultTableCreator) Create(dbConn *sql.DB, info *idrf.DataSetInfo) error {
	tableName := info.DataSetName
	if info.DataSetSchema != "" {
		tableName = info.DataSetSchema + "." + tableName
	}

	query := dataSetToSQLTableDef(tableName, info)
	log.Printf("Creating table with:\n %s", query)

	rows, err := dbConn.Query(query)
	if err != nil {
		return err
	}
	rows.Close()

	log.Printf("Preparing TimescaleDB extension:\n%s", createTimescaleExtensionQuery)
	rows, err = dbConn.Query(createTimescaleExtensionQuery)
	if err != nil {
		return err
	}
	rows.Close()

	hypertableQuery := fmt.Sprintf(createHypertableQueryTemplate, tableName, info.TimeColumn)
	log.Printf("Creating hypertable with: %s", hypertableQuery)
	rows, err = dbConn.Query(hypertableQuery)
	if err != nil {
		return err
	}

	rows.Close()
	return nil
}

func dataSetToSQLTableDef(tableName string, dataSet *idrf.DataSetInfo) string {
	columnDefinitions := make([]string, len(dataSet.Columns))
	for i, column := range dataSet.Columns {
		dataType := idrfToPgType(column.DataType)
		columnDefinitions[i] = fmt.Sprintf(columnDefTemplate, column.Name, dataType)
	}

	columnsString := strings.Join(columnDefinitions, ", ")

	return fmt.Sprintf(createTableQueryTemplate, tableName, columnsString)
}
