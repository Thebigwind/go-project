func NewDBServiceMultiDB(etcdConfig MetaEtcdConfig) error {
	var config DBServiceConfig

	//Publish v2 at 2017-11-31 without multiDB
	dbConfList := etcdConfig.DBConfig
	for _, dbConf := range dbConfList {

		//dbConf := etcdConfig.DBConfig
		config = dbConf
		Logger.Infof("Multi DBConfig, config:%s", config)

		stat := &dbStat{
			queryErr: 0,
			execErr:  0,
		}

		dbService := &dbService{
			server: config.Server,
			port:   config.Port,
			user:   config.User,
			pass:   config.Password,
			dbname: config.DBName,
			driver: config.Driver,
			stat:   stat,
		}

		dbOpts := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", dbService.user,
			dbService.pass, dbService.dbname)

		DBLogger.Errorf("%v", config)

		if dbService.server != "" {
			dbOpts += " host=" + dbService.server
		}
		if dbService.port != "" {
			dbOpts += " port=" + dbService.port
		}
		db, err := sql.Open(dbService.driver, dbOpts)
		if err != nil {
			DBLogger.Errorf("Fail to open database, error %s \n", err.Error())
			return nil
		}
		dbService.db = db

		connected := false
		for i := 0; i <= dbReconnectCount; i++ {
			err = db.Ping()
			if err != nil {
				DBLogger.Errorf("database connection failed, retry...")
				time.Sleep(dbReconnectPeriod)
			} else {
				DBLogger.Infof("database connection successfully! \n")
				connected = true

				break
			}
		}

		if connected == false {
			DBLogger.Infof("give up to connect database, failed...! \n")
		}

		globalDBService = append(globalDBService, dbService)
	}

	return nil
}