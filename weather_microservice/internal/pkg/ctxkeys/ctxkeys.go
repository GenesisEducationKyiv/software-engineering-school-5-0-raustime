package ctxkeys

type contextKey string

func (k contextKey) String() string {
	return "ctxkey:" + string(k)
}

var Logger = contextKey("logger")
