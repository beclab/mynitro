package handler

func makeFileExistDialogHTML(fileName string) string {
	html := `
			<!DOCTYPE html>
			<html>
			<head>
				<title>Download</title>
				<script>
					window.onload = function() {
						alert("Model '` + fileName + `' already downloaded!");
						window.history.back(); // 返回上一页
					};
				</script>
			</head>
			<body>
			</body>
			</html>
		`
	return html
}

func makeProgressorHTML(taskID string, fileName string) string {
	html := `
		<!DOCTYPE html>
		<html>
		<head>
			<title>Download</title>
			<style>
				body {
					background-image: url("` + prefix + `/static/download.jpeg");
					background-size: cover;
					display: flex;
					justify-content: center;
					align-items: center;
					height: 100vh;
				}

				.progressContainer {
					text-align: center;
				}

				.progressTitle {
					font-size: 24px;
					font-weight: bold;
					color: #ffffff;
					margin-bottom: 20px;
				}

				.progress {
					width: 300px;
					height: 20px;
					border: 1px solid #ccc;
					border-radius: 5px;
					overflow: hidden;
					display: inline-block;
				}

				.progressBar {
					width: 0;
					height: 100%;
					background-color: #4CAF50;
				}

				.progressValue {
					font-size: 18px;
					font-weight: bold;
					color: #ffffff;
					margin-top: 10px;
				}
			</style>
			<script>
				function updateProgress() {
					// 发送Ajax请求获取任务进度
					var xhr = new XMLHttpRequest();
					xhr.open("GET", "` + prefix + `/progress?id=` + taskID + `", true);
					xhr.onreadystatechange = function() {
						if (xhr.readyState === 4 && xhr.status === 200) {
							var progress = JSON.parse(xhr.responseText).progress;
							document.getElementById("progressBar").style.width = progress + "%";
							document.getElementById("progressValue").textContent = progress.toFixed(2) + "%";
							if (progress === -1 || progress === 100) {
								clearInterval(interval);
							}
						}
					};
					xhr.send();
				}

				var interval = setInterval(updateProgress, 1000);
			</script>
		</head>
		<body>
			<div class="progressContainer">
				<h1 class="progressTitle">Downloading Model '` + fileName + `'</h1>
				<div class="progress">
					<div id="progressBar" class="progressBar"></div>
				</div>
				<div id="progressValue" class="progressValue"></div>
			</div>
		</body>
		</html>
	`
	return html
}
