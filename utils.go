package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func MakeLog(tag string, message string, args []interface{}) string {
	return fmt.Sprintf("[%s/%s %s %s] %7s : %s", NODE_NAME, LOCAL_IPADDR, GetTodayString(), GetTimeString(), tag, fmt.Sprintf(message, args...))
}

func LogDebug(message string, args ...interface{}) {
	if IS_LOG_ENABLE && LOG_LEVEL > 0 {
		logString := MakeLog("Debug", message, args)
		fmt.Println(logString)
		LogFile(logString)
	}
}

func LogVerbose(message string, args ...interface{}) {
	if IS_LOG_ENABLE && LOG_LEVEL > 1 {
		logString := MakeLog("Verbose", message, args)
		fmt.Println(logString)
		LogFile(logString)
	}
}

func LogInfo(message string, args ...interface{}) {
	if IS_LOG_ENABLE {
		logString := MakeLog("Info", message, args)
		fmt.Println(logString)
		LogFile(logString)
	}
}

func LogFatal(message string, args ...interface{}) {
	if IS_LOG_ENABLE {
		logString := MakeLog("Fatal", message, args)
		fmt.Println(logString)
		LogFile(logString)
	}
}

func LogFile(data string) {
	f, err := os.OpenFile(LOG_PATH, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err == nil {
		f.WriteString(data)
		f.WriteString("\n")
		f.Close()
	}
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, Origin, Accept-Ranges, Range, bytes")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, DELETE, POST")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func Post(url string, headers map[string]string, dataType string, data interface{}) (int, map[string]interface{}, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := http.Client{
		Timeout:   time.Duration(2) * time.Second,
		Transport: tr,
	}

	var reader io.Reader

	if strings.ToUpper(dataType) == "JSON" {
		jsonString, err := json.Marshal(data)
		if err != nil {
			return 0, nil, err
		}

		reader = bytes.NewBuffer(jsonString)
	} else {
		params := structToMap(data)
		reader = strings.NewReader(params.Encode())
	}

	req, _ := http.NewRequest("POST", url, reader)

	if headers != nil {
		for k, v := range headers {
			req.Header.Add(k, v)
		}
	}

	if strings.ToUpper(dataType) == "JSON" {
		req.Header.Add("Content-Type", "application/json")
	} else {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, err
	}

	var jsonData map[string]interface{}
	if err := json.Unmarshal(body, &jsonData); err != nil {
		return 0, nil, err
	}

	return resp.StatusCode, jsonData, nil
}

func GetTodayString() string {
	currentTime := time.Now()
	return currentTime.Format("2006-01-02")
}

func GetTimeString() string {
	currentTime := time.Now()
	return currentTime.Format("15:04:05")
}

func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP
}

func GetIpWithDomainName(name string) net.IP {
	ips, _ := net.LookupIP(name)
	for _, ip := range ips {
		if ipv4 := ip.To4(); ipv4 != nil {
			return ipv4
		}
	}
	return nil
}

func ByteCountSI(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB",
		float64(b)/float64(div), "kMGTPE"[exp])
}

func ByteCountIEC(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB",
		float64(b)/float64(div), "KMGTPE"[exp])
}

func structToMap(i interface{}) (values url.Values) {
	values = url.Values{}
	iVal := reflect.ValueOf(i).Elem()
	typ := iVal.Type()
	for i := 0; i < iVal.NumField(); i++ {
		f := iVal.Field(i)
		var v string
		switch f.Interface().(type) {
		case int, int8, int16, int32, int64:
			v = strconv.FormatInt(f.Int(), 10)
		case uint, uint8, uint16, uint32, uint64:
			v = strconv.FormatUint(f.Uint(), 10)
		case float32:
			v = strconv.FormatFloat(f.Float(), 'f', 4, 32)
		case float64:
			v = strconv.FormatFloat(f.Float(), 'f', 4, 64)
		case []byte:
			v = string(f.Bytes())
		case string:
			v = f.String()
		}
		values.Set(typ.Field(i).Name, v)
	}
	return
}
