package newfs

func ChildXbee(parent Folder) Folder { return parent.ChildFolder(".xbee") }

func xbeeIntern() Folder {
	return ChildXbee(Home)
}

func SSHFolder() Folder { return xbeeIntern().ChildFolder(".ssh") }
func CachedFileForUrl(rawUrl string) File {
	fd, name := CacheArtefacts().SubFolderForLocation(rawUrl)
	return fd.ChildFile(name)
}
func CacheArtefacts() Folder { return xbeeIntern().ChildFolder("cache-artefacts") }
func CacheElements() Folder  { return xbeeIntern().ChildFolder("cache-elements") }
func LogsFolder() Folder     { return xbeeIntern().ChildFolder("logs") }
func Rsa() *RsaGenerator     { return NewRsaGen(SSHFolder()) }
func TmpDir() Folder         { return xbeeIntern().ChildFolder("tmp") }
