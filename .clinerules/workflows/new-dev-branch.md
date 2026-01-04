# New Dev Branch Workflow

当用户要求"开始一个新功能"、"创建新分支"或执行 `/new-dev-branch.md` 时，请遵循以下流程。本指南旨在确保新分支基于正确的公共基线创建，并正确初始化私有开发环境。

## 1. 确定基准分支

**黄金法则**：永远基于公共分支（Public Branch）创建新的特性分支。

*   ✅ **推荐基准**：
    *   `origin/main` (或 `origin/master`)
    *   `origin/feature/xxx` (已发布的特性分支)
*   ❌ **禁止基准**：
    *   `dev/xxx` (其他私有开发分支)
    *   **原因**：基于私有分支创建会导致 Git 历史混乱，且无法通过提交脚本的"纯净度检查"。

**操作示例**：
```bash
# 获取最新代码
git fetch origin

# 基于远程 feature 分支创建本地 dev 分支
# 注意：<my-new-feature> 是您自定义的分支名称
git checkout -b dev/<my-new-feature> origin/feature/<old-feature>
```

## 2. 恢复私有上下文

由于公共分支不包含私有文件（如 `memory-bank/`, `doc/` 等），新创建的分支会缺失这些重要的开发上下文。

请使用 `./submit.sh` 的智能恢复功能一键找回：

**命令**：
```bash
././submit.sh --restore-private [source_ref]
```

**用法说明**：
*   **智能推断 (推荐)**：不带参数运行。脚本会自动分析当前分支的父节点历史，找到最近的 Cherry-pick 来源，并从那里恢复私有文件。
    ```bash
    ././submit.sh --restore-private
    ```
*   **手动指定**：如果智能推断失败，您可以手动指定源分支（通常是上一个开发分支）或具体的 Commit Hash。
    ```bash
    # 从分支恢复
    ././submit.sh --restore-private dev/previous-feature
    
    # 从 Commit Hash 恢复
    ././submit.sh --restore-private a1b2c3d
    ```

## 3. 初始化提交

恢复私有文件后，请立即创建一个初始化提交，以保存私有上下文。

**命令**：
```bash
export COMMIT_MSG_PRIVATE="chore: Initialize private context"
././submit.sh --commit-private
unset COMMIT_MSG_PRIVATE
```

## 4. 开始开发

现在，您的新分支 `dev/<my-new-feature>` 已经准备就绪：
*   它基于干净的公共历史。
*   它拥有完整的私有开发文档。

您可以开始编码了！完成开发后，请参考 `submit.md` 进行提交和发布。
