package cfdns

import (
	"context"
	"errors"
	"io"
)

// Iterator implements an iterator algorithm from a function that fetches
// blocks of data. This allows having fixed-memory usage when reading
// arbitrary-sized structs without leaking implementation details about
// how to paginate consecutive blocks of data.
type Iterator[T any] struct {
	fetchNext fetchFn[T]
	elements  []*T
	isLast    bool
}

type fetchFn[T any] func(ctx context.Context) (batch []*T, last bool, _ error)

// Next fetches the next item. If there are no more records io.EOF will be
// returned.
func (it *Iterator[T]) Next(ctx context.Context) (retElm *T, err error) {
	if len(it.elements) == 0 && !it.isLast {
		var elements []*T

		elements, it.isLast, err = it.fetchNext(ctx)
		if err != nil {
			return nil, err
		}

		it.elements = elements
	}

	if len(it.elements) == 0 {
		return nil, io.EOF
	}

	retElm = it.elements[0]
	it.elements = it.elements[1:]

	return retElm, nil
}

// ReadAll is an utility function that reads all elements from an iterator
// and return them as an array.
func ReadAll[T any](ctx context.Context, it *Iterator[T]) ([]*T, error) {
	ret := []*T{}

	for {
		item, err := it.Next(ctx)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return nil, err
		}

		ret = append(ret, item)
	}

	return ret, nil
}
