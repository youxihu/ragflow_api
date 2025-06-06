#!/bin/bash
source /home/youxihu/secret/aiops/rag_api/info.sh
#work_dir="/home/youxihu/mywork/public/ragflow"
work_dir="/home/workserver/work-docker/docker-compose/ragflow"
cfg_dir="/home/youxihu/secret/aiops/rag_api"
scr_dir="/home/youxihu/mywork/myproject/ragflow_api/script"

# 检查目录是否存在
checkDir() {
    if [ ! -d "$1" ]; then
        echo "错误: 目录不存在: $1"
        exit 1
    fi
}

function stopRagFlow() {
    checkDir "$work_dir/docker"
    cd "$work_dir/docker" || {
        echo "无法进入目录: $work_dir/docker"
        exit 1
    }

    if ! docker-compose -f docker-compose.yml down; then
        echo "停止RagFlow实例失败，请检查错误"
        exit 1
    fi
}

function zipRagEsData() {
    local tarTime=$(date "+%m%d-%H%M")
    packFile="es-out-${tarTime}.zip"
    packMeta="last_pack_name.txt"

    cd "$work_dir" || {
        echo "无法进入目录: $work_dir"
        exit 1
    }

    if ! zip -qr "$packFile" volumes/; then
        echo "打包RagFlowRoot数据失败，请检查错误"
        exit 1
    fi

    echo "RagFlowRoot 已成功打包为: $packFile"
    touch last_pack_name.txt
    echo "$packFile" > "$packMeta"

    scpOut "$packFile" "$packMeta"
}

function scpOut() {
    local packFile="$1"
    local packMeta="$2"
    local scrDeploySh="$scr_dir"/deploy_parse.sh
    local scrOutSh="$scr_dir"/move_out.sh
    local parseExe="$scr_dir"/parsedocs
    local areAllDoneExe="$scr_dir"/arealldone
    local authCfg="$cfg_dir/auth.yaml"
    local remote_host="$tmp_addr"
    local remote_path="$tmp_path"

    cd "$work_dir" || exit 1

    # 文件存在性检查
    [ -f "$packFile" ] || { echo "备份文件不存在: $packFile"; exit 1; }
    [ -f "$packMeta" ] || { echo "元信息文件不存在: $packMeta"; exit 1; }
    [ -f "$authCfg" ] || { echo "配置文件不存在: $authCfg"; exit 1; }
    [ -f "$scrDeploySh" ] || { echo "部署脚本不存在: $scrDeploySh"; exit 1; }
    [ -f "$scrOutSh" ] || { echo "导出脚本不存在: $scrOutSh"; exit 1; }
    [ -f "$parseExe" ] || { echo "解析程序不存在: $parseExe"; exit 1; }
    [ -f "$areAllDoneExe" ] || { echo "判断回迁程序不存在: $areAllDoneExe"; exit 1; }

    local files_out_scp=("$packFile" "$packMeta" "$authCfg" "$scrDeploySh" "$scrOutSh" "$parseExe" "$areAllDoneExe")

    for file in "${files_out_scp[@]}"; do
        filename=$(basename "$file")
        scp "$file" "root@${remote_host}:${remote_path}" || {
            echo "文件传输失败: $filename"
            exit 1
        }
    done

    if ! ssh "root@${remote_host}" "bash ${remote_path}/deploy_to_parse.sh"; then
        echo "[迁出阶段move_out]远程部署脚本执行失败:deploy_to_parse.sh"
        exit 1
    fi
}

function main() {
    stopRagFlow
    zipRagEsData
}

main