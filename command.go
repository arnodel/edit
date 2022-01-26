package edit

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

type CommandAction interface {
	Apply(w *App)
}
