package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
)

const (
	outputDir = "output"

	minDepth = 1
	maxDepth = 4

	minTrashPerBranch = 2
	maxTrashPerBranch = 3
)

var themedBranches = []string{
	"hr_department",
	"trash",
	"server_configs",
	"old_projects",
}

var itemRoles = []string{
	"correct_code",
	"trap_1",
	"trap_2",
	"trap_3",
}

var branchTrashCategory = map[string]string{
	"hr_department":  "hr",
	"trash":          "generic",
	"server_configs": "configs",
	"old_projects":   "projects",
}

var subfolderPool = []string{
	"drafts",
	"archive",
	"old",
	"misc",
	"backup",
	"2023",
	"tmp",
	"staging",
}

type FileEntry struct {
	Path       string `json:"path"`
	ContentKey string `json:"content_key"`
}

type TreeManifest struct {
	TreeID string      `json:"tree_id"`
	Files  []FileEntry `json:"files"`
}

func main() {
	countStr := os.Getenv("STUDENT_COUNT")
	if countStr == "" {
		fmt.Fprintln(os.Stderr, "Ошибка: переменная окружения STUDENT_COUNT не задана.")
		fmt.Fprintln(os.Stderr, "Пример запуска: STUDENT_COUNT=22 ./home-structure-generator")
		os.Exit(1)
	}

	count, err := strconv.Atoi(countStr)
	if err != nil || count <= 0 {
		fmt.Fprintf(os.Stderr, "Ошибка: STUDENT_COUNT должен быть положительным целым числом, получено: %q\n", countStr)
		os.Exit(1)
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка создания папки %s: %v\n", outputDir, err)
		os.Exit(1)
	}

	for i := 1; i <= count; i++ {
		treeID := fmt.Sprintf("tree_%d", i)
		manifest := generateTree(treeID)

		if err := writeManifest(manifest); err != nil {
			fmt.Fprintf(os.Stderr, "Ошибка записи манифеста для %s: %v\n", treeID, err)
			os.Exit(1)
		}
	}

	fmt.Printf("Готово. Сгенерировано %d манифестов в %s/\n", count, outputDir)
}

func generateTree(treeID string) TreeManifest {
	var files []FileEntry

	shuffledRoles := shuffleStrings(itemRoles)

	for idx, branch := range themedBranches {
		role := shuffledRoles[idx]
		itemPath := buildRandomPath(branch, itemFileName(role))
		files = append(files, FileEntry{Path: itemPath, ContentKey: role})

		trashCount := minTrashPerBranch + rand.Intn(maxTrashPerBranch-minTrashPerBranch+1)
		category := branchTrashCategory[branch]
		for t := 0; t < trashCount; t++ {
			trashKey := fmt.Sprintf("trash:%s:%d", category, t+1)
			trashPath := buildRandomPath(branch, trashFileName(category, t+1))
			files = append(files, FileEntry{Path: trashPath, ContentKey: trashKey})
		}
	}

	beaglePath := filepath.Join(".keys", "activation.txt")
	files = append(files, FileEntry{Path: beaglePath, ContentKey: "beagle_code"})

	return TreeManifest{TreeID: treeID, Files: files}
}

func buildRandomPath(branch, filename string) string {
	depth := minDepth + rand.Intn(maxDepth-minDepth+1)
	parts := []string{branch}

	extraFolders := depth - 1
	lastFolder := ""
	for i := 0; i < extraFolders; i++ {
		next := pickDifferentFolder(lastFolder)
		parts = append(parts, next)
		lastFolder = next
	}
	parts = append(parts, filename)

	return filepath.Join(parts...)
}

func pickDifferentFolder(exclude string) string {
	for {
		candidate := subfolderPool[rand.Intn(len(subfolderPool))]
		if candidate != exclude {
			return candidate
		}
	}
}

func itemFileName(role string) string {
	switch role {
	case "correct_code":
		return "activation.txt"
	case "trap_1":
		return "old_code.txt"
	case "trap_2":
		return "personal_note.txt"
	case "trap_3":
		return "test_access.txt"
	default:
		return "note.txt"
	}
}

func trashFileName(category string, n int) string {
	return fmt.Sprintf("%s_note_%d.txt", category, n)
}

func shuffleStrings(in []string) []string {
	out := make([]string, len(in))
	copy(out, in)
	rand.Shuffle(len(out), func(i, j int) {
		out[i], out[j] = out[j], out[i]
	})
	return out
}

func writeManifest(m TreeManifest) error {
	path := filepath.Join(outputDir, m.TreeID+".json")
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(m)
}
