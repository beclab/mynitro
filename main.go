package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"math"
	"mynitro/handler"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"time"
)

func convertToBytes(value string) (int64, error) {
	re := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*([KMGT]?B?)`)
	matches := re.FindStringSubmatch(value)
	if matches == nil {
		num, err := strconv.ParseInt(value, 10, 64)
		return num, err
	}

	numStr, unit := matches[1], matches[2]

	num, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0, err
	}

	switch unit {
	case "B":
		return int64(num), nil
	case "K", "KB":
		return int64(num * math.Pow(1024, 1)), nil
	case "M", "MB":
		return int64(num * math.Pow(1024, 2)), nil
	case "G", "GB":
		return int64(num * math.Pow(1024, 3)), nil
	case "T", "TB":
		return int64(num * math.Pow(1024, 4)), nil
	default:
		return 0, fmt.Errorf("unsupported unit: %s", unit)
	}
}

func StartNitro() {
	cmd := exec.Command("./nitro", "1", "0.0.0.0", "3928")
	cmd.Dir = "nitro/build"

	logFile, err := os.OpenFile("nitro.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Failed to open log file:", err)
		return
	}

	cmd.Stdout = logFile
	cmd.Stderr = logFile

	err = cmd.Start()
	if err != nil {
		fmt.Println("Failed to start nitro:", err)
		logFile.Close()
		return
	} else {
		fmt.Println("Nitro started successfully.")
	}

	handler.NitroPid = cmd.Process.Pid
	fmt.Println("Nitro started successfully. PID:", handler.NitroPid)

	logSize := os.Getenv("LOG_SIZE")
	if logSize == "" {
		logSize = "15M"
	}
	logBytes, err := convertToBytes(logSize)
	if err != nil {
		fmt.Println("Parse log file size failed:", err)
		fmt.Println("Will use default log file size limit.")
		logBytes = 15 * 1024 * 1024
	}
	fmt.Println("Nitro log size limit:", logBytes, " Bytes")

	go func() {
		defer logFile.Close() // 在 goroutine 结束时关闭日志文件

		for {
			time.Sleep(10 * time.Second)
			fileInfo, err := logFile.Stat()
			if err != nil {
				fmt.Println("Failed to get log file info:", err)
				continue
			}
			if fileInfo.Size() > logBytes {
				fmt.Println("Clearing log file:", fileInfo.Name())
				_, err = logFile.Seek(0, 0)
				if err != nil {
					fmt.Println("Failed to seek to the beginning of the log file:", err)
					continue
				}
				err = logFile.Truncate(0)
				if err != nil {
					fmt.Println("Failed to truncate the log file:", err)
					continue
				}
			}
		}
	}()
}

func handleModelRequest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	fmt.Println("id=", id)
	action := vars["action"]
	fmt.Println("action=", action)
	if r.Method == http.MethodPost && action == "" {
		action = "install"
	}
	if r.Method == http.MethodDelete && action == "" {
		action = "delete"
	}
	fmt.Println(id, action)

	llmUtil := os.Getenv("LLM_UTIL")

	switch action {
	case "":
		newURL := "/progress?id=" + id
		fmt.Println(newURL)

		// 创建新的 http.Request 对象
		newRequest, err := http.NewRequest(http.MethodGet, newURL, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// 将原始请求的上下文传递给新请求
		newRequest = newRequest.WithContext(r.Context())

		// 调用 HandleProgress 处理函数
		handler.HandleProgress(w, newRequest) // GET /model/:id
	case "load":
		newURL := "/load?option=" + id
		fmt.Println(newURL)

		// 创建新的 http.Request 对象
		newRequest, err := http.NewRequest(http.MethodGet, newURL, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// 将原始请求的上下文传递给新请求
		newRequest = newRequest.WithContext(r.Context())

		// 调用 HandleProgress 处理函数
		if llmUtil == "WASM" {
			handler.HandleWASMLoad(w, newRequest) // POST /load?option=3B
		} else {
			handler.HandleLoad(w, newRequest)
		}
	case "stop":
		newURL := "/unload?option=" + id
		fmt.Println(newURL)

		// 创建新的 http.Request 对象
		newRequest, err := http.NewRequest(http.MethodGet, newURL, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// 将原始请求的上下文传递给新请求
		newRequest = newRequest.WithContext(r.Context())

		// 调用 HandleProgress 处理函数
		if llmUtil == "WASM" {
			handler.HandleWASMUnload(w, newRequest) // POST /model/:id/stop
		} else {
			handler.HandleUnload(w, newRequest)
		}
	case "install":
		newURL := "/download?progressor=false&option=" + id
		fmt.Println(newURL)

		// 创建新的 http.Request 对象
		newRequest, err := http.NewRequest(http.MethodGet, newURL, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// 将原始请求的上下文传递给新请求
		newRequest = newRequest.WithContext(r.Context())

		// 调用 HandleProgress 处理函数
		handler.HandleDownload(w, newRequest) // POST /model/:id
	case "delete":
		newURL := "/delete?option=" + id
		fmt.Println(newURL)

		// 创建新的 http.Request 对象
		newRequest, err := http.NewRequest(http.MethodGet, newURL, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// 将原始请求的上下文传递给新请求
		newRequest = newRequest.WithContext(r.Context())

		// 调用 HandleProgress 处理函数
		handler.HandleDelete(w, newRequest) // DELETE /model/:id
	default:
		http.NotFound(w, r)
		return
	}
}

func main() {
	// 检查 nitro 文件是否存在
	if _, err := os.Stat("nitro/build/nitro"); err == nil {
		// 如果存在，启动 nitro 服务
		StartNitro()
	} else {
		fmt.Println("Nitro executable not found in the 'nitro' folder.")
	}

	// 检查 model 文件夹是否存在，如果不存在则创建
	if _, err := os.Stat("model"); os.IsNotExist(err) {
		err := os.Mkdir("model", 0755)
		if err != nil {
			log.Fatal("Failed to create 'model' directory:", err)
		}
	}
	handler.InitializeTasks()
	handler.InitializeWhisperTasks()
	handler.InitializeCustomTasks()

	// 创建路由器
	r := mux.NewRouter()

	// 注册路由
	r.HandleFunc("/model/{id}", handleModelRequest)
	r.HandleFunc("/model/{id}/{action}", handleModelRequest)
	r.HandleFunc("/model", handler.HandleProgress)
	r.HandleFunc("/", handler.HandleIndex)
	r.HandleFunc("/running_type", handler.HandleRunningType)
	r.HandleFunc("/download", handler.HandleDownload)
	r.HandleFunc("/nitro_load", handler.HandleLoad)
	r.HandleFunc("/nitro_unload", handler.HandleUnload)
	r.HandleFunc("/wasm_load", handler.HandleWASMLoad)
	r.HandleFunc("/wasm_unload", handler.HandleWASMUnload)
	llmUtil := os.Getenv("LLM_UTIL")
	if llmUtil == "WASM" {
		r.HandleFunc("/load", handler.HandleWASMLoad)
		r.HandleFunc("/unload", handler.HandleWASMUnload)
	} else {
		r.HandleFunc("/load", handler.HandleLoad)
		r.HandleFunc("/unload", handler.HandleUnload)
	}
	r.HandleFunc("/progress", handler.HandleProgress)
	r.HandleFunc("/delete", handler.HandleDelete)
	r.HandleFunc("/model_config", handler.ModelConfigHandler)
	r.HandleFunc("/model_config/submit", handler.ModelConfigSubmitHandler)
	r.HandleFunc("/model_config/cancel", handler.ModelConfigCancelHandler)
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	r.HandleFunc("/restart", handler.RestartHandler)
	r.HandleFunc("/view_nitro_log", handler.ViewNitroLogHandler)
	r.HandleFunc("/nitro_log", handler.NitroLogHandler)
	r.HandleFunc("/view_wasm_log", handler.ViewWASMLogHandler)
	r.HandleFunc("/wasm_log", handler.WASMLogHandler)

	// 创建服务器并指定路由器
	server := &http.Server{
		Addr:    "0.0.0.0:3900",
		Handler: r,
	}
	log.Println("Server started on http://0.0.0.0:3900")
	log.Fatal(server.ListenAndServe())
}
