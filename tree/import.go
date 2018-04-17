package tree

type Import struct {
	Alias string
	Path  string
}

func (i Import) Format() string {
	if len(i.Alias) > 0 {
		return i.Alias + ` "` + i.Path + `"`
	}
	return `"` + i.Path + `"`
}
