# New Dev Branch Workflow

当用户要求创建新分支或执行 `/newdev.md` 时，请遵循以下流程。

## 1. 创建分支

```bash
./submit.sh --new-dev-branch <feature-name> <base>
```

| 参数 | 说明 | 示例 |
|------|------|------|
| `<feature-name>` | 功能名称（**不含** `dev/` 前缀） | `auth-refactor` |
| `<base>` | 公共基准分支 | `main`、`origin/main`、`feature/xxx` |

**结果**：创建并切换到 `dev/<feature-name>` 分支（对应公共分支 `feature/<feature-name>`）

## 2. 恢复私有文件

新创建的私有分支不含私有文件，需要恢复：

```bash
./submit.sh --restore-private
```

## 3. 初始化私有提交

```bash
COMMIT_MSG_PRIVATE='chore: initialize private context' ./submit.sh --commit-private
```

## 4. 开始开发

分支已就绪，完成开发后参考 `submit.md` 提交。
