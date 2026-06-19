package yaml2

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/iodasolutions/xbee-common/cmd"
	"github.com/iodasolutions/xbee-common/newfs"
	"github.com/iodasolutions/xbee-common/template"
	"github.com/iodasolutions/xbee-common/util"
	"gopkg.in/yaml.v3"
)

func CloneNode(n *yaml.Node) *yaml.Node {
	if n == nil {
		return nil
	}

	clone := *n // copie de la structure

	if n.Content != nil {
		clone.Content = make([]*yaml.Node, len(n.Content))
		for i, child := range n.Content {
			clone.Content[i] = CloneNode(child)
		}
	}

	if n.Alias != nil {
		clone.Alias = CloneNode(n.Alias)
	}

	return &clone
}

func LoadYamlDocument(f newfs.File) (*yaml.Node, *cmd.XbeeError) {
	data, err := os.ReadFile(f.String())
	if err != nil {
		return nil, cmd.Error("Error reading file", err)
	}

	var doc yaml.Node

	err = yaml.Unmarshal(data, &doc)
	if err != nil {
		return nil, cmd.Error("Error parsing yaml2", err)
	}

	return &doc, nil
}
func NewNodeFromFile(f newfs.File) (*yaml.Node, *cmd.XbeeError) {
	doc, err := LoadYamlDocument(f)
	if err != nil {
		return nil, err
	}
	if len(doc.Content) > 0 {
		return doc.Content[0], nil
	}
	return doc, nil
}
func FindNodeNoError(root *yaml.Node, path string) *yaml.Node {
	y, _ := FindNode(root, path)
	return y
}
func FindNode(root *yaml.Node, path string) (*yaml.Node, *cmd.XbeeError) {

	parts := strings.Split(path, ".")
	current := root

	for _, part := range parts {

		index := -1

		// gestion des tableaux : skills[1]
		if strings.Contains(part, "[") {
			p := strings.Index(part, "[")
			idx := strings.Index(part, "]")

			if p < 0 || idx < 0 {
				return nil, cmd.Error("invalid path syntax: %s", part)
			}

			i, err := strconv.Atoi(part[p+1 : idx])
			if err != nil {
				return nil, cmd.Error("invalid path syntax: %s", part)
			}

			index = i
			part = part[:p]
		}

		// mapping
		if part != "" {
			if current.Kind != yaml.MappingNode {
				return nil, cmd.Error("not a mapping node at %s", part)
			}

			found := false

			for i := 0; i < len(current.Content); i += 2 {

				key := current.Content[i]
				value := current.Content[i+1]

				if key.Value == part {
					current = value
					found = true
					break
				}
			}

			if !found {
				return nil, cmd.Error("key not found: %s", part)
			}
		}

		// séquence
		if index >= 0 {

			if current.Kind != yaml.SequenceNode {
				return nil, cmd.Error("not a sequence for index access")
			}

			if index >= len(current.Content) {
				return nil, cmd.Error("index out of range")
			}

			current = current.Content[index]
		}
	}

	return current, nil
}

func EnsurePathNoError(root *yaml.Node, path string) *yaml.Node {
	y, err := FindNode(root, path)
	if err != nil {
		panic(err)
	}
	return y
}
func EnsurePath(root *yaml.Node, path string) (*yaml.Node, *cmd.XbeeError) {

	current := root

	parts := strings.Split(path, ".")

	for _, part := range parts {

		index := -1

		// gestion des tableaux : skills[2]
		if strings.Contains(part, "[") {
			p := strings.Index(part, "[")
			idx := strings.Index(part, "]")

			if p < 0 || idx < 0 {
				return nil, cmd.Error("invalid path syntax: %s", part)
			}

			i, err := strconv.Atoi(part[p+1 : idx])
			if err != nil {
				return nil, cmd.Error("invalid path syntax: %s", part)
			}

			index = i
			part = part[:p]
		}

		// --- mapping (clé) ---
		if part != "" {

			// si current n'est pas un mapping, on le transforme
			if current.Kind != yaml.MappingNode {
				current.Kind = yaml.MappingNode
				current.Tag = "!!map"
				current.Content = []*yaml.Node{}
			}

			var next *yaml.Node
			found := false

			for i := 0; i < len(current.Content); i += 2 {
				k := current.Content[i]
				v := current.Content[i+1]

				if k.Value == part {
					next = v
					found = true
					break
				}
			}

			// si absent → création
			if !found {

				keyNode := &yaml.Node{
					Kind:  yaml.ScalarNode,
					Tag:   "!!str",
					Value: part,
				}

				valueNode := &yaml.Node{}

				current.Content = append(current.Content, keyNode, valueNode)

				next = valueNode
			}

			current = next
		}

		// --- séquence ---
		if index >= 0 {

			// transformer en séquence si nécessaire
			if current.Kind != yaml.SequenceNode {
				current.Kind = yaml.SequenceNode
				current.Tag = "!!seq"
				current.Content = []*yaml.Node{}
			}

			// agrandir si nécessaire
			for len(current.Content) <= index {
				current.Content = append(current.Content, &yaml.Node{})
			}

			current = current.Content[index]
		}
	}
	if current == nil {
		panic("This should never happen")
	} else {
		if current.Kind == 0 {
			current.Kind = yaml.MappingNode
			current.Tag = "!!map"
		}
	}
	return current, nil
}

