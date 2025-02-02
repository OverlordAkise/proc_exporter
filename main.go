package main

import (
	"flag"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log/slog"
	"maps"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Localonly bool
	Port      int
	LogPath   string
}

func main() {
	starttime := time.Now()
	var logger *slog.Logger
	var config Config

	//config
	// var allowed string
	flag.BoolVar(&config.Localonly, "local", false, "listen on localhost only")
	flag.IntVar(&config.Port, "port", 4885, "port to listen on")
	flag.StringVar(&config.LogPath, "log", "stdout", "where to log to, e.g. `./procexp.log`, default='stdout'")
	flag.Parse()

	//log
	if config.LogPath == "stdout" {
		logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	} else {
		f, err := os.OpenFile(config.LogPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		logger = slog.New(slog.NewTextHandler(f, nil))
	}

	//web
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK")
	})

	cpuTimeTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "proc",
			Name:      "cpu_time_total",
			Help:      "clock ticks per process",
		},
		[]string{"pname"},
	)
	prometheus.MustRegister(cpuTimeTotal)

	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		for name, cputime := range GetStats() {
			cpuTimeTotal.WithLabelValues(name).Add(cputime)
		}
		promhttp.Handler().ServeHTTP(w, r)
		//logger.Info("request", "url", "/metrics", "remoteaddr", r.RemoteAddr)
	})

	listenHost := ":" + strconv.Itoa(config.Port)
	if config.Localonly {
		listenHost = "127.0.0.1" + listenHost
	}
	donetime := time.Now()
	logger.Info("Startup finished", "timetaken", donetime.Sub(starttime).String(), "listen", listenHost)
	panic(http.ListenAndServe(listenHost, nil))
}

// https://man7.org/linux/man-pages/man5/proc_pid_stat.5.html
var re = regexp.MustCompile(`\(.+[ ].+\)`)

func ParseStatForNameAndCPU(in string) (string, float64) {
	var t int
	var tt string
	var utime float64
	var stime float64
	var name string
	// Extra logic because of "(tmux: server)" name messing up the whitespace parsing
	if re.MatchString(in) {
		orig := re.FindString(in)
		repl := strings.ReplaceAll(orig, " ", "_")
		in = strings.ReplaceAll(in, orig, repl)
	}
	_, err := fmt.Sscanf(in, "%d %s %s %d %d %d %d %d %d %d %d %d %d %f %f", &t, &name, &tt, &t, &t, &t, &t, &t, &t, &t, &t, &t, &t, &utime, &stime)
	if err != nil {
		panic(err)
	}
	return strings.Trim(name, "()"), utime + stime
}

var oldValues = map[string]float64{}

func GetStats() map[string]float64 {
	ret := map[string]float64{}
	newOldValues := map[string]float64{}
	folders, err := os.ReadDir("/proc")
	if err != nil {
		panic(err)
	}
	procPIDs := []string{}
	for _, dir := range folders {
		fName := dir.Name()
		if _, err := strconv.Atoi(fName); err == nil {
			procPIDs = append(procPIDs, fName)
		}
	}

	for _, pid := range procPIDs {
		bytes, err := os.ReadFile("/proc/" + pid + "/stat")
		if err != nil {
			continue
		}
		name, cpu := ParseStatForNameAndCPU(string(bytes))
		newOldValues[name] = cpu
		if _, exist := ret[name]; !exist {
			ret[name] = 0
		}
		if val, exist := oldValues[name]; exist && val < cpu {
			ret[name] += (cpu - val)
		} else {
			ret[name] += cpu
		}
	}
	oldValues = maps.Clone(newOldValues)
	return ret
}
