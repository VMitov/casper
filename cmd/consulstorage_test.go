package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/hashicorp/consul/api"
	"github.com/miracl/casper/lib/diff"
)

func TestNewConsulStorage(t *testing.T) {
	testCases := []struct {
		addr string
		ok   bool
	}{
		{"", true},
		{"localhost:8500", true},
		{"http://192.168.0.%31/", false},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Case%v", i), func(t *testing.T) {
			_, err := NewConsulStorageConfig(map[string]interface{}{"addr": tc.addr})
			if tc.ok != (err == nil) {
				if err != nil {
					t.Fatal(err)
				} else {
					t.Fatal("Should fail")
				}
			}
		})
	}
}

type ConsulKVMock struct {
	list    api.KVPairs
	listErr error

	puts api.KVPairs
	dels []string
}

func (kv *ConsulKVMock) List(prefix string, q *api.QueryOptions) (api.KVPairs, *api.QueryMeta, error) {
	return kv.list, nil, kv.listErr
}

func (kv *ConsulKVMock) Put(p *api.KVPair, q *api.WriteOptions) (*api.WriteMeta, error) {
	kv.puts = append(kv.puts, p)
	return nil, nil
}

func (kv *ConsulKVMock) Delete(key string, w *api.WriteOptions) (*api.WriteMeta, error) {
	kv.dels = append(kv.dels, key)
	return nil, nil
}

var ConsulKVMockError = errors.New("ConsulKVMockError")

func TestConsulStorageString(t *testing.T) {
	testCases := []struct {
		list    api.KVPairs
		listErr error
		str     string
		err     error
	}{
		{
			api.KVPairs{
				&api.KVPair{Key: "key1", Value: []byte("val1")},
			},
			nil,
			`{"key1":"val1"}`,
			nil,
		},
		{
			api.KVPairs{
				&api.KVPair{Key: "key1", Value: []byte("val1")},
			},
			ConsulKVMockError,
			``,
			ConsulKVMockError,
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Case%v", i), func(t *testing.T) {
			s := &ConsulStorage{&ConsulKVMock{list: tc.list, listErr: tc.listErr}, []string{"jsonraw"}}
			str, err := s.String("jsonraw")
			if err != tc.err {
				t.Fatalf("Got %v; want %v", err, tc.err)
			}

			if str != tc.str {
				t.Errorf("Got `%v`; want `%v`", str, tc.str)
			}
		})
	}
}

func TestConsulStoragePush(t *testing.T) {
	testCases := []struct {
		list   api.KVPairs
		config string
		diff   string
		puts   api.KVPairs
		dels   []string
	}{
		{
			api.KVPairs{
				&api.KVPair{Key: "key1", Value: []byte("val1")},
				&api.KVPair{Key: "key2", Value: []byte("val2")},
				&api.KVPair{Key: "key3", Value: []byte("val")},
			},
			`{"key1":"val1","key3":"val3","key4":"val4"}`,
			"" +
				"-key2=val2\n" +
				"-key3=val\n" +
				"+key3=val3\n" +
				"+key4=val4\n",
			api.KVPairs{
				&api.KVPair{Key: "key3", Value: []byte("val3")},
				&api.KVPair{Key: "key4", Value: []byte("val4")},
			},
			[]string{"key2"},
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Case%v", i), func(t *testing.T) {
			s := &ConsulStorage{&ConsulKVMock{list: tc.list}, []string{"jsonraw	"}}

			cs, err := s.GetChanges([]byte(tc.config), "jsonraw", "")
			if err != nil {
				t.Fatal(err)
			}

			diff := s.Diff(cs, false)
			if diff != tc.diff {
				t.Errorf("Got `%v`; want `%v`", diff, tc.diff)
			}

			err = s.Push(cs)
			if err != nil {
				t.Fatal(err)
			}

			kv := s.kv.(*ConsulKVMock)
			if !reflect.DeepEqual(kv.dels, tc.dels) {
				t.Errorf("Got `%v`; want `%v`", kv.dels, tc.dels)
			}

			if len(kv.puts) != len(tc.puts) {
				t.Errorf("Got `%v`; want `%v`", kv.puts, tc.puts)
			}
			for _, e := range tc.puts {
				found := false
				for _, p := range kv.puts {
					if p.Key == e.Key && bytes.Compare(p.Value, e.Value) == 0 {
						found = true
						break
					}
				}

				if !found {
					if len(kv.puts) != len(tc.puts) {
						t.Errorf("%v missing from %v", e, kv.puts)
					}
				}
			}

		})
	}
}

