package common

import (
	//"errors"
	"bytes"
	"strconv"
	"strings"
	"sync"
)

type UserCacheMgr struct {
	// User LRU cache, key is user and value is uid
	UserLRU  *LRU
	UserLock *sync.RWMutex
	// Group LRU cache, key is group and value is gid
	GroupLRU  *LRU
	GroupLock *sync.RWMutex
	// Uid LRU cache, key is uid and value is user
	UidLRU  *LRU
	UidLock *sync.RWMutex
	// Gid LRU cache, key is gid and value is group
	GidLRU  *LRU
	GidLock *sync.RWMutex
	// RelevantUids LRU cache, key is uid and value is RelevantUids
	RelevantUidsLRU  *LRU
	RelevantUidsLock *sync.RWMutex
	// Gids LRU cache, key is uid and value is gids
	GidsLRU  *LRU
	GidsLock *sync.RWMutex
}

const (
	//use Put() function,not check len; newLRU need a positive number
	CACHE_USER_LRU_SIZE          int = 1024
	CACHE_GROUP_LRU_SIZE         int = 1024
	CACHE_UID_LRU_SIZE           int = 1024
	CACHE_GID_LRU_SIZE           int = 1024
	CACHE_GIDS_LRU_SIZE          int = 1024
	CACHE_RELEVANT_UIDS_LRU_SIZE int = 1024
)

var GlobalUserCacheMgr *UserCacheMgr = nil

func GetUserCacheMgr() *UserCacheMgr {
	return GlobalUserCacheMgr
}

func NewUserCacheMgr() *UserCacheMgr {

	Logger.Println("======User Cache Mananger======")

	userLRU, err := NewLRU(CACHE_USER_LRU_SIZE, nil)
	if err != nil {
		Logger.Error("NewLRU for user failed %s", err.Error())
		return nil
	}
	userLock := new(sync.RWMutex)

	groupLRU, err := NewLRU(CACHE_GROUP_LRU_SIZE, nil)
	if err != nil {
		Logger.Error("NewLRU for group failed %s", err.Error())
		return nil
	}
	groupLock := new(sync.RWMutex)

	uidLRU, err := NewLRU(CACHE_UID_LRU_SIZE, nil)
	if err != nil {
		Logger.Error("NewLRU for uid failed %s", err.Error())
		return nil
	}
	uidLock := new(sync.RWMutex)

	gidLRU, err := NewLRU(CACHE_GID_LRU_SIZE, nil)
	if err != nil {
		Logger.Error("NewLRU for gid failed %s", err.Error())
		return nil
	}
	gidLock := new(sync.RWMutex)

	relevantUidsLRU, err := NewLRU(CACHE_RELEVANT_UIDS_LRU_SIZE, nil)
	if err != nil {
		Logger.Error("NewLRU for relevantUids failed %s", err.Error())
		return nil
	}
	relevantUidsLock := new(sync.RWMutex)

	gidsLRU, err := NewLRU(CACHE_GIDS_LRU_SIZE, nil)
	if err != nil {
		Logger.Error("NewLRU for gids failed %s", err.Error())
		return nil
	}
	gidsLock := new(sync.RWMutex)

	userCacheMgr := &UserCacheMgr{
		UserLRU:          userLRU,
		UserLock:         userLock,
		GroupLRU:         groupLRU,
		GroupLock:        groupLock,
		UidLRU:           uidLRU,
		UidLock:          uidLock,
		GidLRU:           gidLRU,
		GidLock:          gidLock,
		RelevantUidsLRU:  relevantUidsLRU,
		RelevantUidsLock: relevantUidsLock,
		GidsLRU:          gidsLRU,
		GidsLock:         gidsLock,
	}

	GlobalUserCacheMgr = userCacheMgr
	return userCacheMgr
}

func (mgr *UserCacheMgr) GetUser(uid int) (error, string) {
	var ok bool
	var found bool
	var value interface{}
	var user string

	mgr.UidLock.Lock()
	defer mgr.UidLock.Unlock()
	value, found = mgr.UidLRU.Get(uid)
	if found {
		user, ok = value.(string)
		if ok == false {
			Logger.Errorf("get user by uid %d from cache is not string", uid)
			return NewXtErrorInfo(DATA_TYPE_PARSE_ERR, ""), ""
		}
	} else {
		return NewXtErrorInfo(USER_NOT_EXIST_ERR, ""), ""
	}
	return nil, user
}

