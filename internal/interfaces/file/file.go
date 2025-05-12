package file

type File interface {
	Write(data, fName string) (rErr error)
	Read(fName string) (string, error)
}
