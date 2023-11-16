package cfdns

import (
	"context"
	"errors"
)

const itemsPerPage = 500

var Done = errors.New("done")

// Iterator implements an iterator algorithm from a function that fetches
// blocks of data. This allows having fixed-memory usage when reading
// arbitrary-sized structs without leaking implementation details about
// how to paginate consecutive blocks of data.
type Iterator[T any] struct {
	fetchNext FetchFn[T]
	elements  []T
	isLast    bool
}

type FetchFn[T any] func(ctx context.Context) (batch []T, last bool, _ error)

func (it *Iterator[T]) Next(ctx context.Context) (retElm T, err error) {
	if len(it.elements) == 0 && !it.isLast {
		var elements []T
		elements, it.isLast, err = it.fetchNext(ctx)
		if err != nil {
			return
		}

		it.elements = elements
	}

	if len(it.elements) == 0 {
		err = Done
		return
	}

	retElm = it.elements[0]
	it.elements = it.elements[1:]
	return retElm, nil
}