func FindCommonScalarLeafKeys(node1, node2 *yaml.Node) []string {
	if node1 == nil || node2 == nil {
		return nil
	}
	keys1 := make(map[string]struct{})
	keys2 := make(map[string]struct{})

	collectScalarLeafKeys(node1, "", keys1)
	collectScalarLeafKeys(node2, "", keys2)

	var common []string
	for k := range keys1 {
		if _, ok := keys2[k]; ok {
			common = append(common, k)
		}
	}

	sort.Strings(common)
	return common
}

func collectScalarLeafKeys(node *yaml.Node, path string, out map[string]struct{}) {
	if node == nil {
		return
	}

	switch node.Kind {
	case yaml.DocumentNode:
		for _, child := range node.Content {
			collectScalarLeafKeys(child, path, out)
		}

	case yaml.MappingNode:
		for i := 0; i+1 < len(node.Content); i += 2 {
			k := node.Content[i]
			v := node.Content[i+1]

			nextPath := k.Value
			if path != "" {
				nextPath = path + "." + k.Value
			}

			collectScalarLeafKeys(v, nextPath, out)
		}

	case yaml.SequenceNode:
		for i, child := range node.Content {
			nextPath := path + "[" + strconv.Itoa(i) + "]"
			collectScalarLeafKeys(child, nextPath, out)
		}

	case yaml.ScalarNode:
		if path != "" {
			out[path] = struct{}{}
		}

	case yaml.AliasNode:
		// au besoin, on suit l'alias
		if node.Alias != nil {
			collectScalarLeafKeys(node.Alias, path, out)
		}
	}
}

func PromoteToMap(node *yaml.Node, key string) {
	// sauvegarde de l'ancienne valeur
	oldValue := node.Content

	// transformation en mapping
	node.Kind = yaml.MappingNode
	node.Tag = "!!map"

	node.Content = []*yaml.Node{
		{
			Kind:  yaml.ScalarNode,
			Tag:   "!!str",
			Value: key,
		},
		{
			Kind:    yaml.SequenceNode,
			Tag:     "!!seq",
			Content: oldValue,
		},
	}
	// important : vider l'ancien contenu scalaire
	node.Value = ""
}

func PromoteMapToList(node *yaml.Node) {
	if node.Kind != yaml.MappingNode {
		panic("This should never happen")
	}
	node.Kind = yaml.SequenceNode
	node.Tag = "!!seq"
	node.Value = ""
	oldContent := node.Content
	node.Content = []*yaml.Node{
		{
			Kind:    yaml.MappingNode,
			Tag:     "!!map",
			Content: oldContent,
		},
	}
}

func AddStringToMap(node *yaml.Node, key string, value string) {

	keyNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Tag:   "!!str",
		Value: key,
	}

	valueNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Tag:   "!!str",
		Value: value,
	}

	node.Content = append(node.Content, keyNode, valueNode)
}

func PrintNode(node *yaml.Node) {
	out, err := yaml.Marshal(node)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(out))
}

