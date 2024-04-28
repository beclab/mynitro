package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sync"
)

//type ProgressWriter struct {
//	Writer   io.Writer
//	Total    int64
//	Progress func(progress float64)
//	FileName string
//}

//func (pw *ProgressWriter) Write(p []byte) (int, error) {
//	n, err := pw.Writer.Write(p)
//	//fmt.Println(n)
//	if err != nil {
//		return n, err
//	}
//
//	pw.Progress(float64(n) / float64(pw.Total) * 100)
//
//	return n, nil
//}

// 20240423

//type ProgressWriter struct {
//	Writer    io.Writer
//	Total     int64
//	Progress  func(progress float64)
//	FileName  string
//	Written   int64
//	Completed bool
//}
//
//func (pw *ProgressWriter) Write(p []byte) (int, error) {
//	n, err := pw.Writer.Write(p)
//	if err != nil {
//		return n, err
//	}
//
//	pw.Written += int64(n)
//
//	// 计算进度
//	progress := float64(pw.Written) / float64(pw.Total) * 100
//	pw.Progress(progress)
//
//	// 最后一次写入时将进度手动设置为100
//	if pw.Written == pw.Total {
//		pw.Progress(100)
//		pw.Completed = true
//	}
//
//	return n, nil
//}

type ProgressWriter struct {
	Writer    io.Writer
	Total     int64
	FileName  string
	Progress  func(float64)
	Written   int64
	Completed bool
	Option    string
}

func NewProgressWriter(writer io.Writer, total int64, fileName string, option string) *ProgressWriter {
	return &ProgressWriter{
		Writer:    writer,
		Total:     total,
		FileName:  fileName,
		Option:    option,
		Completed: false,
		Written:   0,
	}
}

func (pw *ProgressWriter) Write(p []byte) (int, error) {
	if pw.Completed && pw.Written == 0 {
		return 0, nil // 不允许写入数据，返回写入字节数为 0
	}
	
	n, err := pw.Writer.Write(p)
	if err != nil {
		return n, err
	}

	pw.Written += int64(n)

	// 检查任务状态
	downloadTasksLock.RLock()
	task := downloadTasks[pw.Option]
	downloadTasksLock.RUnlock()

	if pw.Completed {
		if task.Status == "installed" {
			pw.Progress(100)
		} else {
			pw.Progress(0)
		}
	} else {
		if task.Status == "installing" {
			// 更新下载进度
			if pw.Total > 0 && pw.Progress != nil {
				currentProgress := float64(pw.Written) / float64(pw.Total) * 100
				pw.Progress(currentProgress)
			}
			// 最后一次写入时将进度手动设置为100
			if pw.Written == pw.Total {
				pw.Progress(100)
				pw.Completed = true
			}
		} else if task.Status == "no_installed" {
			// 锁定下载进度为0
			if !pw.Completed {
				pw.Progress(0)
				pw.Written = 0
				pw.Completed = true
			}
		}
	}

	return n, nil
}

// Reset 重置进度信息
func (pw *ProgressWriter) Reset() {
	pw.Written = 0
	pw.Completed = false
}

func toJSON(data interface{}) string {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return ""
	}
	return string(jsonData)
}

// 添加全局锁，用于保护下载任务的并发访问
var downloadTasksLock sync.RWMutex

// 添加全局变量，用于存储下载任务及其进度
var downloadTasks = make(map[string]*DownloadTask)

// DownloadTask 结构体
type DownloadTask struct {
	ID         string                 `json:"id"`
	FileName   string                 `json:"file_name"`
	Progress   float64                `json:"progress"`
	Status     string                 `json:"status"`
	FolderPath string                 `json:"folder_path"`
	Model      map[string]interface{} `json:"model"`
	Type       string                 `json:"type"` // 新增的 Type 字段
}

// SSEWriter 结构体
type SSEWriter struct {
	http.ResponseWriter
}

func (w *SSEWriter) Write(data []byte) (int, error) {
	_, err := w.ResponseWriter.Write([]byte("data: " + string(data) + "\n\n"))
	if err != nil {
		return 0, err
	}
	w.ResponseWriter.(http.Flusher).Flush()
	return len(data), nil
}

func sendSSE(w http.ResponseWriter, event string, data interface{}) {
	// 设置 SSE 响应头
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// 构建 SSE 数据
	sseData := fmt.Sprintf("event: %s\n", event)
	sseData += fmt.Sprintf("data: %s\n\n", toJSON(data))
	fmt.Println(sseData)

	// 发送 SSE 数据
	_, _ = w.Write([]byte(sseData))
}

var options []string

