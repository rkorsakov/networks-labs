package logic

type Field struct {
	Width, Height int
}

func NewField(width, height int) *Field {
	return &Field{Width: width, Height: height}
}
