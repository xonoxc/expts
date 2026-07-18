package command

import (
	"errors"
	"strconv"
	"time"

	"github.com/xonoxc/expts/redis-recreation/internal/resp"
	"github.com/xonoxc/expts/redis-recreation/internal/store"
)

var (
	ErrrUnkownCommand        = errors.New("command now found")
	ErrNotEnoughArguements   = errors.New("not enough arguements")
	ErrInvalidExpirationTime = errors.New("invalid expiration time")
	ErrInvalidSyntax         = errors.New("invalid syntax")
)

// this can potentially be an interface
// but for now i don't care lol.
type Dispatcher struct {
	repo *store.Store
}

func NewDispatcher(ste *store.Store) *Dispatcher {
	return &Dispatcher{
		repo: ste,
	}
}

func (dsp *Dispatcher) Dispatch(cmd Command) ([]byte, error) {
	switch cmd.Name {
	case "SET":
		if len(cmd.Args) == 2 {
			dsp.repo.Set(cmd.Args[0], cmd.Args[1])
			return resp.SerializeSimpleString("OK"), nil
		}

		if len(cmd.Args) == 4 {
			if cmd.Args[2] != "EX" {
				return nil, ErrInvalidSyntax
			}

			expInt, err := strconv.Atoi(cmd.Args[3])
			if err != nil {
				return nil, ErrInvalidExpirationTime
			}

			dsp.repo.SetWithExpiration(cmd.Args[0], cmd.Args[1], time.Duration(expInt)*time.Second)
			return resp.SerializeSimpleString("OK"), nil
		}

		return nil, ErrInvalidSyntax

	case "GET":
		if len(cmd.Args) != 1 {
			return nil, ErrNotEnoughArguements
		}

		rsp, exists := dsp.repo.Get(cmd.Args[0])
		if !exists {
			return resp.SerializeNull(), nil
		}

		return resp.SerializeBulkString(rsp.Data), nil

	case "DELETE":
		dsp.repo.Delete(
			cmd.Args[0],
		)
		return resp.SerializeInteger(1), nil

	default:
		return nil, ErrrUnkownCommand
	}
}
