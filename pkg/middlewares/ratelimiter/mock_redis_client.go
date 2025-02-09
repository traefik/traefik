package ratelimiter

import (
	"context"
	"errors"
	"fmt"

	"github.com/mailgun/ttlmap"
	"github.com/redis/go-redis/v9"
	lua "github.com/yuin/gopher-lua"
)

// Mock redis client.
type MockRedisClient struct {
	ttl  int
	keys *ttlmap.TtlMap
}

func NewMockRedisClient(ttl int) Rediser {
	buckets, _ := ttlmap.NewConcurrent(65536)
	return &MockRedisClient{
		ttl:  ttl,
		keys: buckets,
	}
}

func (m *MockRedisClient) EvalSha(ctx context.Context, _ string, keys []string, args ...interface{}) *redis.Cmd {
	state := lua.NewState()
	defer state.Close()

	tableKeys := state.NewTable()
	for _, key := range keys {
		tableKeys.Append(lua.LString(key))
	}
	state.SetGlobal("KEYS", tableKeys)

	tableArgv := state.NewTable()
	for _, arg := range args {
		tableArgv.Append(lua.LString(fmt.Sprint(arg)))
	}
	state.SetGlobal("ARGV", tableArgv)

	mod := state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{
		"call": func(state *lua.LState) int {
			switch state.Get(1).String() {
			case "hset":
				key := state.Get(2).String()
				keyLast := state.Get(3).String()
				last := state.Get(4).String()
				keyTokens := state.Get(5).String()
				tokens := state.Get(6).String()
				table := []string{keyLast, last, keyTokens, tokens}
				_ = m.keys.Set(key, table, m.ttl)
			case "hgetall":
				key := state.Get(2).String()
				value, ok := m.keys.Get(key)
				table := state.NewTable()
				if !ok {
					state.Push(table)
				} else {
					switch v := value.(type) {
					case []string:
						if len(v) != 4 {
							break
						}
						for i := range v {
							table.Append(lua.LString(v[i]))
						}
					default:
						fmt.Printf("Unknown type: %T\n", v)
					}
					state.Push(table)
				}
			case "expire":
			default:
				return 0
			}

			return 1
		},
	})
	state.SetGlobal("redis", mod)
	state.Push(mod)

	cmd := redis.NewCmd(ctx)
	if err := state.DoString(AllowTokenBucketRaw); err != nil {
		cmd.SetErr(err)
		return cmd
	}

	result := state.Get(2)
	resultTable, ok := result.(*lua.LTable)
	if !ok {
		cmd.SetErr(errors.New("unexpected response type: " + result.String()))
		return cmd
	}

	var resultSlice []interface{}
	resultTable.ForEach(func(_ lua.LValue, value lua.LValue) {
		valueNbr, ok := value.(lua.LNumber)
		if !ok {
			valueStr, ok := value.(lua.LString)
			if !ok {
				cmd.SetErr(errors.New("unexpected response value type " + value.String()))
			}
			resultSlice = append(resultSlice, string(valueStr))
			return
		}

		resultSlice = append(resultSlice, int64(valueNbr))
	})

	cmd.SetVal(resultSlice)

	return cmd
}

func (m *MockRedisClient) Eval(ctx context.Context, script string, keys []string, args ...interface{}) *redis.Cmd {
	return m.EvalSha(ctx, script, keys, args...)
}

func (m *MockRedisClient) ScriptExists(ctx context.Context, hashes ...string) *redis.BoolSliceCmd {
	return nil
}

func (m *MockRedisClient) ScriptLoad(ctx context.Context, script string) *redis.StringCmd {
	return nil
}

func (m *MockRedisClient) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	return nil
}

func (m *MockRedisClient) EvalRO(ctx context.Context, script string, keys []string, args ...interface{}) *redis.Cmd {
	return nil
}

func (m *MockRedisClient) EvalShaRO(ctx context.Context, sha1 string, keys []string, args ...interface{}) *redis.Cmd {
	return nil
}
