package shape

import "math"

// 计算两点之间的距离
func CalculateDistance(x1, y1, x2, y2 float64) float64 {
	return math.Sqrt(math.Pow(x2-x1, 2) + math.Pow(y2-y1, 2))
}

// 判断点是否在圆内
// (cx, cy) 是圆心的坐标，radius 是圆的半径，(x, y) 是点的坐标
func IsInsideCircle(cx, cy, radius, x, y float64) bool {
	distance := math.Sqrt(math.Pow(x-cx, 2) + math.Pow(y-cy, 2))
	return distance <= radius
}

/**
 * 获得从基准点到目标点的角度
 * @param baseX
 * @param baseY
 * @param targetX
 * @param targetY
 * @return
 */
func GetAngle(baseX, baseY, targetX, targetY float64) float64 {
	radians := GetRadians(baseX, baseY, targetX, targetY)
	return RadiansToDegrees(radians)
}

/**
 * 获得从基准点到目标点的弧度
 * @param baseX
 * @param baseY
 * @param targetX
 * @param targetY
 * @return
 */
func GetRadians(baseX, baseY, targetX, targetY float64) float64 {
	return math.Atan2(baseY-targetY, targetX-baseX) // flash 的坐标的y是反的
}

/*
* 将弧度转换为度数
 */
func RadiansToDegrees(radians float64) float64 {
	return radians * (180 / math.Pi)
}
