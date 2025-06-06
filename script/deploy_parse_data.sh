#!/bin/bash

work_dir="/home/workserver/work-docker/docker-compose/ragflow"
source /home/sys_bash_send/send_rag_api.sh

function unzipToRunRag() {
    local bakTime=$(date "+%m%d-%H%M")

    cd "$work_dir" || exit 1

    [ -f /tmp/last_pack_name.txt ] || exit 1

    bkFileName=$(cat /tmp/last_pack_name.txt)

    [ -f "/tmp/$bkFileName" ] || exit 1

    # 移动解析后的数据文件
    mv "/tmp/$bkFileName" . || exit 1

    [ -d volumes ] && mv volumes "volumes.$bakTime.bak" || true

    unzip -q "$bkFileName" || exit 1

    cd "$work_dir/docker" && docker-compose -f docker-compose.yml up -d || exit 1

    sendDing "### 事件通知: ragflow数据已迁回
    #### 状态: 已完成
    - 执行时间: $bakTime
    - 备注:【请及时登陆服务器并检查是否可用](http://ragflow.biaobiaoxing.com)"
}

function main() {
    unzipToRunRag
}

main