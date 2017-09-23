// Package diff is a utility to create visual diffs for key/value changes.
package diff

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/sergi/go-diff/diffmatchpatch"
)

// KVChange represent one change in key/value storage.
type KVChange interface {
	Key() string
	Val() string
	String() string
	Pretty() string
}

// KVChanges is a collection of key/value changes.
type KVChanges []KVChange

func (c KVChanges) Len() int {
	return len(c)
}

func (c KVChanges) Less(i, j int) bool {
	return c[i].Key() < c[j].Key()
}

func (c KVChanges) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

// Diff returns a visual representation of the changes.
func Diff(changes KVChanges, pretty bool) string {
	sort.Sort(changes)

	var buffer bytes.Buffer

	for _, c := range changes {
		var s string
		if pretty {
			s = c.Pretty()
		} else {
			s = c.String()
		}
		buffer.WriteString(s)
		buffer.WriteString("\n")
	}

	return buffer.String()
}

// Pair is a single key/value pair.
type Pair struct {
	key string
	val string
}

// Key returns the key of the Pair.
func (p Pair) Key() string {
	return p.key
}

// Val returns the value of the Pair.
func (p Pair) Val() string {
	return p.val
}

// Add is an addition change.
type Add struct {
	Pair
}

// NewAdd create Add.
func NewAdd(key, val string) *Add {
	return &Add{Pair{key, val}}
}

// String returns string representation of the addition.
func (c Add) String() string {
	return fmt.Sprintf("+%v=%v", c.key, quoted(c.val))
}

// Pretty returns colorful string representation of the addition.
func (c Add) Pretty() string {
	return fmt.Sprint(green(c.key), white("="), green(quoted(c.val)))
}

// Remove is an removal change.
type Remove struct {
	Pair
}

// NewRemove creates Remove.
func NewRemove(key, val string) *Remove {
	return &Remove{Pair{key, val}}
}

// String returns string representation of the removal.
func (c Remove) String() string {
	return fmt.Sprintf("-%v=%v", c.key, c.val)
}

// Pretty returns colorful string representation of the removal.
func (c Remove) Pretty() string {
	return fmt.Sprint(red(c.key), white("="), red(c.val))
}

// Update is an update change.
type Update struct {
	Pair
	newVal string
}

// NewUpdate creates update.
func NewUpdate(key, val, newVal string) *Update {
	return &Update{Pair{key, val}, newVal}
}

// NewVal return the new value of the update change.
func (c Update) NewVal() string {
	return c.newVal
}

// String returns string representation of the update.
func (c Update) String() string {
	return fmt.Sprintf("-%v=%v\n+%v=%v", c.key, c.val, c.key, quoted(c.newVal))
}

// Pretty returns colorful string representation of the update.
func (c Update) Pretty() string {
	dmp := diffmatchpatch.New()
	return fmt.Sprint(yellow(c.key), white("="), dmp.DiffPrettyText(dmp.DiffMain(c.val, c.newVal, false)))
}

func quoted(s string) string {
	if s == "" {
		return `""`
	}
	return s
}
