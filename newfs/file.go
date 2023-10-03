package newfs

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/iodasolutions/xbee-common/cmd"
	"github.com/iodasolutions/xbee-common/template"
	"github.com/mholt/archiver/v3"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
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
func (f File) OpenFileForCreation() *os.File {
	f.Dir().Create()
	fd, err := os.Create(string(f))
	if err != nil {
		panic(fmt.Errorf("cannot create file %s : %v", f, err))
	}
	return fd
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
	fp := f.OpenFileForCreation()
	defer fp.Close()
	if err := template.OutputWithTemplate(templateS, fp, data, funcMap); err != nil {
		return cmd.Error("cannot parse [%s] with variables [%s]: %v", templateS, data, err)
	}
	return nil
}

func (f File) SetContentBytes(content []byte) {
	w := f.OpenFileForCreation()
	buf := bytes.NewBuffer(content)
	if _, err := io.Copy(w, buf); err != nil {
		panic(err)
	}
	defer w.Close()
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
func (f File) DoTargGz() File {
	tarGz := f.Dir().ChildFile(fmt.Sprintf("%s.tar.gz", f.Base()))
	tarGz.EnsureDelete()
	if err := archiver.Archive([]string{f.String()}, tarGz.String()); err != nil {
		panic(fmt.Errorf("cannot compress file %s : %v\n", f, err))
	}
	return tarGz
}
func (f File) DoZip() File {
	zipFile := f.Dir().ChildFile(fmt.Sprintf("%s.zip", f.Base()))
	zipFile.EnsureDelete()
	if err := archiver.Archive([]string{f.String()}, zipFile.String()); err != nil {
		panic(fmt.Errorf("cannot compress file %s : %v\n", f, err))
	}
	return zipFile
}

func (f File) ExtractTo(outputDir Folder) *cmd.XbeeError {
	if err := archiver.Unarchive(string(f), outputDir.String()); err != nil {
		return cmd.Error("cannot extract %s to %s: %v", f, outputDir, err)
	}
	return nil
}

func (f File) Untar() *cmd.XbeeError {
	return f.ExtractTo(f.Dir())
}