func AddMap(node *yaml.Node, key string) *yaml.Node {
	keyNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Tag:   "!!str",
		Value: key,
	}

	valueNode := &yaml.Node{
		Kind:    yaml.MappingNode,
		Tag:     "!!map",
		Content: []*yaml.Node{},
	}

	node.Content = append(node.Content, keyNode, valueNode)

	return valueNode
}

func AddMapValue(node *yaml.Node, key string, value *yaml.Node) *yaml.Node {
	keyNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Tag:   "!!str",
		Value: key,
	}
	node.Content = append(node.Content, keyNode, value)
	return value
}

func indexOf(node *yaml.Node, key string) int {
	for i := 0; i < len(node.Content); i += 2 {
		k := node.Content[i]
		if k.Value == key {
			return i
		}
	}
	return -1
}

func AddOrReplaceSequence(node *yaml.Node, key string, values []*yaml.Node) *yaml.Node {
	if node.Kind != yaml.MappingNode {
		panic("This should never happen")
	}
	valueNode := &yaml.Node{
		Kind:    yaml.SequenceNode,
		Tag:     "!!seq",
		Content: values,
	}
	index := indexOf(node, key)
	if index == -1 {
		keyNode := &yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!str",
			Value: key,
		}
		node.Content = append(node.Content, keyNode, valueNode)
	} else {
		node.Content[index+1] = valueNode
	}

	return valueNode
}

func PromoteMapToMap(node *yaml.Node, newKey string) {
	oldContent := node.Content

	node.Kind = yaml.MappingNode
	node.Tag = "!!map"
	node.Value = ""
	node.Anchor = ""
	node.Alias = nil

	node.Content = []*yaml.Node{
		{
			Kind:  yaml.ScalarNode,
			Tag:   "!!str",
			Value: newKey,
		},
		{
			Kind:    yaml.MappingNode,
			Tag:     "!!map",
			Content: oldContent,
		},
	}
}

func IsStringSequence(node *yaml.Node) bool {

	if node == nil || node.Kind != yaml.SequenceNode {
		return false
	}

	for _, child := range node.Content {
		if child.Kind != yaml.ScalarNode {
			return false
		}
		if child.Tag != "!!str" && child.Tag != "" {
			return false
		}
	}

	return true
}

func FirstPropertyAmong(node *yaml.Node, keys ...string) string {
	set := util.SetFromStringSlice(keys)
	for i := 0; i < len(node.Content); i += 2 {
		key := node.Content[i].Value
		if set.HasElement(key) {
			return key
		}
	}
	return ""
}

func AddStringList(node *yaml.Node, key string, values []string) {
	// sécurité : on s'assure que c'est bien un mapping
	if node.Kind != yaml.MappingNode {
		panic("node must be a MappingNode")
	}

	// construire la clé
	keyNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Tag:   "!!str",
		Value: key,
	}

	// construire la séquence
	seqNode := &yaml.Node{
		Kind: yaml.SequenceNode,
		Tag:  "!!seq",
	}

	for _, v := range values {
		seqNode.Content = append(seqNode.Content, &yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!str",
			Value: v,
		})
	}

	// ajouter clé + valeur
	node.Content = append(node.Content, keyNode, seqNode)
}

func StringList(node *yaml.Node) (result []string) {
	if node == nil {
		return
	}
	for _, elt := range node.Content {
		result = append(result, elt.Value)
	}
	return
}

func RemovePropertyFromMap(node *yaml.Node, key string) {
	index := -1
	for i := 0; i < len(node.Content); i += 2 {
		theKey := node.Content[i].Value
		if theKey == key {
			index = i
			break
		}
	}
	if index > -1 {
		node.Content = append(node.Content[:index], node.Content[index+2:]...)
	}
}

func FindMapValue(node *yaml.Node, key string) (*yaml.Node, int) {
	if node.Kind != yaml.MappingNode {
		return nil, -1
	}
	for i := 0; i < len(node.Content); i += 2 {
		k := node.Content[i]
		if k.Value == key {
			return node.Content[i+1], i + 1
		}
	}
	return nil, -1
}

func SetMapValue(node *yaml.Node, key string, value *yaml.Node) bool {
	if node.Kind != yaml.MappingNode {
		return false
	}
	for i := 0; i < len(node.Content); i += 2 {
		k := node.Content[i]
		if k.Value == key {
			node.Content[i+1] = value
			return true
		}
	}
	return false
}

