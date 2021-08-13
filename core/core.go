package core

import "context"

// Pipeline types
type Sourcer func(context.Context) (<-chan *Node, <-chan error, error)
type Transformer func(context.Context, <-chan *Node) (<-chan *Node, <-chan error, error)
type Sinker func(context.Context, <-chan *Node) (<-chan error, error)
