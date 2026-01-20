# Submit Workflow

当用户要求提交代码或执行 `/submit.md` 时，请遵循以下流程。

## 分支说明

| 私有分支 (开发用) | 公共分支 (发布用) |
|------------------|------------------|
| `develop`        | `main`/`master`  |
| `dev/xxx`        | `feature/xxx`    |

私有分支包含所有文件；公共分支仅包含可公开的文件。

## 1. 前置检查

```bash
./submit.sh --check-prerequisites
```

检查分支命名和父分支纯净度。失败时参考 `newdev.md` 重建分支。

## 2. 查看变更

| 类型     | 命令                               | 说明         |
|----------|-----------------------------------|--------------|
| 私有文件 | `./submit.sh --print-private-show` | 状态 + diff |
| 公共文件 | `./submit.sh --print-public-show`  | 状态 + diff |

其他选项：`--print-xxx-status`（仅状态）、`--print-xxx-diff`（仅 diff）、`--print-xxx-files`（文件列表）。

## 3. 提交

根据变更生成符合 [Conventional Commits](https://www.conventionalcommits.org/) 规范的消息：

```bash
# 私有变更
export COMMIT_MSG_PRIVATE="feat(memory-bank): update progress"
./submit.sh --commit-private

# 公共变更
export COMMIT_MSG_PUBLIC="fix(biss): correct CRC calculation"
./submit.sh --commit-public
```

## 4. 推送

| 命令                              | 说明                                     |
|----------------------------------|------------------------------------------|
| `./submit.sh --push-private` | 推送私有分支 → 仅私有仓库                  |
| `./submit.sh --push-public`  | 同步公共提交 → 公共分支 → 所有仓库         |

推送时自动检测仓库可见性，跳过不可达或不适用的仓库。

## 5. 清理

```bash
unset COMMIT_MSG_PRIVATE COMMIT_MSG_PUBLIC
```

## 6. 报告

向用户报告：
- 当前分支
- 私有/公共提交的 Hash 和 Message
- 远程同步状态
- 异常警告（如需手动解决的冲突）
