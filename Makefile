.PHONY: build clean

# 默认 build 提示信息
build:
	@echo "需要指定构建编号"
	@echo "1: 列出所有文件夹内文件信息"
	@echo "2: 解析所有文件夹中未解析的文件"
	@echo "3: 上传指定文件夹下的文件"
	@echo "4: 判断是否全部解析完成"
	@echo "5: 列出所有文件夹的ID"
	@echo "6: 停止解析所有文件"
	@echo "例如: make build-1"

# 编号构建逻辑
build-%:
	@case '$*' in \
		1) APP_DIR=internal/fileManagement/listdocs ;; \
		2) APP_DIR=internal/fileManagement/parsedocs ;; \
		3) APP_DIR=internal/fileManagement/uploadocs ;; \
		4) APP_DIR=internal/fileManagement/arealldone ;; \
		5) APP_DIR=internal/datasetManagement/listdatasets ;; \
		6) APP_DIR=internal/fileManagement/stopparsedocs ;; \
		*) echo "无效构建编号" ; exit 1 ;; \
	esac && \
	APP_NAME=$$(basename $$APP_DIR) && \
	echo "Building project: $$APP_NAME from directory: $$APP_DIR" && \
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./script/$$APP_NAME cmd/$$APP_NAME/main.go && \
	upx ./script/$$APP_NAME


test:
	@echo "需要指定测试编号"
	@echo "1: go test -v -run TestAreAllDatasetsDone ragflow_api/internal/fileManagement/listdocs"
	@echo "2: go test -v -run TestParse ragflow_api/internal/fileManagement/parsedocs"
	@echo "3: go test -v -run TestUpload ragflow_api/internal/fileManagement/uploadocs"
	@echo "4: go test -v -run TestList ragflow_api/internal/fileManagement/listdocs"
	@echo "5: go test -v -run TestListSets ragflow_api/internal/datasetManagement/listdatasets"
	@echo "6: go test -v -run TestStopParsing ragflow_api/internal/fileManagement/parsedocs"
	@echo "例如:  make test-1"


test-%:
	@case '$*' in \
		1) cd internal/fileManagement/listdocs && go test -v -run TestAreAllDatasetsDone ;; \
		2) cd internal/fileManagement/parsedocs  && go test -v -run TestParse ;; \
		3) cd internal/fileManagement/uploadocs  && go test -v -run TestUpload ;; \
		4) cd internal/fileManagement/listdocs  && go test -v -run TestList ;; \
		5) cd internal/datasetManagement/listdatasets &&  go test -v -run TestListSets ;;\
		6) cd internal/fileManagement/parsedocs &&  go test -v -run TestStopParsing ;;\
		*) echo "无效选项" ;; \
	esac