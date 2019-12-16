package db

import (
	"log"
	"time"

	"github.com/globalsign/mgo"
)




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
	return globalDBService
}


func NewDBService(etcdConfig projectConfig) error {
	maxOpenConns := etcdConfig.RestConfig.MaxOpenConns
	if maxOpenConns <= 0 {
		maxOpenConns = 64
	}
	maxIdleConns := etcdConfig.RestConfig.MaxIdleConns
	if maxIdleConns <= 0 {
		maxIdleConns = 15
	}

	var config MongoDBConfig
	config = etcdConfig.MongoConfig
	Logger.Infof("Single DBConfig, config:%v", config)

	dbService, err := createNewDBService(config, maxOpenConns, maxIdleConns)
	if err != nil {
		Logger.Errorf("Connect db service failed!\n")
		return nil
	}
	globalDBService = dbService

	return nil
}


func createNewDBService(config PostgresDBConfig, maxOpenConns, maxIdleConns int) (*dbService, error) {


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
	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
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




var globalS *mgo.Session

func createNewDBService(config, maxOpenConns, maxIdleConns) {
	dialInfo := &mgo.DialInfo{
		Addrs:     []string{dbhost},
		Timeout:   timeout,
		Source:    authdb,
		Username:  authuser,
		Password:  authpass,
		PoolLimit: poollimit,
	}

	s, err := mgo.DialWithInfo(dialInfo)
	if err != nil {
		log.Fatalf("Create Session: %s\n", err)
	}
	globalS = s
}

func connect(db, collection string) (*mgo.Session, *mgo.Collection) {
	ms := globalS.Copy()
	c := ms.DB(db).C(collection)
	ms.SetMode(mgo.Monotonic, true)
	return ms, c
}

func getDb(db string) (*mgo.Session, *mgo.Database) {
	ms := globalS.Copy()
	return ms, ms.DB(db)
}

func IsEmpty(db, collection string) bool {
	ms, c := connect(db, collection)
	defer ms.Close()
	count, err := c.Count()
	if err != nil {
		log.Fatal(err)
	}
	return count == 0
}

func Count(db, collection string, query interface{}) (int, error) {
	ms, c := connect(db, collection)
	defer ms.Close()
	return c.Find(query).Count()
}



///////////////////////////////////
const (
	dbhost    = "127.0.0.1:27017"
	authdb    = "admin"
	authuser  = "user"
	authpass  = "123456"
	timeout   = 60 * time.Second
	poollimit = 4096
)

var globalS *mgo.Session

func init() {
	dialInfo := &mgo.DialInfo{
		Addrs:     []string{dbhost},
		Timeout:   timeout,
		Source:    authdb,
		Username:  authuser,
		Password:  authpass,
		PoolLimit: poollimit,
	}

	s, err := mgo.DialWithInfo(dialInfo)
	if err != nil {
		log.Fatalf("Create Session: %s\n", err)
	}
	globalS = s
}

func connect(db, collection string) (*mgo.Session, *mgo.Collection) {
	ms := globalS.Copy()
	c := ms.DB(db).C(collection)
	ms.SetMode(mgo.Monotonic, true)
	return ms, c
}

func getDb(db string) (*mgo.Session, *mgo.Database) {
	ms := globalS.Copy()
	return ms, ms.DB(db)
}

func IsEmpty(db, collection string) bool {
	ms, c := connect(db, collection)
	defer ms.Close()
	count, err := c.Count()
	if err != nil {
		log.Fatal(err)
	}
	return count == 0
}

func Count(db, collection string, query interface{}) (int, error) {
	ms, c := connect(db, collection)
	defer ms.Close()
	return c.Find(query).Count()
}

func Insert(db, collection string, docs ...interface{}) error {
	ms, c := connect(db, collection)
	defer ms.Close()

	return c.Insert(docs...)
}

func FindOne(db, collection string, query, selector, result interface{}) error {
	ms, c := connect(db, collection)
	defer ms.Close()

	return c.Find(query).Select(selector).One(result)
}

func FindAll(db, collection string, query, selector, result interface{}) error {
	ms, c := connect(db, collection)
	defer ms.Close()

	return c.Find(query).Select(selector).All(result)
}

func FindPage(db, collection string, page, limit int, query, selector, result interface{}) error {
	ms, c := connect(db, collection)
	defer ms.Close()

	return c.Find(query).Select(selector).Skip(page * limit).Limit(limit).All(result)
}

func FindIter(db, collection string, query interface{}) *mgo.Iter {
	ms, c := connect(db, collection)
	defer ms.Close()

	return c.Find(query).Iter()
}

func Update(db, collection string, selector, update interface{}) error {
	ms, c := connect(db, collection)
	defer ms.Close()

	return c.Update(selector, update)
}

func Upsert(db, collection string, selector, update interface{}) error {
	ms, c := connect(db, collection)
	defer ms.Close()

	_, err := c.Upsert(selector, update)
	return err
}

func UpdateAll(db, collection string, selector, update interface{}) error {
	ms, c := connect(db, collection)
	defer ms.Close()

	_, err := c.UpdateAll(selector, update)
	return err
}

func Remove(db, collection string, selector interface{}) error {
	ms, c := connect(db, collection)
	defer ms.Close()

	return c.Remove(selector)
}

func RemoveAll(db, collection string, selector interface{}) error {
	ms, c := connect(db, collection)
	defer ms.Close()

	_, err := c.RemoveAll(selector)
	return err
}

//insert one or multi documents
func BulkInsert(db, collection string, docs ...interface{}) (*mgo.BulkResult, error) {
	ms, c := connect(db, collection)
	defer ms.Close()
	bulk := c.Bulk()
	bulk.Insert(docs...)
	return bulk.Run()
}

func BulkRemove(db, collection string, selector ...interface{}) (*mgo.BulkResult, error) {
	ms, c := connect(db, collection)
	defer ms.Close()

	bulk := c.Bulk()
	bulk.Remove(selector...)
	return bulk.Run()
}

func BulkRemoveAll(db, collection string, selector ...interface{}) (*mgo.BulkResult, error) {
	ms, c := connect(db, collection)
	defer ms.Close()
	bulk := c.Bulk()
	bulk.RemoveAll(selector...)
	return bulk.Run()
}

func BulkUpdate(db, collection string, pairs ...interface{}) (*mgo.BulkResult, error) {
	ms, c := connect(db, collection)
	defer ms.Close()
	bulk := c.Bulk()
	bulk.Update(pairs...)
	return bulk.Run()
}

func BulkUpdateAll(db, collection string, pairs ...interface{}) (*mgo.BulkResult, error) {
	ms, c := connect(db, collection)
	defer ms.Close()
	bulk := c.Bulk()
	bulk.UpdateAll(pairs...)
	return bulk.Run()
}

func BulkUpsert(db, collection string, pairs ...interface{}) (*mgo.BulkResult, error) {
	ms, c := connect(db, collection)
	defer ms.Close()
	bulk := c.Bulk()
	bulk.Upsert(pairs...)
	return bulk.Run()
}

func PipeAll(db, collection string, pipeline, result interface{}, allowDiskUse bool) error {
	ms, c := connect(db, collection)
	defer ms.Close()
	var pipe *mgo.Pipe
	if allowDiskUse {
		pipe = c.Pipe(pipeline).AllowDiskUse()
	} else {
		pipe = c.Pipe(pipeline)
	}
	return pipe.All(result)
}

func PipeOne(db, collection string, pipeline, result interface{}, allowDiskUse bool) error {
	ms, c := connect(db, collection)
	defer ms.Close()
	var pipe *mgo.Pipe
	if allowDiskUse {
		pipe = c.Pipe(pipeline).AllowDiskUse()
	} else {
		pipe = c.Pipe(pipeline)
	}
	return pipe.One(result)
}

func PipeIter(db, collection string, pipeline interface{}, allowDiskUse bool) *mgo.Iter {
	ms, c := connect(db, collection)
	defer ms.Close()
	var pipe *mgo.Pipe
	if allowDiskUse {
		pipe = c.Pipe(pipeline).AllowDiskUse()
	} else {
		pipe = c.Pipe(pipeline)
	}

	return pipe.Iter()

}

func Explain(db, collection string, pipeline, result interface{}) error {
	ms, c := connect(db, collection)
	defer ms.Close()
	pipe := c.Pipe(pipeline)
	return pipe.Explain(result)
}
func GridFSCreate(db, prefix, name string) (*mgo.GridFile, error) {
	ms, d := getDb(db)
	defer ms.Close()
	gridFs := d.GridFS(prefix)
	return gridFs.Create(name)
}

func GridFSFindOne(db, prefix string, query, result interface{}) error {
	ms, d := getDb(db)
	defer ms.Close()
	gridFs := d.GridFS(prefix)
	return gridFs.Find(query).One(result)
}

func GridFSFindAll(db, prefix string, query, result interface{}) error {
	ms, d := getDb(db)
	defer ms.Close()
	gridFs := d.GridFS(prefix)
	return gridFs.Find(query).All(result)
}

func GridFSOpen(db, prefix, name string) (*mgo.GridFile, error) {
	ms, d := getDb(db)
	defer ms.Close()
	gridFs := d.GridFS(prefix)
	return gridFs.Open(name)
}

func GridFSRemove(db, prefix, name string) error {
	ms, d := getDb(db)
	defer ms.Close()
	gridFs := d.GridFS(prefix)
	return gridFs.Remove(name)
}
