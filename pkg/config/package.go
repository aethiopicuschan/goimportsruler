package config

type Package string

func (p Package) String() string {
	return string(p)
}

func (p Package) toDTO() pack {
	return pack(p)
}

type pack string

func (p pack) toPackage() Package {
	return Package(p)
}
