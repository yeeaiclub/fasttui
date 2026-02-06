package fasttui

type Component interface {
	/**
	 * 根据给定的视口宽度将组件渲染为行
	 * @param width - 当前视口宽度
	 * @returns 字符串数组，每个字符串代表一行
	 */
	Render(width int) []string

	/**
	 * 组件获得焦点时的可选键盘输入处理器
	 */
	HandleInput(data string)

	/**
	 * 如果为 true，组件将接收按键释放事件（Kitty 协议）。
	 * 默认值为 false - 释放事件会被过滤掉。
	 */
	WantsKeyRelease() bool

	/**
	 * 使任何缓存的渲染状态失效。
	 * 在主题更改或组件需要从头重新渲染时调用。
	 */
	Invalidate()
}
