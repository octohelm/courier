package internal

func MinInt(bitSize uint) int64 {
	return -1 << (bitSize - 1)
}

func MaxInt(bitSize uint) int64 {
	return 1<<(bitSize-1) - 1
}

func MaxUint(bitSize uint) uint64 {
	return 1<<bitSize - 1
}
