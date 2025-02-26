#!/bin/bash

. scripts/list_app.sh

get_app_list

readonly root_path=`pwd`
for app_path in ${app_list[*]}; do
    # 确保切换到应用目录
    app_dir="${root_path}/${app_path}"
    echo "Running app from: ${app_dir}"

    # 如果 app 目录存在，则切换到该目录并执行 go run
    if [ -d "${app_dir}" ]; then
        cd "${app_dir}" || { echo "Failed to cd into ${app_dir}"; exit 1; }
        
        log_file="../../logs/${app_path}_nohup.out"
        # 执行 go run
        nohup go run . > "../../logs/nohup.out" 2>&1 &
    else
        echo "Directory ${app_dir} not found."
    fi
done