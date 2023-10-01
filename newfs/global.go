package newfs

var xbeeGlobal Folder

func init() {
	xbeeGlobal = ChildXbee(Home)
}

func ChildXbee(parent Folder) Folder { return parent.ChildFolder(".xbee") }

type XbeeGlobal interface {
	SSHFolder() Folder
}

func XbeeIntern() XbeeGlobal {
	return globalXbeeFolder{}
}

type globalXbeeFolder struct {
}

func (gxf globalXbeeFolder) SSHFolder() Folder { return xbeeGlobal.ChildFolder(".ssh") }

//func GlobalXbeeFolder() globalXbeeFolder {
//	internalDir := fromConf("internaldir").GetString()
//	if internalDir != "" {
//		return globalXbeeFolder{Folder: Folder(internalDir)}
//	}
//	return globalXbeeFolder{Folder: Home}
//}

func (gxf globalXbeeFolder) CacheElements() Folder   { return xbeeGlobal.ChildFolder("cache-elements") }
func (gxf globalXbeeFolder) CacheArtefacts() Folder  { return xbeeGlobal.ChildFolder("cache-artefacts") }
func (gxf globalXbeeFolder) CacheExports() Folder    { return xbeeGlobal.ChildFolder("cache-exports") }
func (gxf globalXbeeFolder) EnvsFolder() Folder      { return xbeeGlobal.ChildFolder("envs") }
func (gxf globalXbeeFolder) LogsFolder() Folder      { return xbeeGlobal.ChildFolder("logs") }
func (gxf globalXbeeFolder) VolumesFolder() Folder   { return xbeeGlobal.ChildFolder("volumes") }
func (gxf globalXbeeFolder) ContainerFolder() Folder { return xbeeGlobal.ChildFolder("container") }
func (gxf globalXbeeFolder) Rsa() *RsaGenerator      { return NewRsaGen(gxf.SSHFolder()) }
func (gxf globalXbeeFolder) ProviderFolder(provider string) Folder {
	return gxf.EnvsFolder().ChildFolder(provider)
}
func (gxf globalXbeeFolder) CachedFileForUrl(rawUrl string) File {
	fd, name := gxf.CacheArtefacts().SubFolderForLocation(rawUrl)
	return fd.ChildFile(name)
}
