package constant

var (
	Video    *video
	Resp     *resp
	Location *location
	Operator *operator
)

type video struct{}

func (*video) SuffixList() []string {
	return []string{".mkv", ".mp4", ".flv", ".avi"}
}

func (*video) PageSize() int {
	return 8
}

// ---------------------------------------------------

type resp struct{}

func (*resp) OK() int {
	return 0
}

func (*resp) Fail() int {
	return -1
}

type location struct{}

func (*location) PageSize() int {
	return 18
}

type operator struct{}

func (*operator) NothingToDo() int {
	return 0
}
func (*operator) FileRemoved() int {
	return 1
}
func (*operator) FileChanged() int {
	return 2
}