func ParseYAMLStringNoError(s string) *yaml.Node {
	node, err := ParseYAMLString(s)
	if err != nil {
		panic(err)
	}
	return node
}

func ParseYAMLString(input string) (*yaml.Node, *cmd.XbeeError) {
	var root yaml.Node
	if err := yaml.Unmarshal([]byte(input), &root); err != nil {
		return nil, cmd.Error("failed to parse YAML: %w", err)
	}
	return &root, nil
}

func HasProperty(y *yaml.Node, key string) bool {
	if y.Kind != yaml.MappingNode {
		return false
	}
	for i := 0; i < len(y.Content); i += 2 {
		if y.Content[i].Value == key {
			return true
		}
	}
	return false
}

func PropertyStringValueFrom(node *yaml.Node, key string) string {
	aChild := Child(node, key)
	if aChild != nil {
		return aChild.Value
	}
	return ""
}

// MergeNodes fusionne src dans dst (modifie dst)
func MergeNodes(dst, src *yaml.Node) {
	if dst == nil || src == nil {
		return
	}

	// Si types différents → on remplace complètement
	if dst.Kind != src.Kind {
		*dst = *CloneNode(src)
		return
	}

	switch dst.Kind {

	case yaml.MappingNode:
		mergeMapping(dst, src)

	case yaml.SequenceNode:
		// concat
		for _, n := range src.Content {
			dst.Content = append(dst.Content, CloneNode(n))
		}

	case yaml.ScalarNode:
		dst.Value = src.Value
		dst.Tag = src.Tag

	default:
		// fallback : remplacement
		*dst = *CloneNode(src)
	}
}

func mergeMapping(dst, src *yaml.Node) {
	for i := 0; i < len(src.Content); i += 2 {
		srcKey := src.Content[i]
		srcVal := src.Content[i+1]

		found := false

		for j := 0; j < len(dst.Content); j += 2 {
			dstKey := dst.Content[j]
			dstVal := dst.Content[j+1]

			if dstKey.Value == srcKey.Value {
				MergeNodes(dstVal, srcVal)
				found = true
				break
			}
		}

		if !found {
			dst.Content = append(dst.Content,
				CloneNode(srcKey),
				CloneNode(srcVal),
			)
		}
	}
}

// SaveNodeToFile sérialise un noeud YAML et l'écrit dans un fichier.
func SaveNodeToFile(node *yaml.Node, f newfs.File) *cmd.XbeeError {
	if node == nil {
		return cmd.Error("node is nil")
	}

	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)

	if err := enc.Encode(node); err != nil {
		_ = enc.Close()
		return cmd.Error("encode yaml2: %w", err)
	}
	if err := enc.Close(); err != nil {
		return cmd.Error("close encoder: %w", err)
	}

	if err := os.WriteFile(f.String(), buf.Bytes(), 0o644); err != nil {
		return cmd.Error("write file %s: %w", f.String(), err)
	}

	return nil
}

func NodeTo[T any](node *yaml.Node) (*T, *cmd.XbeeError) {
	var out T
	if node == nil {
		return &out, cmd.Error("node is nil")
	}
	if err := node.Decode(&out); err != nil {
		return &out, cmd.Error("decode yaml: %w", err)
	}
	return &out, nil
}
func NodeToInterfaceNoError[T any](node *yaml.Node) T {
	out, err := NodeTo[T](node)
	if err != nil {
		panic(err)
	}
	return *out
}

func NodeToString(node *yaml.Node) string {
	out, err := yaml.Marshal(node)
	if err != nil {
		panic(err)
	}
	return string(out)
}

// CollectKeys retourne toutes les clés sous forme de slice de string
func LeaveKeys(node *yaml.Node) []string {
	var keys []string
	collectKeysRecursive(node, &keys)
	return keys
}

