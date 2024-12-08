package logutil

import (
	"context"
	"errors"
	"testing"

	"github.com/RyanBard/go-service-ex/internal/ctxutil"
	"github.com/stretchr/testify/assert"
)

func TestLogAttrSVC(t *testing.T) {
	t.Parallel()

	t.Run("should set the 'svc' key to the passed in service name", func(t *testing.T) {
		input := "foo"

		actual := LogAttrSVC(input)

		assert.Equal(t, "svc", actual.Key)
		assert.Equal(t, input, actual.Value.String())
	})
}

func TestLogAttrReqID(t *testing.T) {
	t.Parallel()

	t.Run("should set the 'reqID' key to the ContextKeyReqID in context", func(t *testing.T) {
		input := "foo"
		ctx := context.WithValue(context.Background(), ctxutil.ContextKeyReqID{}, input)

		actual := LogAttrReqID(ctx)

		assert.Equal(t, "reqID", actual.Key)
		assert.Equal(t, input, actual.Value.String())
	})

	t.Run("should set the 'reqID' key to empty string if ContextKeyReqID is not in context", func(t *testing.T) {
		ctx := context.Background()

		actual := LogAttrReqID(ctx)

		assert.Equal(t, "reqID", actual.Key)
		assert.Equal(t, "", actual.Value.String())
	})
}

func TestLogAttrFN(t *testing.T) {
	t.Parallel()

	t.Run("should set the 'fn' key to the passed in function name", func(t *testing.T) {
		input := "foo"

		actual := LogAttrFN(input)

		assert.Equal(t, "fn", actual.Key)
		assert.Equal(t, input, actual.Value.String())
	})
}

func TestLogAttrLoggedInUserID(t *testing.T) {
	t.Parallel()

	t.Run("should set the 'loggedInUserID' key to the ContextKeyReqID in context", func(t *testing.T) {
		input := "foo"
		ctx := context.WithValue(context.Background(), ctxutil.ContextKeyUserID{}, input)

		actual := LogAttrLoggedInUserID(ctx)

		assert.Equal(t, "loggedInUserID", actual.Key)
		assert.Equal(t, input, actual.Value.String())
	})

	t.Run("should set the 'reqID' key to empty string if ContextKeyReqID is not in context", func(t *testing.T) {
		ctx := context.Background()

		actual := LogAttrLoggedInUserID(ctx)

		assert.Equal(t, "loggedInUserID", actual.Key)
		assert.Equal(t, "", actual.Value.String())
	})
}

func TestLogAttrError(t *testing.T) {
	t.Parallel()

	t.Run("should set the 'error' key to the string representation of the passed in error", func(t *testing.T) {
		input := "foo"
		err := errors.New(input)

		actual := LogAttrError(err)

		assert.Equal(t, "error", actual.Key)
		assert.Equal(t, input, actual.Value.String())
	})
}

func TestParseLevel(t *testing.T) {
	t.Parallel()

	t.Run("should understand 'debug'", func(t *testing.T) {
		lvl, err := ParseLevel("debug")
		assert.Nil(t, err)
		assert.NotNil(t, lvl)
	})

	t.Run("should understand 'DEBUG'", func(t *testing.T) {
		lvl, err := ParseLevel("DEBUG")
		assert.Nil(t, err)
		assert.NotNil(t, lvl)
	})

	t.Run("should understand 'info'", func(t *testing.T) {
		lvl, err := ParseLevel("info")
		assert.Nil(t, err)
		assert.NotNil(t, lvl)
	})

	t.Run("should understand 'INFO'", func(t *testing.T) {
		lvl, err := ParseLevel("INFO")
		assert.Nil(t, err)
		assert.NotNil(t, lvl)
	})

	t.Run("should understand 'warn'", func(t *testing.T) {
		lvl, err := ParseLevel("warn")
		assert.Nil(t, err)
		assert.NotNil(t, lvl)
	})

	t.Run("should understand 'WARN'", func(t *testing.T) {
		lvl, err := ParseLevel("WARN")
		assert.Nil(t, err)
		assert.NotNil(t, lvl)
	})

	t.Run("should understand 'warning'", func(t *testing.T) {
		lvl, err := ParseLevel("warning")
		assert.Nil(t, err)
		assert.NotNil(t, lvl)
	})

	t.Run("should understand 'WARNING'", func(t *testing.T) {
		lvl, err := ParseLevel("WARNING")
		assert.Nil(t, err)
		assert.NotNil(t, lvl)
	})

	t.Run("should understand 'error'", func(t *testing.T) {
		lvl, err := ParseLevel("error")
		assert.Nil(t, err)
		assert.NotNil(t, lvl)
	})

	t.Run("should understand 'ERROR'", func(t *testing.T) {
		lvl, err := ParseLevel("ERROR")
		assert.Nil(t, err)
		assert.NotNil(t, lvl)
	})

	t.Run("should not understand 'trace'", func(t *testing.T) {
		lvl, err := ParseLevel("trace")
		assert.NotNil(t, err)
		assert.Nil(t, lvl)
	})

	t.Run("should not understand 'fatal'", func(t *testing.T) {
		lvl, err := ParseLevel("fatal")
		assert.NotNil(t, err)
		assert.Nil(t, lvl)
	})

	t.Run("should not understand 'panic'", func(t *testing.T) {
		lvl, err := ParseLevel("panic")
		assert.NotNil(t, err)
		assert.Nil(t, lvl)
	})
}
