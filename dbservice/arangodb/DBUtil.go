package dbservice_arangodb

import (
	"context"
	driver "github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
	"strings"
	"sync"
)

type dbServiceConfig struct {
	Endpoints    []string
	User         string
	Password     string
	DBName       string
	MaxOpenConns int
}

type dbService struct {
	db     driver.Database
	client driver.Client
	next   *dbService
}

func GetDBService() *dbService {
	globalDBLimit <- struct{}{}

	globalDBServiceLock.Lock()
	defer globalDBServiceLock.Unlock()
	if globalDBServiceHead == nil {
		// the queue is empty, create a new db connection
		db, err := createNewDBService(globalDBServiceConfig)
		if err != nil {
			DBLogger.Errorf("createNewDBService failed: %s", err)
			return nil
		}
		return db
	} else {
		db := globalDBServiceHead
		globalDBServiceHead = db.next
		db.next = nil
		if globalDBServiceHead == nil {
			globalDBServiceTail = nil
		}
		return db
	}
}

func ReleaseDBService(db *dbService) {
	globalDBServiceLock.Lock()
	if globalDBServiceTail == nil {
		globalDBServiceTail = db
		globalDBServiceHead = db
	} else {
		globalDBServiceTail.next = db
		globalDBServiceTail = db
	}
	globalDBServiceLock.Unlock()

	<-globalDBLimit
}

var globalDBLimit chan struct{} // 限制与数据库的最大连接
var globalDBServiceConfig dbServiceConfig

var globalDBServiceHead *dbService
var globalDBServiceTail *dbService
var globalDBServiceLock *sync.RWMutex

func createNewDBService(config dbServiceConfig) (*dbService, error) {
	//Connecting to ArangoDB
	conn, err := http.NewConnection(http.ConnectionConfig{
		Endpoints: config.Endpoints,
		/*TLSConfig: &tls.Config{},*/
	})
	if err != nil {
		return nil, err
	}
	c, err := driver.NewClient(driver.ClientConfig{
		Connection:     conn,
		Authentication: driver.BasicAuthentication(config.User, config.Password),
	})
	if err != nil {
		return nil, err
	}
	ctx := context.Background()

	//Opening a database
	db, err := c.Database(ctx, config.DBName)
	if err != nil {
		return nil, err
	}
	return &dbService{db: db, client: c, next: nil}, nil
}

func NewDBService(config ArangoDBConfig) error {
	globalDBServiceConfig.Endpoints = strings.Split(config.Server, ",")
	globalDBServiceConfig.User = config.User
	globalDBServiceConfig.Password = config.Password
	globalDBServiceConfig.DBName = config.DBName

	if globalDBServiceConfig.MaxOpenConns <= 0 {
		globalDBServiceConfig.MaxOpenConns = 64
	}
	DBLogger.Infof("globalDBServiceConfig is %v", globalDBServiceConfig)
	globalDBLimit = make(chan struct{}, globalDBServiceConfig.MaxOpenConns)

	db, err := createNewDBService(globalDBServiceConfig)
	if err != nil {
		return err
	}

	globalDBServiceHead = db
	globalDBServiceTail = db
	globalDBServiceLock = new(sync.RWMutex)

	return nil
}

func (db *dbService) Query(ctx context.Context, query string, bindVars map[string]interface{}) (driver.Cursor, error) {
	cursor, err := db.db.Query(ctx, query, bindVars)
	return cursor, err
}

func (db *dbService) Exec(ctx context.Context, action string) (interface{}, error) {
	result, err := db.db.Transaction(ctx, action, nil)
	return result, err
}

func (db *dbService) Exec2(action string, readCollections []string, writeCollections []string, params []interface{}) (interface{}, error) {
	ctx := context.Background()

	transOptions := &driver.TransactionOptions{
		//MaxTransactionSize = options.MaxTransactionSize,
		//LockTimeout = options.LockTimeout,
		WaitForSync: true,
		//IntermediateCommitCount: options.IntermediateCommitCount,
		Params: params,
		//IntermediateCommitSize:  options.IntermediateCommitSize,
		ReadCollections:    readCollections,
		WriteCollections:   writeCollections,
		MaxTransactionSize: 100000,
	}
	result, err := db.db.Transaction(ctx, action, transOptions)
	return result, err
}
