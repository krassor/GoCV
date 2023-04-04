package handlers

import (
	//"encoding/json"
	//"fmt"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	//"net/url"

	//"github.com/rs/zerolog/log"

	"github.com/krassor/GoCV/internal/models"
	"github.com/krassor/GoCV/internal/pkg/utils"
	"github.com/rs/zerolog/log"

	
)

type TrainerService interface {
	TrainModel(tenant *models.Tenant) error
}

type FrHandler struct {
	trainer TrainerService
}

func NewFrHandler(t TrainerService) *FrHandler {
	return &FrHandler{
		trainer: t,
	}
}

func (h *FrHandler) LoadTenantFoto(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(32 << 20) // maxMemory 32MB
	if err != nil {
		httpErr := utils.Err(w, http.StatusBadRequest, err)
		if httpErr != nil {
			log.Error().Msgf("Error response: %s", httpErr)
		}
		return
	}
	defer func() {
		if err := r.MultipartForm.RemoveAll(); err != nil {
			log.Error().Msgf("Error remove multipartform temp files: %s", err)
		}
	}()

	tenant := models.Tenant{}
	tenant.Name = r.MultipartForm.Value["name"][0]
	tenant.Surname = r.MultipartForm.Value["surname"][0]

	filePath := path.Join("dataset", fmt.Sprintf("%s %s", tenant.Surname, tenant.Name))
	err = createOrCleanPath(filePath, &tenant)
	if err != nil {
		log.Error().Msgf("Error create path \"%s\": %s", filePath, err)
		errUtil := utils.Err(w, http.StatusInternalServerError, fmt.Errorf("Internal error"))
		if err != nil {
			log.Error().Msgf("Error response: %s", errUtil)
		}
		return
	}

	for _, f := range r.MultipartForm.File["photo"] {
		file, err := f.Open()
		if err != nil {
			log.Error().Msgf("Error receiving file \"%s\": %s", f.Filename, err)
			err = utils.Err(w, http.StatusInternalServerError, fmt.Errorf("Error receiving file \"%s\"", f.Filename))
			if err != nil {
				log.Error().Msgf("Error response: %s", err)
			}
			continue
		}
		defer file.Close()

		dst, err := os.Create(path.Join(filePath, f.Filename))
		if err != nil {
			log.Error().Msgf("Error creating file \"%s\": %s", f.Filename, err)
			errUtil := utils.Err(w, http.StatusInternalServerError, fmt.Errorf("Error receiving file \"%s\"", f.Filename))
			if err != nil {
				log.Error().Msgf("Error response: %s", errUtil)
			}
			continue
		}
		defer dst.Close()

		_, err = io.Copy(dst, file)
		if err != nil {
			log.Error().Msgf("Error saving file \"%s\": %s", f.Filename, err)
			errUtil := utils.Err(w, http.StatusInternalServerError, fmt.Errorf("Error saving file \"%s\"", f.Filename))
			if err != nil {
				log.Error().Msgf("Error response: %s", errUtil)
			}
			continue
		}
	}

	if err := h.trainer.TrainModel(&tenant); err != nil {
		log.Error().Msgf("Error training model \"%s %s\": %s", tenant.Surname, tenant.Name, err)
		err = utils.Err(w, http.StatusInternalServerError, fmt.Errorf("Internal error"))
		if err != nil {
			log.Error().Msgf("Error response: %s", err)
		}
		return
	}

	log.Info().Msgf("Successful train model %s %s", tenant.Surname, tenant.Name)
	var response = make(map[string]string)
	response["name"] = tenant.Name
	response["surname"] = tenant.Surname
	response["trainStatus"] = "ok"
	if err = utils.Json(w, http.StatusOK, response); err != nil {
		log.Error().Msgf("Error response: %s", err)
	}
}

func createOrCleanPath(cleanPath string, tenant *models.Tenant) error {
	if err := os.Mkdir(cleanPath, 0755); err != nil {
		ExistFiles, errListDir := listDir(cleanPath)
		if errListDir != nil {
			return fmt.Errorf("Error read dataset dir \"%s %s\": %w", tenant.Surname, tenant.Name, err)
		}
		for _, ExistFile := range ExistFiles {
			os.Remove(ExistFile)
		}
	}
	return nil
}

// listDir return slice of images path in srcPath, excluding within directories.
// Return error if directory does not exist
func listDir(srcPath string) ([]string, error) {
	var imageFiles []string

	files, err := os.ReadDir(srcPath) //ioutil.ReadDir(srcPath)
	if err != nil {
		return nil, fmt.Errorf("Error read padth: %w", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		} else {
			fileName := file.Name()
			if isImage(fileName) {
				imageFiles = append(imageFiles, path.Join(srcPath, fileName))
			}

		}
	}
	return imageFiles, nil
}

func isImage(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif":
		return true
	}
	return false
}
