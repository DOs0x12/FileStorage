package file

type Extractor interface {
	ExtractFileName(val string) string
	ExtractNumber(val string) (int64, error)
}