func (mgr *UserCacheMgr) GetUid(user string) (error, int) {
	var ok bool
	var found bool
	var value interface{}
	var uid int

	mgr.UserLock.Lock()
	defer mgr.UserLock.Unlock()
	value, found = mgr.UserLRU.Get(user)
	if found {
		uid, ok = value.(int)
		if ok == false {
			Logger.Errorf("get uid by user %s from cache is not int", user)
			return NewXtErrorInfo(DATA_TYPE_PARSE_ERR, ""), 0
		}
	} else {
		return NewXtErrorInfo(USER_NOT_EXIST_ERR, ""), 0

	}
	return nil, uid
}

func (mgr *UserCacheMgr) GetAllUser() (error, map[int]string) {

	userMap := make(map[int]string)

	mgr.UidLock.Lock()
	defer mgr.UidLock.Unlock()

	maps, found := mgr.UidLRU.GetAll()
	if found {
		for k, v := range maps {
			uid, ok := k.(int)
			if !ok {
				Logger.Errorf("get uid %d from cache is not int", uid)
				return NewXtErrorInfo(DATA_TYPE_PARSE_ERR, ""), nil
			}

			user, ok := v.(string)
			if !ok {
				Logger.Errorf("get user %s from cache is not string", user)
				return NewXtErrorInfo(DATA_TYPE_PARSE_ERR, ""), nil
			}
			userMap[uid] = user
		}
		return nil, userMap
	} else {
		return NewXtErrorInfo(GROUP_NOT_EXIST_ERR, ""), nil
	}
}

func (mgr *UserCacheMgr) GetAllGroup() (error, map[int]string) {

	groupMap := make(map[int]string)

	mgr.GidLock.Lock()
	defer mgr.GidLock.Unlock()

	maps, found := mgr.GidLRU.GetAll()
	if found {
		for k, v := range maps {
			gid, ok := k.(int)
			if !ok {
				Logger.Errorf("get gid %d from cache is not int", gid)
				return NewXtErrorInfo(DATA_TYPE_PARSE_ERR, ""), nil
			}

			group, ok := v.(string)
			if !ok {
				Logger.Errorf("get group %s from cache is not string", group)
				return NewXtErrorInfo(DATA_TYPE_PARSE_ERR, ""), nil
			}
			groupMap[gid] = group
		}
		return nil, groupMap
	} else {
		return NewXtErrorInfo(GROUP_NOT_EXIST_ERR, ""), nil
	}
}

func (mgr *UserCacheMgr) GetGroup(gid int) (error, string) {
	var ok bool
	var found bool
	var value interface{}
	var group string

	mgr.GidLock.Lock()
	defer mgr.GidLock.Unlock()
	value, found = mgr.GidLRU.Get(gid)
	if found {
		group, ok = value.(string)
		if ok == false {
			Logger.Errorf("get group by gid %d from cache is not string", gid)
			return NewXtErrorInfo(DATA_TYPE_PARSE_ERR, ""), ""
		}
	} else {
		return NewXtErrorInfo(GROUP_NOT_EXIST_ERR, ""), ""
	}
	return nil, group
}

func (mgr *UserCacheMgr) GetGid(group string) (error, int) {
	var ok bool
	var found bool
	var value interface{}
	var gid int

	mgr.GroupLock.Lock()
	defer mgr.GroupLock.Unlock()
	value, found = mgr.GroupLRU.Get(group)
	if found {
		gid, ok = value.(int)
		if ok == false {
			Logger.Errorf("get gid by group %s from cache is not int", group)
			return NewXtErrorInfo(DATA_TYPE_PARSE_ERR, ""), 0
		}
	} else {
		return NewXtErrorInfo(GROUP_NOT_EXIST_ERR, ""), 0
	}
	return nil, gid
}

func (mgr *UserCacheMgr) GetRelevantUids(uid int) (error, []int) {
	var ok bool
	var found bool
	var value interface{}
	var uids []int

	mgr.RelevantUidsLock.Lock()
	defer mgr.RelevantUidsLock.Unlock()
	value, found = mgr.RelevantUidsLRU.Get(uid)
	if found {
		uids, ok = value.([]int)
		if ok == false {
			Logger.Errorf("get relevant uids by uid %s from cache is not []int", uid)
			return NewXtErrorInfo(DATA_TYPE_PARSE_ERR, ""), nil
		}
	} else {
		return NewXtErrorInfo(USER_NOT_EXIST_ERR, ""), nil
	}
	return nil, uids
}

