package common

import (
	"bytes"
	//"context"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/pkg/term"
	"os/exec"
	"path"

	//"github.com/coreos/etcd/client"
	etcd "github.com/coreos/etcd/client"
	"io"
	"io/ioutil"
	//"log"
	. "github.com/xtao/xt_message/xt_api_message"
	"net/http"
	"os"
	"os/user"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

const (
	POWER_ENV              = "/etc/power.conf"
	POWER_DEFAULT_ENDPOINT = "http://127.0.0.1:2379"
	POWER_DEFAULT_SERVER   = "127.0.0.1:7789"
)

const (
	TEST_INTERVAL   = 1
	METACLI_ENV_DIR = "/tmp"
)

/*
 * If biocli gives endpoints, use endpoints. Otherwise,
 * try default client config file; if all tries fail,
 * just uses http://localhost:2379 as endpoints
 */
func parseEndpoints(endpoints string) []string {
	if endpoints == "" {
		_, err := os.Stat(POWER_ENV)
		if err != nil {
			endpoints = POWER_DEFAULT_ENDPOINT
		} else {
			raw, err := ioutil.ReadFile(POWER_ENV)
			if err == nil {
				var config map[string]interface{}
				err = json.Unmarshal(raw, &config)
				if err == nil {
					svrs, ok := config["endpoints"]
					if ok {
						endpoints = svrs.(string)
					}
				}
			}
		}
	}

	return strings.Split(endpoints, ",")
}

func ParsePowerServer(s string) string {
	var server string

	if s == "" {
		server = os.Getenv("METAVIEW_SERVER")
		if server == "" {
			server = POWER_DEFAULT_SERVER
		}
	} else {
		server = s
	}
	return server
}

func Contain(obj interface{}, target interface{}) (bool, error) {
	targetValue := reflect.ValueOf(target)
	switch reflect.TypeOf(target).Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < targetValue.Len(); i++ {
			if targetValue.Index(i).Interface() == obj {
				return true, nil
			}
		}
	case reflect.Map:
		if targetValue.MapIndex(reflect.ValueOf(obj)).IsValid() {
			return true, nil
		}
	}

	return false, errors.New("not in array")
}

func IndexOf(array interface{}, obj interface{}) (int, error) {
	value := reflect.ValueOf(array)

	switch reflect.TypeOf(array).Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < value.Len(); i++ {
			if value.Index(i).Interface() == obj {
				return i, nil
			}
		}
		return -1, errors.New("not in array")
	default:
		return 0, errors.New("not supported type")
	}
}

func WriteJSON(status int, data interface{}, w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

func CopyFile(src, dst string) error {
	stat, err := os.Stat(src)
	if err != nil {
		return err
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}

	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}

	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	err = os.Chmod(dst, stat.Mode())
	if err != nil {
		return err
	}

	return nil
}

func IsConnectionError(err error) bool {
	reg := regexp.MustCompile("connection?")

	return reg.FindStringIndex(err.Error()) != nil
}

func IndentPrint(indent bool, format string, a ...interface{}) {
	if indent == true {
		fmt.Printf("    ")
	}
	fmt.Printf(format, a...)
}

const (
	GET_JOBID_FROM_MEM int = 1
	GET_JOBID_FROM_DB  int = 1 << 1
	GET_JOBID_FROM_ALL int = GET_JOBID_FROM_MEM | GET_JOBID_FROM_DB
)

const BIOFLOW_TIME_LAYOUT string = "2006-01-02 15:04:05"

func GenBackendId(t string, server string, port string) string {
	return fmt.Sprintf("%s:%s:%s", t, server, port)
}

func KeyNotFound(err error) bool {
	if err != nil {
		if etcdError, ok := err.(etcd.Error); ok {
			if etcdError.Code == etcd.ErrorCodeKeyNotFound {
				return true
			}
		}
	}
	return false
}

func GetTimeStamp(timeString string, timeEnd bool) int {
	/* "2017-05-21 17:15:04" to 1495358104 int*/
	if timeString == "" {
		if timeEnd {
			return 4294967295 //2147483647 max value of int32, 4294967295 max of oid(uint32)
		} else {
			return 0
		}
	}

	loc, _ := time.LoadLocation("Local")
	theTime, _ := time.ParseInLocation("2006-01-02 15:04:05", timeString, loc)
	timeStamp := theTime.Unix()
	return int(timeStamp)
}

func GetSizeUnit(inputSize int64, sizeConv bool) string {
	//Need or not to keep size in origin number format
	if sizeConv {
		return strconv.FormatInt(inputSize, 10)
	}

	var middle float64
	size := float64(inputSize)
	//fmt.Printf("size=%d", size)

	sizeString := "N/A"
	if size >= 0 && size < 1024 {
		middle = size
		middleString := fmt.Sprintf("%.0f B", middle)
		return middleString
	} else if size >= 1024 && size < 1024*1024 {
		middle = size / 1024
		middleString := fmt.Sprintf("%.1f KB", middle)
		return middleString
	} else if size >= 1024*1024 && size < 1024*1024*1024 {
		middle = size / 1024 / 1024
		middleString := fmt.Sprintf("%.1f MB", middle)
		return middleString
	} else if size >= 1024*1024*1024 && size < 1024*1024*1024*1024 {
		middle = size / 1024 / 1024 / 1024
		middleString := fmt.Sprintf("%.1f GB", middle)
		return middleString
	} else if size >= 1024*1024*1024*1024 && size < 1024*1024*1024*1024*1024 {
		middle = size / 1024 / 1024 / 1024 / 1024
		middleString := fmt.Sprintf("%.1f TB", middle)
		return middleString
	} else if size >= 1024*1024*1024*1024*1024 && size < 1024*1024*1024*1024*1024*1024 {
		middle = size / 1024 / 1024 / 1024 / 1024 / 1024
		middleString := fmt.Sprintf("%.1f PB", middle)
		return middleString
	}
	return sizeString
}

