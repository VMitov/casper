package main

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/hashicorp/consul/api"
	"github.com/miracl/casper/lib/consul"
	"github.com/miracl/casper/lib/diff"
	yaml "gopkg.in/yaml.v2"
)

const (
	defaultIgnoreVal = "_ignore"
)

// ConsulKV is interface that consul KV type implements.
// Defined and used mainly for testing.
type ConsulKV interface {
	List(prefix string, q *api.QueryOptions) (api.KVPairs, *api.QueryMeta, error)
	Put(p *api.KVPair, q *api.WriteOptions) (*api.WriteMeta, error)
	Delete(key string, w *api.WriteOptions) (*api.WriteMeta, error)
}

var consulFormats = []string{"json", "yaml", "jsonraw"}

type consulStorage struct {
	kv ConsulKV

	formats   []string
	ignoreVal string
}

type changeError struct {
	c interface{}
}

func (e changeError) Error() string {
	return fmt.Sprintf("Consul: Invalid change type: %T", e.c)
}

func newConsulStorage(addr string) (storage, error) {
	cfg := &api.Config{}

	ignoreVal := defaultIgnoreVal

	if addr != "" {
		addr, err := url.Parse(addr)
		if err != nil {
			return nil, err
		}
		cfg.Address = addr.Host
		cfg.Scheme = addr.Scheme
		cfg.Token = addr.Query().Get("token")

		ignore := addr.Query().Get("ignore")
		if ignore != "" {
			ignoreVal = ignore
		}
	}

	client, err := api.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	return &consulStorage{client.KV(), consulFormats, ignoreVal}, nil
}

func (s consulStorage) String(format string) (string, error) {
	pairs, _, err := s.kv.List("", nil)
	if err != nil {
		return "", err
	}
	return kvPairsToString(pairs, format), nil
}

func (s consulStorage) FormatIsValid(format string) bool {
	for _, f := range s.formats {
		if format == f {
			return true
		}
	}
	return false
}

func (s consulStorage) DefaultFormat() string {
	return s.formats[0]
}

func (s consulStorage) GetChanges(config []byte, format, key string) (changes, error) {
	pairs, _, err := s.kv.List("", nil)
	if err != nil {
		return nil, err
	}

	return getChanges(pairs, config, format, key, s.ignoreVal)
}

func (consulStorage) Diff(cs changes, pretty bool) string {
	return diff.Diff(cs.(diff.KVChanges), pretty)
}

func (s consulStorage) Push(cs changes) error {
	for _, ci := range cs.(diff.KVChanges) {
		if err := s.push(ci); err != nil {
			return err
		}
	}
	return nil
}

func (s consulStorage) push(change interface{}) error {
	switch c := change.(type) {
	case *diff.Add:
		_, err := s.kv.Put(&api.KVPair{Key: c.Key(), Value: []byte(c.Val())}, nil)
		return err
	case *diff.Update:
		_, err := s.kv.Put(&api.KVPair{Key: c.Key(), Value: []byte(c.NewVal())}, nil)
		return err
	case *diff.Remove:
		_, err := s.kv.Delete(c.Key(), nil)
		return err
	}

	return changeError{change}
}

func kvPairsToString(pairs api.KVPairs, format string) string {
	j := consul.KVPairsToMap(pairs)

	var res []byte
	switch format {
	case "json":
		res, _ = json.MarshalIndent(j, "", "  ")
	case "jsonraw":
		res, _ = json.Marshal(j)
	default:
		res, _ = yaml.Marshal(j)

	}

	return string(res)
}

func getChanges(pairs api.KVPairs, config []byte, format, key, ignoreVal string) (changes, error) {
	consulChanges, err := consul.GetChanges(pairs, config, format)
	if err != nil {
		return nil, err
	}

	kvChanges := diff.KVChanges{}
	for _, c := range consulChanges {
		// Skip ignored pairs
		if ignoreVal != "" && c.NewVal == ignoreVal {
			continue
		}

		if key != "" && key != c.Key {
			continue
		}

		switch c.Action {
		case consul.ConsulAdd:
			kvChanges = append(kvChanges, diff.NewAdd(c.Key, c.NewVal))
		case consul.ConsulRemove:
			kvChanges = append(kvChanges, diff.NewRemove(c.Key, c.Val))
		case consul.ConsulUpdate:
			kvChanges = append(kvChanges, diff.NewUpdate(c.Key, c.Val, c.NewVal))
		}

	}

	return kvChanges, nil
}
