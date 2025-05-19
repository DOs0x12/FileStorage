package file

type File interface {
	Write(data []byte, fName string) (rErr error)
	Read(fName string) ([]byte, error)
}