func GetWidthSizeUnit(inputSize int64, sizeUnit string) string {
	//Need or not to keep size in origin number format

	var middle float64
	size := float64(inputSize)
	//fmt.Printf("size=%d", size)

	sizeString := "N/A"
	if sizeUnit == "B" || sizeUnit == "b" {
		middle = size
		middleString := fmt.Sprintf("%.0f", middle) //strconv.FormatFloat(size, 'E', -1, 64)
		return middleString
	} else if sizeUnit == "KB" || sizeUnit == "kb" || sizeUnit == "K" || sizeUnit == "k" {
		middle = size / 1024
		middleString := fmt.Sprintf("%.0f", middle)
		return middleString + " KB"
	} else if sizeUnit == "MB" || sizeUnit == "mb" || sizeUnit == "M" || sizeUnit == "m" {
		middle = size / 1024 / 1024
		middleString := fmt.Sprintf("%.0f", middle)
		return middleString + " MB"
	} else if sizeUnit == "GB" || sizeUnit == "gb" || sizeUnit == "G" || sizeUnit == "g" {
		middle = size / 1024 / 1024 / 1024
		middleString := fmt.Sprintf("%.0f", middle)
		return middleString + " GB"
	} else if sizeUnit == "TB" || sizeUnit == "tb" || sizeUnit == "T" || sizeUnit == "t" {
		middle = size / 1024 / 1024 / 1024 / 1024
		middleString := fmt.Sprintf("%.0f", middle)
		return middleString + " TB"
	} else if sizeUnit == "PB" || sizeUnit == "pb" || sizeUnit == "P" || sizeUnit == "p" {
		middle = size / 1024 / 1024 / 1024 / 1024 / 1024
		middleString := fmt.Sprintf("%.0f", middle)
		return middleString + " PB"
	}
	return sizeString
}

func GetWidthSize(widthSize string) (error, int64, int64, string) {

	//check the widthSize ,the format: number + letter, and number must before letter.
	reg := regexp.MustCompile(`^[1-9][0-9]{0,}\w{1,2}`)
	result := reg.MatchString(widthSize)
	if result == false {
		return errors.New("the format is not correct."), 0, 0, ""
	}
	//get the number.
	reg = regexp.MustCompile(`^[1-9][0-9]{0,}`)
	num := reg.FindString(widthSize)
	//fmt.Println("num:", num)
	number, err := strconv.ParseInt(num, 10, 64)
	if err != nil {
		return errors.New("the format is not correct."), 0, 0, ""
	}

	//get the unit.
	reg = regexp.MustCompile(`[a-zA-Z]{1,2}`)
	unit := reg.FindString(widthSize)
	//fmt.Println("unit:", unit)

	var size int64 = 0
	if unit == "B" || unit == "b" {
		size = number
	}
	if unit == "KB" || unit == "K" || unit == "k" || unit == "kb" {
		size = number * 1024
	} else if unit == "MB" || unit == "M" || unit == "m" || unit == "mb" {
		size = number * 1024 * 1024
	} else if unit == "GB" || unit == "G" || unit == "g" || unit == "gb" {
		size = number * 1024 * 1024 * 1024
	} else if unit == "TB" || unit == "T" || unit == "t" || unit == "tb" {
		size = number * 1024 * 1024 * 1024 * 1024
	} else {
		return errors.New("the format is not correct."), 0, 0, ""
	}

	//fmt.Println("size:", size)
	return nil, size, number, unit
}

func GetSizeFromRange(sizeRange string) (error, int64, int64) {
	var sizeGreaterInt64 int64 = 0
	var sizeLessInt64 int64 = -1 //1073741824
	var err error

	if sizeRange != "" {
		sizeFormatError := false
		if strings.HasSuffix(sizeRange, "]") {
			sizeRange = strings.TrimRight(sizeRange, "]")
		} else {
			sizeFormatError = true
		}

		if strings.HasPrefix(sizeRange, "[") {
			sizeRange = strings.TrimLeft(sizeRange, "[")
		} else {
			sizeFormatError = true
		}

		if sizeFormatError {
			err = errors.New("Size range input invalid, format: \"[0, 1073741824]\"")
			return err, 0, 0
		}

		sizeList := strings.Split(sizeRange, ",")
		if len(sizeList) != 2 {
			fmt.Println("the size format is not correct.format: \"[0, 1073741824]\"")
			return err, 0, 0
		}
		sizeGreater := strings.TrimSpace(sizeList[0])
		sizeLess := strings.TrimSpace(sizeList[1])

		if sizeGreater == "-" && sizeLess == "-" {
			err = errors.New("Please specify the range!")
			return err, 0, 0
		}

		if sizeGreater != "-" && sizeGreater != "" {
			sizeGreaterInt64, err = strconv.ParseInt(sizeGreater, 10, 64)
			if err != nil {
				err = errors.New("Size value is invalid!")
				return err, 0, 0
			}
		}

		if sizeLess != "-" && sizeLess != "" {
			sizeLessInt64, err = strconv.ParseInt(sizeLess, 10, 64)
			if err != nil {
				err = errors.New("Size value is invalid!")
				return err, 0, 0
			}
		}

	}

	//fmt.Printf("sizeGreaterInt64=%s, sizeLessInt64=%s", sizeGreaterInt64, sizeLessInt64)
	return nil, sizeGreaterInt64, sizeLessInt64
}

