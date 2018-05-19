package cache

import (
	"strconv"

	"gopkg.in/redis.v4"
)

func NewCache(host, password string) *redis.Client {
	kv := redis.NewClient(&redis.Options{
		Addr:     host,
		Password: password,
		DB:       0,
	})
	return kv
}

func SliceStringToInt32(l []string) []int32 {
	ids := []int32{}
	for _, i := range l {
		id, err := strconv.Atoi(i)
		if err != nil {
			continue
		}
		ids = append(ids, int32(id))
	}

	return ids
}

func SplitStringSlice(slice []string, offset, count int32) []string {
	total := int32(len(slice))

	if offset < 0 {
		offset = 0
	}
	if count < 0 {
		count = 0
	}
	if offset+count > total {
		count = total - offset
	}

	return slice[offset:(offset + count)]
}
