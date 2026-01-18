# Submit Workflow

当用户要求提交代码或执行 `/submit.md` 时，请遵循以下流程。本工作流支持标准的 `develop` 分支开发，也支持 `dev/xxx` 特性分支开发。

## 1. 前置检查 (Prerequisites)

在执行提交脚本之前，脚本会自动进行一系列**严格检查**。也可以手动执行 `./submit.sh --check-prerequisites` 来仅检查分支命名与父分支纯净度：

1.  **分支命名检查**：
    *   **标准开发**：应处于 `develop` 分支。
    *   **特性开发**：应处于 `dev/xxx` 分支（例如 `dev/motor-control`）。
    *   脚本 `./submit.sh` 会自动识别并映射：`develop` $\rightarrow$ `main` (或 `master`)，`dev/xxx` $\rightarrow$ `feature/xxx`。

2.  **父分支纯净度检查 (Sanity Check)**：
    *   脚本会检查当前分支的**父提交 (Parent Commit)**。
    *   **通过标准**：父提交属于任何**公共分支**（`main`, `master`, `feature/*` 等）。
    *   **失败标准**：父提交**不属于**任何公共分支，但属于其他**私有分支**（`develop` 或 `dev/*`）。这通常意味着您错误地基于另一个私有分支创建了当前分支。
    *   **处理办法**：如果报错，请参考 `.clinerules/workflows/newdev.md`，重新基于公共分支创建分支，并使用 `--restore-private` 找回私有文件。

如果检查不通过，脚本将拒绝执行。

## 2. 分类文件
将文件分为两组：
*   **Private Group**: 只能存在于私有开发分支（`develop` 或 `dev/xxx`），禁止推送到公共发布分支。可通过命令 `./submit.sh --print-private-patterns` 获取的文件列表。
*   **Public Group**:  可以推送到公共发布分支（`main` 或 `master` 或 `feature/xxx`）。可通过命令 `./submit.sh --print-public-patterns` 获取的文件列表。

## 3. 分析变更
分别获取私有文件的变更 和 公共文件的变更。
*   **Private**:
    *   **Status**: `./submit.sh --print-private-status` (打印已变更的私有文件清单)
    *   **Diff**: `./submit.sh --print-private-diff` (打印已变更的私有文件的具体变更内容)
    *   **Show**: `./submit.sh --print-private-show` (打印已变更的私有文件清单和具体的变更内容)
*   **Public**:
    *   **Status**: `./submit.sh --print-public-status` (打印已变更的公共文件清单)
    *   **Diff**: `./submit.sh --print-public-diff` (打印已变更的公共文件的具体变更内容)
    *   **Show**: `./submit.sh --print-public-show` (打印已变更的公共文件清单和具体的变更内容)
根据变更，生成符合 [Conventional Commits](https://www.conventionalcommits.org/) 规范的 Commit Message，分别存储在环境变量 `COMMIT_MSG_PRIVATE` 和 `COMMIT_MSG_PUBLIC`。

## 4. 提交变更
*   **Private Commit**: 如果有私有变更，执行 `./submit.sh --commit-private`。
*   **Public Commit**: 如果有公共变更，执行 `./submit.sh --commit-public`。

## 5. 同步到远程仓库
脚本会自动检测远程仓库的 **可见性 (Visibility)**，即该仓库在托管平台（如 GitHub）上是 **私有 (Private)** 还是 **公开 (Public)**。

*   **Private Push**: 执行 `./submit.sh --push-private`。
    *   **只负责私有分支**：仅推送当前的私有开发分支（`develop` 或 `dev/xxx`）。
    *   **只推送到私有仓库**：脚本会遍历所有远程仓库，**仅**对检测为 **Private** 的仓库执行推送。
    *   **安全保护**：会自动跳过所有 Public 仓库，防止私有分支泄露。
    *   **跳过不可达仓库**：如果检测到远程仓库不可达（如内网仓库），将输出警告并自动跳过，不阻断流程。

*   **Public Push**: 执行 `./submit.sh --push-public`。
    *   **只负责公共分支**：自动将公共提交从当前私有分支同步（cherry-pick）到对应的公共分支（`main` (或 `master`) 或 `feature/xxx`）。
    *   **智能父节点探测**：如果本地不存在对应的 `feature/xxx` 分支，脚本会自动扫描**所有远程公共分支**（排除私有开发分支），寻找与当前开发分支**最近的共同祖先 (Merge Base)**（如 `main`, `master`, `release/*` 或其他 `feature/*`），并基于该祖先创建新分支。这确保了新分支能精准继承历史开发上下文。
    *   **全量推送**：将更新后的公共分支（`main` (或 `master`) 或 `feature/xxx`）推送到 **所有** 远程仓库（无论 Public 还是 Private）。
    *   **连通性检查**：在推送前检查远程仓库连通性，自动跳过不可达的仓库并警告，防止操作长时间卡顿。

## 6. 报告
1.  **清理环境变量**：为了防止污染后续操作，请务必执行以下命令清理环境变量：
    ```bash
    unset COMMIT_MSG_PRIVATE COMMIT_MSG_PUBLIC
    ```
2.  **生成报告**：向用户报告详细的提交结果，包含以下内容：
    *   **当前分支**：说明当前工作的分支名称。
    *   **私有提交**：Commit Hash、Message、包含的文件。
    *   **公共提交**：Commit Hash、Message、包含的文件。
    *   **同步状态**：远程分支是否已同步更新。
    *   **异常警告**：如果在过程中遇到任何非阻碍性错误（如需要手动解决的冲突），请明确指出。
