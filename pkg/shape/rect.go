package shape

// 左上角坐标系
type Rect struct {
	X, Y, Width, Height int64
}

// Contains 判断点是否在矩形内
func (r Rect) Contains(x, y int64) bool {
	return x >= r.X && x <= r.X+r.Width &&
		y >= r.Y && y <= r.Y+r.Height
}

// Contains 判断点是否在矩形内
func (r Rect) ContainsRect(rect2 Rect) bool {
	return rect2.X >= r.X && rect2.X+rect2.Width <= r.X+r.Width &&
		rect2.Y >= r.Y && rect2.Y+rect2.Height <= r.Y+r.Height
}
