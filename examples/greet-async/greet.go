package greetasync

import (
	"context"
	"strings"
	"time"

	"github.com/livetemplate/livetemplate"
)

type State struct {
	Name string
}

type Controller struct{}

func (c *Controller) Greet(s State, ctx *livetemplate.Context) (State, error) {
	name := strings.TrimSpace(ctx.GetString("name"))
	livetemplate.Async(ctx,
		func(context.Context) (string, error) {
			time.Sleep(700 * time.Millisecond)
			return name, nil
		},
		func(s State, name string, err error) (State, error) {
			s.Name = name
			return s, nil
		},
	)
	return s, nil
}
