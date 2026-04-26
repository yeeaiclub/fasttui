package style

// Mermaid diagram rendering depends on external tooling in the original TypeScript stack.
// These stubs keep a compatible API for callers that optionally prerender diagrams.

var mermaidRenderHook func()

// SetMermaidRenderHook registers a callback invoked when new diagrams might be available.
func SetMermaidRenderHook(fn func()) {
	mermaidRenderHook = fn
}

var mermaidCache = map[uint64]string{}

// MermaidASCII returns cached ASCII for a content hash, or empty string if missing.
func MermaidASCII(hash uint64) string {
	return mermaidCache[hash]
}

// SetMermaidASCII stores prerendered ASCII for a hash (replace prior value).
func SetMermaidASCII(hash uint64, ascii string) {
	mermaidCache[hash] = ascii
}

// ClearMermaidCache drops all cached diagrams.
func ClearMermaidCache() {
	mermaidCache = map[uint64]string{}
}

// PrerenderMermaid is a no-op placeholder; extend with a Mermaid → ASCII pipeline as needed.
func PrerenderMermaid(markdown string) {
	_ = markdown
}

// HasPendingMermaid always reports false until a real extractor is wired.
func HasPendingMermaid(markdown string) bool {
	_ = markdown
	return false
}
