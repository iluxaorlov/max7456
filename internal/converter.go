package internal

type Converter interface {
	Decode(filePath string) error
	Encode(directoryPath string) error
}
