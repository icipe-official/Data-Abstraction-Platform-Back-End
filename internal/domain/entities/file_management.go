package entities

const (
	STORAGE_DRIVE_TYPE_LOCAL string = "local"
)

type localStorageDrive struct {
	Path string
}

func LocalStorageDrive() localStorageDrive {
	return localStorageDrive{
		Path: "path",
	}
}
