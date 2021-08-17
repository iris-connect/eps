package eps

type DatastoreDefinition struct {
	Name              string            `json:"name"`
	Description       string            `json:"description"`
	Maker             DatastoreMaker    `json:"-"`
	SettingsValidator SettingsValidator `json:"-"`
}

type DatastoreDefinitions map[string]DatastoreDefinition
type DatastoreMaker func(settings interface{}) (Datastore, error)

type Datastore interface {
	// Write data to the store
	Write(*DataEntry) error
	// Read data from the store
	Read() ([]*DataEntry, error)
	Init() error
}

const (
	NullType = 0
)

type DataEntry struct {
	Type uint8
	ID   []byte
	Data []byte
}
