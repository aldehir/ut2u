package upkg

type header struct {
	Magic           uint32
	Version         uint16
	Licensee        uint16
	PackageFlags    uint32
	NameCount       uint32
	NameOffset      uint32
	ExportCount     uint32
	ExportOffset    uint32
	ImportCount     uint32
	ImportOffset    uint32
	GUID            [16]byte
	GenerationCount uint32
}

type generation struct {
	ExportCount uint32
	NameCount   uint32
}

type index = int

type name struct {
	Str   string
	Flags uint32
}

type import_ struct {
	ClassPackageIndex index
	ClassNameIndex    index
	Package           uint32
	ObjectNameIndex   index
}
