package rc

import (
	"fmt"
	"sort"
)

// Params defines RC parameters.
type Params struct {
	Alpha      uint64
	Beta       uint64
	CSize      uint64
	CCompute   uint64
	CStorage   uint64
	MaxSkewSec int64
	WindowN    int
}

// ValidateGenesis enforces genesis bounds.
func (p Params) ValidateGenesis() error {
	if !within(p.Alpha, 1, 1_000_000) {
		return fmt.Errorf("alpha out of bounds")
	}
	if !within(p.Beta, 1, 1_000_000) {
		return fmt.Errorf("beta out of bounds")
	}
	if !within(p.CSize, 1, 1_000_000) {
		return fmt.Errorf("c_size out of bounds")
	}
	if !within(p.CCompute, 1, 1_000_000) {
		return fmt.Errorf("c_compute out of bounds")
	}
	if !within(p.CStorage, 1, 1_000_000) {
		return fmt.Errorf("c_storage out of bounds")
	}
	if p.WindowN < 1 {
		return fmt.Errorf("window_n must be >= 1")
	}
	if p.MaxSkewSec < 0 {
		return fmt.Errorf("max_skew must be >= 0")
	}
	return nil
}

func within(v, min, max uint64) bool {
	return v >= min && v <= max
}

// EffectiveTime clamps a block timestamp using median of last N timestamps and max skew.
func EffectiveTime(blockTimestamp int64, lastTimestamps []int64, maxSkew int64) int64 {
	if len(lastTimestamps) == 0 {
		return blockTimestamp
	}
	median := Median(lastTimestamps)
	min := median - maxSkew
	max := median + maxSkew
	if blockTimestamp < min {
		return min
	}
	if blockTimestamp > max {
		return max
	}
	return blockTimestamp
}

// Median returns the median of the given timestamps; for empty slice, returns 0.
func Median(timestamps []int64) int64 {
	if len(timestamps) == 0 {
		return 0
	}
	cp := make([]int64, len(timestamps))
	copy(cp, timestamps)
	sort.Slice(cp, func(i, j int) bool { return cp[i] < cp[j] })
	m := len(cp) / 2
	if len(cp)%2 == 1 {
		return cp[m]
	}
	// deterministic median for even count: lower middle
	return cp[m-1]
}

// RCMax computes RC max for a given stake.
func (p Params) RCMax(stake uint64) uint64 {
	if stake == 0 || p.Alpha == 0 {
		return 0
	}
	if stake > (^uint64(0))/p.Alpha {
		return ^uint64(0)
	}
	return p.Alpha * stake
}

// Regen computes regenerated RC based on effective time delta.
func (p Params) Regen(currentRC, stake uint64, lastEffectiveTime, newEffectiveTime int64) (uint64, int64) {
	if stake == 0 {
		return 0, newEffectiveTime
	}
	dt := newEffectiveTime - lastEffectiveTime
	if dt < 0 {
		dt = 0
	}
	regen := uint64(dt)
	if p.Beta != 0 {
		if regen > (^uint64(0))/p.Beta {
			regen = ^uint64(0)
		} else {
			regen *= p.Beta
		}
		if stake > 0 {
			if regen > (^uint64(0))/stake {
				regen = ^uint64(0)
			} else {
				regen *= stake
			}
		}
	}
	rc := currentRC
	if regen > 0 {
		if rc > (^uint64(0))-regen {
			rc = ^uint64(0)
		} else {
			rc += regen
		}
	}
	rcMax := p.RCMax(stake)
	if rc > rcMax {
		rc = rcMax
	}
	return rc, newEffectiveTime
}

// Cost computes RC cost for a transaction.
func (p Params) Cost(sizeBytes uint64, wasmInstructions uint64, stateWrites uint64) uint64 {
	total := uint64(0)
	add := func(a, b uint64) {
		if a == 0 || b == 0 {
			return
		}
		if b > (^uint64(0))/a {
			total = ^uint64(0)
			return
		}
		if total > (^uint64(0))-(a*b) {
			total = ^uint64(0)
			return
		}
		total += a * b
	}
	add(p.CSize, sizeBytes)
	add(p.CCompute, wasmInstructions)
	add(p.CStorage, stateWrites)
	return total
}
