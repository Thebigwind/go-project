//数据库连接池测试
packagemain
 
import(
    "database/sql"
    "fmt"
    _ "github.com/go-sql-driver/mysql"
    "log"
    "net/http"
)
 
var globalMysqlService *dbService

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
    return globalMysqlService
}


func createNewDBService(config MysqlDBConfig) (*dbService, error) {

    dbService := &dbService{
        server: config.Server,
        port:   config.Port,
        user:   config.User,
        pass:   config.Password,
        dbname: config.DBName,
        driver: config.Driver,
    }

    dbOpts := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s dbname=%s?charset=utf8", dbService.user,
        dbService.pass, dbService.server, dbService.port, dbService.dbname)

    db, err := sql.Open(dbService.driver, dbOpts)
    if err != nil {
        DBLogger.Errorf("Fail to open database, error %s \n", err.Error())
        return nil, err
    }

    if config.MaxOpenConns <= 0{
        config.MaxOpenConns = 100
    }
    if config.MaxIdleConns <= 0{
        config.MaxIdleConns = 15
    }

    DB.SetConnMaxLifetime(100*time.Second)  //最大连接周期，超过时间的连接就close
    db.SetMaxOpenConns(config.MaxOpenConns) //最大连接数
    db.SetMaxIdleConns(config.MaxIdleConns) //空闲连接数
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

func NewDBService(config MysqlDBConfig) error {
    
    DBLogger.Infof("Single DBConfig, config:%v", config)

    dbService, err := createNewDBService(config)
    if err != nil {
        Logger.Errorf("Connect db service failed!\n")
        return nil
    }
    globalMysqlService = dbService

    return nil
}



//============================================================//



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





