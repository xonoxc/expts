package command

import (
	"errors"
	"strconv"
	"time"

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
			return nil, nil
		}

		// if the args have ['key' , 'value',  'EX' , 'expiration_time']
		// i don't support PX or anything beyound set with expiration
		// because this is my recreation i don't have time for perfectionism.
		// i just want to learn.
		if len(cmd.Args) == 4 {
			if cmd.Args[2] != "EX" {
				return nil, ErrInvalidSyntax
			}

			expInt, err := strconv.Atoi(cmd.Args[3])
			if err != nil {
				return nil, ErrInvalidExpirationTime
			}

			dsp.repo.SetWithExpiration(cmd.Args[0], cmd.Args[1], time.Duration(expInt)*time.Second)
		}

		return nil, nil

	case "GET":
		if len(cmd.Args) != 1 {
			return nil, ErrNotEnoughArguements
		}

		resp, exists := dsp.repo.Get(cmd.Args[0])
		if !exists {
			return []byte("(nil)"), nil
		}

		return []byte(resp.Data), nil

	case "DELETE":
		dsp.repo.Delete(cmd.Args[0])
		return nil, nil

	default:
		return nil, ErrrUnkownCommand
	}
}
