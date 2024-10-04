package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"time"
)

var (
	Ctx = context.Background()
	Rdb *redis.Client
)

func Init(Address string, Port string, Username string, Password string, ID int) error {
	Rdb = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", Address, Port),
		Username: Username,
		Password: Password,
		DB:       ID,
	})

	err := Rdb.Ping(Ctx).Err()
	if err != nil {
		return err
	}

	return nil
}

/*
функция для проверки существования таблицы в кэше

принимает:

	table - имя таблицы

возвращает:

	bool - true, если таблица существует, иначе false
	error - ошибка, если возникла
*/
func IsExistInCache(hashKey string) (bool, error) {
	exists, err := Rdb.Exists(Ctx, hashKey).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

/*
функция для записи данных в кэш как string, принимает любые данные на вход
*/
func SaveCache(hashKey string, data interface{}, ExTime time.Duration) error {
	err := Rdb.Set(Ctx, hashKey, data, ExTime).Err()
	if err != nil {
		return err
	}

	return nil
}

/*
Функция для чтения значений по хэш-ключу
*/
func ReadCache(hashKey string) (string, error) {
	response, err := Rdb.Get(Ctx, hashKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", nil
		}

		return "", err
	}

	return response, nil
}

/*
функция для записи map в кэш
*/
func SaveMapCache(hashKey string, dataMap any, ExTime time.Duration) error {
	marshaledMap, err := json.Marshal(dataMap)
	if err != nil {
		return err
	}

	err = Rdb.Set(Ctx, hashKey, marshaledMap, ExTime).Err()
	if err != nil {
		return err
	}

	return nil
}

/*
Функция для чтения map по хэш-ключу
*/
func ReadMapCache(hashKey string, response any) error {
	cacheData, err := Rdb.Get(Ctx, hashKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil
		}

		return err
	}

	err = json.Unmarshal([]byte(cacheData), &response)
	if err != nil {
		return err
	}

	return nil
}

/*
Функция для удаления значений по хэш-ключу
*/
func DeleteCache(hashKey string) error {
	err := Rdb.Del(Ctx, hashKey).Err()
	if err != nil {
		return err
	}

	return nil
}

/*
Функция для удаления значений по шаблону

пример pattern: news_category_*, где * - любое подстановочное значение
*/
func DeleteCacheByPattern(pattern string) error {
	var cursor uint64

	for {
		keys, nextCursor, err := Rdb.Scan(Ctx, cursor, pattern, 10).Result()
		if err != nil {
			return err
		}

		if len(keys) > 0 {
			err = Rdb.Del(Ctx, keys...).Err()
			if err != nil {
				return err
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return nil
}

/*
функция для стирания всего кэша

нужна в основном для дэбага
*/
func ClearCache(hashKey string) error {
	err := Rdb.Del(Ctx, hashKey).Err()
	if err != nil {
		return err
	}
	return nil
}

/*
Функция для чтения значений по хэш-ключу

возвращает grpc response
*/
func ReadProtoCache(hashKey string, m protoreflect.ProtoMessage) (protoreflect.ProtoMessage, error) {
	cacheData, err := Rdb.Get(Ctx, hashKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}

		return nil, err
	}

	err = proto.Unmarshal([]byte(cacheData), m)
	if err != nil {
		return nil, err
	}

	return m, nil
}

/*
функция для записи данных в кэш
*/
func SaveProtoCache(hashKey string, data protoreflect.ProtoMessage, ExTime time.Duration) error {
	cacheData, err := proto.Marshal(data)
	if err != nil {
		return nil
	}

	err = Rdb.Set(Ctx, hashKey, cacheData, ExTime).Err()
	if err != nil {
		return err
	}

	return nil
}
