// Copyright 2013 Prometheus Team
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package web

import (
	"flag"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"time"

	pprof_runtime "runtime/pprof"

	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/exp"

	"github.com/prometheus/prometheus/web/api"
	"github.com/prometheus/prometheus/web/blob"
)

// Commandline flags.
var (
	listenAddress  = flag.String("listenAddress", ":9090", "Address to listen on for web interface.")
	useLocalAssets = flag.Bool("useLocalAssets", false, "Read assets/templates from file instead of binary.")
	userAssetsPath = flag.String("userAssets", "", "Path to static asset directory, available at /user")
	enableQuit     = flag.Bool("web.enableRemoteShutdown", false, "Enable remote service shutdown")
)

type WebService struct {
	StatusHandler    *PrometheusStatusHandler
	DatabasesHandler *DatabasesHandler
	MetricsHandler   *api.MetricsService
	AlertsHandler    *AlertsHandler

	QuitDelegate func()
}

func (w WebService) ServeForever() error {
	exp.Handle("/favicon.ico", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "", 404)
	}))

	// TODO(julius): This will need to be rewritten once the exp package provides
	// the coarse mux behaviors via a wrapper function.
	exp.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
	exp.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
	exp.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
	exp.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))

	exp.Handle("/", w.StatusHandler)
	exp.Handle("/databases", w.DatabasesHandler)
	exp.Handle("/alerts", w.AlertsHandler)
	exp.HandleFunc("/graph", graphHandler)
	exp.HandleFunc("/heap", dumpHeap)

	w.MetricsHandler.RegisterHandler()
	exp.Handle("/metrics", prometheus.DefaultHandler)
	if *useLocalAssets {
		exp.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))
	} else {
		exp.Handle("/static/", http.StripPrefix("/static/", new(blob.Handler)))
	}

	if *userAssetsPath != "" {
		exp.Handle("/user/", http.StripPrefix("/user/", http.FileServer(http.Dir(*userAssetsPath))))
	}

	if *enableQuit {
		exp.HandleFunc("/-/quit", w.quitHandler)
	}

	glog.Info("listening on ", *listenAddress)

	return http.ListenAndServe(*listenAddress, exp.DefaultCoarseMux)
}

func (s WebService) quitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.Header().Add("Allow", "POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	fmt.Fprintf(w, "Requesting termination... Goodbye!")

	s.QuitDelegate()
}

func getLocalTemplate(name string) (*template.Template, error) {
	return template.ParseFiles(
		"web/templates/_base.html",
		fmt.Sprintf("web/templates/%s.html", name),
	)
}

func getEmbeddedTemplate(name string) (*template.Template, error) {
	t := template.New("_base")

	file, err := blob.GetFile(blob.TemplateFiles, "_base.html")
	if err != nil {
		glog.Error("Could not read base template: ", err)
		return nil, err
	}
	t.Parse(string(file))

	file, err = blob.GetFile(blob.TemplateFiles, name+".html")
	if err != nil {
		glog.Errorf("Could not read %s template: %s", name, err)
		return nil, err
	}
	t.Parse(string(file))

	return t, nil
}

func getTemplate(name string) (t *template.Template, err error) {
	if *useLocalAssets {
		t, err = getLocalTemplate(name)
	} else {
		t, err = getEmbeddedTemplate(name)
	}

	if err != nil {
		return
	}

	if *userAssetsPath != "" {
		// replace "user_dashboard_link" template
		t.Parse(`{{define "user_dashboard_link"}}<a href="/user">User Dashboard{{end}}`)
	}

	return
}

func executeTemplate(w http.ResponseWriter, name string, data interface{}) {
	tpl, err := getTemplate(name)
	if err != nil {
		glog.Error("Error preparing layout template: ", err)
		return
	}
	err = tpl.Execute(w, data)
	if err != nil {
		glog.Error("Error executing template: ", err)
	}
}

func dumpHeap(w http.ResponseWriter, r *http.Request) {
	target := fmt.Sprintf("/tmp/%d.heap", time.Now().Unix())
	f, err := os.Create(target)
	if err != nil {
		glog.Error("Could not dump heap: ", err)
	}
	fmt.Fprintf(w, "Writing to %s...", target)
	defer f.Close()
	pprof_runtime.WriteHeapProfile(f)
	fmt.Fprintf(w, "Done")
}

func MustBuildServerUrl() string {
	_, port, err := net.SplitHostPort(*listenAddress)
	if err != nil {
		panic(err)
	}
	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("http://%s:%s", hostname, port)
}
