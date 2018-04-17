package examples

type AlreadyAbstracted interface {
	DoSomething(string, int64, float64) error
}

type skippedForGoodReasons struct {
	apples []AppleIface
}

type Implemented struct {
	info    string
	version uint
}

func (i Implemented) String() string {
	return i.info
}
