package handler

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

func HandleDelete(w http.ResponseWriter, r *http.Request) {
	option := r.URL.Query().Get("option")

	downloadTasksLock.RLock()
	task, exists := downloadTasks[option]
	downloadTasksLock.RUnlock()

	if !exists || task.Status == "no_installed" {
		fmt.Fprint(w, "Model not installed!")
		return
	}

	if task.Status == "running" {
		fmt.Fprint(w, "Please stop running model first!")
		return
	}

	// 删除文件
	folderPath := filepath.Join("model", option)
	err := os.RemoveAll(folderPath)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// 更新任务状态
	downloadTasksLock.Lock()
	task, exists = downloadTasks[option]
	if exists {
		task.Status = "no_installed"
		task.Progress = 0
	}
	downloadTasksLock.Unlock()

	// 返回成功消息
	message := fmt.Sprintf("Deleted files for option: %s", option)
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, message)
}