func TestConsulStorageFormats(t *testing.T) {
	testCases := []struct {
		formats []string
		fmt     string
		valid   bool
		def     string
	}{
		{
			[]string{"fmt1", "fmt2"},
			"fmt1", true,
			"fmt1",
		},
		{
			[]string{"fmt1", "fmt2"},
			"fmt2", true,
			"fmt1",
		},
		{
			[]string{"fmt1", "fmt2"},
			"fmt3", false,
			"fmt1",
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Case%v", i), func(t *testing.T) {
			s := &ConsulStorage{&ConsulKVMock{}, tc.formats}

			if s.FormatIsValid(tc.fmt) != tc.valid {
				t.Errorf("%v should have been valid:%v", tc.fmt, tc.valid)
			}

			if s.DefaultFormat() != tc.def {
				t.Errorf("Default format should have been %v, not %v", tc.def, s.DefaultFormat())
			}
		})
	}
}

func TestKVPairsToString(t *testing.T) {
	testCases := []struct {
		pairs  api.KVPairs
		format string
		out    string
	}{
		{nil, "jsonraw", "{}"},
		{
			api.KVPairs{
				&api.KVPair{Key: "key1", Value: []byte("val1")},
			},
			"jsonraw", `{"key1":"val1"}`,
		},
		{
			api.KVPairs{
				&api.KVPair{Key: "key1", Value: []byte("val1")},
			},
			"json",
			"" +
				"{\n" +
				`  "key1": "val1"` + "\n" +
				"}",
		},
		{
			api.KVPairs{
				&api.KVPair{Key: "key1", Value: []byte("val1")},
			},
			"yaml",
			"key1: val1\n",
		},
		{
			api.KVPairs{
				&api.KVPair{Key: "key1", Value: []byte("val1")},
			},
			"invalid",
			"key1: val1\n",
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Case%v", i), func(t *testing.T) {
			out := kvPairsToString(tc.pairs, tc.format)
			if out != tc.out {
				t.Errorf("Got `%v`; want `%v`", out, tc.out)
			}
		})
	}
}

func TestGetChanges(t *testing.T) {
	testCases := []struct {
		pairs  api.KVPairs
		config string
		format string
		key    string
		ch     diff.KVChanges
		ok     bool
	}{
		{nil, "", "", "", nil, false},
		{
			api.KVPairs{
				&api.KVPair{Key: "key1", Value: []byte("val1")},
			},
			`{"key1": "val1"}`, "json", "",
			diff.KVChanges{},
			true,
		},
		{
			api.KVPairs{
				&api.KVPair{Key: "key1", Value: []byte("val1")},
			},
			`{"key1": "val1", "key2": "val2"}`, "json", "",
			diff.KVChanges{
				diff.NewAdd("key2", "val2"),
			},
			true,
		},
		{
			api.KVPairs{
				&api.KVPair{Key: "key1", Value: []byte("val1")},
				&api.KVPair{Key: "key2", Value: []byte("val2")},
			},
			`{"key1": "val1"}`, "json", "",
			diff.KVChanges{
				diff.NewRemove("key2", "val2"),
			},
			true,
		},
		{
			api.KVPairs{
				&api.KVPair{Key: "key1", Value: []byte("val1")},
				&api.KVPair{Key: "key2", Value: []byte("val")},
			},
			`{"key1": "val1", "key2": "val2"}`, "json", "",
			diff.KVChanges{
				diff.NewUpdate("key2", "val", "val2"),
			},
			true,
		},
		{
			api.KVPairs{
				&api.KVPair{Key: "key", Value: []byte("val")},
				&api.KVPair{Key: "key1", Value: []byte("val1")},
				&api.KVPair{Key: "key2", Value: []byte("val")},
			},
			`{"key1": "val1", "key2": "val2", "key3": "val3"}`, "json", "",
			diff.KVChanges{
				diff.NewRemove("key", "val"),
				diff.NewUpdate("key2", "val", "val2"),
				diff.NewAdd("key3", "val3"),
			},
			true,
		},
		{
			api.KVPairs{
				&api.KVPair{Key: "key", Value: []byte("val")},
				&api.KVPair{Key: "key1", Value: []byte("val1")},
				&api.KVPair{Key: "key2", Value: []byte("val")},
			},
			`{"key1": "val1", "key2": "val2", "key3": "val3"}`, "json", "key2",
			diff.KVChanges{
				diff.NewUpdate("key2", "val", "val2"),
			},
			true,
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Case%v", i), func(t *testing.T) {
			ch, err := getChanges(tc.pairs, []byte(tc.config), tc.format, tc.key)

			if tc.ok != (err == nil) {
				if err != nil {
					t.Fatal(err)
				} else {
					t.Fatal("Get should have failed but haven't")
				}
			}

			if tc.ok && diff.Diff(ch.(diff.KVChanges), false) != diff.Diff(tc.ch, false) {
				t.Errorf("Got `%v`; want `%v`", ch, tc.ch)
			}
		})
	}
}
