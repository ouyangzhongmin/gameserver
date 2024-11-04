package path

// 预生成的路径
type SerialPaths struct {
	Id    int         `json:"id"`
	Paths []PointPath `json:"paths"`
}

type PointPath struct {
	Sx    int32     `json:"sx"`
	Sy    int32     `json:"sy"`
	Ex    int32     `json:"ex"`
	Ey    int32     `json:"ey"`
	Paths [][]int32 `json:"paths"`
}
