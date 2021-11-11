package demoServices

import (
	"context"
	"fmt"
)

type Args struct {
	A int
	B int
}

type Reply struct {
	C int
}

func Mul(ctx context.Context, args *Args, reply *Reply) error {
	reply.C = args.A * args.B
	fmt.Printf("call: %d * %d = %d\n", args.A, args.B, reply.C)
	return nil
}

func Add(ctx context.Context, args *Args, reply *Reply) error {
	reply.C = args.A + args.B
	fmt.Printf("call: %d + %d = %d\n", args.A, args.B, reply.C)
	return nil
}
