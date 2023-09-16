package upkg

import "github.com/aldehir/ut2u/pkg/encoding/ue2"

type header struct {
	Magic        uint32
	Version      uint16
	Licensee     uint16
	PackageFlags uint32
	NameCount    uint32
	NameOffset   uint32
	ExportCount  uint32
	ExportOffset uint32
	ImportCount  uint32
	ImportOffset uint32
	GUID         [16]byte
}

type generation struct {
	ExportCount uint32
	NameCount   uint32
}

type name struct {
	Str   string
	Flags uint32
}

type import_ struct {
	ClassPackageIndex ue2.Index
	ClassNameIndex    ue2.Index
	Package           uint32
	ObjectNameIndex   ue2.Index
}