func InitializeTasks() {
	modelConfigPath := "model_config"
	modelFolders, err := ioutil.ReadDir(modelConfigPath)
	if err != nil {
		// 处理读取文件夹错误
		// 可以根据具体需求进行错误处理
		return
	}

	for _, folder := range modelFolders {
		if folder.IsDir() {
			option := folder.Name()
			options = append(options, option)
			folderPath := filepath.Join("model", option)
			var model map[string]interface{}
			modelFilePath := filepath.Join(modelConfigPath, option, "model.json")
			modelFile, err := ioutil.ReadFile(modelFilePath)
			if err != nil {
				// 处理读取文件错误
				// 可以根据具体需求进行错误处理
				continue
			}
			err = json.Unmarshal(modelFile, &model)
			if err != nil {
				// 处理解析JSON错误
				// 可以根据具体需求进行错误处理
				continue
			}
			fileName := path.Base(model["source_url"].(string))
			filePath := filepath.Join(folderPath, fileName)
			_, err = os.Stat(filePath)
			if os.IsNotExist(err) {
				// 文件不存在，创建任务
				downloadTasks[option] = &DownloadTask{
					ID:         option,
					FileName:   fileName,
					Progress:   0,
					Status:     "no_installed",
					FolderPath: folderPath,
					Model:      model,
					Type:       "default", // 默认值为 "default"
				}
			} else {
				// 文件存在，创建任务
				downloadTasks[option] = &DownloadTask{
					ID:         option,
					FileName:   fileName,
					Progress:   100,
					Status:     "installed",
					FolderPath: folderPath,
					Model:      model,
					Type:       "default", // 默认值为 "default"
				}
			}
		}
	}
}

func HandleDownload(w http.ResponseWriter, r *http.Request) {
	option := r.URL.Query().Get("option")
	progressor := r.URL.Query().Get("progressor")

	// 根据选项确定要下载的文件名和 URL
	var (
		fileName    string
		downloadURL string
	)

	// 检查文件是否已经存在
	downloadTasksLock.RLock()
	task, _ := downloadTasks[option]
	fileName = task.FileName
	model := task.Model
	model_type := task.Type
	downloadURL = model["source_url"].(string)
	downloadTasksLock.RUnlock()

	if task.Status != "no_installed" {
		// 文件已存在，返回HTML页面，弹出对话框
		html := makeFileExistDialogHTML(task.FileName)
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, html)
		return
	}

	// 创建文件夹路径
	folderPath := filepath.Join("model", option)
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		// 创建文件夹
		err := os.MkdirAll(folderPath, os.ModePerm)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	// 创建 SSEWriter 对象
	// sseWriter := &SSEWriter{ResponseWriter: w}

	//// 创建下载任务
	//downloadTasksLock.Lock()
	//downloadTasks[option] = &DownloadTask{
	//	ID:         option,
	//	FileName:   fileName,
	//	Progress:   0,
	//	Status:     "installing",
	//	FolderPath: folderPath,
	//	Model:      model,
	//	Type:       model_type,
	//}
	//downloadTasksLock.Unlock()

	// 启动下载任务的 Goroutine
	//go func() {
	//// 下载文件
	//resp, err := http.Get(downloadURL)
	//if err != nil {
	//	fmt.Println(err)
	//	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	//	return
	//}
	//defer resp.Body.Close()
	//
	//filePath := filepath.Join(folderPath, fileName)
	//file, err := os.Create(filePath)
	//if err != nil {
	//	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	//	return
	//}
	//defer file.Close()
	//
	//progressWriter := &ProgressWriter{
	//	Writer:   file,
	//	Total:    resp.ContentLength,
	//	Progress: func(progress float64) { updateTaskProgress(option, progress) },
	//	FileName: fileName,
	//}
	//
	//_, err = io.Copy(progressWriter, resp.Body)
	//if err != nil {
	//	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	//	return
	//}
	//
	//// 手动将进度设置为100
	////if !progressWriter.Completed {
	////	updateTaskProgress(option, 100)
	////	progressWriter.Completed = true
	////	progressWriter.Progress(100)
	////}
	//
	//// 下载完成后更新任务状态
	//updateTaskStatus(option, "installed")

	go func() {
		// 创建下载任务
		downloadTasksLock.Lock()
		downloadTasks[option] = &DownloadTask{
			ID:         option,
			FileName:   fileName,
			Progress:   0,
			Status:     "installing",
			FolderPath: folderPath,
			Model:      model,
			Type:       model_type,
		}
		downloadTasksLock.Unlock()

		// 下载文件
		resp, err := http.Get(downloadURL)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		filePath := filepath.Join(folderPath, fileName)
		file, err := os.Create(filePath)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		defer file.Close()

		progressWriter := NewProgressWriter(file, resp.ContentLength, fileName, option)

		// 重置进度信息
		progressWriter.Reset()

		// 检查任务状态
		downloadTasksLock.RLock()
		task := downloadTasks[option]
		downloadTasksLock.RUnlock()

		if task.Status == "installing" {
			progressWriter.Progress = func(progress float64) {
				updateTaskProgress(option, progress)
			}
		} else if task.Status == "no_installed" {
			updateTaskProgress(option, 0)
			progressWriter.Written = 0
			progressWriter.Completed = true
		}
		_, err = io.Copy(progressWriter, resp.Body)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// 下载完成后更新任务状态
		updateTaskStatus(option, "installed")
	}()

	if progressor == "false" {
		// 返回任务信息给前端
		taskInfo := downloadTasks[option]
		sendSSE(w, "task-update", taskInfo)
	} else {
		// 返回HTML页面，显示下载进度条
		html := makeProgressorHTML(option, fileName)
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, html)
	}
}

func updateTaskProgress(option string, progress float64) {
	downloadTasksLock.Lock()
	defer downloadTasksLock.Unlock()

	task := downloadTasks[option]
	if task != nil {
		task.Progress = progress
	}
}

func updateTaskStatus(option, status string) {
	downloadTasksLock.Lock()
	defer downloadTasksLock.Unlock()

	task := downloadTasks[option]
	if task != nil {
		task.Status = status
		if status == "installed" {
			task.Progress = 100
		}
	}
}
