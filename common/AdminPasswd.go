

package common

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
)

const (
	PROJECT_CLIENT_ADMIN_PASSFILE         = "/etc/.xxx_admin.key" ///etc/.PROJECT_admin.key
	PROJECT_INITIAL_ADMIN_PASSWORD string = "test@lu"
	PROJECT_ADMIN_PASSWORD_SEED    string = "test@lusthefuture"
	PROJECT_ADMIN_PASSWORD_SHADOW  string = "test@luShad0w"
)

func ClientConfDir() string {
	user, err := user.Current()
	if err != nil {
		Logger.Errorf("Fail to get current user: %s\n",
			err.Error())
		return ""
	}
	return user.HomeDir
}

/*
 * encrypt admin passwd by PROJECT_ADMIN_PASSWORD_SHADOW at local
 * encrypt admin passwd by PROJECT_ADMIN_PASSWORD_SEED at ETCD
 * admin skey = encrypt admin passwd by PROJECT_SECURITY_KEY_SEED
 */

func SaveAdminPassword(pass string) error {
	tmpfile, err := ioutil.TempFile(ClientConfDir(), "PROJECT_adminpass")
	if err != nil {
		fmt.Println("Fail to create temp password file", err)
		return err
	}
	defer os.Remove(tmpfile.Name())

	aesPass := NewAesEncrypt(PROJECT_ADMIN_PASSWORD_SHADOW)
	passEnc, err := aesPass.Encrypt(pass)

	if err != nil {
		fmt.Println("Failed to encrypt password!")
		return err
	}

	if _, err := tmpfile.WriteString(passEnc); err != nil {
		fmt.Println("Write temp password error!", err)
		return err
	}

	err = os.Rename(tmpfile.Name(), PROJECT_CLIENT_ADMIN_PASSFILE)
	if err != nil {
		fmt.Println("Failed to rename to PROJECT client admin passfile:%s!\n", err.Error())
		return err
	}

	tmpfile.Close()

	return nil
}

func LoadAdminPassword() (string, error) {
	_, err := os.Stat(PROJECT_CLIENT_ADMIN_PASSFILE)
	if os.IsNotExist(err) {
		return "", nil
	}

	o, err := ioutil.ReadFile(PROJECT_CLIENT_ADMIN_PASSFILE)
	if err != nil {
		return "", errors.New("Failed to read password file!")
	}

	aesPass := NewAesEncrypt(PROJECT_ADMIN_PASSWORD_SHADOW)
	pass, err := aesPass.Decrypt(string(o))
	if err != nil {
		return "", errors.New("Failed to decrypt password.")
	}

	return pass, nil
}

func LoadAdminSkey() (string, error) {

	_, err := os.Stat(PROJECT_CLIENT_ADMIN_PASSFILE)
	if os.IsNotExist(err) {
		return PROJECT_INITIAL_ADMIN_PASSWORD, nil
	}

	o, err := ioutil.ReadFile(PROJECT_CLIENT_ADMIN_PASSFILE)
	if err != nil {
		fmt.Println("Failed to read password file.", err.Error())
		return "", errors.New("Failed to read password file.")
	}

	aesPass := NewAesEncrypt(PROJECT_ADMIN_PASSWORD_SHADOW)
	pass, err := aesPass.Decrypt(string(o))
	if err != nil {
		fmt.Println("Failed to decrypt shadow passward.", err.Error())
		return "", errors.New("Failed to decrypt shadow password.")
	}

	aesSkey := NewAesEncrypt(PROJECT_SECURITY_KEY_SEED)
	skey, err := aesSkey.Encrypt(pass)
	if err != nil {
		fmt.Println("Failed to get skey for admin.", err.Error())
		return "", errors.New("Failed to get security key for admin.")
	}

	return skey, err
}
