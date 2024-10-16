package newfs

import (
	"archive/tar"
	"fmt"
	"github.com/iodasolutions/xbee-common/cmd"
	"github.com/iodasolutions/xbee-common/exec2"
	"github.com/iodasolutions/xbee-common/stringutils"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type Folder string

var Home Folder

func init() {
	homeS, _ := os.UserHomeDir()
	homeS = filepath.ToSlash(homeS)
	Home = Folder(homeS)
}

func CWD() Folder {
	s, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	s = filepath.ToSlash(s)
	return Folder(s)
}

func (fd Folder) Path() Path {
	return Path(fd)
}

func (fd Folder) Exists() bool {
	return Path(fd).Exists()
}
func (fd Folder) EnsureEmpty() {
	fd.EnsureExists()
	fd.DeleteDirContent()
}

func (fd Folder) EnsureExists() bool {
	if !fd.Exists() {
		fd.Create()
		return true
	}
	return false
}

func (fd Folder) Create() Folder {
	if !fd.Exists() {
		err := os.MkdirAll(string(fd), 0755)
		if err != nil {
			panic(fmt.Errorf("Cannot create folder %s : %v\n", string(fd), err))
		}
	}
	return fd
}

func (fd Folder) DeleteDirContent() *cmd.XbeeError {
	if !fd.Exists() {
		return nil
	}
	dir, err := os.Open(fd.String())
	if err != nil {
		return cmd.Error("cannot open %s : %v", fd, err)
	}
	defer dir.Close()
	names, err := dir.Readdirnames(-1)
	if err != nil {
		return cmd.Error("cannot read folder content %s : %v", fd, err)
	}
	for _, name := range names {
		child := fd.ChildFile(name).String()
		err = os.RemoveAll(child)
		if err != nil {
			return cmd.Error("cannot removeAll from %s : %v", child, err)
		}
	}
	return nil
}

func (fd Folder) Delete() *cmd.XbeeError {
	if fd.Exists() {
		if err := fd.DeleteDirContent(); err != nil {
			return err
		}
		if err := os.Remove(fd.String()); err != nil {
			return cmd.Error("cannot remove empty folder %s : %v", fd, err)
		}
	}
	return nil
}

func (fd Folder) ChildFile(name string) File {
	path := Path(fd).Child(name)
	return File(path)
}
func (fd Folder) ChildFolder(name string) Folder {
	path := Path(fd).Child(name)
	return Folder(path)
}
func (fd Folder) ChildFileJson(name string) File {
	return fd.ChildFile(name + jsonExt)
}
func (fd Folder) ChildFileYml(name string) File {
	return fd.ChildFile(name + YamlExt)
}
func (fd Folder) ChildFilesYml() (result []File) {
	return fd.ChildrenFilesEndingWith(YamlExt)
}
func (fd Folder) ChildrenFilesEndingWith(end string) (result []File) {
	theFiles, _ := fd.ChildrenFilesAndFolders()
	for _, child := range theFiles {
		if strings.HasSuffix(child.String(), end) {
			result = append(result, child)
		}
	}
	return
}
func (fd Folder) Empty() bool {
	if !fd.Exists() {
		return true
	}
	fis, err := ioutil.ReadDir(string(fd))
	if err != nil {
		log.Panicf("Cannot not read children files for folder : %s : %v\n", fd, err)
	}
	return len(fis) == 0
}
func (fd Folder) ChildrenFilesAndFolders() (theFiles []File, theFolders []Folder) {
	if !fd.Exists() {
		return
	}
	fis, err := ioutil.ReadDir(string(fd))
	if err != nil {
		panic(fmt.Errorf("Cannot not read children files for folder %s : %v\n", fd, err))
	}
	for _, fi := range fis {
		name := fi.Name()
		if !fi.IsDir() {
			theFiles = append(theFiles, fd.ChildFile(name))
		} else {
			theFolders = append(theFolders, fd.ChildFolder(name))
		}
	}
	return
}

func (fd Folder) Children() (result []Path) {
	fis, err := ioutil.ReadDir(string(fd))
	if err != nil {
		panic(fmt.Errorf("Cannot not read children files for folder %s : %v\n", fd, err))
	}
	for _, fi := range fis {
		name := fi.Name()
		result = append(result, fd.ChildPath(name))
	}
	return
}

func (fd Folder) ChildrenFilesAndFoldersRelativePaths() (result []string) {
	theFiles, theFolders := fd.ChildrenFilesAndFolders()
	for _, f := range theFiles {
		s := strings.TrimPrefix(f.String(), fd.String())
		s = strings.TrimPrefix(s, "/")
		result = append(result, s)
	}
	for _, f := range theFolders {
		s := strings.TrimPrefix(f.String(), fd.String())
		s = strings.TrimPrefix(s, "/")
		result = append(result, s)
	}
	return
}
func (fd Folder) ChildPath(name string) Path {
	return Path(fd).Child(name)
}

func (fd Folder) Sha1() string {
	entries := fd.AllEntries()
	var sha1s []string
	for _, entry := range entries {
		var sha1 string
		absPath := fd.ChildPath(string(entry))
		if !absPath.IsDir() {
			f := File(absPath)
			sha1 = f.Sha1()
		}
		sha1s = append(sha1s, fmt.Sprintf("%s : %s", entry, sha1))
	}
	allSha1s := strings.Join(sha1s, "\n")
	return stringutils.Sha1String(allSha1s)
}

func (fd Folder) AllEntries() (result []Path) {
	gitDir := fd.ChildFolder(".git")
	walkFunc := func(path string, info os.FileInfo, err error) error {
		aPath := Path(path)
		if path != string(fd) &&
			aPath.Base() != ".gitignore" &&
			!aPath.IsDescendantOf(gitDir.Path()) &&
			string(fd.Path().Sha1Path()) != path {
			result = append(result, fd.Path().RelativeFrom(aPath))
		}
		return nil
	}
	if err := filepath.Walk(string(fd), walkFunc); err != nil {
		panic(fmt.Errorf("Cannot list all entries of %s : %v\n", fd, err))
	}
	return
}
func (fd Folder) Base() string {
	return filepath.Base(string(fd))
}
func (fd Folder) CopyDirToDir(dstDir Folder) {
	childDstDir := dstDir.ChildFolder(fd.Base())
	fd.CopyDirContentToDir(childDstDir)
}
func (fd Folder) ChMod(mod os.FileMode) {
	err := os.Chmod(string(fd), mod)
	if err != nil {
		panic(err)
	}
}
func (fd Folder) Dir() Folder {
	return Folder(filepath.ToSlash(filepath.Dir(string(fd))))
}
func (fd Folder) CopyDirContentToDir(dstDir Folder) {
	fd.CopyDirContentToDirKeepOwner(dstDir, true)
}

func (fd Folder) CopyDirContentToDirKeepOwner(dstDir Folder, keepOwner bool) {
	dstDir.Create()
	dstDir.ChMod(fd.Mod())
	if keepOwner {
		uid, gid := fd.Owner()
		if uid != -1 {
			if err := os.Chown(string(dstDir), uid, gid); err != nil {
				panic(fmt.Errorf("Cannot change owner %d for file path %s : %v\n", uid, dstDir, err)) //TODO deals with errors
			}
		}
	}
	entries, err := ioutil.ReadDir(string(fd))
	if err != nil {
		panic(fmt.Errorf("CopyDirContentToDir : Cannot read conent of src dir %s : %v", fd, err))
	}

	for _, entry := range entries {
		srcPath := filepath.Join(string(fd), entry.Name())
		dstPath := filepath.Join(string(dstDir), entry.Name())

		if entry.IsDir() {
			Folder(srcPath).CopyDirContentToDir(Folder(dstPath))
		} else {
			// Skip symlinks.
			if entry.Mode()&os.ModeSymlink != 0 {
				continue
			}
			srcFile := File(srcPath)
			srcFile.CopyToPath(File(dstPath))
		}

	}
	return
}

func (fd Folder) RandomChildFolder() Folder {
	return fd.ChildFolder(stringutils.RandomString())
}

func (fd Folder) ParsePath(data interface{}, funcMap map[string]interface{}) {
	walkFunc := func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			f := File(path)
			if f.CanBeFiltered() {
				template := f.Content()
				if err := f.FillWithTemplate(template, data, funcMap); err != nil {
					panic(fmt.Errorf("cannot parse file %s : %v", f, err))
				}
			}
		}
		return nil
	}
	if err := filepath.Walk(string(fd), walkFunc); err != nil {
		panic(fmt.Errorf("error when walking down folder %s : %v", fd.String(), err))
	}
}

