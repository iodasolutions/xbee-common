package newfs

import (
	"fmt"
	"os"
)

const (
	YamlExt = ".yaml"
	jsonExt = ".json"
)

func NewTestFolder() Folder {
	f := tmpDir.RandomChildFolder().Create()
	if err := os.Chdir(f.String()); err != nil {
		panic(err)
	}
	return f
}

// by chatGTP
func moveDirectory(srcDir string, destDir string) error {
	// Vérifie si le répertoire source existe
	srcInfo, err := os.Stat(srcDir)
	if err != nil {
		return err
	}
	if !srcInfo.IsDir() {
		return fmt.Errorf("%s n'est pas un répertoire", srcDir)
	}

	// Crée le répertoire de destination s'il n'existe pas
	err = os.MkdirAll(destDir, srcInfo.Mode())
	if err != nil {
		return err
	}

	// Ouvre le répertoire source
	dir, err := os.Open(srcDir)
	if err != nil {
		return err
	}
	defer dir.Close()

	// Parcours les fichiers du répertoire source
	files, err := dir.Readdir(-1)
	if err != nil {
		return err
	}
	for _, file := range files {
		// Construit le chemin absolu pour le fichier source
		srcPath := fmt.Sprintf("%s/%s", srcDir, file.Name())
		// Construit le chemin absolu pour le fichier de destination
		destPath := fmt.Sprintf("%s/%s", destDir, file.Name())

		if file.IsDir() {
			// Déplace les sous-répertoires récursivement
			err = moveDirectory(srcPath, destPath)
			if err != nil {
				return err
			}
		} else {

			// Déplace les fichiers vers le répertoire de destination
			err = os.Rename(srcPath, destPath)
			if err != nil {
				return err
			}
		}
	}

	// Supprime le répertoire source vide
	err = os.Remove(srcDir)
	if err != nil {
		return err
	}

	return nil
}
