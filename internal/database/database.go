package database

type Path string

const (
	Prefix        = "tmp/data/"
	AccessCluster = Path("cluster")
	AccessUser    = Path("user")
)

func (p Path) String() string {
	return Prefix + string(p)
}

func (p Path) Append(path string) Path {
	if path[0] == '/' {
		return Path(string(p) + path)
	}

	return Path(string(p) + "/" + path)
}
