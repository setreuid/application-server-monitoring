package main

import (
	"fmt"
	humanize "github.com/dustin/go-humanize"
	"github.com/minio/minio/pkg/disk"
	"time"
)

type HDD struct {
	items []*HDDItem
}

type HDDItem struct {
	Name          string  `json:"name"`
	Path          string  `json:"path"`
	LastCheckTime int64   `json:"last_check_time_seconds"`
	LastCheck     string  `json:"last_check_time"`
	Usage         float64 `json:"usage_percent"`
	Total         string  `json:"total"`
	Used          string  `json:"used"`
	Free          string  `json:"free"`
	IsWarning     bool    `json:"is_warning"`
}

func (self *HDD) Run() {
	// Init
	parentNodeUrl := fmt.Sprintf("%s/hdd", MASTER_NODE_API_HOST)

	for i, v := range HDD_TARGETS {
		self.items = append(self.items, &HDDItem{
			Name: HDD_TARGET_NAMES[i],
			Path: v,
		})
	}

	// Loop
	for RUNNING {
		for _, v := range self.items {
			self.check(v)

			now := time.Now()
			v.LastCheckTime = now.Unix()
			v.LastCheck = time.Unix(0, now.UnixNano()).String()
		}

		data := NodeData{
			Name:     NODE_NAME,
			IpAddr:   LOCAL_IPADDR,
			HddItems: self.items,
		}

		if !IS_MASTER {
			go Post(parentNodeUrl, nil, "json", data)
		} else {
			go ctr.ProcessHdd(data)
		}

		LogDebug("Wait next hdd check %ds", HDD_INTERVAL_SECONDS)
		<-time.After(time.Second * time.Duration(HDD_INTERVAL_SECONDS))
	}
}

func (self *HDD) check(item *HDDItem) error {
	di, err := disk.GetInfo(item.Path)
	if err != nil {
		return err
	}

	percentage := (float64(di.Total-di.Free) / float64(di.Total)) * 100

	item.Usage = percentage
	item.Total = humanize.Bytes(di.Total)
	item.Used = humanize.Bytes(di.Used)
	item.Free = humanize.Bytes(di.Free)

	if percentage > float64(HDD_LIMIT_PERCENT) {
		item.IsWarning = true
	} else {
		item.IsWarning = false
	}

	LogDebug("%s (%s) is %s of %s disk space used (%0.2f%%)",
		item.Name,
		item.Path,
		humanize.Bytes(di.Total-di.Free),
		humanize.Bytes(di.Total),
		percentage,
	)
	return nil
}
