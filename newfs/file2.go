package newfs

import (
	"archive/zip"
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/iodasolutions/xbee-common/cmd"
	"github.com/iodasolutions/xbee-common/template"
	"github.com/ulikunitz/xz"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type File struct {
	Path2
}

func NewFile(path string) File {
	return File{Path2(path)}
}
func (f File) IsYAML() bool {
	return strings.HasSuffix(f.String(), YamlExt)
}
func (f File) IsJSON() bool {
	return strings.HasSuffix(f.String(), ".json")
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
	data, err := os.ReadFile(f.String())
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

func (f File) OpenFileForCreation() (*os.File, *cmd.XbeeError) {
	f.Dir().Create()
	fd, err := os.Create(f.String())
	if err != nil {
		return nil, cmd.Error("cannot create file %s : %v", f, err)
	}
	return fd, nil
}

func (f File) EnsureDelete() *cmd.XbeeError {
	if !f.Exists() {
		return nil
	}
	if err := os.Remove(f.String()); err != nil {
		return cmd.Error("cannot remove %s: %v", f, err)
	}
	return nil
}

func (f File) CopyToPath(targetPath File) File {
	in, err := os.Open(f.String())
	if err != nil {
		panic(fmt.Errorf("CopyToPath : Cannot open source file %s : %v\n", f, err))
	}
	defer in.Close()
	targetPath.Dir().EnsureExists()
	targetPath.FillFromReader(in)
	targetPath.ChMod(f.Mod())
	uid, gid := f.Owner()
	if uid != -1 {
		if err := os.Chown(targetPath.String(), uid, gid); err != nil {
			panic(fmt.Errorf("Cannot change owner %d for file path %s : %v\n", uid, targetPath, err)) //TODO deals with errors
		}
	}
	return targetPath
}

func (f File) FillFromReader(in io.Reader) int64 {
	f.Dir().EnsureExists()
	out, err := os.Create(f.String())
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
	si, err := os.Stat(f.String())
	if err != nil {
		panic(err)
	}
	return si.Mode()
}

func (f File) CopyToDir(targetDir Folder) File {
	targetPath := targetDir.ChildFile(f.Base())
	return f.CopyToPath(targetPath)
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

func (f File) Create() {
	f.SetContent("")
}

func (f File) ZipFile() File {
	return f.Dir().ChildFile(fmt.Sprintf("%s.zip", f.Base()))
}

func (f File) Compress(extension string) (File, *cmd.XbeeError) {
	if extension != "gz" && extension != "zip" {
		return NewFile(""), cmd.Error("compression support only gz or zip, actual is [%s]", extension)
	}
	reader, err := os.Open(f.String())
	if err != nil {
		return NewFile(""), cmd.Error("cannot open file %s : %v", f.String(), err)
	}
	defer reader.Close()
	target := NewFile(f.String() + "." + extension)
	targetWriter, err2 := target.OpenFileForCreation()
	if err2 != nil {
		return NewFile(""), cmd.Error("cannot create file %s : %v", target.String(), err2)
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
			return NewFile(""), cmd.Error("cannot create zip entry %s : %v", f.Base(), err)
		}
		compress_writer = zipEntry
		closer = zipWriter
	}
	defer closer.Close()
	// Copier le contenu du fichier tar dans le writer gzip
	_, err = io.Copy(compress_writer, reader)
	if err != nil {
		return NewFile(""), cmd.Error("cannot add content to file %s : %v", target.String(), err)
	}
	return target, nil
}
func (f File) DecompressTarTo(fd Folder) (File, *cmd.XbeeError) {
	fd.EnsureExists()
	target := fd.ChildFile(f.BaseWithoutExtension())
	return f.DecompressTarToFile(target)
}
func (f File) DecompressTarToFile(target File) (File, *cmd.XbeeError) {
	sourceReader, err := os.Open(f.String())
	if err != nil {
		return NewFile(""), cmd.Error("cannot open file %s : %v", f.String(), err)
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
			return NewFile(""), cmd.Error("cannot create gzip reader for %s : %v", f.String(), err)
		}
		reader = gr
		closer = gr
	case "bzip2", "bz2":
		reader = bzip2.NewReader(sourceReader)
	case "xz":
		// Créer un lecteur gzip
		reader, err = xz.NewReader(sourceReader)
		if err != nil {
			return NewFile(""), cmd.Error("cannot open reader for %s : %v", f.String(), err)
		}
	default:
		return NewFile(""), cmd.Error("unknown compression support only gz, zip, xz or bzip2 are supported, actual is [%s]", extension)
	}
	if closer != nil {
		defer closer.Close()
	}

	// Créer le fichier de destination

	writer, err := os.Create(target.String())
	if err != nil {
		return NewFile(""), cmd.Error("cannot create file %s : %v", target.String(), err)
	}
	defer writer.Close()

	// Copier le contenu décompressé dans le fichier de destination
	_, err = io.Copy(writer, reader)
	if err != nil {
		return NewFile(""), cmd.Error("cannot add content to file %s : %v", target.String(), err)
	}
	return target, nil
}
func (f File) DecompressTar() (File, *cmd.XbeeError) {
	target := f.Dir().ChildFile(f.BaseWithoutExtension())
	return f.DecompressTarToFile(target)
}

func (f File) Unzip() *cmd.XbeeError {
	// Ouvre le fichier ZIP
	dest := f.Dir().String()
	r, err := zip.OpenReader(f.String())
	if err != nil {
		return cmd.Error("cannot open zip reader for %s : %v", f.String(), err)
	}
	defer r.Close()

	// Parcourt chaque fichier dans l'archive ZIP
	for _, file := range r.File {
		// Crée un chemin d'extraction complet pour le fichier/dossier
		fpath := filepath.Join(dest, file.Name)

		// Vérifie si c'est un dossier
		if file.FileInfo().IsDir() {
			// Crée le dossier
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		// Crée les dossiers parents du fichier s'ils n'existent pas déjà
		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return cmd.Error("cannot create dir for %s : %v", fpath, err)
		}

		// Ouvre le fichier dans l'archive ZIP
		rc, err := file.Open()
		if err != nil {
			return cmd.Error("cannot open file for %s : %v", fpath, err)
		}
		defer rc.Close()

		// Crée le fichier sur le système de fichiers local
		outFile, err := os.Create(fpath)
		if err != nil {
			return cmd.Error("cannot create file for %s : %v", fpath, err)
		}
		defer outFile.Close()

		// Copie le contenu du fichier ZIP vers le fichier local
		_, err = io.Copy(outFile, rc)
		if err != nil {
			return cmd.Error("cannot add content to file %s : %v", fpath, err)
		}
	}

	return nil
}
