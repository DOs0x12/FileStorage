package file

type Writer interface {
	Write(data, fName string) (rErr error)
}
