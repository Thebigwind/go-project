package dbservice
import (
	"bytes"
	"database/sql"
	"regexp"

	"errors"
	"fmt"
	_ "github.com/lib/pq"
	"strconv"
	"strings"
	"time"
)
func (db *dbService) ChownDataOwner(NewUserUid, NewUser string,
	Key, Value, Uid string) error {

	// Test if the user have this tag
	querySql := "SELECT xa_id from t_test_main WHERE xa_name = $1 AND xa_uid = " + Uid
	querySql += " AND xa_value = $2 "

	DBLogger.Debugf("querySql:%s", querySql)
	rows, err := db.Query(querySql, Key, Value)
	if err != nil {
		db.QueryErr(err)
		DBLogger.Errorf("DBService: query error %s\n", err.Error())
		return err
	}
	defer rows.Close()
	if !rows.Next() { // the user do not have the tag
		return NewXtErrorInfo(META_NOT_FOUND_ERR, "")
	}

	// Test if the new user is in the blacklist
	querySql = "SELECT xa_id from t_test_main WHERE black_list @> string_to_array($3, '') "
	querySql += " AND xa_name = $1 AND xa_uid = " + Uid
	querySql += " AND xa_value = $2  "

	DBLogger.Debugf("querySql:%s", querySql)
	if rows, err = db.Query(querySql, Key, Value, NewUser); err != nil {
		DBLogger.Errorf("DBService: query error %s\n", err.Error())
		return err
	}
	defer rows.Close()
	if rows.Next() { //the user in the blacklist
		xt_err := NewXtErrorInfo(BLACK_USER_EXIST_ERR, "")
		return xt_err
	}

	// Test if the NewUser already have the tag
	querySql = "SELECT xa_id from t_test_main WHERE xa_name = $1 AND xa_uid = " + NewUserUid
	querySql += " AND xa_value = $2 "

	DBLogger.Debugf("querySql:%s", querySql)
	if rows, err = db.Query(querySql, Key, Value); err != nil {
		db.QueryErr(err)
		DBLogger.Errorf("DBService: query error %s\n", err.Error())
		return err
	}
	defer rows.Close()
	if rows.Next() { //the NewUser already have the tag
		return NewXtErrorInfo(META_ALREADY_EXIST_ERR, "")
	}

	// Update the tag's info
	updateSql := "UPDATE t_test_main SET xa_uid=$3 WHERE xa_name=$1 AND xa_uid = $4 "
	updateSql += "  AND xa_value = $2  "

	DBLogger.Debugf("updateSql:%s", updateSql)
	err, affect := db.Exec2(updateSql, Key, Value, NewUserUid, Uid)
	if err != nil {
		DBLogger.Errorf("DBService: exec error %s\n", err.Error())
		return err
	}

	if affect < 1 {
		return NewXtErrorInfo(META_NOT_FOUND_ERR, "")
	}

	DBLogger.Debugf("Succeed chown meta ")
	return nil
}



func (db *dbService) GetFstype(fs_id string) (error, string) {

	var fs_type string = ""
	querySql := "SELECT fs_type FROM t_file_system_main where fs_id = $1"
	rows, err := db.Query(querySql, fs_id)
	if err != nil {
		db.QueryErr(err)
		DBLogger.Errorf("DBService: query error %s", err.Error())
		return err, fs_type
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&fs_type)
		if err != nil {
			DBLogger.Errorf("DBService: scan error %s", err.Error())
			xt_err := NewXtErrorInfo(DB_SCAN_ERR, "")
			return xt_err, ""
		}
	}
	return err, fs_type
}

/*
 * DBUser DB Functions BEGIN
 */

func (db *dbService) AddUser(userInfo *DBUserEntry) error {
	insertSql := "INSERT INTO t_user_test(user_id,user_name,user_type,user_info,create_ts,modify_ts) "
	insertSql += " VALUES($1,$2,$3,$4,$5,$6)"
	DBLogger.Debugf(" add user sql:%s", insertSql)

	err := db.Exec(
		insertSql,
		userInfo.UserId,
		userInfo.UserName,
		userInfo.PrimaryGroup,
		userInfo.UserInfo,
		time.Now().Unix(),
		time.Now().Unix())

	if err != nil {
		DBLogger.Errorf("Fail to add user %s ", err.Error())
	}

	return err
}


