package rpcx

import "errors"

func ErrOnRefused(address string) error {
	return errors.New("connection refused")
}

func ErrOnNullParams() error {
	return errors.New("null params")
}
