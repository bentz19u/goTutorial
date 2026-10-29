package source

import (
	"maps"
)

func (s *FakeSource) clone() *FakeSource {
	return &FakeSource{orders: maps.Clone(s.orders)}
}
