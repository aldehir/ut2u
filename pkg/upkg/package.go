package upkg

import (
	"strings"
)

type Package struct {
	h       header
	gen     []generation
	names   []name
	imports []import_
}

func (p *Package) PackageDependencies() []string {
	deps := make(map[string]struct{})
	for _, imp := range p.imports {
		depPkg := p.names[imp.ClassPackageIndex].Str
		depCls := p.names[imp.ClassNameIndex].Str
		depName := p.names[imp.ObjectNameIndex].Str

		if strings.EqualFold(depPkg, "Core") && strings.EqualFold(depCls, "Package") && imp.Package == 0 {
			deps[depName] = struct{}{}
		}
	}

	depsSlice := make([]string, 0, len(deps))
	for k := range deps {
		depsSlice = append(depsSlice, k)
	}

	return depsSlice
}