func GetSizeFromRangeReal(sizeRange string) (error, int64, int64) {
	var sizeGreaterInt64 int64 = 0
	var sizeLessInt64 int64 = 0
	var err error

	if sizeRange != "" {
		sizeFormatError := false
		if strings.HasSuffix(sizeRange, "]") {
			sizeRange = strings.TrimRight(sizeRange, "]")
		} else {
			sizeFormatError = true
		}

		if strings.HasPrefix(sizeRange, "[") {
			sizeRange = strings.TrimLeft(sizeRange, "[")
		} else {
			sizeFormatError = true
		}

		if sizeFormatError {
			err = errors.New("Size range input invalid, format: \"[0, 1073741824]\"")
			return err, 0, 0
		}

		sizeList := strings.Split(sizeRange, ",")

		if len(sizeList) != 2 {
			fmt.Println("the size format is not correct.format: \"[0, 1073741824]\"")
			return err, 0, 0
		}

		sizeGreater := strings.TrimSpace(sizeList[0])
		sizeLess := strings.TrimSpace(sizeList[1])

		if sizeGreater == "-" && sizeLess == "-" {
			err = errors.New("Please specify the range!")
			return err, 0, 0
		}

		if sizeGreater != "-" && sizeGreater != "" {
			sizeGreaterInt64, err = strconv.ParseInt(sizeGreater, 10, 64)
			if err != nil {
				err = errors.New("Size value is invalid!")
				return err, 0, 0
			}
		}

		if sizeLess != "-" && sizeLess != "" {
			sizeLessInt64, err = strconv.ParseInt(sizeLess, 10, 64)
			if err != nil {
				err = errors.New("Size value is invalid!")
				return err, 0, 0
			}
		}

	}

	//fmt.Printf("sizeGreaterInt64=%s, sizeLessInt64=%s", sizeGreaterInt64, sizeLessInt64)
	return nil, sizeGreaterInt64, sizeLessInt64
}

func GetTimeStringFromRange(timeRange string) (error, string, string) {
	var fromTime string = ""
	var toTime string = ""
	var err error
	if timeRange != "" {
		timeFormatError := false
		if strings.HasSuffix(timeRange, "]") {
			timeRange = strings.TrimRight(timeRange, "]")
		} else {
			timeFormatError = true
		}

		if strings.HasPrefix(timeRange, "[") {
			timeRange = strings.TrimLeft(timeRange, "[")
		} else {
			timeFormatError = true
		}

		if timeFormatError {
			err = errors.New("Time range input invalid, format: \"[2017-06-01 12:00:00, 2017-07-01 12:00:00]\"")
			return err, "", ""
		}

		timeList := strings.Split(timeRange, ",")

		if len(timeList) != 2 {
			err = errors.New("Time range input invalid, format: \"[2017-06-01 12:00:00, 2017-07-01 12:00:00]\"")
			return err, "", ""
		}

		fromTime = strings.TrimSpace(timeList[0])
		toTime = strings.TrimSpace(timeList[1])

		if strings.Count(fromTime, "")-1 != 19 && fromTime != "-" {
			err = errors.New("Time range input invalid, format: \"[2017-06-01 12:00:00, 2017-07-01 12:00:00]\"")
			return err, "", ""
		}
		if strings.Count(toTime, "")-1 != 19 && toTime != "-" {
			err = errors.New("Time range input invalid, format: \"[2017-06-01 12:00:00, 2017-07-01 12:00:00]\"")
			return err, "", ""
		}

		fromYear := string([]byte(fromTime)[0:4])
		if fromTime != "-" {
			fromY, err := strconv.Atoi(fromYear)
			if err != nil {
				err = errors.New("Time range input invalid, format: \"[2017-06-01 12:00:00, 2017-07-01 12:00:00]\"")
				return err, "", ""
			}

			if fromY < 1970 || fromY > 2106 {
				err = errors.New("year range should between 1970 and 2106")
				return err, "", ""
			}
		}
		toYear := string([]byte(toTime)[0:4])
		if toTime != "-" {
			toY, err := strconv.Atoi(toYear)
			if err != nil {
				err = errors.New("Time range input invalid, format: \"[2017-06-01 12:00:00, 2017-07-01 12:00:00]\"")
				return err, "", ""
			}

			if toY < 1970 || toY > 2106 {
				err = errors.New("year range should between 1970 and 2106")
				return err, "", ""
			}
		}
		//fmt.Println(string([]byte(fromTime)[0:4]))
		//fmt.Println(string([]byte(toTime)[0:4]))

		if fromTime == "-" && toTime == "-" {
			err = errors.New("Please specify the range!")
			return err, "", ""
		}

		if fromTime == "-" || fromTime == "" {
			fromTime = ""
		}

		if toTime == "-" || toTime == "" {
			toTime = ""
		}
	}

	//fmt.Printf("fromTime=%s, toTime=%s", fromTime, toTime)
	return nil, fromTime, toTime
}

func PrettyPrint(b []byte) ([]byte, error) {
	var out bytes.Buffer
	err := json.Indent(&out, b, "", "  ")
	return out.Bytes(), err
}

