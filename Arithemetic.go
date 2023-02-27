package main

type Arith struct{}

type Args struct {
	A, B int
}

func (t *Arith) Add(args *Args, reply *int) error {
	*reply = args.A + args.B
	return nil
}
