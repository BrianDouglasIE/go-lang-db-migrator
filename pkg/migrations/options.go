package migrations

type MigrationOptions struct {
	To    int
	Up    int
	Down  int
	All   bool
	Clean bool
}
