package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type adguardResponse struct {
	Data []struct {
		Answer []struct {
			TTL   int    `json:"ttl"`
			Type  string `json:"type"`
			Value string `json:"value"`
		} `json:"answer"`
		AnswerDnssec bool   `json:"answer_dnssec"`
		Client       string `json:"client"`
		ClientProto  string `json:"client_proto"`
		ElapsedMs    string `json:"elapsedMs"`
		Question     struct {
			Class string `json:"class"`
			Host  string `json:"host"`
			Type  string `json:"type"`
		} `json:"question"`
		Reason   string `json:"reason"`
		Status   string `json:"status"`
		Time     string `json:"time"`
		Upstream string `json:"upstream"`
		FilterID int    `json:"filterId,omitempty"`
		Rule     string `json:"rule,omitempty"`
	} `json:"data"`
	Oldest string `json:"oldest"`
}

func handler(w http.ResponseWriter, req *http.Request) {
	hostChan := make(chan string, 10)
	errChan := make(chan error, 2)
	wg := sync.WaitGroup{}

	var hosts []string
	go func() {
		for host := range hostChan {
			if host == "" {
				continue
			}

			// check to see if we already have the host in the list
			var found bool
			for _, h := range hosts {
				if h == host {
					found = true
				}

			}

			if !found {
				hosts = append(hosts, host)
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		resp, err := http.Get("https://raw.githubusercontent.com/kboghdady/youTube_ads_4_pi-hole/master/black.list")
		if err != nil {
			errChan <- err
			return
		}
		defer resp.Body.Close()

		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			errChan <- err
			return
		}

		for _, host := range strings.Split(string(b), "\n") {
			hostChan <- host
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		r, err := http.NewRequest("GET", fmt.Sprintf("http://%s:%s/control/querylog?search=googlevideo&response_status=all&older_than=", os.Getenv("ADGUARD_HOST"), os.Getenv("ADGUARD_PORT")), nil)
		if err != nil {
			errChan <- err
			return
		}

		r.SetBasicAuth(os.Getenv("ADGUARD_USERNAME"), os.Getenv("ADGUARD_PASSWORD"))

		resp, err := http.DefaultClient.Do(r)
		if err != nil {
			errChan <- fmt.Errorf("AdguardHome request error: %v", err)
			return
		}
		defer resp.Body.Close()

		agr := new(adguardResponse)
		err = json.NewDecoder(resp.Body).Decode(agr)
		if err != nil {
			errChan <- err
			return
		}

		for _, host := range agr.Data {
			hostChan <- host.Question.Host
		}
	}()

	wg.Wait()
	close(errChan)
	close(hostChan)

	for err := range errChan {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("failed to generate youtube adblock: %v\n", err)
	}

	w.Header().Set("content-type", "text/plain")
	w.Write([]byte(strings.Join(hosts, "\n")))
}

func main() {
	srv := &http.Server{
		Addr:         ":8082",
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}

	http.HandleFunc("/", handler)
	log.Fatal(srv.ListenAndServe())
}
