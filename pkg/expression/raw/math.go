package raw

import (
	"strconv"
)

// fixDecimal 修正浮点数精度问题。
func fixDecimal(f float64) float64 {
	res, _ := strconv.ParseFloat(strconv.FormatFloat(f, 'g', 10, 64), 64)
	return res
}
