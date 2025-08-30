package validator

import (
	"errors"
)

// experimenting

type SliceV[T any] struct {
	slice []T
}

type LenCheck func(int) bool

func gt(num int) func(l int) bool {
	return func(l int) bool {
		return l > num
	}
}

func (s SliceV[T]) Length(err error, fs ...LenCheck) {
	for _, f := range fs {
		ok := f(len(s.slice))
		if !ok {
			if err == nil {
				panic("length is invalid")
			}
			panic(err)
		}
	}
}

func NewSliceV[T any](s []T) SliceV[T] {
	return SliceV[T]{slice: s}
}

func testValidation() (val int, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()

	slice := []int{1, 2, 3}
	NewSliceV(slice).Length(errors.New("length not as expected"), gt(2))
	return slice[2], err
}
