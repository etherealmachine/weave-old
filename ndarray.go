package main

import (
	"fmt"
)

type NDArray[V any] struct {
	extents []int
	array   []V
}

func NewNDArray[V any](extents ...int) *NDArray[V] {
	size := extents[0]
	for _, e := range extents[1:] {
		size *= e
	}
	return &NDArray[V]{
		extents: extents,
		array:   make([]V, size),
	}
}

func (a *NDArray[V]) Dims() int {
	return len(a.extents)
}

func (a *NDArray[V]) Size() int {
	return len(a.array)
}

func (a *NDArray[V]) Array() []V {
	return a.array
}

func (a *NDArray[V]) Index(coords ...int) int {
	index := coords[0]
	mult := 1
	for i, c := range coords[1:] {
		mult *= a.extents[i]
		index += c * mult
	}
	return index
}

func (a *NDArray[V]) Coords(index int) []int {
	coords := make([]int, a.Dims())
	mult := len(a.array)
	for i := len(coords) - 1; i >= 0; i-- {
		mult /= a.extents[i]
		coords[i] = index / mult
		index -= coords[i] * mult
	}
	return coords
}

func (a *NDArray[V]) At(coords ...int) V {
	if len(coords) != a.Dims() {
		panic(fmt.Errorf("wrong number of dimensions for ndarray: got %d, need %d", len(coords), a.Dims()))
	}
	return a.array[a.Index(coords...)]
}

func (a *NDArray[V]) Set(v V, dims ...int) {
	a.array[a.Index(dims...)] = v
}
