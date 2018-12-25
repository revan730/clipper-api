package log

type Logger interface {
	Info(i string)
	Error(msg string, err error)
	Fatal(msg string, err error)
}
