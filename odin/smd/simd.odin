package main

import "core:fmt"
import "core:slice"
import "core:time"

FAT_ARRAY_SIZE: uint : 10_00_0003

LANES :: 8

Vec8 :: #simd[LANES]f32

WIDE_LANES :: 16

VecW :: #simd[WIDE_LANES]f32


/*
  without bound checks
*/
sequentialProcessing :: proc(health: []f32, deduction: f32) {
	for &value in health {
		value -= deduction
	}
}


simdProcessing :: proc(health: []f32, deduction: f32) {
	splat: [WIDE_LANES]f32
	for lane in 0 ..< WIDE_LANES {
		splat[lane] = deduction
	}
	damage: VecW = transmute(VecW)splat

	i := 0

	misalign := transmute(uintptr)&health[i] % 32
	if misalign != 0 {
		n := int((32 - misalign) / 4)
		for n > 0 && i < len(health) {
			health[i] -= deduction
			i += 1
			n -= 1
		}
	}

	for i + WIDE_LANES <= len(health) {
		ptr := (^VecW)(&health[i])
		v := ptr^
		v -= damage
		ptr^ = v
		i += WIDE_LANES
	}

	for i < len(health) {
		health[i] -= deduction
		i += 1
	}
}


floatsEqual :: proc(a, b: []f32) -> bool {
	if len(a) != len(b) {
		return false
	}
	for i in 0 ..< len(a) {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}


getBigFatArray :: proc() -> []f32 {
	health := make([]f32, FAT_ARRAY_SIZE)

	for _, idx in health {
		health[idx] = f32(50 + (idx % 100))
	}

	return health
}


/*
 this one hans bound check as and additional overhead
*/
linearProcessing :: proc(health: []f32, deduction: f32) {
	for _, i in health {
		health[i] -= deduction
	}
}


main :: proc() {
	base := getBigFatArray()

	fat_arr := slice.clone(base)
	fat_arr_clone := slice.clone(base)
	fat_arr_sequential := slice.clone(base)
	fat_arr_fixed := slice.clone(base)


	start := time.now()
	linearProcessing(fat_arr_clone, 10.32)
	fmt.println("linearProcessing version :: => ", time.since(start))

	start = time.now()
	sequentialProcessing(fat_arr_sequential, 10.32)
	fmt.println("sequentialProcessing version :: => ", time.since(start))

	start = time.now()
	simdProcessing(fat_arr_fixed, 10.32)
	fmt.println("simd processing version :: => ", time.since(start))

}
