// Copyright (c) 2016-2018 Christian Saide <supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/supernomad/quantum/blob/master/LICENSE

package rest

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/supernomad/quantum/common"
	"github.com/supernomad/quantum/metric"
	"github.com/supernomad/quantum/version"
)

// Rest is a generic rest api struct for exporting internal information and general purpose api settings.
type Rest struct {
	cfg        *common.Config
	stopped    bool
	server     *http.Server
	aggregator *metric.Aggregator
}

func (rest *Rest) returnStats(w http.ResponseWriter, r *http.Request) {
	rest.cfg.Log.Debug.Println("[REST]", "Received an api request:", r)

	header := w.Header()
	header.Set("Content-Type", "application/json")
	header.Set("Server", "quantum v"+version.Version())

	_, err := w.Write(rest.aggregator.Bytes(strings.Contains(r.RequestURI, "pretty")))
	if err != nil {
		rest.cfg.Log.Error.Println("[REST]", "Error writing stats api response:", err.Error())
	}
}

func (rest *Rest) run() {
	http.HandleFunc(rest.cfg.StatsRoute, rest.returnStats)

	for {
		if err := rest.server.ListenAndServe(); err != nil && !rest.stopped {
			rest.cfg.Log.Error.Println("[REST]", "Error initializing stats api:", err.Error())
		}

		time.Sleep(10 * time.Second)
	}
}

// Start will start the rest api up on the specified address and routes.
func (rest *Rest) Start() {
	go rest.run()
}

// Stop will stop the rest api.
func (rest *Rest) Stop() error {
	rest.stopped = true
	return rest.server.Close()
}

// New generates an Rest instance exposing metrics and general purpose routes via a REST api interface.
func New(cfg *common.Config, aggregator *metric.Aggregator) *Rest {
	return &Rest{
		cfg:        cfg,
		server:     &http.Server{Addr: fmt.Sprintf("%s:%d", cfg.StatsAddress, cfg.StatsPort)},
		aggregator: aggregator,
	}
}
