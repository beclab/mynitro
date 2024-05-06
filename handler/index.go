package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
)

var prefix = os.Getenv("PREFIX")
var runningType = ""

func HandleIndex(w http.ResponseWriter, r *http.Request) {
	// 获取 URL 参数中的选项值
	option := r.URL.Query().Get("option")
	var difyURL string
	if prefix != "" {
		difyURL = prefix + "/dify"
	} else {
		difyURL = "http://localhost"
	}
	// 显示选项页面
	html := `
		<!DOCTYPE html>
		<html>
		<head>
			<title>Options</title>
			<style>
				body {
					background-image: url("` + prefix + `/static/background.jpg");
					background-size: cover;
					display: flex;
					justify-content: center;
					align-items: center;
					height: 100vh;
					margin: 0;
				}
				.container {
					background-color: rgba(255, 255, 255, 0.8);
					padding: 20px;
					text-align: center;
				}
				.title {
					font-size: 24px;
					font-weight: bold;
					margin-bottom: 20px;
					text-align: center;
				}
				.options {
					list-style-type: none;
					text-align: left;
					padding-left: 0;
				}
				.options li {
					margin-bottom: 10px;
				}
				.options input[type="radio"] {
					margin-right: 5px;
				}
				.buttons {
					display: flex;
					justify-content: center;
					margin-top: 10px;
				}
				.buttons form {
					margin: 0 10px;
				}
				.buttons input[type="submit"] {
					padding: 10px 20px;
					border: none;
					border-radius: 5px;
					background-color: #4CAF50;
					color: white;
					font-size: 16px;
					cursor: pointer;
					transition: background-color 0.3s;
				}
				.buttons input[type="submit"]:hover {
					background-color: #45a049;
				}
				.buttons input[type="submit"]:not(:last-child) {
					margin-right: 10px;
				}
				.button-wrapper {
        			margin-right: 10px;
    			}
				.button {
					background-color: #007bff;
					border: none;
					color: white;
					padding: 8px 16px;
					text-align: center;
					text-decoration: none;
					display: inline-block;
					font-size: 16px;
					margin-top: 0px;
					border-radius: 5px;
					cursor: pointer;
				}
				.button:hover {
					background-color: #0056b3;
				}
				.button2 {
  					background-color: #ff4500;
  					border: none;
  					color: white;
  					padding: 8px 16px;
  					text-align: center;
  					text-decoration: none;
  					display: inline-block;
  					font-size: 16px;
  					margin-top: 0px;
  					border-radius: 5px;
  					cursor: pointer;
				}

				.button2:hover {
  					background-color: #b32d00;
				}
				.view-log-button {
					background-color: #0077be;
					color: white;
					border: none;
    				padding: 10px 20px;
    				font-size: 16px;
    				position: fixed;
    				top: 20px;
                    right: 250px; /* 调整按钮位置，使其向左移动一些 */
    				border-radius: 4px;
    				box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
    				cursor: pointer;
    				text-decoration: none;
  				}

  				.view-log-button:hover {
					background-color: #00568c;
  				}

				.view-log-button2 {
					background-color: #0077be;
					color: white;
					border: none;
    				padding: 10px 20px;
    				font-size: 16px;
    				position: fixed;
    				top: 20px;
                    right: 120px; /* 调整按钮位置，使其向左移动一些 */
    				border-radius: 4px;
    				box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
    				cursor: pointer;
    				text-decoration: none;
  				}

  				.view-log-button2:hover {
					background-color: #00568c;
  				}
				.restart-button {
    				background-color: #6c4f99;
    				color: white;
    				border: none;
    				padding: 10px 20px;
    				font-size: 16px;
    				position: fixed;
    				top: 20px;
    				right: 20px;
    				border-radius: 4px;
    				box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
    				cursor: pointer;
    				text-decoration: none;
				}

				.restart-button:hover {
    				background-color: #542f73; /* 中等深度的紫色 */
				}
                .left-align {
                    text-align: left;
                }
				.gray {
  					color: #777777; /* 灰色 */
				}

				.red {
  					color: #ff0000; /* 红色 */
				}

				.green {
  					color: #008000; /* 绿色 */
				}

				.blue {
  					color: #0000ff; /* 蓝色 */
				}
			</style>
			<script>
				function validateForm() {
					var selectedOption = document.querySelector('input[name="option"]:checked');
					if (!selectedOption) {
						alert("Please choose a model first!");
						return false;
					}
				}

				const viewLogButton = document.getElementById('viewLogButton');
  				viewLogButton.addEventListener('click', function() {
    				// 在按钮点击时打开一个新窗口并加载日志页面
    				window.open('static/log.html', '_blank');
  				});
			</script>
		</head>
		<body>
			<a class="view-log-button" href="` + prefix + `/view_nitro_log" target="_blank">Nitro Log</a>
			<a class="view-log-button2" href="` + prefix + `/view_wasm_log" target="_blank">WASM Log</a>
			<a class="restart-button" href="` + prefix + `/restart">Restart</a>
			<div class="container">
				<h1 class="title">Choose Your Model</h1>
				<form action="` + prefix + `/download" method="get" onsubmit="return validateForm();">
					<div class="left-align"><b>Current LLM Running By: ` + runningType + `</b></div>
					<br>
					<div class="left-align"><b>Usable Builtin Model IDs:</b></div>
					<ul class="options">` + GenerateOptions(option, "default") + `</ul>
					<br>
					<div class="left-align"><b>Usable Whisper Model IDs:</b></div>
					<ul class="options">` + GenerateOptions(option, "whisper") + `</ul>
					<br>
					<div class="left-align"><b>Usable Custom Model IDs:</b></div>
					<ul class="options">` + GenerateOptions(option, "custom") + `</ul>
					<div class="buttons">
						<div class="button-wrapper">
        					<a class="button2" href="` + prefix + `/model_config">Add New Model</a>
    					</div>
						&nbsp;
						<div class="button-wrapper">
        					<input type="submit" value="Install">
    					</div>
						&nbsp;
    					<div class="button-wrapper">
        					<input type="submit" formaction="` + prefix + `/delete" formmethod="get" formnovalidate value="Delete">
    					</div>
						&nbsp;
    					<div class="button-wrapper">
        					<a class="button" href="` + difyURL + `">Go to dify! →</a>
    					</div>
					</div>
					<div class="buttons">
						<input type="submit" id="nitroLoadButton" formaction="` + prefix + `/nitro_load" formmethod="get" formnovalidate value="Nitro Load">
						<input type="submit" id="nitroUnloadButton" formaction="` + prefix + `/nitro_unload" formmethod="get" formnovalidate value="Nitro Stop">
						<input type="submit" id="wasmLoadButton" formaction="` + prefix + `/load" formmethod="get" formnovalidate value="WASM Load">
						<input type="submit" id="wasmUnloadButton" formaction="` + prefix + `/unload" formmethod="get" formnovalidate value="WASM Stop">
					</div>
				</form>
			</div>
		</body>
		</html>
	`
	fmt.Fprint(w, html)
}

