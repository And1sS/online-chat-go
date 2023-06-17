package util

import (
	"fmt"
)

type RemovalError struct {
	msg string
}

func (re RemovalError) Error() string {
	return re.msg
}

func RemoveSwapElem[T comparable](arr *[]T, elem T) error {
	for i := 0; i < len(*arr); i++ {
		if (*arr)[i] == elem {
			return RemoveSwap(arr, i)
		}
	}

	return &RemovalError{msg: "Element not found"}
}

func RemoveSwap[T any](arr *[]T, idx int) error {
	derefArr := *arr
	length := len(derefArr)
	if idx < 0 {
		return &RemovalError{"Index is less than 0"}
	} else if idx >= length {
		return &RemovalError{fmt.Sprintf("Index: %d is greater than slice length: %d", idx, length)}
	}

	if idx != length-1 {
		derefArr[idx] = derefArr[length-1]
	}

	// erase old value, reasonable in case T is pointer, so we won't end up with memory leak
	var noop T
	derefArr[length-1] = noop

	*arr = derefArr[:length-1]
	return nil
}
