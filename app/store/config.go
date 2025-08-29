package store

import (
	"fmt"
	"log"
)

// requrie even arguments, of key value pair
func (rs *RedisStore) InitConfig(configs ...string) {
	if len(configs)%2 != 0 {
		log.Fatal("init config: requrie even arguments, of key value pair")
	}
	for i := 0; i < len(configs); i += 2 {
		key := configs[i]
		val := configs[i+1]
		rs.Config[key] = val
	}
}

func (rs *RedisStore) ConfigGet(args []string) ([]string, error) {
	result := make([]string, 0, len(args)*2)
	for _, arg := range args {
		val, ok := rs.Config[arg]
		if !ok {
			return nil, fmt.Errorf("unknown config %s", arg)
		}
		result = append(result, arg)
		result = append(result, val)
	}
	return result, nil
}
