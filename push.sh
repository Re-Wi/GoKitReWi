#!/bin/bash

# 定义提交类型及其提示信息
declare -A commit_types_with_hints=(
    ["feat"]="新功能（feature）"
    ["fix"]="修复bug"
    ["docs"]="文档（documentation）"
    ["style"]="格式（不影响代码运行的结果，如缩进、空格等）"
    ["refactor"]="重构（即不是新增功能，也不是修复bug）"
    ["test"]="增加测试、修改测试"
    ["chore"]="构建过程或辅助工具的变动"
    ["perf"]="性能 (提高代码性能的改变)"
    ["build"]="影响构建系统或外部依赖项的更改(maven,gradle,npm 等等)"
    ["ci"]="对 CI 配置文件和脚本的更改"
    ["revert"]="Revert a commit"
    ["types"]="类型定义文件更改"
    ["workflow"]="工作流改进"
    ["wip"]="工作进行中（Work In Progress）"
    ["config"]="配置文件的改动"
    ["locale"]="国际化/本地化的改动"
    ["security"]="安全相关的改动"
)

commit_types=("${!commit_types_with_hints[@]}")
selected_index=0

# 显示提交类型菜单
function show_menu() {
    echo "请选择提交类型："
    for i in "${!commit_types[@]}"; do
        if [[ $i -eq $selected_index ]]; then
            echo "> ${commit_types[i]} (${commit_types_with_hints[${commit_types[i]}]})"
        else
            echo "  ${commit_types[i]} (${commit_types_with_hints[${commit_types[i]}]})"
        fi
    done
}

# 捕获键盘输入并更新选中项
function select_commit_type() {
    while true; do
        clear
        show_menu

        # 读取用户输入（捕获单个字符）
        read -rsn1 input
        case $input in
        $'\x1b')                   # 如果是转义序列开头
            read -rsn2 extra_input # 读取额外两个字符
            case $extra_input in
            "[A") # 上方向键
                selected_index=$(((selected_index - 1 + ${#commit_types[@]}) % ${#commit_types[@]}))
                ;;
            "[B") # 下方向键
                selected_index=$(((selected_index + 1) % ${#commit_types[@]}))
                ;;
            esac
            ;;
        "") # 回车键确认选择
            break
            ;;
        esac
    done
}

# 主程序
echo "Git 提交助手"
select_commit_type

# 获取选中的提交类型
selected_type="${commit_types[$selected_index]}"

# 输入提交信息
read -p "请输入提交信息： " commit_message

# 检查提交信息是否为空
if [[ -z "$commit_message" ]]; then
    echo "错误：提交信息不能为空！"
    exit 1
fi

# 构造完整的提交信息
full_commit_message="${selected_type}: ${commit_message}"

# 执行 Git 提交
echo "提交信息：$full_commit_message"
git add --all
git commit -m "$full_commit_message"

# 提交完成提示
if [[ $? -eq 0 ]]; then
    echo "提交成功！"
else
    echo "提交失败，请检查错误信息。"
fi

read -p "是否推送到远程仓库？(y/n): " push_choice
if [[ "$push_choice" == "y" || "$push_choice" == "Y" ]]; then
    git push
fi
