package handler

import (
	"fmt"
	"net/http"
)

func HandleProgress(w http.ResponseWriter, r *http.Request) {
	// 获取 URL 参数中的任务ID
	taskID := r.URL.Query().Get("id")
	//fmt.Println("taskID=", taskID)

	// 如果没有提供任务ID，则返回所有下载任务及其进度值
	if taskID == "" {
		// 获取所有下载任务的副本
		downloadTasksLock.RLock()
		tasks := make([]*DownloadTask, 0, len(downloadTasks))
		for _, task := range downloadTasks {
			tasks = append(tasks, task)
		}
		downloadTasksLock.RUnlock()

		// 将任务列表转换为JSON格式
		jsonData := toJSON(tasks)

		// 返回JSON响应
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, jsonData)
		return
	}

	// 如果提供了任务ID，则返回特定任务的进度值
	downloadTasksLock.RLock()
	task, exists := downloadTasks[taskID]
	downloadTasksLock.RUnlock()

	if !exists {
		// 任务ID不存在，返回错误响应
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	// 将任务转换为JSON格式
	jsonData := toJSON(task)

	// 返回JSON响应
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, jsonData)
}
