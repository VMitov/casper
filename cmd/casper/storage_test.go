package main

import (
	"fmt"
	"testing"
)

func TestGetStorage(t *testing.T) {
	testCases := []struct {
		t   string
		cfg map[string]interface{}
		err error
	}{
		{"test", nil, nil},
		{"invalid", nil, storageError("invalid")},
	}

	storages = map[string]func(map[string]interface{}) (storage, error){
		"test": newTestStorage,
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Case%v", i), func(t *testing.T) {
			s, err := getStorage(tc.t, tc.cfg)

			if err != nil {
				if err != tc.err {
					t.Fatalf("Got error %v; want %v", err.Error(), tc.err)
				}
				return
			}

			if tc.err != nil {
				t.Errorf("Got no error; want %v", tc.err)
			}

			cfg, _ := s.String("")
			if cfg != "TEST" {
				t.Errorf("Got %v; want %v", cfg, "TEST")
			}
		})
	}
}

type testStorage struct{}

func (testStorage) String(string) (string, error) {
	return "TEST", nil
}

func (s testStorage) FormatIsValid(format string) bool {
	return true
}

func (s testStorage) DefaultFormat() string {
	return "plain"
}

func newTestStorage(config map[string]interface{}) (storage, error) {
	return &testStorage{}, nil
}

type TestChange struct{}

func (c TestChange) Len() int {
	return 0
}

func (testStorage) GetChanges(config []byte, format, key string) (changes, error) {
	return TestChange{}, nil
}

func (testStorage) Diff(cs changes, pretty bool) string {
	return ""
}

func (testStorage) Push(cs changes) error {
	return nil
}
