package main

import (
	"fmt"
)

type stringSliceFlag struct {
	value string
}

func (s *stringSliceFlag) Set(value string) error {
	fmt.Println(value)
	s.value = value
	return nil
}

func (s *stringSliceFlag) String() string {
	return s.value
}
