package main

import (
	"gopkg.in/ini.v1"
)

var (
	api  *Api
	hb   *Heartbeat
	ping *Ping
	hdd  *HDD
	ctr  *Collector
	wh   *WebHook

	RUNNING      bool
	LOCAL_IPADDR string

	IS_MASTER bool
	NODE_NAME string
	PORT      int

	MASTER_NODE_API_HOST string

	IS_LOG_ENABLE bool
	LOG_LEVEL     int
	LOG_PATH      string

	IS_SSL_ENABLE bool
	SSL_CERT_FILE string
	SSL_KEY_FILE  string

	IS_HB_ENABLE        bool
	HB_INTERVAL_SECONDS int
	HB_TIMEOUT_SECONDS  int64
	HB_APPS_CONFIG_JSON string

	IS_PING_ENABLE        bool
	PING_INTERVAL_SECONDS int
	PING_TIMEOUT_SECONDS  int64
	PING_TARGETS          []string
	PING_TARGET_NAMES     []string

	IS_HDD_ENABLE        bool
	HDD_INTERVAL_SECONDS int
	HDD_LIMIT_PERCENT    int
	HDD_TARGETS          []string
	HDD_TARGET_NAMES     []string

	IS_WEB_HOOK_ENABLE      bool
	IS_WEB_HOOK_HB_ENABLE   bool
	IS_WEB_HOOK_PING_ENABLE bool
	IS_WEB_HOOK_HDD_ENABLE  bool
	WEB_HOOK_CONFIG_JSON    string
)

func main() {
	LOCAL_IPADDR = GetOutboundIP().String()

	if cfg, err := ini.Load("Config.ini"); err == nil {
		IS_MASTER = cfg.Section("").Key("IS_MASTER").MustBool(true)
		NODE_NAME = cfg.Section("").Key("NODE_NAME").MustString("NODE")
		PORT = cfg.Section("").Key("PORT").MustInt(20550)
		MASTER_NODE_API_HOST = cfg.Section("").Key("MASTER_NODE_API_HOST").MustString("")

		if !IS_MASTER && MASTER_NODE_API_HOST == "" {
			LogFatal("Need master node api host when running slave mode.")
			return
		}

		IS_LOG_ENABLE = cfg.Section("LOG").Key("IS_ENABLE").MustBool(false)
		LOG_LEVEL = cfg.Section("LOG").Key("LEVEL").MustInt(0)
		LOG_PATH = cfg.Section("LOG").Key("BASIC_PATH").MustString("asm.log")

		IS_SSL_ENABLE = cfg.Section("SSL").Key("IS_ENABLE").MustBool(false)
		SSL_CERT_FILE = cfg.Section("SSL").Key("CERT_PATH").MustString("")
		SSL_KEY_FILE = cfg.Section("SSL").Key("KEY_PATH").MustString("")

		if !IS_MASTER && IS_SSL_ENABLE {
			IS_SSL_ENABLE = false
			LogInfo("Disable ssl when running slave mode.")
		}

		IS_HB_ENABLE = cfg.Section("HEARTBEAT").Key("IS_ENABLE").MustBool(false)
		HB_INTERVAL_SECONDS = cfg.Section("HEARTBEAT").Key("INTERVAL_SECONDS").MustInt(60)
		HB_TIMEOUT_SECONDS = int64(cfg.Section("HEARTBEAT").Key("TIMEOUT_SECONDS").MustInt(120))
		HB_APPS_CONFIG_JSON = cfg.Section("HEARTBEAT").Key("APPS_CONFIG_JSON").MustString("")

		if !IS_MASTER && IS_HB_ENABLE {
			IS_SSL_ENABLE = false
			LogInfo("Disable heart beat checker when running slave mode.")
		}

		IS_PING_ENABLE = cfg.Section("PING").Key("IS_ENABLE").MustBool(false)
		PING_INTERVAL_SECONDS = cfg.Section("PING").Key("INTERVAL_SECONDS").MustInt(60)
		PING_TIMEOUT_SECONDS = int64(cfg.Section("PING").Key("TIMEOUT_SECONDS").MustInt(5))
		PING_TARGETS = cfg.Section("PING").Key("TARGETS").Strings(",")
		PING_TARGET_NAMES = cfg.Section("PING").Key("TARGET_NAMES").Strings(",")

		if len(PING_TARGETS) != len(PING_TARGET_NAMES) {
			LogFatal("Not match ping target, target name items count.")
			return
		}

		IS_HDD_ENABLE = cfg.Section("HDD").Key("IS_ENABLE").MustBool(false)
		HDD_INTERVAL_SECONDS = cfg.Section("HDD").Key("INTERVAL_SECONDS").MustInt(300)
		HDD_LIMIT_PERCENT = cfg.Section("HDD").Key("LIMIT_PERCENT").MustInt(85)
		HDD_TARGETS = cfg.Section("HDD").Key("TARGETS").Strings(",")
		HDD_TARGET_NAMES = cfg.Section("HDD").Key("TARGET_NAMES").Strings(",")

		if len(HDD_TARGETS) != len(HDD_TARGET_NAMES) {
			LogFatal("Not match hdd target, target name items count.")
			return
		}

		IS_WEB_HOOK_ENABLE = cfg.Section("WEB_HOOK").Key("IS_ENABLE").MustBool(false)
		IS_WEB_HOOK_HB_ENABLE = cfg.Section("WEB_HOOK").Key("IS_ENABLE_HEARTBEAT").MustBool(false)
		IS_WEB_HOOK_PING_ENABLE = cfg.Section("WEB_HOOK").Key("IS_ENABLE_PING").MustBool(false)
		IS_WEB_HOOK_HDD_ENABLE = cfg.Section("WEB_HOOK").Key("IS_ENABLE_HDD").MustBool(false)
		WEB_HOOK_CONFIG_JSON = cfg.Section("WEB_HOOK").Key("CONFIG_JSON").MustString("")

		if !IS_MASTER && IS_WEB_HOOK_ENABLE {
			IS_WEB_HOOK_ENABLE = false
			LogInfo("Disable web hook when running slave mode.")
		}
	} else {
		LogFatal("Not found config file '%s'", "Config.ini")
		return
	}

	RUNNING = true

	if IS_MASTER {
		ctr = &Collector{}
		ctr.Init()
	}

	if IS_WEB_HOOK_ENABLE {
		wh = &WebHook{}
		wh.Init()
	}

	if IS_HB_ENABLE {
		hb = &Heartbeat{}
		go hb.Run()
	}

	if IS_PING_ENABLE {
		ping = &Ping{}
		go ping.Run()
	}

	if IS_HDD_ENABLE {
		hdd = &HDD{}
		go hdd.Run()
	}

	api.Run()
}
