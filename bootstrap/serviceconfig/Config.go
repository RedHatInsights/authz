package serviceconfig

import "time"

// Config defines an interface that's capable to provide config for interaction in the domain layer without
// tightly coupling to a config implementation.
// see https://github.com/aellwein/config/blob/master/examples/example.go
// NOTE: this interface is most likely too broad and too generic, so take it as an exmaple.
// we want to look at approaches based on structs using
// a tool such as https://github.com/ilyakaznacheev/cleanenv/ i guess.
type Config interface {
	GetAll() map[string]interface{}
	Get(key string) interface{}
	GetString(key string) string
	GetBool(key string) bool
	GetInt(key string) int
	GetInt32(key string) int32
	GetInt64(key string) int64
	GetUint(key string) uint
	GetUint32(key string) uint32
	GetUint64(key string) uint64
	GetFloat64(key string) float64
	GetDuration(key string) time.Duration
	GetIntSlice(key string) []int
	GetSizeInBytes(key string) uint
	GetTime(key string) time.Time
	GetStringSlice(key string) []string
}
