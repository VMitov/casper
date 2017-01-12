package diff

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/sergi/go-diff/diffmatchpatch"
)

type KVChange interface {
	Key() string
	Val() string
	String() string
	Pretty() string
}

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

type Pair struct {
	key string
	val string
}

func (p Pair) Key() string {
	return p.key
}

func (p Pair) Val() string {
	return p.val
}

type Add struct {
	Pair
}

func NewAdd(key, val string) *Add {
	return &Add{Pair{key, val}}
}

func (c Add) String() string {
	return fmt.Sprintf("+%v=%v", c.key, quoted(c.val))
}

func (c Add) Pretty() string {
	return fmt.Sprint(green(c.key), white("="), green(quoted(c.val)))
}

type Remove struct {
	Pair
}

func NewRemove(key, val string) *Remove {
	return &Remove{Pair{key, val}}
}

func (c Remove) String() string {
	return fmt.Sprintf("-%v=%v", c.key, c.val)
}

func (c Remove) Pretty() string {
	return fmt.Sprint(red(c.key), white("="), red(c.val))
}

type Update struct {
	Pair
	newVal string
}

func NewUpdate(key, val, newVal string) *Update {
	return &Update{Pair{key, val}, newVal}
}

func (c Update) NewVal() string {
	return c.newVal
}

func (c Update) String() string {
	return fmt.Sprintf("-%v=%v\n+%v=%v", c.key, c.val, c.key, quoted(c.newVal))
}

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
