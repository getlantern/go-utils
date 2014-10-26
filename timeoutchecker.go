// timeoutchecker opens a connection, waits some period of time, then tries to
// process an HTTP request on that connection.
package main

import (
	"flag"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/getlantern/golog"
)

var (
	log = golog.LoggerFor("timeoutchecker")

	target    = flag.String("url", "https://news.ycombinator.com/", "url to test")
	keepalive = flag.Int("keepalive", 0, "interval at which to send empty data as keepalives (0 means don't use keepalives)")
)

func main() {
	flag.Parse()
	targetUrl, err := url.Parse(*target)
	if err != nil {
		log.Fatalf("Unable to parse url %s: %s", target, err)
	}

	increment := 1
	if *keepalive > 0 {
		increment = *keepalive
	}
	for i := 1; i < 70; i += increment {
		test(targetUrl, i)
	}
}

func test(targetUrl *url.URL, wait int) {
	sleepTime := time.Duration(wait) * time.Second

	tr := &http.Transport{
		Dial: func(network, addr string) (conn net.Conn, err error) {
			if *keepalive > 0 {
				defer func() {
					if err == nil {
						for i := 0; i < int(sleepTime.Seconds()) / *keepalive; i++ {
							time.Sleep(time.Duration(*keepalive) * time.Second)
							conn.Write([]byte{})
						}
					}
				}()
			} else {
				defer time.Sleep(sleepTime)
			}
			return net.Dial("tcp", addr)
		},
	}

	client := &http.Client{
		Transport: tr,
	}
	req, err := http.NewRequest("GET", targetUrl.String(), nil)
	if err != nil {
		log.Fatalf("Unable to construct request: %s", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Debugf("At %s, Unable to execute request: %s", sleepTime, err)
		return
	}
	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Debugf("At %s, unable to read response: %s", sleepTime, err)
		return
	}

	log.Debugf("At %s, read response of size: %d", sleepTime, len(bytes))
}