func collectKeysRecursive(node *yaml.Node, keys *[]string) {
	if node == nil {
		return
	}

	switch node.Kind {

	case yaml.MappingNode:
		// Mapping = alternance clé / valeur
		for i := 0; i < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valueNode := node.Content[i+1]

			if keyNode.Kind == yaml.ScalarNode {
				*keys = append(*keys, keyNode.Value)
			}

			// continuer sur la valeur
			collectKeysRecursive(valueNode, keys)
		}

	case yaml.SequenceNode:
		// parcourir tous les éléments
		for _, child := range node.Content {
			collectKeysRecursive(child, keys)
		}

	case yaml.DocumentNode:
		for _, child := range node.Content {
			collectKeysRecursive(child, keys)
		}
	}
}

func ChangePropertyValue(node *yaml.Node, key string, value *yaml.Node) {
	index := -1
	for i := 0; i < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		if keyNode.Value == key {
			index = i
			break
		}
	}
	if index > -1 {
		node.Content[index+1] = value
	} else {
		node.Content = append(node.Content, &yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!str",
			Value: key,
		},
			value)
	}

}

func Resolve(y *yaml.Node, ctxMap interface{}) (*yaml.Node, *cmd.XbeeError) {
	if y == nil {
		return nil, nil
	}
	s := NodeToString(y)
	if err := template.Output(&s, ctxMap, nil); err != nil {
		return nil, cmd.Error("cannot parse %s with [%v]: %v", s, ctxMap, err)
	}
	result, err := ParseYAMLString(s)
	if err != nil {
		return nil, cmd.Error("cannot parse %s with [%v]: %v", s, ctxMap, err)
	}
	return result.Content[0], nil
}

func ConvertYamlNode(node *yaml.Node) interface{} {
	switch node.Kind {

	case yaml.MappingNode:
		m := make(map[string]interface{})
		for i := 0; i < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valNode := node.Content[i+1]

			key := keyNode.Value
			m[key] = ConvertYamlNode(valNode)
		}
		return m

	case yaml.SequenceNode:
		var list []interface{}
		for _, n := range node.Content {
			list = append(list, ConvertYamlNode(n))
		}
		return list

	case yaml.ScalarNode:
		return parseScalar(node)

	default:
		return nil
	}
}

// parseScalar convertit correctement les scalaires YAML
func parseScalar(node *yaml.Node) interface{} {
	switch node.Tag {
	case "!!str":
		return node.Value
	case "!!int":
		var i int
		fmt.Sscanf(node.Value, "%d", &i)
		return i
	case "!!bool":
		return node.Value == "true"
	case "!!float":
		var f float64
		fmt.Sscanf(node.Value, "%f", &f)
		return f
	case "!!null":
		return nil
	default:
		return node.Value
	}
}

func KeysForNode(node *yaml.Node) []string {
	var keys []string
	for i := 0; i < len(node.Content); i += 2 {
		keys = append(keys, node.Content[i].Value)
	}
	return keys
}

func MapToYAMLNode(m map[string]interface{}) (*yaml.Node, *cmd.XbeeError) {
	// 1. marshal de la map en YAML
	data, err := yaml.Marshal(m)
	if err != nil {
		return nil, cmd.Error("marshal map to yaml2: %w", err)
	}

	// 2. un document YAML complet est relu dans un yaml2.Node
	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, cmd.Error("unmarshal yaml2 into node: %w", err)
	}

	// 3. doc est un DocumentNode ; son contenu[0] est en général le vrai mapping
	if len(doc.Content) == 0 {
		return nil, cmd.Error("empty YAML document")
	}

	return doc.Content[0], nil
}

func NodeLeavesToString(node *yaml.Node) string {
	if node == nil {
		return ""
	}

	var parts []string
	collectLeaves(node, "", &parts)

	// tri optionnel pour avoir une sortie stable
	sort.Strings(parts)

	return strings.Join(parts, "|")
}

func collectLeaves(node *yaml.Node, path string, parts *[]string) {
	if node == nil {
		return
	}

	switch node.Kind {
	case yaml.DocumentNode:
		for _, child := range node.Content {
			collectLeaves(child, path, parts)
		}

	case yaml.MappingNode:
		for i := 0; i+1 < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valNode := node.Content[i+1]

			key := keyNode.Value
			nextPath := key
			if path != "" {
				nextPath = path + "." + key
			}

			collectLeaves(valNode, nextPath, parts)
		}

	case yaml.SequenceNode:
		for i, child := range node.Content {
			nextPath := fmt.Sprintf("%s[%d]", path, i)
			if path == "" {
				nextPath = fmt.Sprintf("[%d]", i)
			}
			collectLeaves(child, nextPath, parts)
		}

	case yaml.ScalarNode:
		*parts = append(*parts, path+"="+node.Value)

	case yaml.AliasNode:
		if node.Alias != nil {
			collectLeaves(node.Alias, path, parts)
		}
	}
}

