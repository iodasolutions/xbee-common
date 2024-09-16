package newfs

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/iodasolutions/xbee-common/cmd"
	"github.com/iodasolutions/xbee-common/template"
	"github.com/ulikunitz/xz"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

type File string

func (f File) String() string {
	return string(f)
}

func (f File) IsYAML() bool {
	return strings.HasSuffix(f.String(), YamlExt)
}
func (f File) IsJSON() bool {
	return strings.HasSuffix(f.String(), ".json")
}
func (f File) Exists() bool {
	return Path(f).Exists()
}

func (f File) Content() (result string) {
	data := f.ContentBytes()
	if data == nil {
		result = ""
	} else {
		result = string(data)
	}
	return
}

func (f File) ContentBytes() []byte {
	if !f.Exists() {
		return nil
	}
	data, err := ioutil.ReadFile(string(f))
	if err != nil {
		panic(fmt.Sprintf("cannot read content of %s : %v\n", f, err))
	}
	return data
}

func Unmarshal[T any](f File) (T, *cmd.XbeeError) {
	var t T
	if err := json.Unmarshal(f.ContentBytes(), &t); err != nil {
		return t, cmd.Error("cannot unmarshal %s: %s", f, err)
	}
	return t, nil
}

func (f File) BaseWithoutExtension() string {
	return Path(f).BaseWithoutExtension()
}
func (f File) Extension() string {
	return Path(f).Extension()
}

func (f File) Save(outs ...interface{}) {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "    ")
	for _, anOut := range outs {
		if err := enc.Encode(anOut); err != nil {
			panic(fmt.Errorf("cannot encode as json2 to file %s : %v", f, err))
		}
	}
	f.SetContent(buf.String())
}

func (f File) Dir() Folder {
	return Folder(filepath.Dir(f.String()))
}
func (f File) OpenFileForCreation() (*os.File, *cmd.XbeeError) {
	f.Dir().Create()
	fd, err := os.Create(string(f))
	if err != nil {
		return nil, cmd.Error("cannot create file %s : %v", f, err)
	}
	return fd, nil
}

func (f File) EnsureDelete() *cmd.XbeeError {
	if !f.Exists() {
		return nil
	}
	if err := os.Remove(string(f)); err != nil {
		return cmd.Error("cannot remove %s: %v", f, err)
	}
	return nil
}

func (f File) Sha1() string {
	data, err := ioutil.ReadFile(string(f))
	if err != nil {
		panic(fmt.Errorf("failed to read %s : %v", string(f), err))
	}
	hash := sha1.New()
	if _, err := hash.Write(data); err != nil {
		panic(err)
	}
	hashBytes := hash.Sum(nil)
	content := hex.EncodeToString(hashBytes)
	return content
}
func (f File) Sha1File() File {
	return File(fmt.Sprintf("%s.sha1", string(f)))
}

func (f File) CreateSha1() File {
	theSha1 := f.Sha1File()
	theSha1.SetContent(f.Sha1())
	return theSha1
}
func (f File) CopyToPath(targetPath File) File {
	in, err := os.Open(string(f))
	if err != nil {
		panic(fmt.Errorf("CopyToPath : Cannot open source file %s : %v\n", f, err))
	}
	defer in.Close()
	targetPath.Dir().EnsureExists()
	targetPath.FillFromReader(in)
	targetPath.ChMod(f.Mod())
	uid, gid := f.Owner()
	if uid != -1 {
		if err := os.Chown(string(targetPath), uid, gid); err != nil {
			panic(fmt.Errorf("Cannot change owner %d for file path %s : %v\n", uid, targetPath, err)) //TODO deals with errors
		}
	}
	return targetPath
}
func (f File) ChMod(mod os.FileMode) {
	err := os.Chmod(string(f), mod)
	if err != nil {
		panic(err)
	}
}

func (f File) FillFromReader(in io.Reader) int64 {
	f.Dir().EnsureExists()
	out, err := os.Create(string(f))
	if err != nil {
		panic(fmt.Errorf("FillFromReader : Cannot create target file %s : %v\n", f, err))
	}
	defer func() {
		if e := out.Close(); e != nil {
			panic(fmt.Errorf("FillFromReader : Cannot close target file %s : %v\n", f, err))
		}
	}()

	size, err := io.Copy(out, in)
	if err != nil {
		panic(fmt.Errorf("fillFromReader : Cannot copy input stream to file %s : %v\n", f, err))
	}

	err = out.Sync()
	if err != nil {
		panic(fmt.Errorf("fillFromReader : Cannot sync target file %s : %v\n", f, err))
	}
	return size
}

