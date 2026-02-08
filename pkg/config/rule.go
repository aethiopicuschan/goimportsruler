package config

type Rule struct {
	name        string
	description string
	sources     []Package
	disallow    []Package
}

func (r *Rule) Name() string {
	return r.name
}

func (r *Rule) Description() string {
	return r.description
}

func (r *Rule) Sources() []Package {
	return r.sources
}

func (r *Rule) Disallow() []Package {
	return r.disallow
}

func (r *Rule) toDTO() rule {
	sources := make([]pack, len(r.Sources()))
	for i, p := range r.Sources() {
		sources[i] = p.toDTO()
	}
	disallow := make([]pack, len(r.Disallow()))
	for i, p := range r.Disallow() {
		disallow[i] = p.toDTO()
	}
	return rule{
		Name:        r.Name(),
		Description: r.Description(),
		Sources:     sources,
		Disallow:    disallow,
	}
}

type rule struct {
	Name        string `json:"name"`        // rule name
	Description string `json:"description"` // rule description
	Sources     []pack `json:"sources"`     // source packages
	Disallow    []pack `json:"disallow"`    // packages to disallow
}

func (r *rule) toRule() Rule {
	sources := make([]Package, len(r.Sources))
	for i, p := range r.Sources {
		sources[i] = p.toPackage()
	}
	disallow := make([]Package, len(r.Disallow))
	for i, p := range r.Disallow {
		disallow[i] = p.toPackage()
	}
	return Rule{
		name:        r.Name,
		description: r.Description,
		sources:     sources,
		disallow:    disallow,
	}
}
