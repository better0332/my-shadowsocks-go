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
	ts *trafficStat

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
	ts = &trafficStat{m: make(map[string]*trafficStruct, 100)}
	go sendTraffic()
}

func upTraffic(port string, traffic int, ip string) {
	ts.Lock()
	defer ts.Unlock()

	if st, ok := ts.m[port]; ok {
		st.Traffic += traffic
		if ip != "" {
			st.ClientIP = ip
		}
	}
}

func DelTraffic(port string) {
	ts.Lock()
	defer ts.Unlock()

	delete(ts.m, port)
}

func AddTraffic(port string) {
	ts.Lock()
	defer ts.Unlock()

	ts.m[port] = &trafficStruct{}
}

func sendTraffic() {
	for {
		time.Sleep(30 * time.Second)

		ts.Lock()
		if len(ts.m) == 0 {
			ts.Unlock()
			continue
		}
		buf, err := json.Marshal(ts.m)
		ts.Unlock()
		if err != nil {
			log.Println(err)
			continue
		}

		if resp, err := client.PostForm("https://shadowrockets.com/traffic_stat.php",
			url.Values{"traffic": {string(buf)}}); err == nil {
			cont, err := ioutil.ReadAll(resp.Body)
			resp.Body.Close()
			if string(cont) != "success" {
				if err != nil {
					log.Println(err)
				} else {
					log.Printf("%s\n", cont)
				}
				continue
			}
			ts.Lock()
			for k, _ := range ts.m {
				ts.m[k].Traffic = 0
			}
			ts.Unlock()

			Debug.Println("Update Traffic Stat Success")
		}
	}
}