func (fd Folder) Mod() os.FileMode {
	si, err := os.Stat(string(fd))
	if err != nil {
		panic(err)
	}
	return si.Mode()
}

func (fd Folder) ChModRecursive(mod os.FileMode) {
	files, dirs := fd.ChildrenFilesAndFolders()
	for _, aDir := range dirs {
		aDir.ChMod(mod)
		aDir.ChModRecursive(mod)
	}
	for _, aFile := range files {
		aFile.ChMod(mod)
	}
}

func (fd Folder) DescendantFilesYml() (result []File) {
	if !fd.Exists() || fd.Empty() {
		return nil
	}
	walkFunc := func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && (strings.HasSuffix(info.Name(), YamlExt)) {
			result = append(result, File(path))
		}
		return nil
	}
	if err := filepath.Walk(string(fd), walkFunc); err != nil {
		panic(err)
	}
	return
}

func (fd Folder) SubFolderForLocation(location string) (Folder, string) {
	anUrl := Parse(location)
	host, path, name := anUrl.Split()
	return fd.ChildFolder(host).ChildFolder(path), name
}

func (fd Folder) ResolvePath(s string) string {
	if s == "" {
		return fd.String()
	}
	if strings.HasPrefix(s, "~") {
		s = Home.String() + s[1:]
	} else if strings.HasPrefix(s, ".") || strings.HasPrefix(s, "..") {
		s = fd.ChildFolder(s).String()
	} else if strings.HasPrefix(s, "$HOME") {
		s = Home.ChildFile(strings.TrimPrefix(s, "$HOME")).String()
	} else {
		return s
	}
	return filepath.Clean(s)
}

