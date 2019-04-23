package main

import (
	"os"
	"strconv"
	"strings"
)

type App struct {
	patterns  []*actionPattern
	context   *Context
	execCount int
}

type Context struct {
	Path string
	Args []string
	KV   map[string]string
}

func (c *Context) Get(key string) string {
	return c.KV[key]
}

func (c *Context) GetInt(key string) int {
	v, err := strconv.Atoi(c.KV[key])
	if err != nil {
		panic("value type illegal (expect int): " + key + " " + c.KV[key])
	}
	return v
}

type Handler func(*Context)

type actionPattern struct {
	commands []string
	handler  Handler
}

func (p *actionPattern) parse(pattern string, h Handler) *actionPattern {
	words := strings.Split(pattern, " ")
	p.commands = make([]string, len(words))
	for i, word := range words {
		p.commands[i] = word
	}
	p.handler = h
	return p
}

func (p *actionPattern) resolve(args []string) (m map[string]string, ok bool) {
	if len(args) != len(p.commands) {
		return
	}
	m = make(map[string]string)
	for i, cmd := range p.commands {
		arg := args[i]
		if strings.HasPrefix(cmd, ":") {
			m[cmd[1:]] = arg
		} else if arg != cmd {
			return
		}
	}
	ok = true
	return
}

func (app *App) Action(pattern string, h Handler) *App {
	app.patterns = append(app.patterns, new(actionPattern).parse(pattern, h))
	return app
}

func (app *App) Run() *App {
	for _, p := range app.patterns {
		if kv, ok := p.resolve(os.Args[1:]); ok {
			app.context = &Context{Path: os.Args[0], Args: os.Args[1:], KV: kv}
			p.handler(app.context)
			app.execCount++
			break
		}
	}
	if app.execCount == 0 {
		println("no action executed.")
	}
	return app
}
