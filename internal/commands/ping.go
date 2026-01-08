package commands

func Ping(args []any) any {
	if len(args) == 1 {
		return args[0]
	}
	return "PONG"
}
