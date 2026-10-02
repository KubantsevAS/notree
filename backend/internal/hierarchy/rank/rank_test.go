package rank_test

import (
	"math"
	"testing"

	"github.com/KubantsevAS/notree/backend/internal/hierarchy/rank"
	"github.com/stretchr/testify/require"
)

func ptr(v int64) *int64 { return &v }

func TestBetween(t *testing.T) {
	tests := []struct {
		name     string
		prev     *int64
		next     *int64
		wantRank int64
		wantOK   bool
	}{
		{name: "empty list", wantRank: rank.Gap, wantOK: true},
		{name: "append", prev: ptr(1000), wantRank: 1000 + rank.Gap, wantOK: true},
		{name: "prepend", next: ptr(1000), wantRank: 1000 - rank.Gap, wantOK: true},
		{name: "midpoint", prev: ptr(1000), next: ptr(2000), wantRank: 1500, wantOK: true},
		{name: "odd distance rounds down", prev: ptr(1000), next: ptr(1003), wantRank: 1001, wantOK: true},
		{name: "negative ranks", prev: ptr(-2000), next: ptr(-1000), wantRank: -1500, wantOK: true},
		{name: "distance larger than MaxInt64", prev: ptr(math.MinInt64), next: ptr(math.MaxInt64), wantRank: -1, wantOK: true},

		{name: "adjacent values", prev: ptr(1000), next: ptr(1001), wantOK: false},
		{name: "equal values", prev: ptr(1000), next: ptr(1000), wantOK: false},
		{name: "prev after next", prev: ptr(2000), next: ptr(1000), wantOK: false},
		{name: "append overflows", prev: ptr(math.MaxInt64 - rank.Gap + 1), wantOK: false},
		{name: "prepend underflows", next: ptr(math.MinInt64 + rank.Gap - 1), wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := rank.Between(tt.prev, tt.next)

			require.Equal(t, tt.wantOK, ok)
			if !tt.wantOK {
				return
			}
			require.Equal(t, tt.wantRank, got)
			if tt.prev != nil {
				require.Greater(t, got, *tt.prev)
			}
			if tt.next != nil {
				require.Less(t, got, *tt.next)
			}
		})
	}
}

func TestBetween_RepeatedInsertsExhaustGap(t *testing.T) {
	prev, next := int64(0), rank.Gap
	inserts := 0
	for {
		r, ok := rank.Between(&prev, &next)
		if !ok {
			break
		}
		next = r
		inserts++
	}

	require.Equal(t, 32, inserts)
}