func CountChecksum(dataName string, dataSize int64, sizeLimit int64) (error, string, []string) {
	//count check0 to check7 of the record
	if sizeLimit == 0 {
		sizeLimit = 4096
	}

	checkTotal := ""
	checksums := make([]string, 8)
	var pos int64 = 0

	f, err := os.OpenFile(dataName, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0660)
	if err != nil {
		MetaViewLogger.Errorf("OpenFile error: %s\n", err.Error())
		return err, "", nil
	}
	defer f.Close()

	if dataSize < sizeLimit {
		readBuf := make([]byte, dataSize)
		readString, err := f.ReadAt(readBuf, pos)
		if err != nil && err != io.EOF {
			MetaViewLogger.Errorf("ReadAt error: %s\n", err.Error())
			return err, "", nil
		}

		//MetaViewLogger.Infof("pos:%s,readBuf:%s\n", pos, string(readBuf[:readString]))
		ctx := md5.New()
		ctx.Write([]byte(string(readBuf[:readString])))
		checkTotal = hex.EncodeToString(ctx.Sum(nil))

	} else {
		segSize := dataSize / 8
		//MetaViewLogger.Infof("segSize:%d", segSize)

		nums := [8]int64{0, 1, 2, 3, 4, 5, 6, 7}
		checksums = make([]string, 0)
		for _, num := range nums {
			pos = segSize * num
			readBuf := make([]byte, 2<<(uint(num)+2))
			readString, err := f.ReadAt(readBuf, pos)
			if err != nil && err != io.EOF {
				MetaViewLogger.Errorf("ReadAt error: %s\n", err.Error())
				return err, "", nil
			}

			//MetaViewLogger.Infof("pos:%s,readBuf:%s\n", pos, string(readBuf[:readString]))
			ctx := md5.New()
			ctx.Write([]byte(string(readBuf[:readString])))
			check := hex.EncodeToString(ctx.Sum(nil))
			//MetaViewLogger.Infof("check:%s\n", check)
			checksums = append(checksums, check)
		}
	}

	return nil, checkTotal, checksums
}

/*
special deal admin user
*/

func GetCurrentUser(inputUser string) (string, error) {
	u, err := user.Current()
	if err == nil {
		//Only root could specify a filter username
		if u.Username == "root" {
			if inputUser == "" {
				return "", nil
			} else {
				return inputUser, nil
			}
		} else {
			if inputUser != "" {
				var err = errors.New("Only root could specify -u user!")
				return "", err
			} else {
				return u.Username, nil
			}
		}
	}
	return "", err
}

func GetCurrentUid(inputUid string) (string, error) {
	u, err := user.Current()

	if err == nil {
		//Only root could specify a filter username
		if u.Uid == "0" {
			if inputUid == "" {
				return "", nil
			} else {
				return inputUid, nil
			}
		} else {
			if inputUid != "" {
				var err = errors.New("Only root could specify -U uid!")
				return "", err
			} else {
				return u.Uid, err
			}
		}
	}
	return "", err
}

func GetCurrentUidOnly() (string, error) {
	u, err := user.Current()
	if err == nil {
		return u.Uid, err
	}

	return "", err
}

func GetCurrentUsernameOnly() (string, error) {
	u, err := user.Current()
	if err == nil {
		return u.Username, err
	}

	return "", err
}

/*
deal admin user end
*/

func GetGroupName(inputGroup string) (string, error) {
	u, err := user.Current()
	if err == nil {
		//Only root could specify a filter username
		if u.Username == "root" {
			if inputGroup == "" {
				return "", nil
			} else {
				return inputGroup, nil
			}
		} else {
			if inputGroup != "" {
				var err = errors.New("Only root could specify -g groupname!")
				return "", err
			} else {
				return "", err
			}
		}
	}
	return "", err
}

func GetGroupGid(inputGid string) (string, error) {
	u, err := user.Current()
	if err == nil {
		//Only root could specify a filter username
		if u.Uid == "0" {
			if inputGid == "" {
				return "", nil
			} else {
				return inputGid, nil
			}
		} else {
			if inputGid != "" {
				var err = errors.New("Only root could specify -G gid!")
				return "", err
			} else {
				return "", err
			}
		}
	}
	return "", err
}

func IntModeToString(inputMode int) string {
	prefixType := ""
	binString := strconv.FormatInt(int64(inputMode), 2)

	if len(binString) == 13 {
		prefixType = "p"
	} else if len(binString) == 14 {
		prefixType = "c"
	} else if len(binString) == 15 && string(binString[1]) == "0" { //string(binString[0]) == "1" &&
		prefixType = "d"
	} else if len(binString) == 15 && string(binString[1]) == "1" { //string(binString[0]) == "1" &&
		prefixType = "b"
	} else if len(binString) == 16 && string(binString[2]) == "0" { //string(binString[0]) == "1" &&
		prefixType = "-"
	} else if len(binString) == 16 && string(binString[2]) == "1" { //string(binString[0]) == "1" &&
		prefixType = "l"
	}

	basicString := []string{"r", "w", "x", "r", "w", "x", "r", "w", "x"}
	for i := 1; i <= 9; i++ {
		last := len(binString) - 1
		lastPos := string(binString[last])
		binString = binString[:last]

		if lastPos == "0" {
			replaceIndex := 9 - i
			basicString[replaceIndex] = "-"
		}
	}
	return prefixType + strings.Join(basicString, "") //+ "."
}

func ModeToStringFormat(inputMode string) string {
	modeInt, _ := strconv.Atoi(inputMode)

	return IntModeToString(modeInt)
}

