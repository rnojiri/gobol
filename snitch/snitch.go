package snitch

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/robfig/cron"
	"go.uber.org/zap"
)

// Stats holds several informations.
// The timeseries backend address and port. The POST interval. The default tags
// to be added to all points and a map of all points.
type Stats struct {
	logger              *zap.Logger
	cron                *cron.Cron
	address             string
	port                int
	tags                map[string]string
	proto               string
	timeout             time.Duration
	postInt             time.Duration
	points              sync.Map
	hBuffer             []message
	receiver            chan message
	raiseDebugVerbosity bool
	terminate           bool

	mtx sync.RWMutex
}

// New creates a new stats
func New(logger *zap.Logger, settings Settings) (*Stats, error) {

	if settings.Address == "" {
		return nil, errors.New("address is required")
	}
	if settings.Port == 0 {
		return nil, errors.New("port is required")
	}
	if settings.Protocol != "http" && settings.Protocol != "udp" {
		return nil, errors.New("protocol supported: udp and http")
	}

	var dur, postInt time.Duration
	var err error
	if settings.Protocol == "http" {
		dur, err = time.ParseDuration(settings.HTTPTimeout)
		if err != nil {
			return nil, err
		}
		postInt, err = time.ParseDuration(settings.HTTPPostInterval)
		if err != nil {
			return nil, err
		}
	}

	tags := map[string]string{}
	for k, v := range settings.Tags {
		tags[k] = v
	}

	if tags["ksid"] == "" {
		return nil, errors.New("tag ksid is mandatory")
	}

	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	tags["host"] = hostname

	stats := &Stats{
		cron:                cron.New(),
		address:             settings.Address,
		port:                settings.Port,
		proto:               settings.Protocol,
		timeout:             dur,
		postInt:             postInt,
		logger:              logger,
		tags:                tags,
		points:              sync.Map{},
		hBuffer:             []message{},
		receiver:            make(chan message),
		raiseDebugVerbosity: settings.RaiseDebugVerbosity,
		terminate:           false,
	}
	go stats.start(settings.Runtime)
	return stats, nil
}

// Terminate - terminates the instance
func (st *Stats) Terminate() {

	st.terminate = true
}

func (st *Stats) start(runtime bool) {

	if st == nil {
		return
	}

	st.mtx.RLock()
	st.points.Range(func(_, v interface{}) bool {
		p := v.(*CustomPoint)
		st.cron.AddJob(p.interval, p)
		return true
	})

	st.mtx.RUnlock()

	if st.proto == "udp" {
		go st.clientUDP()
	} else {
		go st.clientHTTP()
	}
	st.cron.Start()
	if runtime {
		go st.runtimeLoop()
	}
}

func (st *Stats) runtimeLoop() {

	for {
		<-time.After(30 * time.Second)
		if st.terminate {
			st.logger.Info("terminating the runtime loop")
			return
		}
		st.ValueAdd(
			"runtime.goroutines.count",
			st.tags, "max", "@every 1m", false, true,
			float64(runtime.NumGoroutine()),
		)
	}
}

func (st *Stats) clientUDP() {
	conn, err := net.Dial("udp", fmt.Sprintf("%s:%d", st.address, st.port))
	if err != nil {
		st.logger.Error("connect", zap.Error(err))
	} else {
		defer conn.Close()
	}

	for {
		if st.terminate {
			st.logger.Info("terminating the udp client loop")
			return
		}

		messageData := <-st.receiver
		payload, err := json.Marshal(messageData)
		if err != nil {
			st.logger.Error("marshal", zap.Error(err))
		}

		if conn != nil {
			_, err = conn.Write(payload)
			if err != nil {
				st.logger.Error("write", zap.Error(err))
			}
		} else {
			conn, err = net.Dial("udp", fmt.Sprintf("%s:%d", st.address, st.port))
			if err != nil {
				st.logger.Error("connect", zap.Error(err))
			} else {
				defer conn.Close()
			}
		}
	}
}

func (st *Stats) clientHTTP() {
	client := &http.Client{
		Timeout: st.timeout,
	}

	url := fmt.Sprintf("%s://%s:%d/api/put", st.proto, st.address, st.port)
	ticker := time.NewTicker(st.postInt)
	for {
		if st.terminate {
			st.logger.Info("terminating the http client loop")
			return
		}

		select {
		case messageData := <-st.receiver:
			st.hBuffer = append(st.hBuffer, messageData)
		case <-ticker.C:
			payload, err := json.Marshal(st.hBuffer)
			if err != nil {
				st.logger.Error("", zap.Error(err))
				break
			}
			req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
			if err != nil {
				st.logger.Error("", zap.Error(err))
				break
			}
			resp, err := client.Do(req)
			if err != nil {
				st.logger.Error("", zap.Error(err))
				break
			}
			if resp.StatusCode != http.StatusNoContent {
				reqResponse, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					st.logger.Error("", zap.Error(err))
				}
				if st.raiseDebugVerbosity {
					st.logger.Debug(string(reqResponse))
				}
			}
			st.hBuffer = []message{}
			resp.Body.Close()
		}
	}
}
