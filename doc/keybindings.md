### 1. EditorAction 类型定义
将 TypeScript 的联合类型转换为 Go 的常量，包含 31 个编辑器动作：
- **光标移动**: cursorUp, cursorDown, cursorLeft, cursorRight, cursorWordLeft, cursorWordRight, cursorLineStart, cursorLineEnd, jumpForward, jumpBackward, pageUp, pageDown
- **删除操作**: deleteCharBackward, deleteCharForward, deleteWordBackward, deleteWordForward, deleteToLineStart, deleteToLineEnd
- **文本输入**: newLine, submit, tab
- **选择/自动完成**: selectUp, selectDown, selectPageUp, selectPageDown, selectConfirm, selectCancel
- **剪贴板**: copy
- **Kill ring**: yank, yankPop
- **撤销**: undo
- **工具输出**: expandTools
- **会话**: toggleSessionPath, toggleSessionSort, renameSession, deleteSession, deleteSessionNoninvasive

### 2. EditorKeybindingsConfig 类型
```go
type EditorKeybindingsConfig map[EditorAction][]string
```

### 3. 默认键绑定配置
完整保留了所有默认键绑定，包括：
- 光标移动的多种快捷键（如 left/ctrl+b, right/ctrl+f）
- 删除操作的快捷键（如 backspace, ctrl+w, alt+d）
- 会话管理的快捷键（如 ctrl+p, ctrl+s, ctrl+r）

### 4. EditorKeybindingsManager 结构体和方法
```go
type EditorKeybindingsManager struct {
    actionToKeys map[EditorAction][]string
}
```

**方法**:
- `NewEditorKeybindingsManager(config)` - 创建新的管理器
- `buildMaps(config)` - 构建键映射（私有方法）
- `Matches(data, action)` - 检查输入是否匹配指定动作
- `GetKeys(action)` - 获取绑定到动作的键列表
- `SetConfig(config)` - 更新配置

### 5. 全局单例
```go
var globalEditorKeybindings *EditorKeybindingsManager

func GetEditorKeybindings() *EditorKeybindingsManager
func SetEditorKeybindings(manager *EditorKeybindingsManager)
```

## 迁移特点

1. **类型安全** - 使用 Go 的类型系统确保编译时检查
2. **内存安全** - 返回切片时使用 `append([]string{}, keys...)` 创建副本，避免意外修改
3. **空值处理** - 正确处理 nil 配置和未找到的动作
4. **兼容性** - 完美兼容 keys.go 中的 `MatchesKey` 函数

代码已通过编译验证，功能与 TypeScript 版本完全一致。