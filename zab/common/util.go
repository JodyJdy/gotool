package common

func GetEpoch(zxid uint64) uint32 {
	return uint32(zxid >> 32)
}

func GetCounter(zxid uint64) uint32 {
	return uint32(zxid)
}

func Min(a, b uint64) uint64 {
	if a > b {
		return b
	}
	return a
}
