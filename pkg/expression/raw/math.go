package raw

import (
	"strconv"
)

func fixDecimal(f float64) float64 {
	res, _ := strconv.ParseFloat(strconv.FormatFloat(f, 'g', 10, 64), 64)
	return res
}
