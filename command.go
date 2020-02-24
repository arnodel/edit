package main

type ArgType interface {
	FromString() interface{}
}

type Parameter struct {
	Name        string
	Description string
	Type        ArgType
}

type Command struct {
	Name        string
	Description string
	Parameters  []Parameter
	Action      CommandAction
}

type Invocation struct {
	*Command
	Arguments []interface{}
}

type CommandAction func(w *App, args []interface{})

var insertRune = &Command{
	Name:        "insert-rune",
	Description: "Insert a rune",
	Action: func(a *App, args []interface{}) {
		a.window.InsertRune(args[0].(rune))
	},
}
