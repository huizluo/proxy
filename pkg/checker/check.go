package checker

import (
	"io/ioutil"
	"log"
	"net"
	"net/url"
	"strings"
	"time"

	"proxy/pkg/utils"
	conn2 "proxy/pkg/conn"
)

type Checker struct {
	data       utils.ConcurrentMap
	blockedMap utils.ConcurrentMap
	directMap  utils.ConcurrentMap
	interval   int64
	timeout    int
}
type CheckerItem struct {
	IsHTTPS      bool
	Method       string
	URL          string
	Domain       string
	Host         string
	Data         []byte
	SuccessCount uint
	FailCount    uint
}

//NewChecker args:
//timeout : tcp timeout milliseconds ,connect to host
//interval: recheck domain interval seconds
func NewChecker(timeout int, interval int64, blockedFile, directFile string) Checker {
	ch := Checker{
		data:     utils.NewConcurrentMap(),
		interval: interval,
		timeout:  timeout,
	}
	ch.blockedMap = ch.loadMap(blockedFile)
	ch.directMap = ch.loadMap(directFile)
	if !ch.blockedMap.IsEmpty() {
		log.Printf("blocked file loaded , domains : %d", ch.blockedMap.Count())
	}
	if !ch.directMap.IsEmpty() {
		log.Printf("direct file loaded , domains : %d", ch.directMap.Count())
	}
	ch.start()
	return ch
}

func (c *Checker) loadMap(f string) (dataMap utils.ConcurrentMap) {
	dataMap = utils.NewConcurrentMap()
	if utils.PathExists(f) {
		_contents, err := ioutil.ReadFile(f)
		if err != nil {
			log.Printf("load file err:%s", err)
			return
		}
		for _, line := range strings.Split(string(_contents), "\n") {
			line = strings.Trim(line, "\r \t")
			if line != "" {
				dataMap.Set(line, true)
			}
		}
	}
	return
}
func (c *Checker) start() {
	go func() {
		for {
			for _, v := range c.data.Items() {
				go func(item CheckerItem) {
					if c.isNeedCheck(item) {
						//log.Printf("check %s", item.Domain)
						var conn net.Conn
						var err error
						if item.IsHTTPS {
							conn, err = conn2.ConnectHost(item.Host, c.timeout)
							if err == nil {
								conn.SetDeadline(time.Now().Add(time.Millisecond))
								conn.Close()
							}
						} else {
							err = conn2.HTTPGet(item.URL, c.timeout)
						}
						if err != nil {
							item.FailCount = item.FailCount + 1
						} else {
							item.SuccessCount = item.SuccessCount + 1
						}
						c.data.Set(item.Host, item)
					}
				}(v.(CheckerItem))
			}
			time.Sleep(time.Second * time.Duration(c.interval))
		}
	}()
}
func (c *Checker) isNeedCheck(item CheckerItem) bool {
	var minCount uint = 5
	if (item.SuccessCount >= minCount && item.SuccessCount > item.FailCount) ||
		(item.FailCount >= minCount && item.SuccessCount > item.FailCount) ||
		c.domainIsInMap(item.Host, false) ||
		c.domainIsInMap(item.Host, true) {
		return false
	}
	return true
}
func (c *Checker) IsBlocked(address string) (blocked bool, failN, successN uint) {
	if c.domainIsInMap(address, true) {
		//log.Printf("%s in blocked ? true", address)
		return true, 0, 0
	}
	if c.domainIsInMap(address, false) {
		//log.Printf("%s in direct ? true", address)
		return false, 0, 0
	}

	_item, ok := c.data.Get(address)
	if !ok {
		//log.Printf("%s not in map, blocked true", address)
		return true, 0, 0
	}
	item := _item.(CheckerItem)

	return item.FailCount >= item.SuccessCount, item.FailCount, item.SuccessCount
}
func (c *Checker) domainIsInMap(address string, blockedMap bool) bool {
	u, err := url.Parse("http://" + address)
	if err != nil {
		log.Printf("blocked check , url parse err:%s", err)
		return true
	}
	domainSlice := strings.Split(u.Hostname(), ".")
	if len(domainSlice) > 1 {
		subSlice := domainSlice[:len(domainSlice)-1]
		topDomain := strings.Join(domainSlice[len(domainSlice)-1:], ".")
		checkDomain := topDomain
		for i := len(subSlice) - 1; i >= 0; i-- {
			checkDomain = subSlice[i] + "." + checkDomain
			if !blockedMap && c.directMap.Has(checkDomain) {
				return true
			}
			if blockedMap && c.blockedMap.Has(checkDomain) {
				return true
			}
		}
	}
	return false
}
func (c *Checker) Add(address string, isHTTPS bool, method, URL string, data []byte) {
	if c.domainIsInMap(address, false) || c.domainIsInMap(address, true) {
		return
	}
	if !isHTTPS && strings.ToLower(method) != "get" {
		return
	}
	var item CheckerItem
	u := strings.Split(address, ":")
	item = CheckerItem{
		URL:     URL,
		Domain:  u[0],
		Host:    address,
		Data:    data,
		IsHTTPS: isHTTPS,
		Method:  method,
	}
	c.data.SetIfAbsent(item.Host, item)
}
