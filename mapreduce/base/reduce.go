package base

import (
	"encoding/json"
	"log"
	"os"
	"sort"
)

func Reduce(files []*json.Decoder, args ReduceTaskArgs,
	reduceF func(key string, value []Value) string) {

	kvs := make(map[string][]Value)

	for _, file := range files {
		for {
			var pair Pair
			err := file.Decode(&pair)
			if err != nil {
				break
			}
			_, ok := kvs[pair.Key]
			if !ok {
				kvs[pair.Key] = []Value{}
			}
			kvs[pair.Key] = append(kvs[pair.Key], pair.Val)
		}
	}
	// 排序
	var keys []string
	for k := range kvs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	resultFileName := GetOutputFileName(args.TaskId, args.ReducePhaseIndex)
	file, err := os.Create(resultFileName)
	if err != nil {
		log.Println("文件创建失败")
	}
	enc := json.NewEncoder(file)
	for _, k := range keys {
		res := reduceF(k, kvs[k])
		enc.Encode(Pair{k, res})
	}
	file.Close()
}
