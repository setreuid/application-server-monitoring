package main

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"time"
)

type Heartbeat struct {
	items []*HeartbeatItem
	table map[string]*HeartbeatItem
}

type HeartbeatItem struct {
	Name          string `json:"name"`
	ID            string `json:"id"`
	LastCheckTime int64  `json:"last_check_time_seconds"`
	LastCheck     string `json:"last_check_time"`
	IsOnline      bool   `json:"is_online"`
}

type HeartbeatJsonItem struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

func (self *Heartbeat) Run() {
	// Init
	b, err := ioutil.ReadFile(HB_APPS_CONFIG_JSON)
	if err != nil {
		LogFatal("Heartbeat initialing failed.")
		LogFatal("Can not read %s file unmarshal (%v).", HB_APPS_CONFIG_JSON, err)
		return
	}

	var data []HeartbeatJsonItem
	if err := json.Unmarshal(b, &data); err != nil {
		LogFatal("Heartbeat initialing failed.")
		LogFatal("Error on %s file unmarshal (%v).", HB_APPS_CONFIG_JSON, err)
		return
	}

	self.table = make(map[string]*HeartbeatItem)
	for _, v := range data {
		hbi := &HeartbeatItem{
			Name: v.Name,
			ID:   v.ID,
		}

		self.items = append(self.items, hbi)
		self.table[v.ID] = hbi
	}

	LogInfo("Heartbeat checker has loaded %d apps.", len(self.items))
	<-time.After(time.Second * time.Duration(HB_INTERVAL_SECONDS))

	for RUNNING {
		for _, v := range self.items {
			now := time.Now().Unix()

			if now-v.LastCheckTime > HB_TIMEOUT_SECONDS {
				LogDebug("Heartbeat check fail on %s (%s), %s", v.Name, v.ID, v.LastCheck)
				v.IsOnline = false
			} else {
				LogDebug("Heartbeat check success on %s (%s), %s", v.Name, v.ID, v.LastCheck)
				v.IsOnline = true
			}
		}

		go ctr.ProcessHeartbeat(NodeData{
			Name:           NODE_NAME,
			IpAddr:         LOCAL_IPADDR,
			HeartbeatItems: self.items,
		})

		LogDebug("Wait next heartbeat check %ds", HB_INTERVAL_SECONDS)
		<-time.After(time.Second * time.Duration(HB_INTERVAL_SECONDS))
	}
}

func (self *Heartbeat) Ping(c *gin.Context) {
	id := c.Param("id")
	if id != "" {
		item := self.table[id]
		if item == nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "No matching watching app."})
		} else {
			now := time.Now()
			item.LastCheckTime = now.Unix()
			item.LastCheck = time.Unix(0, now.UnixNano()).String()

			c.JSON(http.StatusOK, item)
			LogDebug("Heartbeat ping on %s (%s), %s", item.Name, item.ID, item.LastCheck)
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"message": "No id."})
	}
}
