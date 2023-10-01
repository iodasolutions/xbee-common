package provider

type GenericVolume struct {
	Name string
	Size int
}

func FromVolume(req *Volume) GenericVolume {
	return GenericVolume{
		Name: req.Name,
		Size: req.Size,
	}
}
