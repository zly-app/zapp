package utils

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGo_GoAndWait(t *testing.T) {
	err := Go.GoAndWait()
	require.Nil(t, err)

	err = Go.GoAndWait(func() error {
		return nil
	})
	require.Nil(t, err)

	err = Go.GoAndWait(func() error {
		return nil
	}, func() error {
		return errors.New("2")
	}, func() error {
		return errors.New("3")
	})
	require.NotNil(t, err)
}
