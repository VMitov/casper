package consul

// Consul nested key separator
import "strings"

var (
	keySep       = "/"
	folderValKey = "_value"
)

// NestedMap is a map type for parsing Consul key/values in nested structure
type NestedMap map[string]interface{}

// Add value to NestedMap with nested key separated with keySep
func (j NestedMap) Add(path, value string) {
	isFolder := strings.HasSuffix(path, keySep)
	path = strings.TrimRight(path, keySep)
	j.add(path, value, isFolder)
	return
}

func (j NestedMap) add(path, value string, isFolder bool) {
	keyAndPath := strings.SplitN(path, keySep, 2)
	key := keyAndPath[0]

	// Recursively add the key if it is not leaf node
	if len(keyAndPath) > 1 {
		j.addPath(key, keyAndPath[1], value, isFolder)
		return
	}

	// Key is leaf node - add the value to the map

	// Check if key exists
	v, ok := j[key]
	if ok {
		switch nv := v.(type) {
		case NestedMap:
			nv[folderValKey] = value
			return
		}
	}

	if isFolder {
		j[key] = NestedMap{folderValKey: value}
		return
	}
	j[key] = value
}

// Add to key value with nested path
func (j NestedMap) addPath(key, path, value string, isFolder bool) {
	var n NestedMap

	// Check if key exists
	v, ok := j[key]
	if ok {
		switch nv := v.(type) {
		case NestedMap:
			n = nv
		default:
			n = NestedMap{folderValKey: nv}
		}
	} else {
		n = NestedMap{}
	}

	n.add(path, value, isFolder)
	j[key] = n
	return
}
