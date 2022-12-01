package main

import (
	"log"
	"reflect"
	"testing"
)

func Test1DArray(t *testing.T) {
	a := NewNDArray[int](10)
	if got, want := a.Dims(), 1; got != want {
		t.Fatalf("wrong dimensions, got %d, want %d", got, want)
	}
	if got, want := a.Size(), 10; got != want {
		t.Fatalf("wrong size, got %d, want %d", got, want)
	}
	a.Set(8, 4)
	if got, want := a.At(4), 8; got != want {
		t.Fatalf("a[4], got %d, want %d", got, want)
	}
}

func Test2DArray(t *testing.T) {
	a := NewNDArray[int](2, 3)
	if got, want := a.Dims(), 2; got != want {
		t.Fatalf("wrong dimensions, got %d, want %d", got, want)
	}
	if got, want := a.Size(), 6; got != want {
		t.Fatalf("wrong size, got %d, want %d", got, want)
	}
	a.Set(8, 1, 2)
	if got, want := a.At(1, 2), 8; got != want {
		t.Fatalf("a[1, 2], got %d, want %d", got, want)
	}
	if got, want := a.Index(0, 1), 2; got != want {
		t.Fatalf("wrong index, got %d, want %d", got, want)
	}
	if got, want := a.Coords(2), []int{0, 1}; !reflect.DeepEqual(got, want) {
		t.Fatalf("wrong coords, got %v, want %v", got, want)
	}
}

func Test3DArray(t *testing.T) {
	a := NewNDArray[int](4, 3, 2)
	if got, want := a.Dims(), 3; got != want {
		t.Fatalf("wrong dimensions, got %d, want %d", got, want)
	}
	if got, want := a.Size(), 24; got != want {
		t.Fatalf("wrong size, got %d, want %d", got, want)
	}
	a.Set(8, 3, 2, 1)
	if got, want := a.At(3, 2, 1), 8; got != want {
		t.Fatalf("a[1, 2], got %d, want %d", got, want)
	}
	if got, want := a.Index(3, 2, 1), 23; got != want {
		t.Fatalf("wrong index, got %d, want %d", got, want)
	}
	if got, want := a.Coords(23), []int{3, 2, 1}; !reflect.DeepEqual(got, want) {
		t.Fatalf("wrong coords, got %v, want %v", got, want)
	}
}

func Test4DArray(t *testing.T) {
	a := NewNDArray[int](2, 3, 4, 5)
	if got, want := a.Dims(), 4; got != want {
		t.Fatalf("wrong dimensions, got %d, want %d", got, want)
	}
	if got, want := a.Size(), 120; got != want {
		t.Fatalf("wrong size, got %d, want %d", got, want)
	}
	a.Set(8, 1, 2, 3, 4)
	if got, want := a.At(1, 2, 3, 4), 8; got != want {
		t.Fatalf("a[1, 2, 3, 4], got %d, want %d", got, want)
	}
	if got, want := a.Coords(a.Index(0, 1, 2, 3)), []int{0, 1, 2, 3}; !reflect.DeepEqual(got, want) {
		t.Fatalf("wrong coords, got %v, want %v", got, want)
	}
	for i, v := range a.Array() {
		log.Println(a.Coords(i), i, v)
	}
}
