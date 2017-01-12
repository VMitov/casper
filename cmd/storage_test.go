package cmd

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
		"test": NewTestStorage,
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

type TestStorage struct{}

func (TestStorage) String(string) (string, error) {
	return "TEST", nil
}

func (s TestStorage) FormatIsValid(format string) bool {
	return true
}

func (s TestStorage) DefaultFormat() string {
	return "plain"
}

func NewTestStorage(config map[string]interface{}) (storage, error) {
	return &TestStorage{}, nil
}

type TestChange struct{}

func (c TestChange) Len() int {
	return 0
}

func (TestStorage) GetChanges(config []byte, format, key string) (changes, error) {
	return TestChange{}, nil
}

func (TestStorage) Diff(cs changes, pretty bool) string {
	return ""
}

func (TestStorage) Push(cs changes) error {
	return nil
}
