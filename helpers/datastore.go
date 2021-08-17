package helpers

import (
	"github.com/iris-connect/eps"
)

func InitializeDatastore(settings *eps.DatastoreSettings, definitions *eps.Definitions) (eps.Datastore, error) {
	definition := definitions.DatastoreDefinitions[settings.Type]
	return definition.Maker(settings.Settings)
}
