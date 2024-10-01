package migrations

type MigrationSlice []*Migration

func (m MigrationSlice) Len() int {
	return len(m)
}

func (m MigrationSlice) Less(i, j int) bool {
	return m[i].Version < m[j].Version
}

func (m MigrationSlice) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}
