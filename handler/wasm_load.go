package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"time"
)

var WASMPid = 0

func StartWASM(wasmCmd []string) error {
	//cmd := exec.Command("./nitro", "1", "0.0.0.0", "3928")
	cmd := exec.Command(wasmCmd[0], wasmCmd[1:]...)
	cmd.Dir = "wasm" // 设置工作目录为 nitro 文件夹下

	wd, err := os.Getwd()
	if err != nil {
		fmt.Println("Failed to get current working directory:", err)
		return err
	}

	fmt.Println("Current working directory:", wd)

	// 创建日志文件
	logFile, err := os.Create("wasm.log")
	if err != nil {
		fmt.Println("Failed to create log file:", err)
		return err
	}
	defer logFile.Close()

	// 将标准输出和标准错误输出导入日志文件
	cmd.Stdout = logFile
	cmd.Stderr = logFile

	// 启动命令
	err = cmd.Start()
	if err != nil {
		fmt.Println("Failed to start WASM:", err)
		return err
	} else {
		fmt.Println("WASM started successfully.")
	}

	// 记录命令的 PID
	WASMPid = cmd.Process.Pid
	fmt.Println("WASM started successfully. PID:", WASMPid)
	return nil
}

func KillProcess(pid int) error {
	// 查找进程对象
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process: %v", err)
	}

	// 终止进程
	err = process.Kill()
	if err != nil {
		return fmt.Errorf("failed to kill process: %v", err)
	}

	return nil
}

func setWasmDifyModel(r *http.Request) (int, string, error) {
	url := "http://dify/console/api/workspaces/current/model-providers/openai_api_compatible/models"

	cValue := os.Getenv("C_VALUE")

	if cValue == "" {
		cValue = "1024" // 默认值为 4096
	}

	// 构建请求体数据
	payload := map[string]interface{}{
		"model":      "wasm",
		"model_type": "llm",
		"credentials": map[string]interface{}{
			"mode":                    "chat",
			"context_size":            cValue, // "4096",
			"max_tokens_to_sample":    "1024",
			"function_calling_type":   "no_call",
			"stream_function_calling": "not_supported",
			"vision_support":          "no_support",
			"stream_mode_delimiter":   "\\n\\n",
			"endpoint_url":            "http://127.0.0.1:8081/v1",
			"api_key":                 "[__HIDDEN__]",
		},
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("JSON marshal error:", err)
		return 0, "", err
	}

	// 创建 HTTP 请求
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		fmt.Println("HTTP request creation error:", err)
		return 0, "", err
	}
	req.Header = r.Header
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	// 发送 HTTP 请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("HTTP request error:", err)
		return 0, "", err
	}
	defer resp.Body.Close()

	// 读取响应内容
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Read response error:", err)
		return 0, "", err
	}

	// 打印响应结果
	fmt.Println(string(body))
	return resp.StatusCode, string(body), nil
}

func HandleWASMLoad(w http.ResponseWriter, r *http.Request) {
	if runningType == "Nitro" {
		fmt.Fprintf(w, "Nitro LLM Model Running! Can't run two llm models at the same time. Please Nitro Stop first!")
		return
	}

	option := r.URL.Query().Get("option")

	downloadTasksLock.RLock()
	task, exists := downloadTasks[option]
	downloadTasksLock.RUnlock()

	if !exists || (task.Status != "installed" && task.Status != "running") {
		fmt.Fprint(w, "Model not installed!")
		return
	}

	fmt.Print(task.Type)
	if task.Type == "whisper" {
		fmt.Fprintf(w, "Whisper model can't be loaded by WASM for the time being. Please load it by Nitro Load.")
		return
	}

	var llamaModelPath string
	cValue := os.Getenv("C_VALUE")
	nglValue := os.Getenv("NGL_VALUE")

	if cValue == "" {
		cValue = "1024" // 默认值为 4096
	}

	if nglValue == "" {
		nglValue = "0" // 默认值为 20
	}

	//unloadCmd := exec.Command("curl", "http://localhost:3928/inferences/llamacpp/unloadmodel")
	//err := unloadCmd.Run()
	//if err != nil {
	//	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	//	return
	//}
	if WASMPid != 0 {
		err := KillProcess(WASMPid)
		if err != nil {
			fmt.Println("Failed to kill process:", err)
			fmt.Fprintf(w, "Failed to kill process:%s", err)
			return
		} else {
			fmt.Println("Process killed successfully.")
		}
	}

	// 将所有任务状态为 "running" 的任务状态置为 "installed"
	downloadTasksLock.Lock()
	for _, t := range downloadTasks {
		if t.Status == "running" && t.Type != "whisper" {
			t.Status = "installed"
		}
	}
	downloadTasksLock.Unlock()

	task, _ = downloadTasks[option]
	fileName := task.FileName
	llamaModelPath = "../model/" + option + "/" + fileName
	fmt.Println(llamaModelPath)

	//curlCmd := exec.Command("curl", "http://localhost:3928/inferences/llamacpp/loadmodel",
	//	"-H", "Content-Type: application/json",
	//	"-d", fmt.Sprintf(`{
	//	"llama_model_path": "%s",
	//	"ctx_len": %s,
	//	"ngl": %s
	//}`, llamaModelPath, cValue, nglValue))
	//fmt.Print(curlCmd)

	//output, err := curlCmd.Output()
	//fmt.Println(output)

	wasmCmd := []string{
		"/root/.wasmedge/bin/wasmedge",
		"--dir",
		".:.",
		"--nn-preload",
		"default:GGML:AUTO:" + llamaModelPath,
		"llama-api-server.wasm",
		"--prompt-template",
		"llama-2-chat",
		"--model-name",
		"wasm",
		"--socket-addr",
		"0.0.0.0:8081",
		"-c",
		cValue,
		"-n",
		"1024",
		"-g",
		nglValue,
	}
	err := StartWASM(wasmCmd)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// 更新任务状态为 "running"
	downloadTasksLock.Lock()
	task.Status = "running"
	downloadTasksLock.Unlock()

	// DIFY设置MODEL
	modelStatus := 0
	setResp := ""
	for modelStatus != 200 {
		modelStatus, setResp, err = setWasmDifyModel(r)
		time.Sleep(1 * time.Second)
	}

	runningType = "WASM"

	// 返回结果
	fmt.Fprintf(w, "WASM Load option: %s\n", option)
	//fmt.Fprintf(w, "CURL output:\n%s\n", output)
	if err != nil {
		fmt.Println("Dify model set failed. Please retry or manually set it.")
		fmt.Fprintf(w, "Dify model set failed. Please retry or manually set it. Rsep body: %s\n", setResp)
	} else {
		fmt.Println("Dify model set successfully!")
		fmt.Fprintf(w, "Dify model set successfully! Resp body: %s\n", setResp)
	}
}