func StringModeToInt(inputMode string) (error, int) {
	if inputMode == "" {
		return nil, 0
	}

	if len(inputMode) != 10 {
		//var err = errors.New("inputMode invalid!")
		xt_err := NewXtErrorInfo(INVALID_MODE_ERR, "")
		return xt_err, 0
	}

	modeArray := make([]int, 16, 16)
	modeArray[0] = 1

	if string(inputMode[0]) == "l" {
		modeArray[2] = 1
	} else if string(inputMode[0]) == "d" {
		modeArray[0] = 0
		modeArray[1] = 1
	} else if string(inputMode[0]) == "c" {
		modeArray[0] = 0
		modeArray[2] = 1
	} else if string(inputMode[0]) == "b" {
		modeArray[0] = 0
		modeArray[1] = 1
		modeArray[2] = 1
	} else if string(inputMode[0]) == "p" {
		modeArray[0] = 0
		modeArray[3] = 1
	}

	for i := 1; i <= 9; i++ {
		if string(inputMode[i]) != "-" {
			modeArray[i+6] = 1
		}
	}

	var modeInt int = 0
	for j := 0; j <= 15; j++ {
		if modeArray[15-j] == 1 {
			if j == 0 {
				modeInt += 1
			} else {
				modeInt += 2 << (uint(j - 1))
			}
		}
	}

	modeInt = modeInt & (0xf000 | 0x1ff) //(S_IFMAT | ACCESSPERMS)
	//fmt.Println("modeArray=%s\n,modeInt=%d\n", modeArray, modeInt)
	return nil, modeInt
}

func IntAclToString(inputAcl int) string {
	aclStr := ""
	if inputAcl == 0 {
		aclStr = "rgw_perm_none"
	} else if inputAcl == 1 {
		aclStr = "rgw_perm_read"
	} else if inputAcl == 2 {
		aclStr = "rgw_perm_write"
	} else if inputAcl == 4 {
		aclStr = "rgw_perm_read_acp"
	} else if inputAcl == 8 {
		aclStr = "rgw_perm_wirte_acp"
	} else if inputAcl == 10 {
		aclStr = "rgw_perm_read_objs"
	} else if inputAcl == 20 {
		aclStr = "rgw_perm_write_objs"
	} else if inputAcl == 15 {
		aclStr = "rgw_perm_full_control"
	} else {
		aclStr = strconv.Itoa(inputAcl)
	}
	return aclStr
}

func IntAclToStringTest(inputAcl int) string {

	var aclMap = map[int]string{
		0: "rgw_perm_none", 1: "rgw_perm_read", 4: "rgw_perm_read_acp", 8: "rgw_perm_wirte_acp", 10: "rgw_perm_read_objs", 20: "rgw_perm_write_objs", 15: "rgw_perm_full_control",
	}

	//aclStr = strconv.Itoa(inputAcl)

	return aclMap[inputAcl]
}
func StringAclToInt(aclStr string) (error, int) {

	var err error = nil
	var acl int
	if aclStr == "rgw_perm_none" {
		acl = 0
	} else if aclStr == "rgw_perm_read" {
		acl = 1
	} else if aclStr == "rgw_perm_write" {
		acl = 2
	} else if aclStr == "rgw_perm_read_acp" {
		acl = 4
	} else if aclStr == "rgw_perm_wirte_acp" {
		acl = 8
	} else if aclStr == "rgw_perm_read_objs" {
		acl = 10
	} else if aclStr == "rgw_perm_write_objs" {
		acl = 20
	} else if aclStr == "rgw_perm_full_control" {
		acl = 15
	} else {
		errors.New("no exist the acl,please check format.")
	}

	//fmt.Println("modeArray=%s\n,modeInt=%d\n", modeArray, modeInt)
	return err, acl
}

func StringAclToIntTest(aclStr string) (error, int) {

	var aclMap = map[string]int{
		"rgw_perm_none": 0, "rgw_perm_read": 1, "rgw_perm_read_acp": 4, "rgw_perm_wirte_acp": 8, "rgw_perm_read_objs": 10, "rgw_perm_write_objs": 20, "rgw_perm_full_control": 15,
	}

	var err error = nil

	//errors.New("no exist the acl,please check format.")

	return err, aclMap[aclStr]
}

func GetCountFromRange(countRange string) (error, int64, int64) {
	var countGreaterInt64 int64 = 2
	var countLessInt64 int64 = 1000000
	var err error

	if countRange != "" {
		countFormatError := false
		if strings.HasSuffix(countRange, "]") {
			countRange = strings.TrimRight(countRange, "]")
		} else {
			countFormatError = true
		}

		if strings.HasPrefix(countRange, "[") {
			countRange = strings.TrimLeft(countRange, "[")
		} else {
			countFormatError = true
		}

		if countFormatError {
			err = errors.New("count range input invalid, format: \"[2, 10000]\"")
			return err, 0, 0
		}

		countList := strings.Split(countRange, ",")
		countGreater := strings.TrimSpace(countList[0])
		countLess := strings.TrimSpace(countList[1])

		if countGreater == "-" && countLess == "-" {
			err = errors.New("Please specify the range!")
			return err, 0, 0
		}
		if countGreater != "-" && countGreater != "" {
			countGreaterInt64, err = strconv.ParseInt(countGreater, 10, 64)
			if countGreaterInt64 < 2 {
				err = errors.New("count value must >= 2!")
				return err, 0, 0
			}
			if err != nil {
				err = errors.New("count value is invalid!")
				return err, 0, 0
			}
		}
		if countLess != "-" && countLess != "" {
			countLessInt64, err = strconv.ParseInt(countLess, 10, 64)
			if countLessInt64 < 2 {
				err = errors.New("count value must >= 2!")
				return err, 0, 0
			}
			if err != nil {
				err = errors.New("count value is invalid!")
				return err, 0, 0
			}
		}
	}

	return nil, countGreaterInt64, countLessInt64
}

type MetaviewEnvSetting struct {
	Logfile       string
	Configfile    string
	Queue         string
	EtcdEndPoints []string
	AccountMethod string
}

