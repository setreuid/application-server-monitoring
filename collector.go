package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type Collector struct {
	items []*NodeData
	table map[string]*NodeData
}

type NodeData struct {
	Name           string           `json:"name"`
	IpAddr         string           `json:"ip_addr"`
	HeartbeatItems []*HeartbeatItem `json:"heartbeat_items"`
	PingItems      []*PingItem      `json:"ping_items"`
	HddItems       []*HDDItem       `json:"hdd_items"`
}

func (self *Collector) Init() {
	self.table = make(map[string]*NodeData)
}

func (self *Collector) Ping(c *gin.Context) {
	var data NodeData
	if err := c.BindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "No arguments."})
	} else {
		c.JSON(http.StatusOK, data)
		self.ProcessPing(data)
	}
}

func (self *Collector) Hdd(c *gin.Context) {
	var data NodeData
	if err := c.BindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "No arguments."})
	} else {
		c.JSON(http.StatusOK, data)
		self.ProcessHdd(data)
	}
}

func (self *Collector) ProcessHeartbeat(data NodeData) {
	item := self.table[data.Name]
	if item == nil {
		self.table[data.Name] = &data
		self.items = append(self.items, self.table[data.Name])
	} else {
		self.table[data.Name].HeartbeatItems = data.HeartbeatItems
	}
	wh.CheckHeartBeat(&data)
}

func (self *Collector) ProcessPing(data NodeData) {
	item := self.table[data.Name]
	if item == nil {
		self.table[data.Name] = &data
		self.items = append(self.items, self.table[data.Name])
	} else {
		self.table[data.Name].PingItems = data.PingItems
	}
	wh.CheckPing(&data)
}

func (self *Collector) ProcessHdd(data NodeData) {
	item := self.table[data.Name]
	if item == nil {
		self.table[data.Name] = &data
		self.items = append(self.items, self.table[data.Name])
	} else {
		self.table[data.Name].HddItems = data.HddItems
	}
	wh.CheckHdd(&data)
}

func (self *Collector) Status(c *gin.Context) {
	c.JSON(http.StatusOK, self.items)
}
