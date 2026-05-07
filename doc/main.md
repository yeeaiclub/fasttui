这是硬件光标泄漏到 UI 上的问题。让我追踪光标在 Select 组件获得焦点时的行为。


```go
marker := ""
if e.focused && !e.IsShowingAutocomplete() {
    marker = CURSOR_MARKER   // ← 无焦点时不插硬件光标标记 ✓
}

// 但下面这段不管有没有焦点都执行：
cursor := "\x1b[7m" + firstGrapheme + "\x1b[27m"  // ← 反白可见光标始终渲染 ✗
displayText = before + marker + cursor + restAfter
```

**根因**：Editor 失去焦点后，`CURSOR_MARKER` 不再插入（硬件光标会被隐藏），但**可见的反白光标字符 `\x1b[7m...\x1b[27m` 仍然被渲染到输出中**。这个反白块就显示在了 Select 列表旁边。
        
            
toolName: edit_file_search_replace
            
status: success
          
            
filePath: /Users/yangzhengwang/Desktop/code/fasttui/components/editor.go
          
changes: undefined
          

        
            
toolName: run_command
            
status: success
          
command: cd /Users/yangzhengwang/Desktop/code/fasttui/_examples/chat && go build -o /dev/null . 2>&1 undefined
          

        
            
toolName: run_command
            
status: running
          
command: cd /Users/yangzhengwang/Desktop/code/fasttui/_examples/chat && go run . undefined
          
编译通过，已启动。

### 修复内容

**文件**：[editor.go:449-470](file:///Users/yangzhengwang/Desktop/code/fasttui/components/editor.go#L449-L470)

**问题**：Editor 失去焦点后，`CURSOR_MARKER` 不再插入（硬件光标被隐藏 ✓），但 **可见反白光标 `\x1b[7m...\x1b[27m` 仍然渲染到输出中 ✗**，导致 Select 列表右侧出现一个白色光标块。

**修复**：统一用 `showVisibleCursor` 变量控制两种光标的显示：

```
之前（有 bug）：
  CURSOR_MARKER → 仅 focused 时插入  ✓
  反白可见光标 → 始终渲染            ✗

现在（修复后）：
  showVisibleCursor = e.focused && !e.IsShowingAutocomplete()
  CURSOR_MARKER     → showVisibleCursor 时插入  ✓
  反白可见光标      → showVisibleCursor 时渲染   ✓
```

现在输入 `/` 触发 slash command 选择列表，Select 右侧应该不再出现多余的光标了。试试看。