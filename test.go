package main

import (
	"fmt"
)

// mergeMaps fusionne deux maps de type map[string]interface{}. Si des clés existent dans les deux maps,
// et si les valeurs associées sont des maps, elles sont fusionnées récursivement.
// Sinon, les valeurs de map2 écrasent celles de map1.
func mergeMaps(map1, map2 map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// Ajouter tous les éléments de map1 au résultat
	for k, v := range map1 {
		result[k] = v
	}

	// Ajouter ou fusionner les éléments de map2
	for k, v := range map2 {
		if v1, ok := result[k]; ok {
			// Si la clé existe déjà et que les deux valeurs sont des maps, on les fusionne récursivement
			if map1Nested, ok1 := v1.(map[string]interface{}); ok1 {
				if map2Nested, ok2 := v.(map[string]interface{}); ok2 {
					result[k] = mergeMaps(map1Nested, map2Nested)
					continue
				}
			}
		}
		// Sinon, on remplace directement la valeur
		result[k] = v
	}

	return result
}

func main() {
	// Exemple de données YAML chargées sous forme de maps
	map1 := map[string]interface{}{
		"key1": "value1",
		"key2": map[string]interface{}{
			"subkey1": "subvalue1",
			"subkey2": "subvalue2",
		},
		"key3": "value3",
	}

	map2 := map[string]interface{}{
		"key2": map[string]interface{}{
			"subkey2": "new_subvalue2",
			"subkey3": "subvalue3",
		},
		"key4": "value4",
	}

	// Fusion des deux maps
	mergedMap := mergeMaps(map1, map2)

	// Afficher le résultat
	fmt.Printf("Merged Map: %+v\n", mergedMap)
}