func (f File) Mod() os.FileMode {
	si, err := os.Stat(string(f))
	if err != nil {
		panic(err)
	}
	return si.Mode()
}

func (f File) CopyToDir(targetDir Folder) File {
	targetPath := targetDir.ChildFile(f.Base())
	return f.CopyToPath(targetPath)
}
func (f File) Base() string {
	return filepath.Base(string(f))
}

func (f File) CanBeFiltered() bool {
	var parseExcludes = []string{".zip", ".war", ".tar", ".tar.gz", ".tar.xz"}
	name := f.Base()
	for _, exclude := range parseExcludes {
		if strings.HasSuffix(name, exclude) {
			return false
		}
	}
	return true
}

func (f File) FillWithTemplate(templateS string, data interface{}, funcMap map[string]interface{}) *cmd.XbeeError {
	fp, err := f.OpenFileForCreation()
	if err != nil {
		return err
	}
	defer fp.Close()
	if err := template.OutputWithTemplate(templateS, fp, data, funcMap); err != nil {
		return cmd.Error("cannot parse [%s] with variables [%s]: %v", templateS, data, err)
	}
	return nil
}

func (f File) SetContentBytes(content []byte) *cmd.XbeeError {
	w, err := f.OpenFileForCreation()
	if err != nil {
		return err
	}
	buf := bytes.NewBuffer(content)
	if _, err := io.Copy(w, buf); err != nil {
		return cmd.Error("cannot copy content to %s : %v", f, err)
	}
	defer w.Close()
	return nil
}
func (f File) SetContent(content string) {
	f.SetContentBytes([]byte(content))
}
func (f File) Path() Path {
	return Path(f)
}
func (f File) Create() {
	f.SetContent("")
}

func (f File) ZipFile() File {
	return f.Dir().ChildFile(fmt.Sprintf("%s.zip", f.Base()))
}

func (f File) EnsureExists() {
	if !f.Exists() {
		f.Create()
	}
}

func (f File) Compress(extension string) (File, *cmd.XbeeError) {
	if extension != "gz" && extension != "zip" {
		return "", cmd.Error("compression support only gz or zip, actual is [%s]", extension)
	}
	reader, err := os.Open(f.String())
	if err != nil {
		return "", cmd.Error("cannot open file %s : %v", f.String(), err)
	}
	defer reader.Close()
	target := File(f.String() + "." + extension)
	targetWriter, err2 := target.OpenFileForCreation()
	if err2 != nil {
		return "", cmd.Error("cannot create file %s : %v", target.String(), err2)
	}
	defer targetWriter.Close()

	var compress_writer io.Writer
	var closer io.Closer
	switch extension {
	case "gz":
		zw := gzip.NewWriter(targetWriter)
		compress_writer = zw
		closer = zw
	default: //zip
		// Créer un writer gzip avec la compression par défaut
		zipWriter := zip.NewWriter(targetWriter)
		// Créer une entrée zip pour le fichier
		zipEntry, err := zipWriter.Create(f.Base())
		if err != nil {
			return "", cmd.Error("cannot create zip entry %s : %v", f.Base(), err)
		}
		compress_writer = zipEntry
		closer = zipWriter
	}
	defer closer.Close()
	// Copier le contenu du fichier tar dans le writer gzip
	_, err = io.Copy(compress_writer, reader)
	if err != nil {
		return "", cmd.Error("cannot add content to file %s : %v", target.String(), err)
	}
	return target, nil
}
func (f File) ExtractTo(fd Folder) (File, *cmd.XbeeError) {
	fd.EnsureExists()
	target := fd.ChildFile(f.BaseWithoutExtension())
	return f.ExtractToFile(target)
}
func (f File) ExtractToFile(target File) (File, *cmd.XbeeError) {
	sourceReader, err := os.Open(f.String())
	if err != nil {
		return "", cmd.Error("cannot open file %s : %v", f.String(), err)
	}
	defer sourceReader.Close()

	var reader io.Reader
	var closer io.Closer

	extension := f.Extension()
	switch extension {
	case "gz":
		// Créer un lecteur gzip
		gr, err := gzip.NewReader(sourceReader)
		if err != nil {
			return "", cmd.Error("cannot create gzip reader for %s : %v", f.String(), err)
		}
		reader = gr
		closer = gr
	case "zip":
		r, err := zip.OpenReader(f.String())
		if err != nil {
			return "", cmd.Error("cannot open reader for %s : %v", f.String(), err)
		}
		contentFile, err := r.File[0].Open()
		if err != nil {
			return "", cmd.Error("cannot open reader for %s : %v", f.String(), err)
		}
		reader = contentFile
		closer = contentFile
	case "bzip2", "bz2":
		reader = bzip2.NewReader(sourceReader)
	case "xz":
		// Créer un lecteur gzip
		reader, err = xz.NewReader(sourceReader)
		if err != nil {
			return "", cmd.Error("cannot open reader for %s : %v", f.String(), err)
		}
	default:
		return "", cmd.Error("unknown compression support only gz, zip, xz or bzip2 are supported, actual is [%s]", extension)
	}
	if closer != nil {
		defer closer.Close()
	}

	// Créer le fichier de destination

	writer, err := os.Create(target.String())
	if err != nil {
		return "", cmd.Error("cannot create file %s : %v", target.String(), err)
	}
	defer writer.Close()

	// Copier le contenu décompressé dans le fichier de destination
	_, err = io.Copy(writer, reader)
	if err != nil {
		return "", cmd.Error("cannot add content to file %s : %v", target.String(), err)
	}
	return target, nil
}
func (f File) Extract() (File, *cmd.XbeeError) {
	target := f.Dir().ChildFile(f.BaseWithoutExtension())
	return f.ExtractToFile(target)
}

