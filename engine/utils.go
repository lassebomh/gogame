package engine

import (
	"cmp"
	"fmt"
	"slices"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func IterMapSorted[K cmp.Ordered, V any](m map[K]V) func(yield func(K, V) bool) {
	return func(yield func(K, V) bool) {

		keys := make([]K, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		slices.Sort(keys)

		for _, k := range keys {

			if !yield(k, m[k]) {
				break
			}
		}

	}
}

func IterMapSortedLeftValuePairs[K cmp.Ordered, V any](left map[K]*V, right map[K]*V) func(yield func(*V, *V) bool) {
	return func(yield func(*V, *V) bool) {

		keys := make([]K, 0, len(left))
		for k := range left {
			keys = append(keys, k)
		}
		slices.Sort(keys)

		for _, k := range keys {

			l := left[k]
			r, ok := right[k]

			if !ok {
				r = l
			}

			if !yield(l, r) {
				break
			}
		}

	}
}

func Lerp(a, b, t float32) float32 {
	return a + t*(b-a)
}

func Vector2Lerp(a rl.Vector2, b rl.Vector2, t float32) rl.Vector2 {
	tv := rl.Vector2{X: t, Y: t}
	return rl.Vector2Add(a, rl.Vector2Multiply(tv, (rl.Vector2Subtract(b, a))))
}

func Dbg(args ...any) {
	for _, arg := range args {
		fmt.Printf("%+v ", arg)
	}
	fmt.Print("\n")
}