func EnvUtilsParseEnvSetting() *MetaviewEnvSetting {
	envSetting := &MetaviewEnvSetting{
		Logfile:    "/var/log/metaview/metaview.log",
		Configfile: "metaview.json",
	}
	logFile := os.Getenv("LOGFILE")
	if logFile != "" {
		envSetting.Logfile = logFile
	}
	configFile := os.Getenv("CONFIGFILE")
	if configFile != "" {
		envSetting.Configfile = configFile
	}
	etcdHosts := os.Getenv("ETCD_CLUSTER")
	if etcdHosts == "" {
		etcdHosts = "http://localhost:2379"
	}
	queue := os.Getenv("QUEUE")
	envSetting.Queue = queue

	//if queue == "" {
	//	envSetting.Queue = 0
	//} else {
	//	queueNum, err := strconv.Atoi(queue)
	//	if err != nil {
	//		Logger.Errorf("Queue env %s set wrong: %s\n",
	//			queue, err.Error())
	//		envSetting.Queue = 0
	//	} else {
	//		envSetting.Queue = queueNum
	//	}
	//}

	envSetting.AccountMethod = "ldap"
	accoutMethod := os.Getenv("METAVIEW_ACCOUNT_METHOD")
	//fmt.Println("accountMethod:", accoutMethod)
	if accoutMethod == "etcd" {
		envSetting.AccountMethod = "etcd"
	}

	hostItems := strings.Split(etcdHosts, ",")
	endPoints := make([]string, 0, len(hostItems))
	for i := 0; i < len(hostItems); i++ {
		if hostItems[i] != "" {
			endPoints = append(endPoints, hostItems[i])
		}
	}

	envSetting.EtcdEndPoints = endPoints
	return envSetting
}

const (
	TIOCGWINSZ     = 0x5413
	TIOCGWINSZ_OSX = 1074295912
)

type window struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}

func TerminalWidth() (int, error) {
	w := new(window)
	tio := syscall.TIOCGWINSZ
	if runtime.GOOS == "darwin" {
		tio = TIOCGWINSZ_OSX
	}
	res, _, err := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(syscall.Stdin),
		uintptr(tio),
		uintptr(unsafe.Pointer(w)),
	)
	if int(res) == -1 {
		return 0, err
	}
	return int(w.Col), nil
}

func GetWidth(width string) (error, int64, int64, string) {

	//check the widthSize ,the format: number + letter, and number must before letter.
	reg := regexp.MustCompile(`^[1-9][0-9]{0,}\w{1,2}`)
	result := reg.MatchString(width)
	if result == false {
		return errors.New("the format is not correct."), 0, 0, ""
	}
	//get the number.
	reg = regexp.MustCompile(`^[1-9][0-9]{0,}`)
	num := reg.FindString(width)
	//fmt.Println("num:", num)

	number, err := strconv.ParseInt(num, 10, 64)
	if err != nil {
		return errors.New("the format is not correct."), 0, 0, ""
	}

	//get the unit.
	reg = regexp.MustCompile(`[a-zA-Z]{1,}`)
	unit := reg.FindString(width)
	//fmt.Println("unit:", unit)

	//fmt.Println("length_unit:", strings.Count(unit, ""))
	//fmt.Println("unit:", unit)
	if strings.Count(unit, "")-1 <= 2 {
		var size int64 = 0
		if unit == "B" || unit == "b" {
			size = number
		}
		if unit == "KB" || unit == "K" || unit == "k" || unit == "kb" {
			size = number * 1024
		} else if unit == "MB" || unit == "M" || unit == "m" || unit == "mb" {
			size = number * 1024 * 1024
		} else if unit == "GB" || unit == "G" || unit == "g" || unit == "gb" {
			size = number * 1024 * 1024 * 1024
		} else if unit == "TB" || unit == "T" || unit == "t" || unit == "tb" {
			size = number * 1024 * 1024 * 1024 * 1024
		} else {
			return errors.New("the format is not correct."), 0, 0, ""
		}

		//fmt.Println("size:", size)
		return nil, size, number, unit
	} else {
		var time int64 = 0
		if unit == "second" {
			time = number
		} else if unit == "min" {
			time = number * 60
		} else if unit == "hour" {
			time = number * 60 * 60
		} else if unit == "day" {
			time = number * 60 * 60 * 24
		} else if unit == "week" {
			time = number * 60 * 60 * 24 * 7
		} else if unit == "month" {
			time = number * 60 * 60 * 24 * 30 //wait update
		} else if unit == "year" {
			time = number * 60 * 60 * 24 * 365 //wait update
		} else {
			return errors.New("the format is not correct."), 0, 0, ""
		}

		return nil, time, number, unit
	}
}

