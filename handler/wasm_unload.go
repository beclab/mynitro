package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func unsetWASMDifyModel(r *http.Request) (int, string, error) {
	url := DifyHost + "/console/api/workspaces/current/model-providers/openai_api_compatible/models"

	// 构建请求体数据
	payload := map[string]interface{}{
		"model":      "wasm",
		"model_type": "llm",
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("JSON marshal error:", err)
		return 0, "", err
	}

	// 创建 HTTP 请求
	req, err := http.NewRequest("DELETE", url, bytes.NewBuffer(jsonPayload))
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
		return 404, "", err
	}
	defer resp.Body.Close()

	// 读取响应内容
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Read response error:", err)
		return 0, "", err
	}

	// 打印响应结果
	fmt.Println(resp.StatusCode, string(body))
	return resp.StatusCode, string(body), nil
}

func HandleWASMUnload(w http.ResponseWriter, r *http.Request) {
	option := r.URL.Query().Get("option")

	downloadTasksLock.RLock()
	task, _ := downloadTasks[option]
	downloadTasksLock.RUnlock()

	if task.Type == "whisper" {
		fmt.Fprint(w, "Whisper Model cannot be stopped by wasm!")
		return
	}

	if WASMPid == 0 {
		fmt.Fprint(w, "WASM LLM Model not running!")
		return
	}

	// 处理卸载逻辑，使用选项值"option"
	////unloadCmd := exec.Command("curl", "http://localhost:3928/inferences/llamacpp/unloadmodel")
	////err := unloadCmd.Run()
	//if err != nil {
	//	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	//	return
	//}

	err := KillProcess(WASMPid)
	if err != nil {
		fmt.Println("Failed to kill process:", err)
		fmt.Fprintf(w, "Failed to kill process:%s", err)
		return
	} else {
		fmt.Println("Process killed successfully.")
	}

	// 将所有任务状态为 "running" 的任务状态置为 "installed"
	downloadTasksLock.Lock()
	for _, t := range downloadTasks {
		if t.Status == "running" && t.Type != "whisper" {
			t.Status = "installed"
		}
	}
	downloadTasksLock.Unlock()

	// DIFY设置MODEL
	modelStatus, unsetResp, err := unsetWASMDifyModel(r)

	runningType = ""

	fmt.Fprintf(w, "WASM Unload option: %s\n", option)
	if modelStatus == 404 {
		fmt.Println("Difyfusion not installed!")
	} else if err != nil {
		fmt.Println("Dify model unset failed. Please retry or manually unset it.")
		fmt.Fprintf(w, "Dify model unset failed. Please retry or manually unset it. Rsep body: %s\n", unsetResp)
	} else {
		fmt.Println("Dify model unset successfully!")
		fmt.Fprintf(w, "Dify model unset successfully!")
	}
}
