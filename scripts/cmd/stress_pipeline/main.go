// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/njcx/libbeat_v8/beat"
	"github.com/njcx/libbeat_v8/common"
	_ "github.com/njcx/libbeat_v8/outputs/console"
	_ "github.com/njcx/libbeat_v8/outputs/elasticsearch"
	_ "github.com/njcx/libbeat_v8/outputs/fileout"
	_ "github.com/njcx/libbeat_v8/outputs/logstash"
	"github.com/njcx/libbeat_v8/publisher/pipeline/stress"
	_ "github.com/njcx/libbeat_v8/publisher/queue/memqueue"
	conf "github.com/elastic/elastic-agent-libs/config"
	logpcfg "github.com/elastic/elastic-agent-libs/logp/configure"
	"github.com/elastic/elastic-agent-libs/paths"
	"github.com/elastic/elastic-agent-libs/service"
)

var (
	duration   time.Duration // -duration <duration>
	overwrites = conf.SettingFlag(nil, "E", "Configuration overwrite")
)

type config struct {
	Path    paths.Path
	Logging *conf.C
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	info := beat.Info{
		Beat:     "stresser",
		Version:  "0",
		Name:     "stresser.test",
		Hostname: "stresser.test",
	}

	flag.DurationVar(&duration, "duration", 0, "Test duration (default 0)")
	flag.Parse()

	files := flag.Args()
	fmt.Println("load config files:", files)

	cfg, err := common.LoadFiles(files...)
	if err != nil {
		return err
	}

	service.BeforeRun()
	defer service.Cleanup()

	if err := cfg.Merge(overwrites); err != nil {
		return err
	}

	config := config{}
	if err := cfg.Unpack(&config); err != nil {
		return err
	}

	if err := paths.InitPaths(&config.Path); err != nil {
		return err
	}
	if err = logpcfg.Logging("test", config.Logging); err != nil {
		return err
	}

	common.PrintConfigDebugf(cfg, "input config:")

	return stress.RunTests(info, duration, cfg, nil)
}

func startHTTP(bind string) {
	go func() {
		http.ListenAndServe(bind, nil)
	}()
}
