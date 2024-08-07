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

var NitroPid = 0
var DifyHost = os.Getenv("DIFY_HOST")
var LLMHost = os.Getenv("LLM_HOST")

func setDifyModel(r *http.Request) (int, string, error) {
	url := DifyHost + "/console/api/workspaces/current/model-providers/openai_api_compatible/models"

	cValue := os.Getenv("C_VALUE")

	if cValue == "" {
		cValue = "1024" // 默认值为 4096
	}

	// 构建请求体数据
	payload := map[string]interface{}{
		"model":      "nitro",
		"model_type": "llm",
		"credentials": map[string]interface{}{
			"mode":                    "chat",
			"context_size":            cValue, // "4096",
			"max_tokens_to_sample":    "1024",
			"function_calling_type":   "no_call",
			"stream_function_calling": "not_supported",
			"vision_support":          "no_support",
			"stream_mode_delimiter":   "\\n\\n",
			"endpoint_url":            LLMHost + "/nitro/model_server/v1", // "http://127.0.0.1:3928/v1",
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

func HandleLoad(w http.ResponseWriter, r *http.Request) {
	if runningType == "WASM" {
		fmt.Fprintf(w, "WASM LLM Model Running! Can't run two LLM models at the same time. Please WASM Stop first!")
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
	if task.Type != "whisper" {
		var llamaModelPath string
		cValue := os.Getenv("C_VALUE")
		nglValue := os.Getenv("NGL_VALUE")

		if cValue == "" {
			cValue = "1024" // 默认值为 4096
		}

		if nglValue == "" {
			nglValue = "0" // 默认值为 20
		}

		unloadCmd := exec.Command("curl", "http://localhost:3928/inferences/llamacpp/unloadmodel")
		err := unloadCmd.Run()
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
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
		llamaModelPath = "../../model/" + option + "/" + fileName
		fmt.Println(llamaModelPath)

		curlCmd := exec.Command("curl", "http://localhost:3928/inferences/llamacpp/loadmodel",
			"-H", "Content-Type: application/json",
			"-d", fmt.Sprintf(`{
				"llama_model_path": "%s",
				"ctx_len": %s,
				"ngl": %s
			}`, llamaModelPath, cValue, nglValue))
		fmt.Print(curlCmd)

		output, err := curlCmd.Output()
		fmt.Println(output)
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
			modelStatus, setResp, err = setDifyModel(r)
			time.Sleep(1 * time.Second)

			if modelStatus == 404 {
				fmt.Println("Difyfusion not installed!")
				break
			}
		}

		runningType = "Nitro"

		// 返回结果
		fmt.Fprintf(w, "Nitro Load option: %s\n", option)
		fmt.Fprintf(w, "CURL output:\n%s\n", output)
		if err != nil {
			fmt.Println("Dify model set failed. Please retry or manually set it.")
			fmt.Fprintf(w, "Dify model set failed. Please retry or manually set it. Rsep body: %s\n", setResp)
		} else {
			fmt.Println("Dify model set successfully!")
			fmt.Fprintf(w, "Dify model set successfully! Resp body: %s\n", setResp)
		}
	} else {
		var modelPath string

		task, _ = downloadTasks[option]
		fileName := task.FileName
		modelPath = "../../model/" + option + "/" + fileName
		fmt.Println(modelPath)

		curlCmd := exec.Command("curl", "http://localhost:3928/v1/audio/load_model",
			"-H", "Content-Type: application/json",
			"-d", fmt.Sprintf(`{
		"model_path": "%s",
		"model_id": "whisper.cpp"
	}`, modelPath))
		fmt.Print(curlCmd)

		output, err := curlCmd.Output()
		fmt.Println(output)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// 更新任务状态为 "running"
		downloadTasksLock.Lock()
		task.Status = "running"
		downloadTasksLock.Unlock()

		// 返回结果
		fmt.Fprintf(w, "Load whisper model: %s\n", option)
		fmt.Fprintf(w, "CURL output:\n%s\n", output)
	}
}
