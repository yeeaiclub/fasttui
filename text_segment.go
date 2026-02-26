package fasttui

type segmentType int

const (
	segmentTypeAnsi segmentType = iota
	segmentTypeGrapheme
)

type textSegment struct {
	segType segmentType
	value   string
}

type SliceResult struct {
	text  string
	width int
}

func GetSegmenter() any {
	return nil
}

func GraphemeWidth(s string) int {
	if len(s) == 0 {
		return 0
	}
	return len(s)
}
