// Package consul contains helper functions for Consul storage.
package consul

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/consul/api"
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

// Action represent possible Consul changes.
type Action int

// Actions for Consul storage.
const (
	ConsulAdd Action = iota
	ConsulUpdate
	ConsulRemove
)

// Change represents single Consul change.
type Change struct {
	Action Action
	Key    string
	Val    string
	NewVal string
}

// KVPairsToMap creates NestedMap from Consul KVPairs.
func KVPairsToMap(pairs api.KVPairs) NestedMap {
	j := NestedMap{}
	for _, p := range pairs {
		j.Add(p.Key, string(p.Value))
	}
	return j
}

// GetChanges creates collection of changes from Consul KVPairs.
func GetChanges(pairs api.KVPairs, config []byte, format string) ([]Change, error) {
	kv, err := stringToMap(config, format)
	if err != nil {
		return nil, err
	}

	changes := []Change{}

	// index current config and collect removals
	curKV := map[string]string{}
	for _, p := range pairs {
		curKV[p.Key] = string(p.Value)

		// check for removal
		if _, ok := kv[p.Key]; !ok {
			changes = append(changes, Change{ConsulRemove, p.Key, string(p.Value), ""})
		}

	}

	// collect additions
	for k, v := range kv {
		curVal, ok := curKV[k]
		if !ok {
			changes = append(changes, Change{ConsulAdd, k, "", v})
			continue
		}

		if v == curVal {
			continue
		}

		changes = append(changes, Change{ConsulUpdate, k, curVal, v})

	}

	return changes, nil
}

func stringToMap(config []byte, format string) (map[string]string, error) {
	j := &map[string]interface{}{}
	switch format {
	case "json":
		if err := json.Unmarshal(config, j); err != nil {
			return nil, errors.Wrap(err, "parsing json failed")
		}
	case "yaml", "yml":
		if err := yaml.Unmarshal(config, j); err != nil {
			return nil, errors.Wrap(err, "parsing yaml failed")
		}
	default:
		return nil, errors.New("Consul: unsupported template format")
	}

	kv := map[string]string{}
	if err := flatten(*j, []string{}, &kv); err != nil {
		return nil, err
	}
	return kv, nil
}

func flatten(pairs map[string]interface{}, prefixes []string, kv *map[string]string) error {
	for k, v := range pairs {
		switch val := v.(type) {
		case string:
			if k == "_value" {
				k = ""
			}
			(*kv)[strings.Join(append(prefixes, k), "/")] = val
		case float64:
			if k == "_value" {
				k = ""
			}
			(*kv)[strings.Join(append(prefixes, k), "/")] = strconv.FormatFloat(val, 'f', -1, 64)
		case bool:
			if k == "_value" {
				k = ""
			}
			(*kv)[strings.Join(append(prefixes, k), "/")] = strconv.FormatBool(val)
		case map[string]interface{}:
			if err := flatten(val, append(prefixes, k), kv); err != nil {
				return err
			}
		case map[interface{}]interface{}:
			pairs := map[string]interface{}{}
			for ki, vi := range val {
				key, ok := ki.(string)
				if !ok {
					return fmt.Errorf("key %#v not convertible to string", ki)
				}
				pairs[key] = vi
			}
			if err := flatten(pairs, append(prefixes, k), kv); err != nil {
				return err
			}
		default:
			key := strings.Join(append(prefixes, k), ".")
			return fmt.Errorf("type of the value of key %v:%#v not supported", key, v)
		}
	}
	return nil
}
