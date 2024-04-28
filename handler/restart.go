package handler

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
)

func RestartHandler(w http.ResponseWriter, r *http.Request) {
	// 处理卸载逻辑，使用选项值"option"
	unloadCmd := exec.Command("curl", "http://localhost:3928/inferences/llamacpp/unloadmodel")
	err := unloadCmd.Run()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// 将所有任务状态为 "running" 的任务状态置为 "installed"
	downloadTasksLock.Lock()
	for _, t := range downloadTasks {
		if t.Status == "running" {
			t.Status = "installed"
		}
	}
	downloadTasksLock.Unlock()

	err = restartProcess()
	if err != nil {
		log.Println("Failed to restart process:", err)
		http.Error(w, "Failed to restart process", http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(w, "Process restarted successfully")
}

func restartProcess() error {
	// 停止之前的进程
	err := stopProcess()
	if err != nil {
		return err
	}

	// 启动新的进程
	err = startProcess()
	if err != nil {
		return err
	}

	return nil
}

func stopProcess() error {
	// 执行命令 "pkill -9 -f nitro" 来杀死进程
	cmd := exec.Command("pkill", "-9", "-f", "nitro")
	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func startProcess() error {
	cmd := exec.Command("./nitro", "1", "0.0.0.0", "3928")
	cmd.Dir = "nitro/build" // 设置工作目录为 nitro 文件夹下

	// 打开日志文件（追加写入模式）
	logFile, err := os.OpenFile("nitro.log", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("Failed to open log file:", err)
		return err
	}
	defer logFile.Close()

	// 将标准输出和标准错误输出导入日志文件
	cmd.Stdout = logFile
	cmd.Stderr = logFile

	// 启动命令
	err = cmd.Start()
	if err != nil {
		fmt.Println("Failed to start nitro:", err)
		return err
	} else {
		fmt.Println("Nitro started successfully.")
		return nil
	}
}
