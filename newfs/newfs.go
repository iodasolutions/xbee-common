package newfs

import (
	"fmt"
	"github.com/iodasolutions/xbee-common/cmd"
	"os"
)

const (
	YamlExt = ".yaml"
	jsonExt = ".json"
)

// by chatGTP
func moveDirectory(srcDir string, destDir string) *cmd.XbeeError {
	// Vérifie si le répertoire source existe
	srcInfo, err := os.Stat(srcDir)
	if err != nil {
		return cmd.Error("cannot access properties of %s: %v", srcDir, err)
	}
	if !srcInfo.IsDir() {
		return cmd.Error("%s n'est pas un répertoire", srcDir)
	}

	// Crée le répertoire de destination s'il n'existe pas
	err = os.MkdirAll(destDir, srcInfo.Mode())
	if err != nil {
		return cmd.Error("cannot create directories %s: %v", destDir, err)
	}

	// Ouvre le répertoire source
	dir, err := os.Open(srcDir)
	if err != nil {
		return cmd.Error("cannot open directory %s: %v", srcDir, err)
	}
	defer dir.Close()

	// Parcours les fichiers du répertoire source
	files, err := dir.Readdir(-1)
	if err != nil {
		return cmd.Error("cannot read directory %s: %v", srcDir, err)
	}
	for _, file := range files {
		// Construit le chemin absolu pour le fichier source
		srcPath := fmt.Sprintf("%s/%s", srcDir, file.Name())
		// Construit le chemin absolu pour le fichier de destination
		destPath := fmt.Sprintf("%s/%s", destDir, file.Name())

		if file.IsDir() {
			// Déplace les sous-répertoires récursivement
			err2 := moveDirectory(srcPath, destPath)
			if err2 != nil {
				return err2
			}
		} else {
			// Déplace les fichiers vers le répertoire de destination
			err = os.Rename(srcPath, destPath)
			if err != nil {
				return cmd.Error("cannot rename %s to %s: %v", srcPath, destPath, err)
			}
		}
	}

	// Supprime le répertoire source vide
	err = os.Remove(srcDir)
	if err != nil {
		return cmd.Error("cannot remove directory %s: %v", srcDir, err)
	}

	return nil
}
