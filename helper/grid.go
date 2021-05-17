package helper

import "math"

// GetGridRange returns a lower value and an upper value of the grid that closed to the target value
func GetGridRange(target float64, lowerNum float64, upperNum float64, grids float64) (float64, float64, float64) {
	if target <= lowerNum || lowerNum >= upperNum || grids < 2 {
		return 0, 0, 0
	}

	lower := lowerNum
	upper := upperNum
	zone := upper - lower
	grid := zone / grids

	if math.Mod(grids, 5) == 0 {
		div := zone / 5

		if lower+div*4 < target {
			lower += div * 4
		} else if lower+div*3 < target {
			lower += div * 3
		} else if lower+div*2 < target {
			lower += div * 2
		} else if lower+div < target {
			lower += div
		}

		if upper-div*4 > target {
			upper -= div * 4
		} else if upper-div*3 > target {
			upper -= div * 3
		} else if upper-div*2 > target {
			upper -= div * 2
		} else if upper-div > target {
			upper -= div
		}
	} else if math.Mod(grids, 4) == 0 {
		div := zone / 4

		if lower+div*3 < target {
			lower += div * 3
		} else if lower+div*2 < target {
			lower += div * 2
		} else if lower+div < target {
			lower += div
		}

		if upper-div*3 > target {
			upper -= div * 3
		} else if upper-div*2 > target {
			upper -= div * 2
		} else if upper-div > target {
			upper -= div
		}
	} else if math.Mod(grids, 3) == 0 {
		div := zone / 3

		if lower+div*2 < target {
			lower += div * 2
		} else if lower+div < target {
			lower += div
		}

		if upper-div*2 > target {
			upper -= div * 2
		} else if upper-div > target {
			upper -= div
		}
	} else if math.Mod(grids, 2) == 0 {
		div := zone / 2

		if lower+div < target {
			lower += div
		}

		if upper-div > target {
			upper -= div
		}
	}

	for i := 0; i < int(grids); i++ {
		if lower+grid < target {
			lower += grid
		} else {
			break
		}
	}

	for i := 0; i < int(grids); i++ {
		if upper-grid > target {
			upper -= grid
		} else {
			break
		}
	}

	return lower, upper, grid
}
