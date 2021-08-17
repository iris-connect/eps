package datastores

import (
	"github.com/go-redis/redis"
	"github.com/iris-connect/eps"
	"github.com/kiprotect/go-helpers/forms"
	"sync"
	"time"
)

type Redis struct {
	client   redis.UniversalClient
	options  redis.UniversalOptions
	settings *RedisSettings
	mutex    sync.Mutex
	index    int64
}

type RedisSettings struct {
	MasterName string   `json:"master_name"`
	Addresses  []string `json:"addresses`
	Database   int64    `json:"database"`
	Key        string   `json:"key"`
	Password   string   `json:"password"`
}

var RedisForm = forms.Form{
	ErrorMsg: "invalid data encountered in the Redis config form",
	Fields: []forms.Field{
		{
			Name: "addresses",
			Validators: []forms.Validator{
				forms.IsRequired{},
				forms.IsStringList{},
			},
		},
		{
			Name: "key",
			Validators: []forms.Validator{
				forms.IsOptional{Default: "entries"},
				forms.IsString{},
			},
		},
		{
			Name: "master_name",
			Validators: []forms.Validator{
				forms.IsOptional{Default: ""},
				forms.IsString{},
			},
		},
		{
			Name: "database",
			Validators: []forms.Validator{
				forms.IsOptional{Default: 0},
				forms.IsInteger{Min: 0, Max: 100},
			},
		},
		{
			Name: "password",
			Validators: []forms.Validator{
				forms.IsRequired{},
				forms.IsString{},
			},
		},
	},
}

func ValidateRedisSettings(settings map[string]interface{}) (interface{}, error) {
	if params, err := RedisForm.Validate(settings); err != nil {
		return nil, err
	} else {
		redisSettings := &RedisSettings{}
		if err := RedisForm.Coerce(redisSettings, params); err != nil {
			return nil, err
		}
		return redisSettings, nil
	}
}

func MakeRedis(settings interface{}) (eps.Datastore, error) {

	redisSettings := settings.(RedisSettings)

	options := redis.UniversalOptions{
		MasterName:   redisSettings.MasterName,
		Password:     redisSettings.Password,
		ReadTimeout:  time.Second * 1.0,
		WriteTimeout: time.Second * 1.0,
		Addrs:        redisSettings.Addresses,
		DB:           int(redisSettings.Database),
	}

	client := redis.NewUniversalClient(&options)

	if _, err := client.Ping().Result(); err != nil {
		return nil, err
	} else {
		eps.Log.Info("Ping to Redis succeeded!")
	}

	datastore := &Redis{
		options:  options,
		client:   client,
		settings: &redisSettings,
		index:    -1,
	}

	return datastore, nil

}

func (d *Redis) Write(entry *eps.DataEntry) error {
	bytes := ToBytes(entry)
	if err := d.Client().RPush(d.settings.Key, string(bytes)).Err(); err != nil {
		return err
	}
	return nil
}

func (d *Redis) Read() ([]*eps.DataEntry, error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	results, err := d.Client().LRange(d.settings.Key, d.index+1, -1).Result()
	if err != nil {
		return nil, err
	} else {
		eps.Log.Tracef("Read %d new entries", len(results))
		entries := make([]*eps.DataEntry, 0, len(results))
		for _, result := range results {
			if entry, err := FromBytes([]byte(result)); err != nil {
				return nil, err
			} else {
				entries = append(entries, entry)
			}
		}
		d.index += int64(len(results))
		return entries, nil
	}
}

func (d *Redis) Init() error {
	return nil
}

func (d *Redis) Client() redis.Cmdable {
	return d.client
}

func (d *Redis) Open() error {
	return nil
}

func (d *Redis) Close() error {
	return d.client.Close()
}

/*
#	result, err := r.db.Client().HGetAll(string(r.fullKey)).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, NotFound
		}
		return nil, err
	}
	byteMap := map[string][]byte{}
	for k, v := range result {
		byteMap[k] = []byte(v)
	}
	return byteMap, nil
*/
