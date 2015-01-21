package shadowsocks

import (
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"
)

var (
	Traffic *trafficStat

	tr     = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	client = &http.Client{Transport: tr}
)

type trafficStruct struct {
	Traffic  int
	ClientIP string
}

type trafficStat struct {
	sync.Mutex
	m map[string]*trafficStruct
}

func NewTraffic() {
	Traffic = &trafficStat{m: make(map[string]*trafficStruct, 100)}
	go Traffic.sendTraffic()
}

func (t *trafficStat) upTraffic(port string, traffic int, ip string) {
	t.Lock()
	defer t.Unlock()

	if st, ok := t.m[port]; ok {
		st.Traffic += traffic
		if ip != "" {
			st.ClientIP = ip
		}
	}
}

func (t *trafficStat) DelTraffic(port string) {
	t.Lock()
	defer t.Unlock()

	delete(t.m, port)
}

func (t *trafficStat) sendTraffic() {
	for {
		time.Sleep(30 * time.Second)

		t.Lock()
		if len(t.m) == 0 {
			continue
		}
		buf, err := json.Marshal(t.m)
		t.Unlock()
		if err != nil {
			log.Println(err)
			continue
		}

		if resp, err := client.PostForm("https://www.webscan8.com/traffic_stat.php",
			url.Values{"traffic": {string(buf)}}); err == nil {
			defer resp.Body.Close()
			cont, err := ioutil.ReadAll(resp.Body)
			if string(cont) != "success" {
				if err != nil {
					log.Println(err)
				} else {
					log.Println(cont)
				}
				continue
			}
			t.Lock()
			for k, _ := range t.m {
				t.m[k].Traffic = 0
			}
			t.Unlock()

			Debug.Println("Update Traffic Stat Success")
		}
	}
}
