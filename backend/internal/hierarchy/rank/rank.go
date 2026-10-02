package rank

import "math"

const Gap int64 = 1 << 32

func Between(prev, next *int64) (rank int64, ok bool) {
	switch {
	case prev == nil && next == nil:
		return Gap, true

	case prev == nil:
		if *next < math.MinInt64+Gap {
			return 0, false
		}
		return *next - Gap, true

	case next == nil:
		if *prev > math.MaxInt64-Gap {
			return 0, false
		}
		return *prev + Gap, true

	default:
		if *prev >= *next {
			return 0, false
		}

		diff := uint64(*next) - uint64(*prev)
		if diff < 2 {
			return 0, false
		}
		return *prev + int64(diff/2), true
	}
}
