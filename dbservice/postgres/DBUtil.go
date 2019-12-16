package dbservice

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"sync/atomic"
	"time"

	"regexp"

	. "github.com/bigwind/goWeb/common" //DBLogger
	_ "github.com/lib/pq"
)


var globalPostgresService *dbService

var dbReconnectPeriod = time.Second * 5
var dbReconnectCount = 10000


type dbService struct {
	server string
	port   string
	user   string
	pass   string
	dbname string
	driver string
	db     *sql.DB
}

func GetDBService() *dbService {
	return globalPostgresService
}



func createNewDBService(config PostgresDBConfig) (*dbService, error) {


	dbService := &dbService{
		server: config.Server,
		port:   config.Port,
		user:   config.User,
		pass:   config.Password,
		dbname: config.DBName,
		driver: config.Driver,
	}

	dbOpts := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", dbService.user,
		dbService.pass, dbService.dbname)
	if dbService.server != "" {
		dbOpts += " host=" + dbService.server
	}
	if dbService.port != "" {
		dbOpts += " port=" + dbService.port
	}
	db, err := sql.Open(dbService.driver, dbOpts)
	if err != nil {
		DBLogger.Errorf("Fail to open database, error %s \n", err.Error())
		return nil, err
	}
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	dbService.db = db

	//DB service blocking retry dbReconnectCount * dbReconnectPeriod until connect successfully or give up.
	connected := false
	for i := 0; i <= dbReconnectCount; i++ {
		err = db.Ping()
		if err != nil {
			DBLogger.Errorf("database(%s:%s) connection failed, retry...",
				dbService.server, dbService.port)
			time.Sleep(dbReconnectPeriod)
		} else {
			DBLogger.Infof("database(%s:%s) connection successfully! \n",
				dbService.server, dbService.port)
			connected = true
			break
		}
	}

	if connected == false {
		DBLogger.Infof("give up to connect database(%s:%s), failed...! \n",
			dbService.server, dbService.port)
	}
	return dbService, nil
}

func NewDBService(config PostgresDBConfig) error {
	
	DBLogger.Infof("Single DBConfig, config:%v", config)

	dbService, err := createNewDBService(config)
	if err != nil {
		Logger.Errorf("Connect db service failed!\n")
		return nil
	}
	globalPostgresService = dbService

	return nil
}



func (db *dbService) Query(query string, args ...interface{}) (*sql.Rows, error) {
	var err error
	var rows *sql.Rows

	rows, err = db.db.Query(query, args...)
	if err != nil {
		if IsConnectionError(err) != true {
			return nil, err
		} else {
			DBLogger.Errorf("database connection failed, retry...")
		}
	}

	return rows, err
}

func (db *dbService) Exec(sqlStr string, args ...interface{}) error {
	stmt, err := db.db.Prepare(sqlStr)
	if err != nil {
		DBLogger.Errorf("DBService: prepare error %s\n", err.Error())
		return nil, err
	}

	defer stmt.Close()

	_, err = stmt.Exec(args...)
	if err != nil {
		if IsConnectionError(err) {
			DBLogger.Errorf("database connection failed, retry...")
		}

		return nil, err
	}

	return err
}

func (db *dbService) Exec2(sqlStr string, args ...interface{}) (error, int64) {
	stmt, err := db.db.Prepare(sqlStr)
	if err != nil {
		DBLogger.Errorf("DBService: prepare error %s\n", err.Error())
		return nil, err
	}

	defer stmt.Close()

	result, err := stmt.Exec(args...)
	if err != nil {
		if IsConnectionError(err) {
			DBLogger.Errorf("database connection failed, retry...")
		}

		return nil, err
	}

	affect, _ := result.RowsAffected()

	return err, affect
}


func (db *dbService) TxExec(tx *sql.Tx, sqlStr string, args ...interface{}) error {
	var err error = nil

	for i := 0; i <= dbReconnectCount; i++ {
		_, err = tx.Exec(sqlStr, args...)
		if err != nil {
			db.ExecErr(err)
			if IsConnectionError(err) != true {
				break
			} else {
				DBLogger.Errorf("database connection failed, retry...")
				time.Sleep(dbReconnectPeriod)
			}
		} else {
			break
		}
	}
	return err
}

func (db *dbService) TxExec2(tx *sql.Tx, sqlStr string, args ...interface{}) (error, int64) {

	result, err := tx.Exec(sqlStr, args...)
	if err != nil {
		return err, 0
	}

	affect, _ := result.RowsAffected()
	return err, affect
}

func (db *dbService) TxQuery(tx *sql.Tx, sqlStr string, args ...interface{}) error {
	var err error = nil

	for i := 0; i <= dbReconnectCount; i++ {
		_, err = tx.Query(sqlStr, args...)
		if err != nil {
			if IsConnectionError(err) != true {
				break
			} else {
				DBLogger.Errorf("database connection failed, retry...")
				time.Sleep(dbReconnectPeriod)
			}
		} else {
			break
		}
	}
	return err
}

func (db *dbService) Begin() (*sql.Tx, error) {
	return db.db.Begin()
}






/*
func (db *dbService) Query(query string, args ...interface{}) (*sql.Rows, error) {
	var err error
	var rows *sql.Rows
	for i := 0; i <= dbReconnectCount; i++ {
		rows, err = db.db.Query(query, args...)
		if err != nil {
			if IsConnectionError(err) != true {
				return nil, err
			} else {
				DBLogger.Errorf("database connection failed, retry...")
				time.Sleep(dbReconnectPeriod)
			}
		} else {
			break
		}
	}
	return rows, err
}

func (db *dbService) Exec(sqlStr string, args ...interface{}) error {
	stmt, err := db.db.Prepare(sqlStr)
	if err != nil {
		DBLogger.Errorf("DBService: prepare error %s\n", err.Error())
		return err
	}

	defer stmt.Close()

	for i := 0; i <= dbReconnectCount; i++ {
		_, err = stmt.Exec(args...)
		if err != nil {
			db.ExecErr(err)
			if IsConnectionError(err) != true {
				break
			} else {
				DBLogger.Errorf("database connection failed, retry...")
				time.Sleep(dbReconnectPeriod)
			}
		} else {
			break
		}
	}
	return err
}
*/
