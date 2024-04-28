package handler

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
)

func InitializeWhisperTasks() {
	whisperModelConfigPath := "whisper_model_config"
	whisperModelFolders, err := ioutil.ReadDir(whisperModelConfigPath)
	if err != nil {
		// 处理读取文件夹错误
		// 可以根据具体需求进行错误处理
		return
	}

	for _, folder := range whisperModelFolders {
		if folder.IsDir() {
			option := folder.Name()
			//options = append(options, option)
			folderPath := filepath.Join("model", option)
			var model map[string]interface{}
			modelFilePath := filepath.Join(whisperModelConfigPath, option, "model.json")
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
					Type:       "whisper", // 默认值为 "default"
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
					Type:       "whisper", // 默认值为 "default"
				}
			}
		}
	}
}
