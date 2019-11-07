package hydrautil

type logger interface {
	Printf(string, ...interface{})
	Println(...interface{})
}

type debugLogger interface {
	logger
	Debugln(...interface{})
	Debugf(string, ...interface{})
}

var lg logger
var lgd debugLogger

// SetLogger sets a logger on the package that will print messages.
//
// The logger must at least implement methods with the same signatures
// as the Println and Printf methods of the log package
// (https://golang.org/pkg/log/).
//
// Additionally if the logger implements two methods called Debugf and
// Debugln with corresponding signatures to the Print* methods debug
// messages will also be logged
func SetLogger(l logger) {
	if dl, ok := l.(debugLogger); ok {
		lgd = dl
	}
	lg = l
}

func debug(a ...interface{}) {
	if lgd == nil {
		return
	}
	lgd.Debugln(a...)
}

func debugf(f string, a ...interface{}) {
	if lgd == nil {
		return
	}
	lgd.Debugf(f, a...)
}

func logf(f string, a ...interface{}) {
	if lg == nil {
		return
	}
	lg.Printf(f, a...)
}

func log(a ...interface{}) {
	if lg == nil {
		return
	}
	lg.Println(a...)
}
