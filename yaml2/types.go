package yaml2

import (
	"strconv"

	"gopkg.in/yaml.v3"
)

func IsString(node *yaml.Node) bool {
	if node == nil {
		return false
	}

	if node.Kind != yaml.ScalarNode {
		return false
	}

	// Cas explicite
	if node.Tag == "!!string" {
		return true
	}

	// Cas implicite : on considère que tout scalaire non typé
	// ou typé comme string est un string
	switch node.Tag {
	case "", "!!str":
		return true
	}

	return false
}

func IsMap(node *yaml.Node) bool {
	if node == nil || node.Kind != yaml.MappingNode {
		return false
	}

	// Une map doit contenir un nombre pair d’éléments (clé/valeur)
	return len(node.Content)%2 == 0
}

func IsSequence(node *yaml.Node) bool {
	if node == nil || node.Kind != yaml.SequenceNode {
		return false
	}

	// Une séquence peut être vide, donc pas de contrainte stricte ici
	for i, child := range node.Content {
		if child == nil {
			return false // important pour éviter ton panic yaml.Marshal
		}
		_ = i // utile si tu veux logger le chemin plus tard
	}

	return true
}

func IsNumber(node *yaml.Node) bool {
	if node == nil || node.Kind != yaml.ScalarNode {
		return false
	}

	// Cas standard YAML
	if node.Tag == "!!int" || node.Tag == "!!float" {
		return true
	}

	// Fallback : tentative de parsing
	if _, err := strconv.ParseInt(node.Value, 10, 64); err == nil {
		return true
	}

	if _, err := strconv.ParseFloat(node.Value, 64); err == nil {
		return true
	}

	return false
}

func IsInt(node *yaml.Node) bool {
	if node == nil || node.Kind != yaml.ScalarNode {
		return false
	}
	if node.Tag == "!!int" {
		return true
	}
	_, err := strconv.ParseInt(node.Value, 10, 64)
	return err == nil
}

func IsFloat(node *yaml.Node) bool {
	if node == nil || node.Kind != yaml.ScalarNode {
		return false
	}
	if node.Tag == "!!float" {
		return true
	}
	_, err := strconv.ParseFloat(node.Value, 64)
	return err == nil
}

func IsPositiveNumber(node *yaml.Node) bool {
	if node == nil || node.Kind != yaml.ScalarNode {
		return false
	}

	// Cas YAML typé
	switch node.Tag {
	case "!!int":
		v, err := strconv.ParseInt(node.Value, 10, 64)
		return err == nil && v > 0

	case "!!float":
		v, err := strconv.ParseFloat(node.Value, 64)
		return err == nil && v > 0
	}

	// Fallback si pas de tag fiable
	if v, err := strconv.ParseInt(node.Value, 10, 64); err == nil {
		return v > 0
	}

	if v, err := strconv.ParseFloat(node.Value, 64); err == nil {
		return v > 0
	}

	return false
}

func Int(node *yaml.Node) int {
	if !IsInt(node) {
		panic("int node is not an integer")
	}
	result, _ := strconv.Atoi(node.Value)
	return result
}