func GetStride(stride string) (error, int64, int64, string, bool, bool) {
	if stride == "latest" {
		return nil, 0, 0, "", true, true
	}

	//check the widthSize ,the format: number + letter, and number must before letter.
	reg := regexp.MustCompile(`^[1-9][0-9]{0,}\w{1,2}`)
	result := reg.MatchString(stride)
	if result == false {
		return errors.New("the format is not correct."), 0, 0, "", false, false
	}
	//get the number.
	reg = regexp.MustCompile(`^[1-9][0-9]{0,}`)
	num := reg.FindString(stride)
	//fmt.Println("num:", num)

	number, err := strconv.ParseInt(num, 10, 64)
	if err != nil {
		return errors.New("the format is not correct."), 0, 0, "", false, false
	}

	//get the unit.
	reg = regexp.MustCompile(`[a-zA-Z]{1,}`)
	unit := reg.FindString(stride)
	//fmt.Println("unit:", unit)

	//fmt.Println("length_unit:", strings.Count(unit, ""))
	//fmt.Println("unit:", unit)
	if strings.Count(unit, "")-1 <= 2 {
		var size int64 = 0
		if unit == "B" || unit == "b" {
			size = number
		}
		if unit == "KB" || unit == "K" || unit == "k" || unit == "kb" {
			size = number * 1024
		} else if unit == "MB" || unit == "M" || unit == "m" || unit == "mb" {
			size = number * 1024 * 1024
		} else if unit == "GB" || unit == "G" || unit == "g" || unit == "gb" {
			size = number * 1024 * 1024 * 1024
		} else if unit == "TB" || unit == "T" || unit == "t" || unit == "tb" {
			size = number * 1024 * 1024 * 1024 * 1024
		} else {
			return errors.New("the format is not correct."), 0, 0, "", false, false
		}

		//fmt.Println("size:", size)
		return nil, size, number, unit, false, false
	} else {
		var time int64 = 0
		if unit == "second" {
			time = number
		} else if unit == "min" {
			time = number * 60
		} else if unit == "hour" {
			time = number * 60 * 60
		} else if unit == "day" {
			time = number * 60 * 60 * 24
		} else if unit == "week" {
			time = number * 60 * 60 * 24 * 7
		} else if unit == "month" {
			time = number * 60 * 60 * 24 * 30 //wait update
		} else if unit == "year" {
			time = number * 60 * 60 * 24 * 365 //wait update
		} else {
			return errors.New("the format is not correct."), 0, 0, "", false, false
		}

		return nil, time, number, unit, true, false
	}

}

func GetStampTimeStringInt64(timeStamp int64) string {

	stampTimeString := time.Unix(timeStamp, 0).Format("2006-01-02 15:04:05")

	return stampTimeString
}

func GetClientUserInfo() *UserAccountInfo {
	curUser, err := user.Current()
	if err != nil {
		fmt.Printf("Fail to get current user: %s\n",
			err.Error())
		return &UserAccountInfo{}
	}

	info := &UserAccountInfo{
		Username: curUser.Username,
		Uid:      curUser.Uid,
		Gid:      curUser.Gid,
		Umask:    fmt.Sprintf("%d", syscall.Umask(0)),
	}

	return info
}

func StringModeToInt2(inputMode string, modeType string) (error, int) {

	if modeType != "dir" && modeType != "file" {
		return nil, 0
	}
	if inputMode == "" {
		return nil, 0
	}

	if modeType == "dir" {
		inputMode = "d" + inputMode
	}

	if modeType == "file" {
		inputMode = "-" + inputMode
	}

	if len(inputMode) != 10 {
		var err = errors.New("inputMode invalid!")
		return err, 0
	}

	modeArray := make([]int, 16, 16)
	modeArray[0] = 1

	if string(inputMode[0]) == "l" {
		modeArray[2] = 1
	} else if string(inputMode[0]) == "d" {
		modeArray[0] = 0
		modeArray[1] = 1
	}

	for i := 1; i <= 9; i++ {
		if string(inputMode[i]) != "-" {
			modeArray[i+6] = 1
		}
	}

	var modeInt int = 0
	for j := 0; j <= 15; j++ {
		if modeArray[15-j] == 1 {
			if j == 0 {
				modeInt += 1
			} else {
				modeInt += 2 << (uint(j - 1))
			}
		}
	}

	//fmt.Println("modeInt:", modeInt)
	//fmt.Println("modeArray=%s\n,modeInt=%d\n", modeArray, modeInt)
	return nil, modeInt
}

func GetStampTimeString(timeStamp int) string {
	//1495358104 int to "2017-05-21 17:15:04"
	int64TimeStamp := int64(timeStamp)
	stampTimeString := time.Unix(int64TimeStamp, 0).Format("2006-01-02 15:04:05")

	//DBLogger.Infof("stampTimeString:", stampTimeString)
	return stampTimeString
}

func GetStampTimeString2(timeStr string) string {
	//1495358104 int to "2017-05-21 17:15:04"
	int64TimeStamp, err := strconv.ParseInt(timeStr, 10, 64)
	if err != nil {
		return timeStr
	}
	stampTimeString := time.Unix(int64TimeStamp, 0).Format("2006-01-02 15:04:05")

	return stampTimeString
}

func ExecShell(s string) (error, string) {
	cmd := exec.Command("/bin/bash", "-c", s)
	var out bytes.Buffer

	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		Logger.Infof("exec_shell:", err.Error())
		return err, ""
	}
	//fmt.Printf("%s", out.String())
	return err, strings.Trim(out.String(), "\n")
}

/*
 * 将RestfulAPI中`dirPath`和`fsName`拼接成合法的全路径
 * fsName和dirPath都为空时返回"/"W
 * <注>依赖golang/path实现，在除linux以外的平台会出问题
 */
func SpliceToFullPath(fsName, dirPath string) string {
	return path.Join("/", fsName, dirPath)
}

/*
 * Golang 自实现的三元运算符, https://studygolang.com/articles/3248
 * 例: name := IF(age > 18, "man", "child").(string)
 */
func IF(condition bool, trueVal, falseVal interface{}) interface{} {
	if condition {
		return trueVal
	}
	return falseVal
}

/*
 * 阻塞监听控制台的输入事件，并处理翻页事件。
 * 如果用户按下'n'、'N'、'Page down'、'Enter'、'↓'、'→'、那么会将`Cursor + 1`并返回
 * 如果用户按下'p'、'P'、'Page up'、'↑'、'←'那么会将`Cursor - 1`返回
 * 如果用户按下'q'、'ESC'、'Ctrl+C'，那么会退出监听，并且返回-1
 */
