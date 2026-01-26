package recursion

import (
	"context"
	"fmt"
	"slices"
	"strings"
)

type stackType int

const (
	stackKey stackType = iota
)

func CheckRecursion(ctx context.Context, itemType, itemName string) (context.Context, error) {
	currentStack, ok := ctx.Value(stackKey).([]string)
	if !ok {
		currentStack = []string{}
	}
	name := itemType + ":" + itemName
	if slices.Contains(currentStack, name) {
		return ctx, fmt.Errorf("could not instantiate %s %s: recursion detected in %s", itemType, itemName, strings.Join(append(currentStack, name), "->"))
	}
	return context.WithValue(ctx, stackKey, append(currentStack, name)), nil
}
