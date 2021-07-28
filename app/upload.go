package app

import (
	"io"
	"io/ioutil"
	"net/http"
	"path"
)

func UploadHandler(w http.ResponseWriter, req *http.Request) {
	userId := req.FormValue("userId")
	file, header, err := req.FormFile("avatarFile")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// file is an io.Reader
	data, err := ioutil.ReadAll(file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	filename := path.Join("avatars", userId+path.Ext(header.Filename))
	err = ioutil.WriteFile(filename, data, 0777) // 0777 means file permissions
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	io.WriteString(w, "Successful")
}