func Child(y *yaml.Node, key string) *yaml.Node {
	if y == nil {
		return nil
	}
	if HasProperty(y, key) {
		return FindNodeNoError(y, key)
	}
	return nil
}

func EnsureChildSequence(node *yaml.Node, key string) {
	if !HasProperty(node, key) {
		AddOrReplaceSequence(node, key, nil)
	} else {
		child := Child(node, key)
		if child.Kind != yaml.SequenceNode {
			panic("expected a sequence node")
		}
	}
}

// GetLeafPaths retourne tous les chemins vers les feuilles
func GetLeafPaths(node *yaml.Node) []string {
	var paths []string
	walk(node, "", &paths)
	return paths
}

func walk(node *yaml.Node, currentPath string, paths *[]string) {
	switch node.Kind {

	case yaml.MappingNode:
		// map = clé/valeur alternées
		for i := 0; i < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valueNode := node.Content[i+1]

			newPath := keyNode.Value
			if currentPath != "" {
				newPath = currentPath + "." + keyNode.Value
			}

			walk(valueNode, newPath, paths)
		}

	case yaml.SequenceNode:
		for i, item := range node.Content {
			newPath := fmt.Sprintf("%s[%d]", currentPath, i)
			walk(item, newPath, paths)
		}

	case yaml.ScalarNode:
		// feuille atteinte
		*paths = append(*paths, currentPath)
	}
}

func EnsureDefault(y *yaml.Node, path string, value string) {
	n := FindNodeNoError(y, path)
	if n.Value == "" {
		n.Value = value
	}
}

// EnsureScalarWithDefaultPath garantit qu'un chemin (format "a.b.c") existe,
// que la feuille est un scalaire, et qu'elle contient une valeur.
// Si vide, la valeur par défaut est appliquée.
func EnsureScalarWithDefaultPath(root *yaml.Node, path string, defaultValue string) *cmd.XbeeError {
	if root == nil {
		return cmd.Error("root is nil")
	}

	// Nettoyage du path
	path = strings.TrimSpace(path)
	if path == "" {
		return cmd.Error("path is empty")
	}

	parts := strings.Split(path, ".")

	// Si root non initialisé → map
	if root.Kind == 0 {
		root.Kind = yaml.MappingNode
		root.Tag = "!!map"
	}

	if root.Kind != yaml.MappingNode {
		return cmd.Error("root is not a mapping node")
	}

	current := root

	for i, key := range parts {
		key = strings.TrimSpace(key)
		if key == "" {
			return cmd.Error("invalid empty key in path at position %d", i)
		}

		if current.Kind != yaml.MappingNode {
			return cmd.Error("node at %v is not a mapping", parts[:i])
		}

		var next *yaml.Node

		// Recherche de la clé
		for j := 0; j < len(current.Content); j += 2 {
			k := current.Content[j]
			v := current.Content[j+1]

			if k.Value == key {
				next = v
				break
			}
		}

		// Création si absent
		if next == nil {
			keyNode := &yaml.Node{
				Kind:  yaml.ScalarNode,
				Tag:   "!!str",
				Value: key,
			}

			if i == len(parts)-1 {
				// feuille
				next = &yaml.Node{
					Kind:  yaml.ScalarNode,
					Tag:   "!!str",
					Value: defaultValue,
				}
			} else {
				next = &yaml.Node{
					Kind: yaml.MappingNode,
					Tag:  "!!map",
				}
			}

			current.Content = append(current.Content, keyNode, next)
		}

		// Si feuille
		if i == len(parts)-1 {
			if next.Kind != yaml.ScalarNode {
				next.Kind = yaml.ScalarNode
				next.Tag = "!!str"
				next.Content = nil
			}

			if next.Value == "" {
				next.Value = defaultValue
			}

			return nil
		}

		current = next
	}

	return nil
}

