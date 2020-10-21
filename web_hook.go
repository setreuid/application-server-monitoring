package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
)

type WebHook struct {
	isEnabled bool
	items     []WebHookItem
}

type WebHookItem struct {
	EndPoint string                 `json:"end_point"`
	Headers  map[string]string      `json:"headers"`
	DataType string                 `json:"data_type"`
	Data     map[string]interface{} `json:"data"`
}

func (self *WebHook) Init() {
	// Init
	b, err := ioutil.ReadFile(WEB_HOOK_CONFIG_JSON)
	if err != nil {
		LogFatal("Web hook initialing failed.")
		LogFatal("Can not read %s file unmarshal (%v).", WEB_HOOK_CONFIG_JSON, err)
		return
	}

	var data []WebHookItem
	if err := json.Unmarshal(b, &data); err != nil {
		LogFatal("Web hook initialing failed.")
		LogFatal("Error on %s file unmarshal (%v).", WEB_HOOK_CONFIG_JSON, err)
		return
	}

	self.items = data
	self.isEnabled = true
}

func (self *WebHook) SendHeartBeatWarning(nodeName string, nodeIpAddr string, item *HeartbeatItem) {
	self.Send(fmt.Sprintf("[%s/%s] %s (%s) Heartbeat 다운 경고\n\n- 마지막 동작 시간\n%s",
		nodeName,
		nodeIpAddr,
		item.Name,
		item.ID,
		item.LastCheck,
	))
}

func (self *WebHook) SendPingWarning(nodeName string, nodeIpAddr string, item *PingItem) {
	self.Send(fmt.Sprintf("[%s/%s] %s (%s) Ping 다운 경고\n\n- 마지막 확인 시간\n%s\n- 마지막 RTT\n%d",
		nodeName,
		nodeIpAddr,
		item.Name,
		item.IpAddr,
		item.LastCheck,
		item.LastRTT,
	))
}

func (self *WebHook) SendHddWaring(nodeName string, nodeIpAddr string, item *HDDItem) {
	self.Send(fmt.Sprintf("[%s/%s] %s (%s) Disk 사용량 경고\n\n- 마지막 확인 시간\n%s\n- 사용율\n%0.2f%%\n- 총 공간\n%s\n- 사용중인 공간\n%s\n- 잔여 공간\n%s",
		nodeName,
		nodeIpAddr,
		item.Name,
		item.Path,
		item.LastCheck,
		item.Usage,
		item.Total,
		item.Used,
		item.Free,
	))
}

func (self *WebHook) CheckHeartBeat(node *NodeData) {
	for _, x := range node.HeartbeatItems {
		if !x.IsOnline {
			wh.SendHeartBeatWarning(node.Name, node.IpAddr, x)
		}
	}
}

func (self *WebHook) CheckPing(node *NodeData) {
	for _, x := range node.PingItems {
		if !x.IsOnline {
			wh.SendPingWarning(node.Name, node.IpAddr, x)
		}
	}
}

func (self *WebHook) CheckHdd(node *NodeData) {
	for _, x := range node.HddItems {
		if x.IsWarning {
			wh.SendHddWaring(node.Name, node.IpAddr, x)
		}
	}
}

func (self *WebHook) Send(content string) {
	if !self.isEnabled {
		return
	}

	for _, item := range self.items {
		data := make(map[string]interface{})

		for k, v := range item.Data {
			switch v := v.(type) {
			case string:
				data[k] = self.VarMatching(v, content)
			case fmt.Stringer:
				data[k] = self.VarMatching(v.String(), content)
			default:
				data[k] = v
			}
		}

		go Post(item.EndPoint, item.Headers, item.DataType, data)
	}
}

func (self *WebHook) VarMatching(target string, content string) string {
	target = strings.ReplaceAll(target, "%NODE_NAME%", NODE_NAME)
	target = strings.ReplaceAll(target, "%IP_ADDR%", LOCAL_IPADDR)
	target = strings.ReplaceAll(target, "%CONTENT%", content)
	return target
}
