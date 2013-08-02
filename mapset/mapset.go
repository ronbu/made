package mapset

import ()

type MapSet map[interface{}]interface{}

func NewSet(elems []interface{}) MapSet {
	s := make(MapSet, len(elems))
	for _, e := range elems {
		s[e] = struct{}{}
	}
	return s
}

func (m MapSet) AddF(elem interface{}) MapSet {
	c := m.Copy()
	return c.Add(elem)
}

func (m MapSet) Add(elem interface{}) MapSet {
	m[elem] = struct{}{}
	return m
}

func (m MapSet) Contains(elem interface{}) bool {
	_, ok := m[elem]
	return ok
}

func (m MapSet) Equals(o MapSet) bool {
	if len(m) != len(o) {
		return false
	}
	for k, _ := range o {
		if _, ok := m[k]; !ok {
			return false
		}
	}
	return true
}

func (m MapSet) Union(o MapSet) MapSet {
	s := make(MapSet)
	for k, v := range m {
		s[k] = v
	}
	for k, v := range o {
		s[k] = v
	}
	return s
}

func (m MapSet) Copy() MapSet {
	return m.Union(m)
}

func (m MapSet) Inter(o MapSet) MapSet {
	s := make(MapSet)
	for k, v := range m {
		if _, ok := o[k]; ok {
			s[k] = v
		}
	}
	return s
}

func (m MapSet) Diff(o MapSet) MapSet {
	s := make(MapSet)
	for k, v := range m {
		s[k] = v
	}
	for rk, _ := range o {
		if _, ok := m[rk]; ok {
			delete(s, rk)
		}
	}
	return s
}

func (m MapSet) SymDiff(o MapSet) MapSet {
	s := m.Union(o)
	inter := m.Inter(o)
	return s.Diff(inter)
}

func (m MapSet) Super(o MapSet) bool {
	for k, _ := range o {
		if _, ok := m[k]; !ok {
			return false
		}
	}
	return true
}

func (m MapSet) Sub(o MapSet) bool {
	if len(m.Diff(o)) == 0 {
		return true
	} else {
		return false
	}
}
