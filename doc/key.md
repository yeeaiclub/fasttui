我来帮你将 TypeScript 的 keys.ts 实现迁移到 Go 的 keys.go 文件。首先让我读取源文件。
        
## 迁移的功能

### 1. 全局状态管理
- `kittyProtocolActive` - Kitty 协议激活状态
- `lastEventType` - 最后一个按键事件类型

### 2. 类型定义
- `KeyEventType` - 按键事件类型（press/repeat/release）
- `ParsedKittySequence` - 解析后的 Kitty 序列结构体

### 3. 常量定义
- 修饰符常量（Shift/Alt/Ctrl）
- 键码点常量（Escape/Tab/Enter/Space/Backspace/KPEnter）
- 箭头键码点常量
- 功能键码点常量

### 4. 映射表
- `symbolKeys` - 符号键集合（32个符号）
- `legacyKeySequences` - 传统键序列映射
- `legacyShiftSequences` - 传统 Shift 修饰键序列
- `legacyCtrlSequences` - 传统 Ctrl 修饰键序列
- `legacySequenceKeyIDs` - 传统序列到键 ID 的映射

### 5. 核心功能函数
- `SetKittyProtocolActive()` - 设置 Kitty 协议状态
- `IsKittyProtocolActive()` - 查询 Kitty 协议状态
- `IsKeyRelease()` - 检查是否为按键释放
- `IsKeyRepeat()` - 检查是否为按键重复
- `ParseKittySequence()` - 解析 Kitty 协议序列
- `MatchesKey()` - 匹配输入数据与键标识符
- `ParseKey()` - 解析输入数据并返回键标识符

### 6. 辅助函数
- `parseEventType()` - 解析事件类型
- `matchesKittySequence()` - 匹配 Kitty 序列
- `matchesModifyOtherKeys()` - 匹配 xterm modifyOtherKeys 格式
- `rawCtrlChar()` - 获取控制字符
- `parseKeyId()` - 解析键标识符
- `matchesLegacySequence()` - 匹配传统序列
- `matchesLegacyModifierSequence()` - 匹配传统修饰键序列