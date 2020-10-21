package main

import (
	"fmt"
	"github.com/tatsushid/go-fastping"
	"net"
	"time"
)

type Ping struct {
	items []*PingItem
	table map[string]*PingItem
}

type PingItem struct {
	Name          string `json:"name"`
	IpAddr        string `json:"ip_addr"`
	LastCheckTime int64  `json:"last_check_time_seconds"`
	LastCheck     string `json:"last_check_time"`
	LastRTT       int64  `json:"last_check_rtt_mills"`
	IsOnline      bool   `json:"is_online"`
}

func (self *Ping) Run() {
	// Init
	parentNodeUrl := fmt.Sprintf("%s/ping", MASTER_NODE_API_HOST)
	self.table = make(map[string]*PingItem)

	for i, v := range PING_TARGETS {
		pi := &PingItem{
			Name:   PING_TARGET_NAMES[i],
			IpAddr: v,
		}

		self.items = append(self.items, pi)

		ipaddr := GetIpWithDomainName(v).String()
		self.table[ipaddr] = pi
	}

	// Loop
	go self.checkLoop()

	<-time.After(time.Second * time.Duration(PING_TIMEOUT_SECONDS))
	for RUNNING {
		for _, v := range self.items {
			now := time.Now().Unix()

			if now-v.LastCheckTime > PING_TIMEOUT_SECONDS {
				LogDebug("Ping check fail on %s (%s), %s", v.Name, v.IpAddr, v.LastCheck)
				v.IsOnline = false
			} else {
				LogDebug("Ping check success on %s (%s), RTT %dms (%s)", v.Name, v.IpAddr, v.LastRTT, v.LastCheck)
				v.IsOnline = true
			}
		}

		data := NodeData{
			Name:      NODE_NAME,
			IpAddr:    LOCAL_IPADDR,
			PingItems: self.items,
		}

		if !IS_MASTER {
			go Post(parentNodeUrl, nil, "json", data)
		} else {
			go ctr.ProcessPing(data)
		}

		LogDebug("Wait next ping check %ds", PING_INTERVAL_SECONDS)
		<-time.After(time.Second * time.Duration(PING_INTERVAL_SECONDS))
	}
}

func (self *Ping) checkLoop() error {
	p := fastping.NewPinger()

	for _, v := range PING_TARGETS {
		ra, err := net.ResolveIPAddr("ip4:icmp", v)
		if err != nil {
			return err
		}
		p.AddIPAddr(ra)
	}

	p.MaxRTT = time.Second * time.Duration(PING_TIMEOUT_SECONDS)
	p.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
		LogVerbose("IP Addr: %s receive, RTT: %v", addr.String(), rtt)
		item := self.table[addr.String()]

		if item != nil {
			now := time.Now()
			item.LastRTT = rtt.Milliseconds()
			item.LastCheckTime = now.Unix()
			item.LastCheck = time.Unix(0, now.UnixNano()).String()
		}
	}
	p.OnIdle = func() {
		//
	}

	p.RunLoop()
	return nil
}
