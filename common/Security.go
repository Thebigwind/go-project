
package common

import (
	"encoding/json"
	"strings"
)

const (
	SECURITY_MODE_ISOLATION int = 1
	SECURITY_MODE_STRICT    int = 2
)

const (
	UID_GLOBAL string = "Anoynmous"
)

type UserAccountInfo struct {
	Username  string `json:"user"`
	Groupname string `json:"group"`
	Uid       string `json:"uid"`
	Gid       string `json:"gid"`
	Umask     string `json:"umask"`
}

type SecurityContext struct {
	secID      string
	privileged bool
	account    *UserAccountInfo
}

func NewSecurityContext(account *UserAccountInfo) *SecurityContext {
	if account == nil {
		return &SecurityContext{
			secID: UID_GLOBAL,
		}
	}

	ctxt := &SecurityContext{
		secID:   account.Username,
		account: account,
	}

	if strings.ToUpper(account.Username) == "ROOT" {
		ctxt.SetPrivileged(true)
	}

	return ctxt
}

func (ctxt *SecurityContext) SetUserInfo(account *UserAccountInfo) {
	if account == nil {
		return
	}

	ctxt.account = account
	ctxt.secID = account.Username
	if strings.ToUpper(account.Username) == "ROOT" {
		ctxt.SetPrivileged(true)
	}
}

func (ctxt *SecurityContext) GetUserInfo() *UserAccountInfo {
	return ctxt.account
}

func (ctxt *SecurityContext) IsPrivileged() bool {
	return ctxt.privileged
}

func (ctxt *SecurityContext) SetPrivileged(pri bool) {
	ctxt.privileged = pri
}

func (ctxt *SecurityContext) ID() string {
	return ctxt.secID
}

func (ctxt *SecurityContext) CheckGroup(userGroups map[string]bool, pipelineOwnerGroups map[string]bool) bool {
	if ctxt.IsPrivileged() {
		return true
	}

	for pipelineOwnerGid, _ := range pipelineOwnerGroups {
		if _, ok := userGroups[pipelineOwnerGid]; ok {
			return true
		} else {
			continue
		}
	}

	return false
}

func (ctxt *SecurityContext) CheckSecID(id string) bool {
	if ctxt.IsPrivileged() {
		return true
	}

	return ctxt.secID == id
}

func (ctxt *SecurityContext) ToJSON() (error, string) {
	body, err := json.Marshal(ctxt.account)
	if err != nil {
		Logger.Errorf("Fail to marshal security context\n")
		return err, ""
	}

	return nil, string(body)
}

func (ctxt *SecurityContext) FromJSON(jsonData string) error {
	account := &UserAccountInfo{}
	err := json.Unmarshal([]byte(jsonData), account)
	if err != nil {
		return err
	}

	ctxt.SetUserInfo(account)
	return nil
}
