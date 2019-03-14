package navcoind

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/NavPool/navpool-api/config"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type rpcClient struct {
	serverAddr string
	user       string
	password   string
	httpClient *http.Client
}

type rpcRequest struct {
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	Id      int64       `json:"id"`
	JsonRpc string      `json:"jsonrpc"`
}

type rpcResponse struct {
	Id     int64           `json:"id"`
	Result json.RawMessage `json:"result"`
	Err    interface{}     `json:"error"`
}

func newClient(host string, port int, user, password string) (c *rpcClient, err error) {
	if len(host) == 0 {
		err = errors.New("bad call missing argument host")
		return
	}

	var httpClient *http.Client
	httpClient = &http.Client{}

	c = &rpcClient{serverAddr: fmt.Sprintf("http://%s:%d", host, port), user: user, password: password, httpClient: httpClient}

	return
}

func (c *rpcClient) doTimeoutRequest(timer *time.Timer, req *http.Request) (*http.Response, error) {
	type result struct {
		resp *http.Response
		err  error
	}
	done := make(chan result, 1)
	go func() {
		resp, err := c.httpClient.Do(req)
		done <- result{resp, err}
	}()

	select {
	case r := <-done:
		return r.resp, r.err
	case <-timer.C:
		return nil, errors.New("timeout reading data from server")
	}
}

func (c *rpcClient) call(method string, params interface{}) (rr rpcResponse, err error) {
	log.Printf("Navcoind: Method(%s) Params(%s)", method, params)
	connectTimer := time.NewTimer(RPCCLIENT_TIMEOUT * time.Second)
	rpcR := rpcRequest{method, params, time.Now().UnixNano(), "1.0"}
	payloadBuffer := &bytes.Buffer{}
	jsonEncoder := json.NewEncoder(payloadBuffer)
	err = jsonEncoder.Encode(rpcR)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", c.serverAddr, payloadBuffer)
	log.Printf("Navcoind: Request(%s)", payloadBuffer)

	if err != nil {
		return
	}

	req.Header.Add("Content-Type", "application/json;charset=utf-8")
	req.Header.Add("Accept", "application/json")

	if len(c.user) > 0 || len(c.password) > 0 {
		if config.Get().Debug == true {
			log.Printf("Navcoind: Username(%s), Password(%s)", c.user, c.password)
		}
		req.SetBasicAuth(c.user, c.password)
	}

	resp, err := c.doTimeoutRequest(connectTimer, req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	log.Printf("Navcoind: Response(%s)", data)
	if err != nil {
		return
	}

	err = json.Unmarshal(data, &rr)

	return
}
