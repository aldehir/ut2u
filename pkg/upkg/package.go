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

func (p *Package) GUID() []byte {
	// The package guid is stored as 4 dwords in little endian. The header
	// struct reads it an as a byte array so we need to reorder the dwords to
	// big endian.
	guid := make([]byte, 16)

	for i := 0; i < 4; i++ {
		guid[(i * 4)] = p.h.GUID[(i*4)+3]
		guid[(i*4)+1] = p.h.GUID[(i*4)+2]
		guid[(i*4)+2] = p.h.GUID[(i*4)+1]
		guid[(i*4)+3] = p.h.GUID[(i * 4)]
	}

	return guid
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
