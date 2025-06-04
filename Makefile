# 默认值：如果没有指定 APP_DIR，默认构建 uploadocs
APP_DIR ?= parsedocs
APP_NAME = $(APP_DIR)

# 映射表：数字 -> 子目录名称
DIR_MAP = \
    1:listdocs \
    2:parsedocs \
    3:uploadocs

# 根据用户输入的数字选择子目录
ifeq ($(APP_DIR),)
    # 如果没有直接指定 APP_DIR，检查是否有传入数字参数
    ifneq ($(1),)
        # 获取用户输入的数字
        DIR_INDEX := $(1)
        # 从映射表中查找对应的子目录
        APP_DIR := $(word $(DIR_INDEX),$(subst :, ,$(value $(filter $(DIR_INDEX):,$(DIR_MAP)))))
    endif
endif

.PHONY: build clean

# 构建目标
build:
	@echo "Building project in directory: $(APP_DIR)"
	rm -rf ./bin
	mkdir -p bin/
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o  ./bin/$(APP_NAME)  main.go
	upx ./bin/*

# 清理目标
clean:
	rm -rf ./bin