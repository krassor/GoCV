package handlers

import (
	//"encoding/json"
	//"fmt"
	"net/http"
	//"net/url"

	//"github.com/rs/zerolog/log"
	//services "github.com/serverStandMonitor/internal/services/devices"
	"github.com/krassor/GoCV/internal/dto"
)

type TrainerHandlers interface {
	TrainModelByFoto(w http.ResponseWriter, r *http.Request)
}

type trainerHandler struct {
	deviceService services.DevicesRepoService
}

func NewDeviceHandler(deviceService services.DevicesRepoService) DeviceHandlers {
	return &deviceHandler{
		deviceService: deviceService,
	}
}

func (d *deviceHandler) CreateDevice(w http.ResponseWriter, r *http.Request) {
	
}
