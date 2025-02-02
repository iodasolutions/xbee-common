package newfs

func ChildXbee(parent Folder) Folder { return parent.ChildFolder(".xbee") }

func GlobalXbeeFolder() Folder {
	return ChildXbee(Home)
}

func SSHFolder() Folder { return GlobalXbeeFolder().ChildFolder(".ssh") }
func CachedFileForUrl(rawUrl string) File {
	fd, name := CacheArtefacts().SubFolderForLocation(rawUrl)
	return fd.ChildFile(name)
}
func CacheArtefacts() Folder { return GlobalXbeeFolder().ChildFolder("cache-artefacts") }
func CacheElements() Folder  { return GlobalXbeeFolder().ChildFolder("cache-elements") }
func LogsFolder() Folder     { return GlobalXbeeFolder().ChildFolder("logs") }
func Rsa() *RsaGenerator     { return NewRsaGen(SSHFolder()) }
func TmpDir() Folder         { return GlobalXbeeFolder().ChildFolder("tmp") }
func Volumes() Folder        { return GlobalXbeeFolder().ChildFolder("volumes") }
