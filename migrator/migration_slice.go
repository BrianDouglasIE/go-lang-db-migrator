package migrator

type Slice []*Migration

func (m Slice) Len() int {
	return len(m)
}

func (m Slice) Less(i, j int) bool {
	return m[i].Version < m[j].Version
}

func (m Slice) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}