func EnsureStringList(node *yaml.Node, key string) (*yaml.Node, *cmd.XbeeError) {
	if node == nil {
		return nil, cmd.Error("node is nil")
	}

	if node.Kind != yaml.MappingNode {
		return nil, cmd.Error("node is not a mapping node")
	}

	// Parcours des clés
	for i := 0; i < len(node.Content); i += 2 {
		k := node.Content[i]
		v := node.Content[i+1]

		if k.Value == key {
			switch v.Kind {

			case yaml.ScalarNode:
				// Transformer string -> []string
				newSeq := &yaml.Node{
					Kind: yaml.SequenceNode,
					Tag:  "!!seq",
					Content: []*yaml.Node{
						{
							Kind:  yaml.ScalarNode,
							Tag:   "!!str",
							Value: v.Value,
						},
					},
				}
				node.Content[i+1] = newSeq
				return newSeq, nil

			case yaml.SequenceNode:
				// Vérifier que tous les éléments sont des strings
				for _, item := range v.Content {
					if item.Kind != yaml.ScalarNode {
						return nil, cmd.Error("sequence contains non-scalar value for key %s", key)
					}
				}
				return v, nil

			default:
				return nil, cmd.Error("invalid type for key %s: expected string or list", key)
			}
		}
	}

	// Clé absente → on la crée avec une liste vide
	result := &yaml.Node{
		Kind:    yaml.SequenceNode,
		Tag:     "!!seq",
		Content: []*yaml.Node{},
	}
	node.Content = append(node.Content,
		&yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!str",
			Value: key,
		},
		result,
	)

	return result, nil
}

func Leaves(node *yaml.Node) []*yaml.Node {
	var result []*yaml.Node

	var walk func(n *yaml.Node)
	walk = func(n *yaml.Node) {
		if n == nil {
			return
		}

		switch n.Kind {
		case yaml.DocumentNode:
			if len(n.Content) == 0 {
				result = append(result, n)
				return
			}
			for _, child := range n.Content {
				walk(child)
			}

		case yaml.MappingNode:
			if len(n.Content) == 0 {
				result = append(result, n)
				return
			}

			// Dans un mapping, Content contient :
			// key0, value0, key1, value1, ...
			// Les clés ne sont généralement pas considérées comme des feuilles utiles.
			for i := 1; i < len(n.Content); i += 2 {
				walk(n.Content[i])
			}

		case yaml.SequenceNode:
			if len(n.Content) == 0 {
				result = append(result, n)
				return
			}
			for _, child := range n.Content {
				walk(child)
			}

		default:
			result = append(result, n)
		}
	}

	walk(node)
	return result
}

type Leaf struct {
	Path  string
	Value string
	Node  *yaml.Node
}

func LeafValues(node *yaml.Node) []Leaf {
	var result []Leaf

	var walk func(n *yaml.Node, path string)
	walk = func(n *yaml.Node, path string) {
		if n == nil {
			return
		}

		switch n.Kind {
		case yaml.DocumentNode:
			if len(n.Content) > 0 {
				walk(n.Content[0], path)
			}

		case yaml.MappingNode:
			if len(n.Content) == 0 {
				result = append(result, Leaf{Path: path, Value: "", Node: n})
				return
			}

			for i := 0; i+1 < len(n.Content); i += 2 {
				key := n.Content[i]
				value := n.Content[i+1]

				childPath := key.Value
				if path != "" {
					childPath = path + "." + key.Value
				}

				walk(value, childPath)
			}

		case yaml.SequenceNode:
			if len(n.Content) == 0 {
				result = append(result, Leaf{Path: path, Value: "", Node: n})
				return
			}

			for i, child := range n.Content {
				childPath := fmt.Sprintf("%s[%d]", path, i)
				walk(child, childPath)
			}

		case yaml.ScalarNode, yaml.AliasNode:
			result = append(result, Leaf{
				Path:  path,
				Value: n.Value,
				Node:  n,
			})
		default:
			result = append(result, Leaf{
				Path:  path,
				Value: n.Value,
				Node:  n,
			})
		}
	}

	walk(node, "")
	return result
}
