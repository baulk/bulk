package archive

// https://ethw.org/History_of_Lossless_Data_Compression_Algorithms

// Extractor todo
type Extractor interface {
	Extract(destination string) error
	Close() error
}

// Archiver todo
type Archiver interface {
	Close() error
}
