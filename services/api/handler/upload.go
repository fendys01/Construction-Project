package handler

import (
	"Contruction-Project/lib/upload"
	"Contruction-Project/lib/utils"
	"fmt"
	"io"
	"math/rand"
	"mime/multipart"
	"net/http"
	"os"
	"time"
)

// UploadAct ...
func (h *Contract) UploadAct(w http.ResponseWriter, r *http.Request) {
	fm, fInfo, err := upload.Info{MaxSize: 1}.MultipartHandler(
		w, r, "uploadfile", []string{"png", "jpg", "jpeg"},
	)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	_, fname, err := h.toS3(fm, fInfo, "/images/itinerary")
	if err != nil {
		h.SendBadRequest(w, "Something problem when saving file: "+err.Error())
		return
	}

	h.SendSuccess(w, map[string]interface{}{
		"img": h.Config.GetString("aws.s3.public_url") + fname,
	}, nil)
}

// toS3 handle upload file from local to s3
func (h *Contract) toS3(file multipart.File, fInfo upload.FileInfo, path string) (*os.File, string, error) {
	rand.Seed(time.Now().UnixNano())
	rTail, _ := utils.Generate(`[a-zA-Z0-9]{15}`)
	newname := fmt.Sprintf("/uploads/%s/%s.%s/", path, rTail, fInfo.FileExt)
	newpath := h.Config.GetString("upload_path") + newname

	f, err := os.OpenFile(newpath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return f, newname, err
	}
	defer f.Close()
	_, err = io.Copy(f, file)
	if err != nil {
		return f, newname, err
	}

	// upload to s3
	if err = upload.PushS3ByPath(f.Name(), upload.S3Info{
		Key:      h.Config.GetString("aws.s3.key"),
		Secret:   h.Config.GetString("aws.s3.secret"),
		Region:   h.Config.GetString("aws.s3.region"),
		Bucket:   h.Config.GetString("aws.s3.bucket"),
		Filename: newpath,
		Filemime: fInfo.FileMime,
		Filesize: fInfo.FileSize,
	}); err != nil {
		return f, newname, err
	}

	return f, newname, nil
}
