**`GIT_COLLABORATION_GUIDE.md`** 文件。它整合了原有 .docx 规范文档的核心内容，并采用 Markdown 格式优化了结构和可读性。

---

# **Git/GitHub 项目协作规范 (v1.2)**

> **最后更新**：2026-04-18
> 
> **维护者**：项目负责人
> 
> **说明**：本文档定义了本项目的版本控制与协作流程，务必遵守。

## 1. 分支策略

采用精简高效的 **Git Flow** 模型。

| 分支类型       | 命名规范                | 创建自    | 合并到              | 说明与生命周期                                                                                  |
| :------------- | :---------------------- | :-------- | :------------------ | :---------------------------------------------------------------------------------------------- |
| **主分支**     | `main`                  | -         | -                   | **稳定版**。受保护，仅接受来自 `develop` 的 PR。代表可发布的里程碑。                            |
| **开发分支**   | `develop`               | `main`    | `main`              | **集成测试版**。所有新功能的集成分支，是 `main` 的唯一来源。                                    |
| **功能分支**   | `feature/姓名-功能简述` | `develop` | `develop`           | **个人开发分支**。例如 `feature/alac19-config-parser`。功能完成后，通过 **Pull Request** 合并。 |
| **热修复分支** | `hotfix/问题简述`       | `main`    | `main` 与 `develop` | 用于紧急修复 `main` 分支上的生产Bug。                                                           |

## 2. 核心工作流程

### 2.1 开始新功能开发
1.  **同步**：确保本地 `develop` 分支与远程同步。
    ```bash
    git checkout develop
    git pull origin develop
    ```
2.  **开分支**：基于最新的 `develop` 创建功能分支。
    ```bash
    git checkout -b feature/你的名字-功能简述
    ```

### 2.2 日常开发与提交
1.  在功能分支上进行开发。
2.  遵循**提交信息规范**进行原子提交。
3.  定期将本地分支推送到远程以备份：
    ```bash
    git push -u origin feature/你的名字-功能简述
    ```

### 2.3 完成功能：发起 Pull Request
1.  将最终代码推送到远程仓库。
2.  访问 GitHub 仓库页面，点击 **"Compare & pull request"**。
3.  创建 PR：
    *   **Base**: `develop`
    *   **Compare**: `你的功能分支`
    *   **标题**：使用规范前缀，如 `feat: 添加配置文件解析模块`
    *   **描述**：简要说明修改内容、测试情况。
4.  **请求队友进行代码审查**。
5.  根据审查意见在本地修改，然后推送更新（PR会自动同步）。

### 2.4 合并与清理
1.  审查通过后，由合并者选择 **"Squash and merge"** 选项合并PR。
    > **为何使用 Squash and merge？** 它将功能分支的所有提交合并为一条清晰的记录，保持 `develop` 分支历史线性整洁。
2.  在 GitHub 上合并后，**删除远程功能分支**。
3.  **本地清理**：
    ```bash
    # 切换回 develop 并拉取合并后的最新代码
    git checkout develop
    git pull origin develop

    # 删除已合并的本地功能分支
    git branch -d feature/你的名字-功能简述
    ```

## 3. 提交信息规范

每次提交必须使用以下格式：
```
<类型>: <简短描述>
```

### 常用类型

| 类型       | 说明                   | 示例                             |
| :--------- | :--------------------- | :------------------------------- |
| `feat`     | 新增功能               | `feat: 添加用户登录接口`         |
| `fix`      | 修复Bug                | `fix: 修复配置文件路径读取错误`  |
| `docs`     | 文档更新               | `docs: 更新API接口文档`          |
| `refactor` | 代码重构（不改变行为） | `refactor: 优化配置加载函数结构` |
| `chore`    | 构建/工具变动          | `chore: 更新Cargo.toml依赖版本`  |
| `test`     | 测试相关               | `test: 为配置模块添加单元测试`   |

**要求**：描述部分语言保持统一（全英或者全中），清晰说明本次提交的**目的**。

## 4. 黄金准则

1.  **确认分支**：开始编码前，务必用 `git branch` 确认所在分支正确。
2.  **先同步，后操作**：在任何 `pull` 或 `push` 操作前，先执行 `git pull origin develop` 拉取最新代码，这是避免冲突的最重要原则。
3.  **PR是质量的防火墙**：所有功能开发必须通过 Pull Request 合并，禁止直接向 `develop` 或 `main` 推送代码。
4.  **即时清理分支**：功能分支合并后，立即删除远程和本地的该分支，保持仓库整洁。
5.  **沟通先行**：如果某项修改可能影响他人，请在开发前或PR中提前说明。

## 5. 使用示例与图解

> 提示：部分关键操作的工作截图，以助于清晰地阐述步骤。如下：
> ![](./media/git-standard-operations-1.png)
> ![](./media/git-standard-operations-2.png)
> ![](./media/git-standard-operations-3.png)
> ![](./media/git-standard-operations-4.png)
> ![](./media/git-standard-operations-5.png)

完整的操作流程图和关键步骤截图，请参阅项目中的 [GIT_COLLABORATION_GUIDE.docx](./GIT_COLLABORATION_GUIDE.docx) 文档。

后续修改直接在该 **`GIT_COLLABORATION_GUIDE.md`** 文件修改即可

---