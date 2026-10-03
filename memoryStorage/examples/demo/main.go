package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/tanmay21k/goUtils/memoryStorage/stores"
)

func main() {
	ctx := context.Background()
	store, err := stores.NewkvStore(10)
	if err != nil {
		panic(err)
	}

	logger := log.New(os.Stdout, "[user] ", log.LstdFlags)
	out := make(chan error, len(example))
	Do(ctx, store, out, logger, example)

	for err := range out {
		logger.Println(err)
	}
}

// Do executes each operation as a slice of arguments against the provided store.
func Do(ctx context.Context, store stores.Store, out chan<- error, logger *log.Logger, ops [][]string) {
	if out == nil {
		return
	}

	var wg sync.WaitGroup
	for _, args := range ops {
		if len(args) == 0 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if logger != nil {
					logger.Println("empty operation")
				}
				out <- fmt.Errorf("empty operation")
			}()
			continue
		}

		wg.Add(1)
		go func(command []string) {
			defer wg.Done()
			if err := runCommand(ctx, store, command); err != nil {
				if logger != nil {
					logger.Println(err)
				}
				out <- err
			}
		}(args)
	}

	wg.Wait()
	close(out)
}

func runCommand(ctx context.Context, store stores.Store, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("empty operation")
	}

	command := strings.ToUpper(strings.TrimSpace(args[0]))
	params := args[1:]

	switch command {
	case "GET":
		if len(params) != 1 {
			return fmt.Errorf("GET requires 1 arg")
		}
		value, err := store.Get(ctx, params[0])
		if err != nil {
			return err
		}
		fmt.Printf("GET %s => %s\n", params[0], value)
		return nil

	case "SET":
		if len(params) != 2 {
			return fmt.Errorf("SET requires 2 args")
		}
		if err := store.Set(ctx, params[0], stores.Value(params[1])); err != nil {
			return err
		}
		fmt.Printf("SET %s => %s\n", params[0], params[1])
		return nil

	case "DEL":
		if len(params) != 1 {
			return fmt.Errorf("DEL requires 1 arg")
		}
		return store.Del(ctx, params[0])

	case "RENAME":
		if len(params) != 2 {
			return fmt.Errorf("RENAME requires 2 args")
		}
		keyStore, ok := store.(stores.KeyStore)
		if !ok {
			return stores.ErrUnsupportedOperation
		}
		return keyStore.Rename(ctx, params[0], params[1])

	case "POP":
		if len(params) != 1 {
			return fmt.Errorf("POP requires 1 arg")
		}
		keyStore, ok := store.(stores.KeyStore)
		if !ok {
			return stores.ErrUnsupportedOperation
		}
		value, err := keyStore.Pop(ctx, params[0])
		if err != nil {
			return err
		}
		fmt.Printf("POP %s => %s\n", params[0], value)
		return nil

	case "EXISTS":
		if len(params) != 1 {
			return fmt.Errorf("EXISTS requires 1 arg")
		}
		exister, ok := store.(stores.Exister)
		if !ok {
			return stores.ErrUnsupportedOperation
		}
		okVal, err := exister.Exists(ctx, params[0])
		if err != nil {
			return err
		}
		fmt.Printf("EXISTS %s => %t\n", params[0], okVal)
		return nil

	case "SCAN":
		if len(params) != 1 {
			return fmt.Errorf("SCAN requires 1 arg")
		}
		scanner, ok := store.(stores.Scanner)
		if !ok {
			return stores.ErrUnsupportedOperation
		}
		return scanner.Scan(ctx, params[0], func(key string) error {
			fmt.Printf("SCAN => %s\n", key)
			return nil
		})

	case "SETNX":
		if len(params) != 2 {
			return fmt.Errorf("SETNX requires 2 args")
		}
		conditionalStore, ok := store.(stores.ConditionalStore)
		if !ok {
			return stores.ErrUnsupportedOperation
		}
		okVal, err := conditionalStore.SetNX(ctx, params[0], stores.Value(params[1]))
		if err != nil {
			return err
		}
		fmt.Printf("SETNX %s => %t\n", params[0], okVal)
		return nil

	case "SETXX":
		if len(params) != 2 {
			return fmt.Errorf("SETXX requires 2 args")
		}
		conditionalStore, ok := store.(stores.ConditionalStore)
		if !ok {
			return stores.ErrUnsupportedOperation
		}
		okVal, err := conditionalStore.SetXX(ctx, params[0], stores.Value(params[1]))
		if err != nil {
			return err
		}
		fmt.Printf("SETXX %s => %t\n", params[0], okVal)
		return nil

	default:
		return fmt.Errorf("unsupported operation: %q", command)
	}
}
