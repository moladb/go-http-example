/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/moladb/go-http-example/cmd/kv-server/service/v0"
	"github.com/moladb/go-http-example/pkg/rest"
	"github.com/moladb/go-http-example/pkg/version"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "kv-server"
	app.Version = fmt.Sprintf("\nversion: %s\nbuild_date: %s\ngo_version: %s",
		version.VERSION,
		version.BUILDDATE,
		version.GOVERSION)
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name: "enable-pprof",
		},
		cli.BoolFlag{
			Name: "enable-metrics",
		},
		cli.StringFlag{
			Name:  "bind-addr",
			Value: "0.0.0.0:8500",
		},
	}
	app.Usage = "kv-server"
	app.Action = func(c *cli.Context) error {
		srv := rest.NewServer(rest.Config{
			BindAddr:              c.String("bind-addr"),
			EnablePProf:           c.Bool("enable-pprof"),
			EnableAPIMetrics:      c.Bool("enable-metrics"),
			GraceShutdownTimeoutS: 60,
		})

		// register services
		srv.RegisterService(v0.NewKVService())

		go func() {
			quit := make(chan os.Signal, 1)
			signal.Notify(quit, os.Interrupt)
			<-quit
			srv.Shutdown()
		}()

		if err := srv.Run(); err != nil {
			fmt.Println("err:", err)
			os.Exit(1)
		}
		return nil
	}

	app.Run(os.Args)
}
