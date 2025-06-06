#!/bin/bash

source /home/sys_bash_send/send_rag_api.sh
source /home/youxihu/secret/aiops/rag_api/info.sh

work_dir="/workserver/work-docker/docker-volume/ragflow"

function checkAndZip() {
    local tarTime=$(date "+%m%d-%H%M")
    backupFile="es-back-${tarTime}.zip"
    backupMeta="last_backup_name.txt"

    cd "$work_dir" || {
        echo "无法进入目录: $work_dir"
        exit 1
    }

    # 获取文档解析状态
    document_status=$(/tmp/arealldone)

    # 检查是否所有文档解析完成
    if [ "$document_status" != "DOCUMENT_ALL_DONE=true" ]; then
        echo "文档解析未完成，当前状态: $document_status"
        exit 1
    fi

    # 打包解析结果
    if ! zip -qr "$backupFile" volumes/; then
        echo "打包解析后的RagFlowRoot数据失败，请检查错误"
        exit 1
    fi

    scpBack "$backupFile" "$backupMeta"
}

function scpBack() {
    local backup_File="$1"
    local backup_Meta="$2"
    local remote_host="$local_addr"
    local remote_path="$tmp_path"

    cd "$work_dir" || exit 1

    # 文件存在性检查
    [ -f "$backup_File" ] || { echo "解析后的数据文件不存在: $backup_File"; exit 1; }
    [ -f "$backup_Meta" ] || { echo "解析后的数据元信息文件不存在: $backup_Meta"; exit 1; }

    local files_back_scp=("$backup_File" "$backup_Meta")

    for file in "${files_back_scp[@]}"; do
        filename=$(basename "$file")
        scp "$file" "root@${remote_host}:${remote_path}" || {
            echo "文件传输失败: $filename"
            exit 1
        }
    done

    if ! ssh "root@${remote_host}" "bash ${remote_path}/deploy_parse_data.sh"; then
        echo "[回迁阶段move_back]远程部署脚本执行失败: deploy_parse_data.sh"
        exit 1
    fi
}

function main() {
    checkAndZip
}

main