package consul

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/consul/api"
	yaml "gopkg.in/yaml.v2"
)

type ConsulAction int

const (
	ConsulAdd ConsulAction = iota
	ConsulUpdate
	ConsulRemove
)

type ConsulChange struct {
	Action ConsulAction
	Key    string
	Val    string
	NewVal string
}

func KVPairsToMap(pairs api.KVPairs) *NestedMap {
	j := &NestedMap{}
	for _, p := range pairs {
		j.Add(p.Key, string(p.Value))
	}
	return j
}

func GetChanges(pairs api.KVPairs, config []byte, format string) ([]ConsulChange, error) {
	kv, err := stringToMap(config, format)
	if err != nil {
		return nil, err
	}

	changes := []ConsulChange{}

	// Index current config and collect removals
	curKV := map[string]string{}
	for _, p := range pairs {
		curKV[p.Key] = string(p.Value)

		// Check for removal
		if _, ok := kv[p.Key]; !ok {
			changes = append(changes, ConsulChange{ConsulRemove, p.Key, string(p.Value), ""})
		}

	}

	// Collect additions
	for k, v := range kv {
		curVal, ok := curKV[k]
		if !ok {
			changes = append(changes, ConsulChange{ConsulAdd, k, "", v})
			continue
		}

		if v == curVal {
			continue
		}

		changes = append(changes, ConsulChange{ConsulUpdate, k, curVal, v})

	}

	return changes, nil
}

var errInvalidType = errors.New("Consul: unsupported template format")

func stringToMap(config []byte, format string) (map[string]string, error) {
	var err error
	j := &map[string]interface{}{}
	switch format {
	case "json", "jsonraw":
		err = json.Unmarshal(config, j)
	case "yaml":
		err = yaml.Unmarshal(config, j)
	default:
		err = errInvalidType
	}
	if err != nil {
		return nil, err
	}

	kv := map[string]string{}
	if err := flatten(*j, []string{}, &kv); err != nil {
		return nil, err
	}
	return kv, nil
}

type keyError struct {
	k interface{}
}

func (e keyError) Error() string {
	return fmt.Sprintf("Key %#v not convertible to string", e.k)
}

type valError struct {
	k string
	v interface{}
}

func (e valError) Error() string {
	return fmt.Sprintf("Type of the value of key %v:%#v not supported", e.k, e.v)
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
					return keyError{ki}
				}
				pairs[key] = vi
			}
			if err := flatten(pairs, append(prefixes, k), kv); err != nil {
				return err
			}
		default:
			key := strings.Join(append(prefixes, k), ".")
			return valError{key, v}
		}
	}
	return nil
}
