package handler

import (
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
)

func ViewNitroLogHandler(w http.ResponseWriter, r *http.Request) {
	data := struct {
		LogURL string
	}{
		LogURL: prefix + "/nitro_log",
	}
	tmpl := template.Must(template.ParseFiles("static/log.html"))
	tmpl.Execute(w, data)
}

func NitroLogHandler(w http.ResponseWriter, r *http.Request) {
	// 打开日志文件
	file, err := ioutil.ReadFile("nitro.log")
	if err != nil {
		log.Println("Failed to open log file:", err)
		http.Error(w, "Failed to open log file", http.StatusInternalServerError)
		return
	}

	// 设置响应的Content-Type为text/plain
	w.Header().Set("Content-Type", "text/plain")

	// 将日志文件内容写入响应
	_, err = w.Write(file)
	if err != nil {
		log.Println("Failed to write log file to response:", err)
		http.Error(w, "Failed to write log file to response", http.StatusInternalServerError)
		return
	}
}

func ViewWASMLogHandler(w http.ResponseWriter, r *http.Request) {
	data := struct {
		LogURL string
	}{
		LogURL: prefix + "/wasm_log",
	}
	tmpl := template.Must(template.ParseFiles("static/log.html"))
	tmpl.Execute(w, data)
}

func WASMLogHandler(w http.ResponseWriter, r *http.Request) {
	// 打开日志文件
	file, err := ioutil.ReadFile("wasm.log")
	if err != nil {
		log.Println("Failed to open log file:", err)
		http.Error(w, "Failed to open log file", http.StatusInternalServerError)
		return
	}

	// 设置响应的Content-Type为text/plain
	w.Header().Set("Content-Type", "text/plain")

	// 将日志文件内容写入响应
	_, err = w.Write(file)
	if err != nil {
		log.Println("Failed to write log file to response:", err)
		http.Error(w, "Failed to write log file to response", http.StatusInternalServerError)
		return
	}
}
