package config

type Exclude struct {
	name        string
	description string
	sources     []Package
}

func (e *Exclude) Name() string {
	return e.name
}

func (e *Exclude) Description() string {
	return e.description
}

func (e *Exclude) Sources() []Package {
	return e.sources
}

func (e *Exclude) toDTO() exclude {
	sources := make([]pack, len(e.Sources()))
	for i, p := range e.Sources() {
		sources[i] = p.toDTO()
	}

	return exclude{
		Name:        e.Name(),
		Description: e.Description(),
		Sources:     sources,
	}
}

type exclude struct {
	Name        string `json:"name" yaml:"name"`               // exclude name
	Description string `json:"description" yaml:"description"` // rule description
	Sources     []pack `json:"sources" yaml:"sources"`         // source package
}

func (e *exclude) toExclude() Exclude {
	sources := make([]Package, len(e.Sources))
	for i, p := range e.Sources {
		sources[i] = p.toPackage()
	}

	return Exclude{
		name:        e.Name,
		description: e.Description,
		sources:     sources,
	}
}
