package common

func GetEpoch(zxid uint64) uint32 {
	return uint32(zxid >> 32)
}

func GetCounter(zxid uint64) uint32 {
	return uint32(zxid)
}