func (db *dbService) AddUsers(usersInfo []DBUserEntry) error {
	valueStrings := []string{}
	valueArgs := []interface{}{}

	for i, user := range usersInfo {

		index1 := strconv.Itoa(6*i + 1)
		index2 := strconv.Itoa(6*i + 2)
		index3 := strconv.Itoa(6*i + 3)
		index4 := strconv.Itoa(6*i + 4)
		index5 := strconv.Itoa(6*i + 5)
		index6 := strconv.Itoa(6*i + 6)

		var st bytes.Buffer
		st.WriteString("(")
		st.WriteString("$")
		st.WriteString(index1)
		st.WriteString(",")
		st.WriteString("$")
		st.WriteString(index2)
		st.WriteString(",")
		st.WriteString("$")
		st.WriteString(index3)
		st.WriteString(",")
		st.WriteString("$")
		st.WriteString(index4)
		st.WriteString(",")
		st.WriteString("$")
		st.WriteString(index5)
		st.WriteString(",")
		st.WriteString("$")
		st.WriteString(index6)
		st.WriteString(")")

		valueStrings = append(valueStrings, st.String())
		//valueStrings = append(valueStrings, "(?, ?, ?, ?, ?, ?)")

		userId, _ := strconv.Atoi(user.UserId)

		if userId == 0 {
			//fmt.Println(user)
			user.UserName = "root"
			user.PrimaryGroup = "0"
		}
		if user.PrimaryGroup == "" {
			return errors.New("the primary group empty.")
		}

		valueArgs = append(valueArgs, userId)
		valueArgs = append(valueArgs, user.UserName)
		valueArgs = append(valueArgs, user.PrimaryGroup)
		valueArgs = append(valueArgs, user.UserInfo)
		valueArgs = append(valueArgs, time.Now().Unix())
		valueArgs = append(valueArgs, time.Now().Unix())
	}

	smt := "INSERT INTO t_user_test(user_id,user_name,user_type,user_info,create_ts,modify_ts) "
	smt += "     VALUES %s ON CONFLICT (user_id)  "
	smt += " DO nothing "
	//smt += "  DO UPDATE SET user_id = excluded.user_id, user_name = excluded.user_name, user_type = excluded.user_type "
	smt = fmt.Sprintf(smt, strings.Join(valueStrings, ","))

	DBLogger.Debugf("import users:%s", smt)

	tx, _ := db.Begin()
	_, err := tx.Exec(smt, valueArgs...)

	if err != nil {
		tx.Rollback()
		DBLogger.Errorf("Fail to add users %s ", err.Error())
	}

	return tx.Commit()
}



func (db *dbService) GetDBUsers() (error, []DBUserEntry) {
	rows, err := db.Query("SELECT * FROM t_user_test ORDER BY user_id")
	if err != nil {
		db.QueryErr(err)
		DBLogger.Errorf("DBService: query error %s", err.Error())
		return err, nil
	}
	defer rows.Close()

	keys := make([]DBUserEntry, 0, MAX_DB_COUNT)
	for rows.Next() {
		var createTsStamp int
		var modifyTsStamp int
		key := DBUserEntry{}
		err = rows.Scan(
			&key.UserId,
			&key.UserName,
			&key.PrimaryGroup,
			&key.UserInfo,
			&createTsStamp,
			&modifyTsStamp)

		if err != nil {
			DBLogger.Errorf("DBService: get dbusers keys fail %s",
				err.Error())
			xt_err := NewXtErrorInfo(DB_SCAN_ERR, "")
			return xt_err, nil
		}

		key.CreateTs = GetStampTimeString(createTsStamp)
		key.ModifyTs = GetStampTimeString(modifyTsStamp)
		keys = append(keys, key)
	}

	return nil, keys
}


