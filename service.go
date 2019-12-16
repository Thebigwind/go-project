package project

import (
	"errors"
	"fmt"
	"strings"

	postgres "github.com/bigwind//project/dbservice/postgres"
	. "github.com/bigwind/project/common"
	arangodb "github.com/bigwind/project/dbservice/arangodb"
	"github.com/bigwind/project/routine"
	project "github.com/bigwind/project/server"
	. "github.com/bigwind/project/server/manager"
)

const (
	DEFAULT_ETCD_SERVER string = "http://127.0.0.1:2379"
	DEFAULT_REST_PORT   string = "7789"
	DEFAULT_LOG_FILE    string = "/var/log/project/project.log"
	DEFAULT_LOG_LEVEL   int    = 0

	ETCD_KEY_CONFIG_project string = "/datamgmt/project"
)

/*
 *
 * project Daemon Start
 *
 */
func ServerStart(arguments map[string]string) error {
	log_file := arguments["log-file"]
	log_level := arguments["log-level"]
	conf_file := arguments["config-file"]

	/*
	 * Step 1: initialize the logger facility.
	 */
	var loglvl int
	switch log_level {
	case "INFO":
		loglvl = LOG_LEVEL_INFO //int = 0
	case "DEBUG":
		loglvl = LOG_LEVEL_DEBUG //int = 1
	case "WARN":
		loglvl = LOG_LEVEL_WARN //int = 2
	case "ERROR":
		loglvl = LOG_LEVEL_ERROR //int = 3
	default:
		return errors.New("Invalid log level")
	}
	loggerConfig := LoggerConfig{
		Logfile:  log_file,
		LogLevel: loglvl,
	}
	err := LoggerInit(&loggerConfig)
	if err != nil {
		return err
	}

	/*
	 * Step 2: parse configure infomation from config file.
	 */
	projectConfig := ParseprojectConfig(conf_file)
	if projectConfig == nil {
		return errors.New("Invalid configure file")
	}
	etcdHostItems := strings.Split(projectConfig.EtcdList, ",")
	etcdEndPoints := make([]string, 0, len(etcdHostItems))
	for i := 0; i < len(etcdHostItems); i++ {
		if etcdHostItems[i] != "" {
			etcdEndPoints = append(etcdEndPoints, etcdHostItems[i])
		}
	}

	/*
	 * Step 3: initialize xxx logger if enable.
	 */

	/*
	 * Step 4: project connect to postgres.
	 * if db connection failed, would keep retry, always at step 2.
	 */
	Logger.Infof("ServerStart NewDBService")
	if projectConfig.DBConfig.Driver == "postgres" {
		projectConfig.DBType = DBTYPE_POSTGRES
		err = postgres.NewDBService(*projectConfig)
	} else if projectConfig.DBConfig.Driver == "arangodb" {
		projectConfig.DBType = DBTYPE_ARANGODB
		err = arangodb.NewDBService(*projectConfig)
	}
	if err != nil {
		Logger.Errorf("Fail to init DB service %s", err)
		return errors.New("Fail to init database")
	}

	/*
	 * Step 5: init cache manager
	 */
	Logger.Infof("project now init cache manager.")
	cacheMgr := NewCacheMgr()
	if cacheMgr == nil {
		Logger.Errorf("Fail to init cache manager")
		return errors.New("Fail to init cache manager")
	}
	//init user cache manager
	Logger.Infof("project now init usercache manager.")
	usercacheMgr := NewUserCacheMgr()
	if usercacheMgr == nil {
		Logger.Errorf("Fail to init usercache manager")
		return errors.New("Fail to init usercache manager")
	}

	/*
	 * Step 8: init user manager and start watch routine
	 */
	err = InitGlobalContext(etcdEndPoints)
	if err != nil {
		Logger.Errorf("Fail to init etcd config.")
		return errors.New("Fail to init etcd config")
	}

	if projectConfig.UserConfig.EnableLdap == true {
		//get user from ldap
		Logger.Infof("project now init ldap manager for sync user.")
		ldapMgr := NewLdapMgr()
		go func() {
			Logger.Infof("project start init ldap user at once.")
			err := ldapMgr.ExportFromLdap()
			if err != nil {
				Logger.Errorf("Fail to export user from ldap.")
				//return errors.New("Fail to export user from ldap.")
			}
		}()

		ldapMgr.Start()
	} else {
		//get user from etcd
		Logger.Infof("project now init power manager for sync user.")
		powerMgr := NewPowerMgr()

		go func() {
			Logger.Infof("project start init etc user at once.")
			err := powerMgr.ExportFromEtcd()
			if err != nil {
				Logger.Errorf("Fail to export user from etcd.")
				//return errors.New("Fail to export user from etcd.")
			}
		}()

		powerMgr.Start()
	}



	/*
		Local development test, no ldap environment, import local users to the cache
	*/
	if projectConfig.Test {
		err := ExportFromLocal()
		if err != nil {
			Logger.Infof("ExportFromLocal fail :", err)
		} else {
			Logger.Infof("ExportFromLocal succeed.")
		}
	}
	/*
	 * Step last: start restful-api server.
	 */
	port := projectConfig.RestConfig.Port
	if port == "" {
		port = DEFAULT_REST_PORT
	}
	restAddr := fmt.Sprintf(":%s", port)
	Logger.Infof("Start REST Server on %s", restAddr)
	Server := project.NewRESTServer(restAddr)
	Server.StartRESTServer(projectConfig.RestConfig.ServerLimit)

	return nil
}
