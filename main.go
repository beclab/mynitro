package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"mynitro/handler"
	"net/http"
	"os"
	"os/exec"
)

ARCH := os.Getenv("ARCH")

func StartNitro() {
	cmd := exec.Command("./nitro", "1", "0.0.0.0", "3928")
	cmd.Dir = "nitro/build" // 设置工作目录为 nitro 文件夹下

	// 创建日志文件
	logFile, err := os.Create("nitro.log")
	if err != nil {
		fmt.Println("Failed to create log file:", err)
		return
	}
	defer logFile.Close()

	// 将标准输出和标准错误输出导入日志文件
	cmd.Stdout = logFile
	cmd.Stderr = logFile

	// 启动命令
	err = cmd.Start()
	if err != nil {
		fmt.Println("Failed to start nitro:", err)
	} else {
		fmt.Println("Nitro started successfully.")
	}

	// 记录命令的 PID
	handler.NitroPid = cmd.Process.Pid
	fmt.Println("Nitro started successfully. PID:", handler.NitroPid)
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
		handler.HandleWASMLoad(w, newRequest) // POST /load?option=3B
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
		handler.HandleWASMUnload(w, newRequest) // POST /model/:id/stop
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
	if ARCH == "WASM" {
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
