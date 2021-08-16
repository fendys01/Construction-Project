package upload

import (
	"mime/multipart"
	"net/http"
	"strings"
)

const (
	// MB size constant
	MB = 1 << 20
)

// Sizer ...
type Sizer interface {
	Size() int64
}

// Info ...
type Info struct {
	MaxSize int64
}

// FileInfo ...
type FileInfo struct {
	Filename string
	FileSize int64
	FileMime string
	FileExt  string
}

// MultipartHandler handle multipart form data file upload
func (fu Info) MultipartHandler(w http.ResponseWriter,
	r *http.Request, key string, AllowedExt []string,
) (multipart.File, FileInfo, error) {
	if err := r.ParseMultipartForm(fu.MaxSize * MB); err != nil {
		return nil, FileInfo{}, err
	}

	// Limit upload size
	r.Body = http.MaxBytesReader(w, r.Body, fu.MaxSize*MB) // 1 Mb

	// get the file informations
	file, multipartFileHeader, err := r.FormFile(key)
	if err != nil {
		return nil, FileInfo{}, err
	}

	// Create a buffer to store the header of the file in
	// And copy the headers into the FileHeader buffer
	fileHeader := make([]byte, 512)
	if _, err := file.Read(fileHeader); err != nil {
		return nil, FileInfo{}, err
	}

	// set position back to start.
	if _, err := file.Seek(0, 0); err != nil {
		return nil, FileInfo{}, err
	}

	defer file.Close()

	mime := http.DetectContentType(fileHeader)
	ext := strings.Split(mime, "/")

	return file, FileInfo{
		Filename: multipartFileHeader.Filename,
		FileSize: file.(Sizer).Size(),
		FileMime: mime,
		FileExt:  ext[1],
	}, nil
}