func (mgr *UserCacheMgr) GetGids(uid int) (error, []int) {
	var ok bool
	var found bool
	var value interface{}
	var gids []int
	//var err error
	mgr.GidsLock.Lock()
	defer mgr.GidsLock.Unlock()
	value, found = mgr.GidsLRU.Get(uid)
	if found {
		gids, ok = value.([]int)
		if ok == false {
			Logger.Errorf("get gids by uid %s from cache is not []int", gids)
			return NewXtErrorInfo(DATA_TYPE_PARSE_ERR, ""), nil
		}
	} else {
		return NewXtErrorInfo(USER_NOT_EXIST_ERR, ""), nil
	}
	return nil, gids
}

func (mgr *UserCacheMgr) ClearUserCache() {
	Logger.Infof("ClearUserCache begin")
	mgr.UserLock.Lock()
	mgr.UserLRU.Purge()
	mgr.UserLock.Unlock()

	mgr.UidLock.Lock()
	mgr.UidLRU.Purge()
	mgr.UidLock.Unlock()

	mgr.GroupLock.Lock()
	mgr.GroupLRU.Purge()
	mgr.GroupLock.Unlock()

	mgr.GidLock.Lock()
	mgr.GidLRU.Purge()
	mgr.GidLock.Unlock()

	mgr.GidsLock.Lock()
	mgr.GidsLRU.Purge()
	mgr.GidsLock.Unlock()

	mgr.RelevantUidsLock.Lock()
	mgr.RelevantUidsLRU.Purge()
	mgr.RelevantUidsLock.Unlock()
	Logger.Infof("ClearUserCache finished")
}

func (mgr *UserCacheMgr) UpdateUserCache(ldapUsers map[string]UserEntry, ldapGroups map[string]GroupEntry) {

	locker := GetGlobeLocker()
	locker.Lock()
	defer locker.Unlock()

	Logger.Infof("UpdateUserCache begin")
	//add read lock,but can write.
	mgr.UserLock.RLock()
	mgr.UserLRU.Purge()

	mgr.UidLock.RLock()
	mgr.UidLRU.Purge()

	mgr.GroupLock.RLock()
	mgr.GroupLRU.Purge()

	mgr.GidLock.RLock()
	mgr.GidLRU.Purge()

	mgr.GidsLock.RLock()
	mgr.GidsLRU.Purge()

	mgr.RelevantUidsLock.RLock()
	mgr.RelevantUidsLRU.Purge()

	/*1 deal userLRU, uidLRU */
	for key, value := range ldapUsers {
		uid, _ := strconv.Atoi(value.Uid)

		mgr.UidLRU.Put(uid, key)
		mgr.UserLRU.Put(key, uid)
	}

	mgr.UserLock.RUnlock()
	mgr.UidLock.RUnlock()

	/*2 deal groupLRU, gidLRU */

	for gkey, gvalue := range ldapGroups {
		gid, _ := strconv.Atoi(gvalue.Gid)
		mgr.GidLRU.Put(gid, gkey)
		mgr.GroupLRU.Put(gkey, gid)
	}

	mgr.GroupLock.RUnlock()
	mgr.GidLock.RUnlock()

	/* deal gidsLRU*/

	for key, value := range ldapUsers {

		gids := make([]int, 0, len(ldapGroups))

		mgid, _ := strconv.Atoi(value.Gid)
		gids = append(gids, mgid) //ldapgroups中已包含主组的组,但组内不包含该用户

		for _, gvalue := range ldapGroups {
			ok := IsExistInArray(key, gvalue.Users)
			if ok {
				gid, _ := strconv.Atoi(gvalue.Gid)
				gids = append(gids, gid)
			}
		}
		uid, _ := strconv.Atoi(value.Uid)
		mgr.GidsLRU.Put(uid, gids)
	}
	mgr.GidsLock.RUnlock()

	/* deal relevantUidsLRU */
	for key, value := range ldapUsers {

		users := make([]string, 0, len(ldapUsers))
		groups := make([]string, 0, len(ldapGroups))
		relevantUids := make([]int, 0, len(ldapUsers))

		//找到用户所在的组 因为groups的名字是key,因此这里找的是group的名字
		for gkey, gvalue := range ldapGroups {
			//用户的主组
			if gvalue.Gid == value.Gid {
				groups = append(groups, gkey)
			} else { //
				ok := IsExistInArray(key, gvalue.Users)
				if ok {
					groups = append(groups, gkey)
				}
			}

		}

		//找到这些组的用户，拼接数组
		for _, gkey := range groups {
			users = append(users, ldapGroups[gkey].Users...)
		}

		//users过滤去重
		uniqueUsers := DeleteRepeat(users)

		//将users转换为uids
		for _, key := range uniqueUsers {
			relevantUid, _ := strconv.Atoi(ldapUsers[key].Uid)
			relevantUids = append(relevantUids, relevantUid)
		}

		uid, _ := strconv.Atoi(value.Uid)
		mgr.RelevantUidsLRU.Put(uid, relevantUids)
	}
	mgr.RelevantUidsLock.RUnlock()

	Logger.Infof("UpdateUserCache end")
}

