package bloom

type F func(string) int64

func NewFuncs() []F {
	m := make([]F, 0)

	var f F
	f = BKDRHash
	m = append(m, f)

	f = SDBMHash
	m = append(m, f)

	f = DJBHash
	m = append(m, f)

	return m
}

func BKDRHash(str string) int64 {
	seed := int64(131) // 31 131 1313 13131 131313 etc..
	hash := int64(0)
	for i := 0; i < len(str); i++ {
		hash = (hash * seed) + int64(str[i])
	}
	return hash & 0x7FFFFFFF
}
func SDBMHash(str string) int64 {
	hash := int64(0)
	for i := 0; i < len(str); i++ {
		hash = int64(str[i]) + (hash << 6) + (hash << 16) - hash
	}
	return hash & 0x7FFFFFFF
}
func DJBHash(str string) int64 {
	hash := int64(0)
	for i := 0; i < len(str); i++ {
		hash = ((hash << 5) + hash) + int64(str[i])
	}
	return hash & 0x7FFFFFFF
}