func (db *dbService) SetMetaObj(key, value, authority, uid, objId, clusterId string) error {

	modeMap := map[string]int{
		"public": 1, "group": 2, "private": 4,
	}
	modeMap2 := map[int]string{
		1: "public", 2: "group", 4: "private",
	}

	querySql := " SELECT xa_mode FROM t_test_main WHERE xa_name = $1 AND xa_value = $2 AND xa_uid = $3"
	DBLogger.Debugf("querySql:%s", querySql)

	rows, err := db.Query(querySql, key, value, uid)
	if err != nil {
		db.QueryErr(err)
		DBLogger.Errorf("DBService: set meta error : %s\n", err.Error())
		return err
	}
	defer rows.Close()

	var xaMode int = 0
	if rows.Next() {
		err = rows.Scan(&xaMode)
		if err != nil {
			xt_err := NewXtErrorInfo(DB_SCAN_ERR, "")
			return xt_err
		}
		if xaMode != modeMap[authority] {
			xaAuthority := modeMap2[xaMode]
			DBLogger.Errorf("DBService: already exists the tag mode is  : " + xaAuthority)
			return NewXtErrorInfo(META_AUTHORITY_ALREADY_EXIST_ERR, "")
		}
	}

	/*
		//insert xa_main, xa_map
		insertSql := " DO $$ "
		insertSql += " DECLARE "
		insertSql += "	v_xa_id bigint; "
		insertSql += " begin "
		insertSql += " INSERT INTO t_test_main (xa_name, xa_value, xa_mode, xa_uid, create_ts) "
		insertSql += " SELECT E'" + filterSqlInject(key) + "', E'" + filterSqlInject(value) + "', " + strconv.Itoa(modeMap[authority]) + "," + uid + ", extract(epoch from now())::bigint"
		insertSql += " WHERE NOT EXISTS ("
		insertSql += "       SELECT xa_id FROM t_test_main "
		insertSql += "        WHERE xa_name = E'" + filterSqlInject(key) + "' AND xa_value = E'" + filterSqlInject(value) + "' AND xa_uid =" + uid
		insertSql += "       );"
		insertSql += " SELECT xa_id into v_xa_id from t_test_main"
		insertSql += "  WHERE xa_name = E'" + filterSqlInject(key) + "' AND xa_value = E'" + filterSqlInject(value) + "' AND xa_uid = " + uid + ";"
		insertSql += " DELETE FROM r_xa_map a "
		insertSql += "  WHERE a.obj_id = '" + objId + "' AND a.fs_id = '" + clusterId + "' AND a.xa_id = v_xa_id; "
		insertSql += " INSERT INTO r_xa_map (xa_id, obj_id, fs_id, create_ts)"
		insertSql += " SELECT v_xa_id, '" + objId + "', '" + clusterId + "', extract(epoch from now())::bigint ;"
		insertSql += " end "
		insertSql += "$$;"
		DBLogger.Debugf("SetMeta Sql: %s", insertSql)
		err = db.Exec(insertSql)
		if err != nil {
			DBLogger.Errorf("err:", err.Error())
		}
		return err
	*/

	err = db.Exec("select * from set_avu($1, $2, $3, $4, $5, $6, $7)", objId, clusterId, key, value, uid, modeMap[authority], time.Now().Unix())
	if err != nil {
		DBLogger.Errorf("err:", err.Error())
	}
	return err

}

func getUserAuthorityCode(constraint bool, uid int, gids []int, isAdmin bool, tableName string) string {
	authoritySql := ""
	if !constraint {
		if !isAdmin {
			gidsStr := DealType(gids)

			authoritySql += " AND ("
			authoritySql += tableName + ".uid = " + strconv.Itoa(uid) + " OR "
			authoritySql += tableName + ".mode::int & (1 << 2) > 0 OR "
			authoritySql += "  (" + tableName + ".mode::int & (1 << 5) > 0 AND ARRAY[" + tableName + ".gid::int] && ARRAY" + gidsStr + "::int[]  = true )"
			authoritySql += ") "
		}
	} else {
		authoritySql += " AND " + tableName + ".uid = " + strconv.Itoa(uid)
	}
	return authoritySql
}

func getTagUserAuthorityCode(uid int, user string, relevantUids []int, tagExtension bool, isAdmin bool, tableName string) string {
	authoritySql := ""

	if tagExtension {
		if !isAdmin {
			relevantUidsStr := DealType(relevantUids)

			authoritySql += " AND ( "
			authoritySql += tableName + ".xa_uid = " + strconv.Itoa(uid)
			authoritySql += " OR " + tableName + ".xa_mode = 1 " //public
			authoritySql += " OR (" + tableName + ".xa_mode = 2 AND ARRAY[" + tableName + ".xa_uid] && ARRAY" + relevantUidsStr + "::int[] = true )"
			authoritySql += " OR " + tableName + ".white_list @> '{" + user + "}') "
			authoritySql += " AND NOT (CASE WHEN " + tableName + ".black_list IS NULL THEN '{}'::TEXT[] ELSE " + tableName + ".black_list END @> '{" + user + "}') "
		}
	} else {
		authoritySql += " AND  " + tableName + ".xa_uid = " + strconv.Itoa(uid)
	}
	return authoritySql
}