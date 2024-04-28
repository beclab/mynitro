package handler

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
)

func InitializeCustomTasks() {
	customModelConfigPath := "custom_model_config"
	modelFolders, err := ioutil.ReadDir(customModelConfigPath)
	if err != nil {
		// 处理读取文件夹错误
		// 可以根据具体需求进行错误处理
		return
	}

	for _, folder := range modelFolders {
		if folder.IsDir() {
			option := folder.Name()

			// 检查是否已存在相应的任务
			if _, ok := downloadTasks[option]; ok {
				continue
			}

			folderPath := filepath.Join("model", option)
			var model map[string]interface{}
			modelFilePath := filepath.Join(customModelConfigPath, option, "model.json")
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
					Type:       "custom", // 设置为 "custom"
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
					Type:       "custom", // 设置为 "custom"
				}
			}
		}
	}
}

func ModelConfigHandler(w http.ResponseWriter, r *http.Request) {
	data := struct {
		BackgroundImageURL string
		Prefix             string
	}{
		BackgroundImageURL: prefix + "/static/model_config.jpg",
		Prefix:             prefix,
	}
	tmpl := template.Must(template.ParseFiles("static/model_config.html"))
	tmpl.Execute(w, data)
}

// 检查模型ID是否已配置
func IsModelConfigured(modelID string) bool {
	for _, task := range downloadTasks {
		if task.ID == modelID {
			return true
		}
	}
	return false
}

func ModelConfigSubmitHandler(w http.ResponseWriter, r *http.Request) {
	modelID := r.FormValue("modelID")
	modelConfig := r.FormValue("modelConfig")

	// 检查模型ID和模型配置是否都非空
	if modelID == "" || modelConfig == "" {
		// 弹出提示框，提示配置未完成
		fmt.Fprintf(w, `<script>alert('Config Not Completed!'); window.history.back();</script>`)
		return
	}

	// 检查模型ID是否已配置
	if IsModelConfigured(modelID) {
		// 弹出提示框，提示模型已配置
		fmt.Fprintf(w, `<script>alert('Model Already Configured!'); window.history.back();</script>`)
		return
	}

	// 创建 custom_model_config 文件夹（如果不存在）
	err := os.MkdirAll("custom_model_config", 0755)
	if err != nil {
		http.Error(w, "Failed to create custom_model_config folder", http.StatusInternalServerError)
		return
	}

	// 创建子文件夹以及 model.json 文件
	err = os.MkdirAll("custom_model_config/"+modelID, 0755)
	if err != nil {
		http.Error(w, "Failed to create model folder", http.StatusInternalServerError)
		return
	}

	file, err := os.Create("custom_model_config/" + modelID + "/model.json")
	if err != nil {
		http.Error(w, "Failed to create model.json file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	_, err = file.WriteString(modelConfig)
	if err != nil {
		http.Error(w, "Failed to write to model.json file", http.StatusInternalServerError)
		return
	}
	InitializeCustomTasks()

	fmt.Fprint(w, "<script>alert('New Custom Model Config Created!'); window.location.href = '"+prefix+"/';</script>")
}

func ModelConfigCancelHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, prefix+"/", http.StatusSeeOther)
}
