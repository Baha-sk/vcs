/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package common

import (
	"strconv"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/trustbloc/edge-core/pkg/log"
)

func TestDBParams(t *testing.T) {
	t.Run("valid params", func(t *testing.T) {
		expected := &DBParameters{
			URL:     "mem://test",
			Prefix:  "prefix",
			Timeout: 30,
		}
		setEnv(t, expected)
		cmd := &cobra.Command{}
		Flags(cmd)
		result, err := DBParams(cmd)
		require.NoError(t, err)
		require.Equal(t, expected, result)
	})

	t.Run("use default timeout", func(t *testing.T) {
		expected := &DBParameters{
			URL:     "mem://test",
			Prefix:  "prefix",
			Timeout: DatabaseTimeoutDefault,
		}
		setEnv(t, expected)
		t.Setenv(DatabaseTimeoutEnvKey, "")
		cmd := &cobra.Command{}
		Flags(cmd)
		result, err := DBParams(cmd)
		require.NoError(t, err)
		require.Equal(t, expected, result)
	})

	t.Run("error if url is missing", func(t *testing.T) {
		expected := &DBParameters{
			Prefix:  "prefix",
			Timeout: 30,
		}
		setEnv(t, expected)
		cmd := &cobra.Command{}
		Flags(cmd)
		_, err := DBParams(cmd)
		require.Error(t, err)
	})

	t.Run("error if prefix is missing", func(t *testing.T) {
		expected := &DBParameters{
			URL:     "mem://test",
			Timeout: 30,
		}
		setEnv(t, expected)
		cmd := &cobra.Command{}
		Flags(cmd)
		_, err := DBParams(cmd)
		require.Error(t, err)
	})

	t.Run("error if timeout has an invalid value", func(t *testing.T) {
		expected := &DBParameters{
			URL:    "mem://test",
			Prefix: "prefix",
		}
		setEnv(t, expected)
		t.Setenv(DatabaseTimeoutEnvKey, "invalid")
		cmd := &cobra.Command{}
		Flags(cmd)
		_, err := DBParams(cmd)
		require.Error(t, err)
	})
}

func TestInitStore(t *testing.T) {
	t.Run("store", func(t *testing.T) {
		t.Run("inits ok", func(t *testing.T) {
			t.Run("mem", func(t *testing.T) {
				s, err := InitStore(&DBParameters{
					URL:     "mem://test",
					Prefix:  "test",
					Timeout: 30,
				}, log.New("test"))
				require.NoError(t, err)
				require.NotNil(t, s)
			})
			t.Run("MongoDB", func(t *testing.T) {
				s, err := InitStore(&DBParameters{
					URL:     "mongodb://test",
					Prefix:  "test",
					Timeout: 30,
				}, log.New("test"))
				require.NoError(t, err)
				require.NotNil(t, s)
			})
		})

		t.Run("error if url format is invalid", func(t *testing.T) {
			_, err := InitStore(&DBParameters{
				URL:     "invalid",
				Prefix:  "test",
				Timeout: 30,
			}, log.New("test"))
			require.Error(t, err)
		})

		t.Run("error if driver is not supported", func(t *testing.T) {
			_, err := InitStore(&DBParameters{
				URL:     "unsupported://test",
				Prefix:  "test",
				Timeout: 30,
			}, log.New("test"))
			require.Error(t, err)
		})

		t.Run("error if cannot connect to store", func(t *testing.T) {
			invalid := []string{
				"mysql://test:secret@tcp(localhost:5984)",
				"couchdb://test:secret@localhost:5984",
			}

			for _, url := range invalid {
				_, err := InitStore(&DBParameters{
					URL:     url,
					Prefix:  "test",
					Timeout: 1,
				}, log.New("test"))
				require.Error(t, err)
			}
		})
	})
}

func setEnv(t *testing.T, values *DBParameters) {
	t.Helper()

	t.Setenv(DatabaseURLEnvKey, values.URL)
	t.Setenv(DatabasePrefixEnvKey, values.Prefix)
	t.Setenv(DatabaseTimeoutEnvKey, strconv.FormatUint(values.Timeout, 10))
}