////////////

func IsExistInArray(value string, arr []string) bool {
	for _, v := range arr {
		if v == value {
			return true
		}
	}
	return false
}

/*
func IsExistItem(value interface{}, arr []interface{})bool{
	switch reflect.TypeOf(arr).kind(){
	case reflect.Slice:
		s := reflect.ValueOf()
		for i := 0; i < s.Len(); i++{
			if reflect.DeepEqual(value,s.Index(i).interface()){
				return true
			}
		}
	}

	return false
}
*/

func ExportFromLocal() error {
	passwd := "/etc/passwd"
	group := "/etc/group"
	var min_uid uint64 = 1000
	var min_gid uint64 = 1000

	users, err := ParseUserFile(passwd)
	if err != nil {
		return err
	}

	groups, err := ParseGroupFile(group)
	if err != nil {
		return err
	}

	delUsers := make(map[string]string)
	for u, e := range users {
		i, err := strconv.ParseUint(e.Uid, 10, 64)
		if err != nil || i < min_uid {
			delete(users, u)
			delUsers[u] = e.Gid
			continue
		}

		i, err = strconv.ParseUint(e.Gid, 10, 64)
		if err != nil || i < min_gid {
			delete(users, u)
			continue
		}
	}

	for g, e := range groups {
		i, err := strconv.ParseUint(e.Gid, 10, 64)
		if err != nil || i < min_gid {
			delete(groups, g)
			continue
		}
	}

	for u, g := range delUsers {
		if grp, ok := groups[g]; ok {
			i, err := IndexOf(grp.Users, u)
			if err == nil {
				users := append(grp.Users[:i], grp.Users[i+1:]...)
				grp.Users = users
			}
		}
	}

	rootUser := UserEntry{
		Uid: "0",
		Gid: "0",
	}

	rootGroup := GroupEntry{
		Gid: "0",
	}
	//add root
	users["root"] = rootUser
	groups["root"] = rootGroup

	mgr := GetUserCacheMgr()

	mgr.UpdateUserCache(users, groups)

	return err
}

//get user from cache, if not exist, return the uid
func GetUsernameFromCache(uid int) string {
	userMgr := GetUserCacheMgr()

	err, user := userMgr.GetUser(uid)
	if err != nil {
		user = strconv.Itoa(uid)
	}
	return user
}

//get group from cache, if not exist, return the gid
func GetGroupnameFromCache(gid int) string {
	userMgr := GetUserCacheMgr()

	err, group := userMgr.GetGroup(gid)
	if err != nil {
		group = strconv.Itoa(gid)
	}
	return group
}

func GetUsernameByUidStr(uidstr string) string {
	userMgr := GetUserCacheMgr()

	uid, err := strconv.Atoi(uidstr)
	if err != nil {
		return uidstr
	}

	err, user := userMgr.GetUser(uid)
	if err != nil {

		return uidstr
	}
	return user
}

func GetGroupnameByGidStr(gidstr string) string {
	userMgr := GetUserCacheMgr()

	gid, err := strconv.Atoi(gidstr)
	if err != nil {
		return gidstr
	}

	err, group := userMgr.GetGroup(gid)
	if err != nil {
		return gidstr
	}
	return group
}

func GetArrayUser(uidstr string) string {

	uidStrs := strings.TrimPrefix(uidstr, "[")
	uidStrs = strings.TrimSuffix(uidStrs, "]")

	uidArr := strings.Split(uidStrs, ",")

	var st bytes.Buffer
	for _, v := range uidArr {
		user := GetUsernameByUidStr(v)
		st.WriteString(user)
		st.WriteString(",")
	}

	users := strings.TrimSuffix(st.String(), ",")

	return users
}

func GetUserArray(uids []int) []string {
	users := make([]string, 0, 20)
	var st bytes.Buffer
	for _, value := range uids {
		user := GetUsernameFromCache(value)
		st.WriteString(user)
		users = append(users, user)
	}

	return users
}
