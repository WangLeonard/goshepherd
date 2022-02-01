package main

import (
	"embed"
	"flag"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

//go:embed static/*
var f embed.FS
var pprofExePath, traceExePath string

var Port = flag.Int("p", 7777, "port")

var currIP string

func init() {
	currIP = getClientIp()
	flag.Parse()
}

func main() {
	welcome()
	initGoToolPath()

	http.Handle("/", &indexHandle{})
	http.Handle("/static/", http.FileServer(http.FS(f)))
	http.Handle("/api", shepherdInst)

	ch := make(chan int)
	go func() {
		select {
		case <-ch:
		case <-time.After(time.Millisecond * 300):
			startHomePage() // only start home page when everything is ready.
		}
	}()

	if err := http.ListenAndServe(":"+strconv.Itoa(*Port), nil); err != nil {
		ch <- 1
		log.Fatal(err)
	}
}

type indexHandle struct{}

func (h *indexHandle) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	tmpl := template.New("index")
	indexFile, err := f.ReadFile("static/index.html")
	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = tmpl.Parse(string(indexFile))
	if err != nil {
		panic(err)
	}

	tmpl.Execute(writer, nil)
}

// initGoToolPath find the path for pprof and trace binaries.
func initGoToolPath() {
	toolPath := strings.TrimRight(strings.Replace(os.Getenv("GOROOT"), `\`, `/`, -1), "/") + "/pkg/tool/"
	goToolPath := toolPath + runtime.GOOS + "_" + runtime.GOARCH

	filepath.Walk(goToolPath, func(path string, info fs.FileInfo, err error) error {
		if strings.HasPrefix(info.Name(), "pprof") {
			pprofExePath = path
		} else if strings.HasPrefix(info.Name(), "trace") {
			traceExePath = path
		}
		return nil
	})

	if pprofExePath == "" {
		log.Fatalf("pprof binary not found in %v\n", goToolPath)
	}

	if traceExePath == "" {
		log.Fatalf("trace binary not found in %v\n", goToolPath)
	}
}

func welcome() {
	// print logo
	fmt.Println("\n  _____        ____   __                __                 __\n / ___/ ___   / __/  / /  ___    ___   / /  ___   ____ ___/ /\n/ (_ / / _ \\ _\\ \\   / _ \\/ -_)  / _ \\ / _ \\/ -_) / __// _  / \n\\___/  \\___//___/  /_//_/\\__/  / .__//_//_/\\__/ /_/   \\_,_/  \n                              /_/                            ")
	fmt.Println("-------------------------------------------------------------")
	fmt.Println("Welcome to GoShepherd!")
}

func startHomePage() {
	homepage := fmt.Sprintf("http://%s:%v", currIP, strconv.Itoa(*Port))
	commands := map[string]string{
		"windows": "explorer",
		"darwin":  "open",
		"linux":   "xdg-open",
	}

	explorer, ok := commands[runtime.GOOS]
	if !ok {
		fmt.Println("Please visit home page manually: ", homepage)
		return
	}
	fmt.Println("Opening home page: ", homepage)

	exec.Command(explorer, homepage).Run()
}

func getClientIp() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}

	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}

	return "127.0.0.1"
}