func (f File) UntarTo(targetFd Folder) *cmd.XbeeError {
	// Ouvrir le fichier .tar
	file, err := os.Open(f.String())
	if err != nil {
		return cmd.Error("cannot open file %s : %v", f.String(), err)
	}
	defer file.Close()

	// Créer un nouveau lecteur tar
	tarReader := tar.NewReader(file)

	// Parcourir les fichiers dans le tar
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			// Fin du tar
			break
		}
		if err != nil {
			return cmd.Error("cannot untar file %s : %v", f.String(), err)
		}

		// Construire le chemin complet du fichier
		path := filepath.Join(targetFd.String(), header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			// Créer les répertoires s'ils n'existent pas
			if err := os.MkdirAll(path, 0755); err != nil {
				return cmd.Error("cannot create directory %s : %v", path, err)
			}
		case tar.TypeReg:
			// Créer les fichiers réguliers
			outFile, err := os.Create(path)
			if err != nil {
				return cmd.Error("cannot create file %s : %v", path, err)
			}
			defer outFile.Close()

			// Copier le contenu du fichier
			if _, err := io.Copy(outFile, tarReader); err != nil {
				return cmd.Error("cannot add content to file %s : %v", path, err)
			}
		case tar.TypeSymlink:
			// Créer des liens symboliques
			if err := os.Symlink(header.Linkname, path); err != nil {
				return cmd.Error("cannot create symlink %s : %v", path, err)
			}
		case tar.TypeLink:
			// Créer des hardlinks
			linkTarget := filepath.Join(targetFd.String(), header.Linkname)
			if err := os.Link(linkTarget, path); err != nil {
				return cmd.Error("cannot create symlink %s : %v", linkTarget, err)
			}
		case tar.TypeBlock:
			// Créer un périphérique de bloc (nécessite les permissions root)
			if err := syscall.Mknod(path, syscall.S_IFBLK|0755, int(header.Devmajor)<<8|int(header.Devminor)); err != nil {
				return cmd.Error("cannot create bloc device %s : %v", path, err)
			}
		case tar.TypeChar:
			// Créer un périphérique de caractère (nécessite les permissions root)
			if err := syscall.Mknod(path, syscall.S_IFCHR|0755, int(header.Devmajor)<<8|int(header.Devminor)); err != nil {
				return cmd.Error("cannot create character device %s : %v", path, err)
			}
		case tar.TypeFifo:
			// Créer un FIFO (pipe nommé)
			if err := syscall.Mkfifo(path, 0755); err != nil {
				return cmd.Error("cannot create fifo device %s : %v", path, err)
			}
		default:
			fmt.Printf("Type de fichier inconnu ou non géré : %s\n", header.Name)
		}
	}

	return nil

}
func (f File) Untar() *cmd.XbeeError {
	return f.UntarTo(f.Dir())
}
