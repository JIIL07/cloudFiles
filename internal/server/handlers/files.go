package handlers

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/JIIL07/jcloud/internal/server/storage"
	"github.com/JIIL07/jcloud/internal/server/utils"
	"net/http"
	"strings"
)

func AddFileHandler(w http.ResponseWriter, r *http.Request) {
	user := utils.ProvideUser(r, w)
	s := utils.ProvideStorage(r, w)

	var files []storage.File
	err := json.NewDecoder(r.Body).Decode(&files)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	tx, err := s.DB.Beginx()
	if err != nil {
		http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
		return
	}

	defer func() {
		if err != nil {
			err = tx.Rollback()
			if err != nil {
				http.Error(w, "Failed to rollback transaction: "+err.Error(), http.StatusInternalServerError)
			}
			http.Error(w, "Failed to add files: "+err.Error(), http.StatusInternalServerError)
		}
	}()

	for _, file := range files {
		file.UserID = user.UserID
		err = s.AddFileTx(tx, &file)
		if err != nil {
			return
		}
	}

	err = tx.Commit()
	if err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Files added successfully")) // nolint:errcheck
}

func DeleteFileHandler(w http.ResponseWriter, r *http.Request) {
	user := utils.ProvideUser(r, w)
	s := utils.ProvideStorage(r, w)

	f := &storage.File{UserID: user.UserID}
	f.Metadata.Name = r.URL.Query().Get("filename")
	err := s.DeleteFile(f)
	if err != nil {
		http.Error(w, "Failed to delete file", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("File deleted")) // nolint:errcheck
}

func DownloadFileHandler(w http.ResponseWriter, r *http.Request) {
	u := utils.ProvideUser(r, w)
	s := utils.ProvideStorage(r, w)

	filename := r.URL.Query().Get("filename")
	if filename == "" {
		http.Error(w, "Filename is required", http.StatusBadRequest)
		return
	}

	file, err := s.GetFile(u.UserID, strings.Split(filename, ".")[0])
	if err != nil {
		http.Error(w, "Failed to download file", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(file.Data)))
	w.Header().Set("Content-Disposition", "attachment; filename="+file.Metadata.Name+"."+file.Metadata.Extension)
	w.Header().Set("Content-Type", "application/octet-stream")

	w.Write(file.Data) // nolint:errcheck
}

func ListFilesHandler(w http.ResponseWriter, r *http.Request) {
	u := utils.ProvideUser(r, w)
	s := utils.ProvideStorage(r, w)

	files, err := s.GetAllFiles(u.UserID)
	if err != nil {
		http.Error(w, "Failed to list files", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(files) // nolint:errcheck
}

func ImageGalleryHandler(w http.ResponseWriter, r *http.Request) {
	user := utils.ProvideUser(r, w)
	s := utils.ProvideStorage(r, w)

	files, err := s.GetImageFiles(user.UserID)
	if err != nil {
		http.Error(w, "Failed to retrieve images", http.StatusInternalServerError)
		return
	}

	html := "<html><body><h1>Image Gallery</h1><div style='display: flex; flex-wrap: wrap;'>"
	for _, file := range files {
		imageDataURL := fmt.Sprintf("data:image/%s;base64,%s", file.Metadata.Extension, base64.StdEncoding.EncodeToString(file.Data))
		html += fmt.Sprintf(
			"<div style='margin: 10px;'><img src='%s' alt='%s' style='width: 200px; height: auto;'></div>",
			imageDataURL,
			file.Metadata.Name,
		)
	}
	html += "</div></body></html>"

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html)) // nolint:errcheck
}

func HashSumHandler(w http.ResponseWriter, r *http.Request) {
	user := utils.ProvideUser(r, w)
	s := utils.ProvideStorage(r, w)

	filename := r.URL.Query().Get("filename")
	if filename == "" {
		http.Error(w, "Filename is required", http.StatusBadRequest)
		return
	}

	file, err := s.GetFile(user.UserID, filename)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	hash := sha256.Sum256(file.Data)
	checksum := hex.EncodeToString(hash[:])

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"filename": filename, "checksum": checksum}) // nolint:errcheck
}

func FileDataHandler(w http.ResponseWriter, r *http.Request) {
	u := utils.ProvideUser(r, w)
	s := utils.ProvideStorage(r, w)

	f := &storage.File{}
	json.NewDecoder(r.Body).Decode(&f)
	f.UserID = u.UserID

	err := s.UpdateFile(f, f.Data)
	if err != nil {
		http.Error(w, "Failed to save file data", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("File saved successfully")) // nolint:errcheck

}

func PartialUpdateHandler(w http.ResponseWriter, r *http.Request) {

}

func FileInfoHandler(w http.ResponseWriter, r *http.Request) {
	s := utils.ProvideStorage(r, w)
	u := utils.ProvideUser(r, w)

	fileName := r.URL.Query().Get("filename")
	if fileName == "" {
		http.Error(w, "Filename is required", http.StatusBadRequest)
		return
	}

	fileInfo, err := s.GetFile(u.UserID, fileName)
	if err != nil {
		http.Error(w, "Failed to retrieve file info", http.StatusInternalServerError)
		return
	}
	if fileInfo == nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(fileInfo) // nolint:errcheck
}

func UpdatePermissionsHandler(w http.ResponseWriter, r *http.Request) {

}

func FilePermissionsHandler(w http.ResponseWriter, r *http.Request) {

}

func ShareFileHandler(w http.ResponseWriter, r *http.Request) {

}

func FileHistoryHandler(w http.ResponseWriter, r *http.Request) {

}

func UpdateMetadataHandler(w http.ResponseWriter, r *http.Request) {
	s := utils.ProvideStorage(r, w)
	u := utils.ProvideUser(r, w)

	if r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	req := struct {
		Filename    string `json:"filename"`
		Extension   string `json:"extension"`
		Description string `json:"description"`
		OldName     string `json:"oldname"`
	}{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := s.UpdateFileMetadata(u.UserID, req)
	if err != nil {
		http.Error(w, "Failed to update metadata", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Metadata updated successfully",
	})
}
