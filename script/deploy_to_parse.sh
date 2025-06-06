#!/bin/bash

work_dir="/workserver/work-docker/docker-volume/ragflow"
cfg_dir="/home/youxihu/secret/aiops/rag_api"

if [ ! -d "$cfg_dir" ]; then
    echo "目录不存在: $cfg_dir,开始创建..."
    mkdir -p "$cfg_dir"
fi

function unzipToDeploy() {
    local bakTime=$(date "+%m%d-%H%M")

    cd "$work_dir" || exit 1

    [ -f /tmp/last_pack_name.txt ] || exit 1

    pkFileName=$(cat /tmp/last_pack_name.txt)

    [ -f "/tmp/$pkFileName" ] || exit 1
    [ -f "/tmp/auth.yaml" ] || exit 1

    # 移动备份文件和配置文件
    mv "/tmp/$pkFileName" . || exit 1
    mv /tmp/auth.yaml "$cfg_dir" || exit 1

    [ -d volumes ] && mv volumes "volumes.$bakTime.bak" || true

    unzip -q "$pkFileName" || exit 1

    cd "$work_dir/docker" && docker-compose -f docker-compose.yml up -d || exit 1
}

function startParse() {
    cd /tmp && parsedocs
}

function main() {
    unzipToDeploy
    startParse
}

main