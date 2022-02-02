package main

import (
	"embed"
	"flag"
	"fmt"
	"html/template"
	"io/fs"
	"io/ioutil"
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
	recoverConfig()
}

func main() {
	welcome()
	initGoToolPath()

	http.Handle("/", &indexHandle{})
	http.Handle("/static/", http.FileServer(http.FS(f)))
	http.Handle("/api", shepherdInst)
	http.Handle("/upload/", &uploadHandler{})

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

const uploadFilePath = "upload"

type uploadHandler struct{}

func (h *uploadHandler) ServeHTTP(writer http.ResponseWriter, r *http.Request) {
	if _, err := os.Stat(uploadFilePath); err != nil {
		err := os.MkdirAll(uploadFilePath, 0755)
		if err != nil {
			log.Println("Error creating directory", err)
			return
		}
	}

	r.ParseForm()                         //解析表单
	rawFile, _, err := r.FormFile("file") //获取文件内容
	if err != nil {
		log.Fatal(err)
	}
	defer rawFile.Close()
	fileHeader := r.MultipartForm.File["file"]
	var proName string
	if len(r.MultipartForm.Value["proName"]) > 0 {
		proName = r.MultipartForm.Value["proName"][0]
	}
	for _, v := range fileHeader {
		var newFileName strings.Builder
		newFileName.WriteString(proName)
		if newFileName.Len() != 0 {
			newFileName.WriteString("_")
		}
		newFileName.WriteString(strconv.Itoa(int(time.Now().Unix())))
		newFileName.WriteString("_")
		newFileName.WriteString(v.Filename)
		newfilePath, err := filepath.Abs(filepath.Join(uploadFilePath, newFileName.String()))
		if err != nil {
			log.Println("abs fielpath err", err)
			writer.Write([]byte("{\"Path\":\"" + "null" + "\"}"))
			return
		}
		fileContext, _ := ioutil.ReadAll(rawFile)
		err = ioutil.WriteFile(newfilePath, fileContext, 0644)
		if err != nil {
			log.Println("WriteFile err:", err)
		}
		writer.Write([]byte("{\"Path\":\"" + newfilePath + "\"}"))
		return
	}
	writer.Write([]byte("{\"Path\":\"" + "null" + "\"}"))
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