//func GenerateOptions(selectedOption string, model_type string) string {
//	var optionsHTML strings.Builder
//
//	// 将任务ID排序
//	var sortedOptionIDs []string
//	for optionID := range downloadTasks {
//		sortedOptionIDs = append(sortedOptionIDs, optionID)
//	}
//	sort.Strings(sortedOptionIDs)
//
//	// 按排序后的顺序生成选项
//	count := 0
//	for _, optionID := range sortedOptionIDs {
//		option := downloadTasks[optionID]
//		if option.Type == model_type {
//			count += 1
//			optionsHTML.WriteString(fmt.Sprintf(`<li><label><input type="radio" name="option" value="%s" %s>%s</label></li>`,
//				option.ID,
//				isChecked(option.ID, selectedOption),
//				option.ID))
//		}
//	}
//	if count == 0 {
//		optionsHTML.WriteString("No model usable for the time being...")
//	}
//	return optionsHTML.String()
//}

func GenerateOptions(selectedOption string, model_type string) string {
	var optionsHTML strings.Builder

	// 将任务ID排序
	var sortedOptionIDs []string
	for optionID := range downloadTasks {
		sortedOptionIDs = append(sortedOptionIDs, optionID)
	}
	sort.Strings(sortedOptionIDs)

	// 按排序后的顺序生成选项
	count := 0
	for _, optionID := range sortedOptionIDs {
		option := downloadTasks[optionID]
		if option.Type == model_type {
			count++
			statusClass := ""
			statusText := ""

			switch option.Status {
			case "no_installed":
				statusClass = "gray"
				statusText = "no_installed"
			case "installing":
				statusClass = "red"
				statusText = fmt.Sprintf("installing: <span class='progress' id='progress-%s'>%.2f</span>%%", option.ID, option.Progress)
			case "installed":
				statusClass = "green"
				statusText = "installed"
			case "running":
				statusClass = "blue"
				statusText = "running"
			}

			idClass := "default"
			switch option.Status {
			case "no_installed":
				idClass = "gray"
			case "installing":
				idClass = "red"
			case "installed":
				idClass = "green"
			case "running":
				idClass = "blue"
			}

			optionsHTML.WriteString(fmt.Sprintf(`<li><label><input type="radio" name="option" value="%s" %s><span class="%s">%s</span>&nbsp;&nbsp;&nbsp;&nbsp;<span class="%s">(%s)</span></label></li>`,
				option.ID,
				isChecked(option.ID, selectedOption),
				idClass,
				option.ID,
				statusClass,
				statusText))
		}
	}
	if count == 0 {
		optionsHTML.WriteString("No model usable for the time being...")
	}

	// 添加 JavaScript 代码，定时更新百分数
	optionsHTML.WriteString(`
        <script>
            setInterval(function() {
    			var progressElements = document.getElementsByClassName('progress');
    			for (var i = 0; i < progressElements.length; i++) {
        			(function() {
            			var element = progressElements[i];
            			var optionID = element.id.split('-').slice(1).join('-'); // 提取选项ID
            			var url = "` + prefix + `/progress?id=" + optionID; // 替换为您的实际接口URL
            			var nextSibling = element.nextElementSibling; // 获取当前选项的下一个兄弟元素
            			fetch(url)
                			.then(response => response.json())
                			.then(data => {
                    			var newProgress = data.progress; // 根据实际接口返回的数据结构获取进度值
                    			element.textContent = newProgress.toFixed(2);
                    			if (newProgress >= 100) {
                        			nextSibling.textContent = "installed";
                        			nextSibling.classList.add("green");
                    			} else {
                        			nextSibling.textContent = ""; // 清空下一个兄弟元素的内容
                        			nextSibling.classList.remove("green"); // 移除可能存在的"green"类
                    			}
                			})
                			.catch(error => {
                    			console.error('Error:', error);
                			});
					})();
				}
			}, 5000); // 5000毫秒 = 5秒
    	</script>
    `)

	return optionsHTML.String()
}

// 辅助函数，如果选项与当前选中的选项匹配，则返回 "checked"
func isChecked(option, selectedOption string) string {
	if option == selectedOption {
		return "checked"
	}
	return ""
}

// 响应结构体
type Response struct {
	RunningType string `json:"runningType"`
}

func HandleRunningType(w http.ResponseWriter, r *http.Request) {
	// 创建响应结构体
	response := Response{RunningType: runningType}

	// 将响应序列化为 JSON
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 设置响应头
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// 发送响应
	w.Write(jsonResponse)
}
