package generator

// RNG is a small SplitMix64 generator whose implementation is owned by the
// project. Its sequence is part of GeneratorVersion and must not be changed
// without incrementing that version.
type RNG struct {
	state uint64
}

func NewRNG(seed uint64) *RNG {
	return &RNG{state: seed}
}

func (r *RNG) Uint64() uint64 {
	r.state += 0x9e3779b97f4a7c15
	z := r.state
	z = (z ^ (z >> 30)) * 0xbf58476d1ce4e5b9
	z = (z ^ (z >> 27)) * 0x94d049bb133111eb
	return z ^ (z >> 31)
}

func (r *RNG) Intn(n int) int {
	if n <= 0 {
		panic("generator: Intn called with non-positive bound")
	}
	return int(r.Uint64() % uint64(n))
}

func (r *RNG) Bool(chancePercent int) bool {
	return r.Intn(100) < chancePercent
}

func (r *RNG) Pick(values []string) string {
	return values[r.Intn(len(values))]
}