func PageFlipEvent(Cursor int, nextPage bool) (int, error) {
	t, err := term.Open("/dev/tty")
	if err != nil {
		return -1, err
	}
	defer func() {
		fmt.Printf("\r\n")
		t.Restore()
		t.Close()
	}()
	term.RawMode(t)

	if nextPage && Cursor == 0 { //First Page
		fmt.Printf("\033[7m Press N/n to Next, q/Q to quit \033[0m")
	} else if nextPage && Cursor > 0 { //Middle Pages
		fmt.Printf("\033[7m Press P/p to Previous, N/n to Next, q/Q to quit \033[0m")
	} else if Cursor > 0 { //Last Page
		fmt.Printf("\033[7m Press P/p to Previous, q/Q to quit \033[0m")
	} else {
		return -1, nil
	}

	for {
		datas := make([]byte, 4)
		_, err = t.Read(datas)
		if err != nil {
			return -1, err
		}

		// 翻转字节
		datas[0], datas[3] = datas[3], datas[0]
		datas[1], datas[2] = datas[2], datas[1]

		switch binary.BigEndian.Uint32(datas) {
		case 113, 3, 27: // 'q'、'Ctrl+C'、'ESC'
			return -1, nil
		case 13, 78, 110, 4348699, 4414235, 2117491483:
			// 'Enter'、'N'、'n'、'↓'、'→'、'page down'
			if nextPage {
				Cursor += 1
				return Cursor, nil
			}
		case 80, 112, 4283163, 4479771, 2117425947:
			// 'P'、'p'、'↑'、'←'、'page up'
			Cursor -= 1
			if Cursor < 0 {
				Cursor = 0
				continue
			}
			return Cursor, nil
		}
	}
}

//str trans to uuid
func Str2UUID(str string) string {
	ctx := md5.New()
	ctx.Write([]byte(str))
	return hex.EncodeToString(ctx.Sum(nil))
}

/*
func DealType2(ids []int) string {
	var sb strings.Builder
	for _, value := range ids {
		sb.WriteString(strconv.Itoa(value))
		sb.WriteString(",")
	}

	resIds := "[" + strings.TrimSuffix(sb.String(), ",") + "]"
	return resIds
}
*/

func DealType(values []int) string {
	var st bytes.Buffer
	for _, value := range values {
		st.WriteString(strconv.Itoa(value))
		st.WriteString(",")
	}

	resIds := "[" + strings.TrimSuffix(st.String(), ",") + "]"
	return resIds
}

func DealTypeStr(values []string) string {
	var st bytes.Buffer
	for _, value := range values {
		st.WriteString("\"")
		st.WriteString(value)
		st.WriteString("\"")
		st.WriteString(",")
	}

	resIds := "[" + strings.TrimSuffix(st.String(), ",") + "]"
	return resIds
}

func DealTypeStr2(values []string) string {
	var st bytes.Buffer
	for _, value := range values {
		st.WriteString(value)
		st.WriteString(",")
	}

	resIds := "[" + strings.TrimSuffix(st.String(), ",") + "]"
	return resIds
}

func JuddgeTime(timeString string) error {
	if timeString != "" {
		loc, _ := time.LoadLocation("Local")
		_, err := time.ParseInLocation("2006-01-02 15:04:05", timeString, loc)
		if err != nil {
			return err
		}
	}

	return nil
}

/*
func JuddgeTimeStamp2(timeString string, timeEnd bool) (error, int) {
	// "2017-05-21 17:15:04" to 1495358104 int
	if timeString == "" {
		if timeEnd {
			return nil, 4294967295 //2147483647 max value of int32, 4294967295 max of oid(uint32)
		} else {
			return nil, 0
		}
	}

	loc, _ := time.LoadLocation("Local")
	theTime, err := time.ParseInLocation("2006-01-02 15:04:05", timeString, loc)
	if err != nil {
		return err, 0
	}
	timeStamp := theTime.Unix()
	return nil, int(timeStamp)
}
*/
func JuddgeTimeStamp(timeString string) (error, string) {
	/* "2017-05-21 17:15:04" to "1495358104" string*/
	if timeString == "" {
		return nil, ""
	}

	loc, _ := time.LoadLocation("Local")
	theTime, err := time.ParseInLocation("2006-01-02 15:04:05", timeString, loc)
	if err != nil {
		return err, ""
	}
	timeStamp := theTime.Unix()
	return nil, strconv.FormatInt(timeStamp, 10)
}

func ToUpperFieldConv(str string) string {
	return strings.ToUpper(str)
}

/** * 字符串首字母转化为大写 ios_bbbbbbbb -> iosBbbbbbbbb */
func StrFirstToUpper(str string) string {
	temp := strings.Split(str, "_")
	var upperStr string
	for y := 0; y < len(temp); y++ {
		vv := []rune(temp[y])
		if y != 0 {
			for i := 0; i < len(vv); i++ {
				if i == 0 {
					vv[i] -= 32
					upperStr += string(vv[i]) // + string(vv[i+1])
				} else {
					upperStr += string(vv[i])
				}
			}
		}
	}
	return temp[0] + upperStr
}

func DeleteRepeat(list []string) []string {
	mapdata := make(map[string]interface{})
	if len(list) <= 0 {
		return nil
	}
	for _, v := range list {
		mapdata[v] = "true"
	}
	var datas []string
	for k, _ := range mapdata {
		if k == "" {
			continue
		}
		datas = append(datas, k)
	}
	return datas
}

func RemoveSliceMap(a []interface{}) (ret []interface{}) {
	n := len(a)
	for i := 0; i < n; i++ {
		state := false
		for j := i + 1; j < n; j++ {
			if j > 0 && reflect.DeepEqual(a[i], a[j]) {
				state = true
				break
			}
		}
		if !state {
			ret = append(ret, a[i])
		}
	}
	return
}
