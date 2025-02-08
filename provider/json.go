package provider

type Action string

const (
	Up             Action = "up"
	Delete         Action = "delete"
	DestroyVolumes Action = "destroyvolumes"
	Infos          Action = "instanceinfos"
	Image          Action = "image"
	Down           Action = "down"
)

type DestroyVolumesRequest struct {
	Volumes []*XbeeVolume `json:"volumes,omitempty"`
}

func (dr *DestroyVolumesRequest) VolumeNames() (result []string) {
	for _, v := range dr.Volumes {
		result = append(result, v.Name)
	}
	return
}
