package util

import (
	"errors"
	"net"
	"net/http"
	"sync"
	"time"
)

type ReqRes struct {
	Request       *http.Request
	Response      chan *http.Response
	RequestGroup  RequestGroup
	ResponseGroup chan []*http.Response
	Error         chan error
}

type ThrottledClient struct {
	interval   time.Duration
	ticker     *time.Ticker
	pool       chan *ReqRes
	maxsize    int
	stopCh     chan bool
	httpClient *http.Client
	wg         *sync.WaitGroup
}

type ClientOptions func(*http.Client)

func NewThrottledClient(interval time.Duration, maxsize int, opts ...ClientOptions) *ThrottledClient {
	httpClient := &http.Client{
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
	for _, o := range opts {
		o(httpClient)
	}
	return &ThrottledClient{
		interval:   interval,
		ticker:     time.NewTicker(interval),
		pool:       make(chan *ReqRes, maxsize),
		maxsize:    maxsize,
		stopCh:     make(chan bool),
		httpClient: httpClient,
		wg:         new(sync.WaitGroup),
	}
}

func (n *ThrottledClient) Start() {
	go n.poll()
}

func (n *ThrottledClient) Stop() {
	close(n.stopCh)
	close(n.pool)
	n.ticker.Stop()
	n.wg.Wait()
}

func (n *ThrottledClient) request(r *ReqRes) {
	resp, err := n.httpClient.Do(r.Request)
	if err != nil {
		go func() {
			r.Error <- err
		}()
		n.wg.Done()
		return
	}
	go func() {
		r.Response <- resp
	}()
	n.wg.Done()
}
func (n *ThrottledClient) requestGroup(r *ReqRes) {
	req := r.RequestGroup.Next(nil)
	responses := make([]*http.Response, 0)
	var err error
	for req != nil && err == nil {
		resp, e := n.httpClient.Do(req)
		if e != nil {
			err = e
			continue
		}
		responses = append(responses, resp)
		req = r.RequestGroup.Next(resp)
	}
	if err != nil {
		go func() {
			r.Error <- err
		}()
		n.wg.Done()
		return
	}
	go func() {
		r.ResponseGroup <- responses
	}()
	n.wg.Done()
}

func (n *ThrottledClient) poll() {
	for {
		select {
		case <-n.ticker.C:
			reqRes, more := <-n.pool
			if reqRes != nil {
				n.wg.Add(1)
				if reqRes.Request != nil {
					go n.request(reqRes)
				} else if reqRes.RequestGroup != nil {
					go n.requestGroup(reqRes)
				}
			}
			if !more {
				goto LOOP_OUT
			}
		case <-n.stopCh:
			goto LOOP_OUT
		}
	}
LOOP_OUT:
	for r := range n.pool {
		go func(rr *ReqRes) {
			rr.Error <- errors.New("network closed")
		}(r)
	}
}

func (n *ThrottledClient) Do(r *http.Request) (*ReqRes, error) {
	select {
	case <-n.stopCh:
		return nil, errors.New("network closed")
	default:
	}
	reqRes := &ReqRes{
		Request:  r,
		Response: make(chan *http.Response),
		Error:    make(chan error),
	}

	// Can panic here due to race conditions. Need to use a slice to manage requests
	// If Stop and Do are called simultaneously
	n.pool <- reqRes
	return reqRes, nil
}

// Do a group of requests together, pass response of previous to the get the next request
type RequestGroup interface {
	// First called with nil and continutes until nil is returned
	Next(*http.Response) *http.Request
}

func (n *ThrottledClient) DoGroup(r RequestGroup) (*ReqRes, error) {
	select {
	case <-n.stopCh:
		return nil, errors.New("network closed")
	default:
	}
	reqRes := &ReqRes{
		RequestGroup:  r,
		ResponseGroup: make(chan []*http.Response),
		Error:         make(chan error),
	}
	n.pool <- reqRes
	return reqRes, nil
}
