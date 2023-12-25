package provider

type Action string

const (
	Up             Action = "up"
	Delete         Action = "delete"
	DestroyVolumes Action = "destroyvolumes"
	InstanceInfos  Action = "instanceinfos"
	Image          Action = "image"
)

type DestroyVolumesRequest struct {
	Volumes []*Volume `json:"volumes,omitempty"`
}

func (dr *DestroyVolumesRequest) VolumeNames() (result []string) {
	for _, v := range dr.Volumes {
		result = append(result, v.Name)
	}
	return
}