func (fd Folder) RandomFile() File {
	return fd.ChildFile(stringutils.RandomString())
}

func (fd Folder) MoveTo(dir Folder) *cmd.XbeeError {
	return moveDirectory(string(fd), string(dir))
}

func (fd Folder) TarToFile(f File) *cmd.XbeeError {
	var args []string
	if fd == "/" {
		args = strings.Split("--exclude=./dev --exclude=./proc --exclude=./sys --exclude=./tmp --exclude=./run --exclude=./mnt --exclude=./media --exclude=./lost+found --exclude=./xbee --exclude=./usr/bin/xbee", " ")
	}
	args = append(args, "-cvf", f.String(), ".")
	aCmd := exec2.NewCommand("tar", args...).WithDirectory(fd.String())
	err := aCmd.Run(nil)
	out := aCmd.Result()
	if err != nil {
		return cmd.Error("This command [%s] failed : %s", aCmd.String(), out)
	}
	return nil
}

func (fd Folder) TarToFolder(target Folder) *cmd.XbeeError {
	target.EnsureExists()
	targetTar := target.ChildFile(fd.Base() + ".tar")
	return fd.TarToFile(targetTar)
}

// Fonction pour ajouter un fichier/dossier à l'archive tar
func addFileToTar(tw *tar.Writer, basePath, path string, info os.FileInfo) error {
	// Ouvrir le fichier à ajouter à l'archive
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Créer l'en-tête tar pour le fichier
	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return err
	}

	// Ajuster le champ Name pour utiliser un chemin relatif
	relPath, err := filepath.Rel(basePath, path)
	if err != nil {
		return err
	}
	if relPath != "." {
		header.Name = "./" + relPath
	} else {
		return nil
	}

	if info.Mode().IsRegular() {
		// Ouvrir le fichier régulier à ajouter à l'archive
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		// Écrire l'en-tête dans l'archive
		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		// Écrire le contenu du fichier dans l'archive
		_, err = io.Copy(tw, file)
		if err != nil {
			return err
		}
	} else if (info.Mode() & os.ModeSymlink) != 0 {
		// Récupérer le chemin cible du lien symbolique
		target, err := os.Readlink(path)
		if err != nil {
			return err
		}
		header.Linkname = target
		header.Typeflag = tar.TypeSymlink

		// Écrire l'en-tête du lien symbolique dans l'archive
		if err := tw.WriteHeader(header); err != nil {
			return err
		}
	} else {
		// Gérer les autres types de fichiers (comme les dossiers) en écrivant seulement l'en-tête
		if err := tw.WriteHeader(header); err != nil {
			return err
		}
	}

	return nil
}

func (fd Folder) Compress(extension string, keepTar bool) (File, *cmd.XbeeError) {
	return fd.CompressToDir(fd.Dir(), extension, keepTar)
}

// CompressToDir supports gz, zip
func (fd Folder) CompressToDir(target Folder, extension string, keepTar bool) (File, *cmd.XbeeError) {
	target.EnsureExists()
	return fd.CompressToPath(target.ChildFile(fd.Base()+"."+extension), keepTar)
}

func (fd Folder) CompressToPath(target File, keepTar bool) (File, *cmd.XbeeError) {
	targetTar := target.Dir().ChildFile(target.BaseWithoutExtension() + ".tar")
	if err := fd.TarToFile(targetTar); err != nil {
		return "", err
	}
	result, err := targetTar.Compress(target.Extension())
	if err != nil {
		return "", err
	}
	if !keepTar {
		if err := targetTar.EnsureDelete(); err != nil {
			return "", err
		}
	}
	return result, nil
}

func (fd Folder) ChangeOwner(uid, gid int) *cmd.XbeeError {
	// Appliquer le changement sur le répertoire racine
	if err := os.Chown(string(fd), uid, gid); err != nil {
		return cmd.Error("failed to change owner for %s: %v", fd, err)
	}
	// Parcourir tous les fichiers et sous-répertoires
	err := filepath.Walk(string(fd), func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Changer l'owner de chaque fichier/sous-répertoire
		if err := os.Chown(p, uid, gid); err != nil {
			return fmt.Errorf("failed to change owner for %s: %v", p, err)
		}
		return nil
	})
	return cmd.Error("%v", err)
}
