package datastores

import (
	"github.com/iris-connect/eps"
)

var Definitions = eps.DatastoreDefinitions{
	"redis": eps.DatastoreDefinition{
		Name:              "Redis Datastore",
		Description:       "For Production Use",
		Maker:             MakeRedis,
		SettingsValidator: ValidateRedisSettings,
	},
	"file": eps.DatastoreDefinition{
		Name:              "File-based Datastore",
		Description:       "An file-based datastore",
		Maker:             MakeFile,
		SettingsValidator: ValidateFileSettings,
	},
}
