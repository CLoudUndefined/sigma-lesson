package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
)

const (
	studentsCSVPath  = "../container-generator/students.csv"
	treesDir         = "../home-structure-generator/output"
	preparedTreesDir = "prepared_trees"

	trashPoolDir = "content/trash_pool"
	trapsDir     = "content/traps"
	lesson2Dir   = "content/lesson2"

	beagleCode = "SHER-L0CK"

	dockerImage      = "sigma-student-base"
	pollIntervalS    = 1
	pollTimeoutS     = 30
	traefikNetwork   = "traefik-public"
	domainEntrypoint = "websecure"
	certResolver     = "myresolver"
)

type Student struct {
	Login    string
	Password string
	Port     string
	Domain   string
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
	students, err := loadStudents(studentsCSVPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка чтения %s: %v\n", studentsCSVPath, err)
		os.Exit(1)
	}

	if err := os.MkdirAll(preparedTreesDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка создания %s: %v\n", preparedTreesDir, err)
		os.Exit(1)
	}

	var startLines []string
	var stopLines []string
	var restartLines []string

	for i, student := range students {
		treeIdx := i + 1
		treePath := filepath.Join(treesDir, fmt.Sprintf("tree_%d.json", treeIdx))

		manifest, err := loadManifest(treePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Ошибка чтения манифеста %s (для %s): %v\n", treePath, student.Login, err)
			os.Exit(1)
		}

		preparedPath := filepath.Join(preparedTreesDir, student.Login)
		if err := buildRealTree(preparedPath, manifest); err != nil {
			fmt.Fprintf(os.Stderr, "Ошибка сборки дерева для %s: %v\n", student.Login, err)
			os.Exit(1)
		}

		if err := addLesson2Content(preparedPath); err != nil {
			fmt.Fprintf(os.Stderr, "Ошибка добавления контента занятия 2 для %s: %v\n", student.Login, err)
			os.Exit(1)
		}

		containerName := "sigma-" + student.Login

		startLines = append(startLines, generateStartBlock(student, containerName, preparedPath))
		stopLines = append(stopLines, fmt.Sprintf("docker stop %s || true", containerName))
		restartLines = append(restartLines, fmt.Sprintf("docker restart %s", containerName))
	}

	if err := writeScript("build.sh", buildScript()); err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка записи build.sh: %v\n", err)
		os.Exit(1)
	}
	if err := writeScript("start_all.sh", startScript(startLines)); err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка записи start_all.sh: %v\n", err)
		os.Exit(1)
	}
	if err := writeScript("stop_all.sh", simpleScript(stopLines)); err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка записи stop_all.sh: %v\n", err)
		os.Exit(1)
	}
	if err := writeScript("restart_all.sh", simpleScript(restartLines)); err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка записи restart_all.sh: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Готово. Подготовлено %d деревьев в %s/, скрипты: build.sh, start_all.sh, stop_all.sh, restart_all.sh\n", len(students), preparedTreesDir)
}

func loadStudents(path string) ([]Student, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(records) < 2 {
		return nil, fmt.Errorf("файл пуст или содержит только заголовок")
	}

	var students []Student
	for _, rec := range records[1:] { // пропускаем заголовок
		if len(rec) < 4 {
			continue
		}
		students = append(students, Student{
			Login:    rec[0],
			Password: rec[1],
			Port:     rec[2],
			Domain:   rec[3],
		})
	}
	return students, nil
}

func loadManifest(path string) (TreeManifest, error) {
	var m TreeManifest
	data, err := os.ReadFile(path)
	if err != nil {
		return m, err
	}
	if err := json.Unmarshal(data, &m); err != nil {
		return m, err
	}
	return m, nil
}

func buildRealTree(outputDir string, manifest TreeManifest) error {
	if err := os.RemoveAll(outputDir); err != nil {
		return err
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	usedPaths := make(map[string]bool)

	for _, entry := range manifest.Files {
		fullPath, err := resolveAndWriteEntry(outputDir, entry, usedPaths)
		if err != nil {
			return fmt.Errorf("путь %s (%s): %w", entry.Path, entry.ContentKey, err)
		}
		usedPaths[fullPath] = true
	}

	return nil
}

func resolveAndWriteEntry(outputDir string, entry FileEntry, usedPaths map[string]bool) (string, error) {
	const maxAttempts = 20

	for range maxAttempts {
		content, extraName, err := resolveContent(entry.ContentKey)
		if err != nil {
			return "", err
		}

		targetPath := entry.Path
		if extraName != "" {
			targetPath = filepath.Join(entry.Path, extraName)
		}

		fullPath := filepath.Join(outputDir, targetPath)

		if usedPaths[fullPath] {
			if extraName == "" {
				return "", fmt.Errorf("неожиданная коллизия пути для не-мусорной записи: %s", fullPath)
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			return "", err
		}
		if err := os.WriteFile(fullPath, content, 0644); err != nil {
			return "", err
		}

		return fullPath, nil
	}

	return "", fmt.Errorf("не удалось найти свободный путь за %d попыток (возможно, в trash_pool слишком мало файлов для данной глубины дерева)", maxAttempts)
}

func resolveContent(key string) (content []byte, extraName string, err error) {
	switch {
	case key == "correct_code":
		return []byte(generateValidCode() + "\n"), "", nil
	case key == "beagle_code":
		return []byte(beagleCode + "\n"), "", nil
	case key == "trap_1" || key == "trap_2" || key == "trap_3":
		data, err := os.ReadFile(filepath.Join(trapsDir, key+".txt"))
		return data, "", err
	case strings.HasPrefix(key, "trash:"):
		parts := strings.Split(key, ":")
		if len(parts) != 3 {
			return nil, "", fmt.Errorf("неверный формат content_key: %s", key)
		}
		category := parts[1]
		return pickRandomTrashFile(category)
	default:
		return nil, "", fmt.Errorf("неизвестный content_key: %s", key)
	}
}

func pickRandomTrashFile(category string) (content []byte, filename string, err error) {
	dir := filepath.Join(trashPoolDir, category)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, "", err
	}
	var files []string
	for _, e := range entries {
		if !e.IsDir() {
			files = append(files, e.Name())
		}
	}
	if len(files) == 0 {
		return nil, "", fmt.Errorf("папка %s пуста", dir)
	}
	chosen := files[rand.Intn(len(files))]
	data, err := os.ReadFile(filepath.Join(dir, chosen))
	return data, chosen, err
}

func generateValidCode() string {
	for {
		val := uint32(rand.Int63n(0xFFFFFFFF))
		if val%7 == 0 {
			hex := fmt.Sprintf("%08X", val)
			return hex[:4] + "-" + hex[4:]
		}
	}
}

func writeScript(path, content string) error {
	return os.WriteFile(path, []byte(content), 0755)
}

func buildScript() string {
	return `#!/usr/bin/env bash
set -e
cd ../student-environment/.. 2>/dev/null || true
docker build -f student-environment/Dockerfile -t ` + dockerImage + ` .
`
}

func generateStartBlock(s Student, containerName, preparedPath string) string {
	var b strings.Builder

	fmt.Fprintf(&b, "\necho \"Запускаю %s (%s)...\"\n", s.Login, s.Domain)
	b.WriteString("docker run -d \\\n")
	fmt.Fprintf(&b, "  --name %s \\\n", containerName)
	fmt.Fprintf(&b, "  --network %s \\\n", traefikNetwork)
	fmt.Fprintf(&b, "  -p %s:22 \\\n", s.Port)
	fmt.Fprintf(&b, "  -e STUDENT_LOGIN=%s \\\n", s.Login)
	fmt.Fprintf(&b, "  -e STUDENT_PASSWORD=%s \\\n", s.Password)
	b.WriteString("  --cap-add=LINUX_IMMUTABLE \\\n")
	b.WriteString("  --label \"traefik.enable=true\" \\\n")
	fmt.Fprintf(&b, "  --label \"traefik.http.routers.%s.rule=Host(\\`%s\\`)\" \\\n", s.Login, s.Domain)
	fmt.Fprintf(&b, "  --label \"traefik.http.routers.%s.entrypoints=%s\" \\\n", s.Login, domainEntrypoint)
	fmt.Fprintf(&b, "  --label \"traefik.http.routers.%s.tls.certresolver=%s\" \\\n", s.Login, certResolver)
	fmt.Fprintf(&b, "  --label \"traefik.http.services.%s.loadbalancer.server.port=80\" \\\n", s.Login)
	fmt.Fprintf(&b, "  %s\n\n", dockerImage)

	fmt.Fprintf(&b, "echo \"Жду готовности пользователя %s внутри контейнера...\"\n", s.Login)
	b.WriteString("READY=0\n")
	fmt.Fprintf(&b, "for i in $(seq 1 %d); do\n", pollTimeoutS)
	fmt.Fprintf(&b, "  if docker exec %s test -d /home/%s 2>/dev/null; then\n", containerName, s.Login)
	b.WriteString("    READY=1\n")
	b.WriteString("    break\n")
	b.WriteString("  fi\n")
	fmt.Fprintf(&b, "  sleep %d\n", pollIntervalS)
	b.WriteString("done\n\n")

	b.WriteString("if [ \"$READY\" -ne 1 ]; then\n")
	fmt.Fprintf(&b, "  echo \"ОШИБКА: пользователь %s не появился за %d секунд, пропускаю docker cp!\" >&2\n", s.Login, pollTimeoutS)
	b.WriteString("else\n")
	fmt.Fprintf(&b, "  docker cp %s/. %s:/home/%s/\n", preparedPath, containerName, s.Login)
	fmt.Fprintf(&b, "  docker exec %s chown -R %s:%s /home/%s\n", containerName, s.Login, s.Login, s.Login)
	fmt.Fprintf(&b, "  docker cp %s/demo.conf %s:/etc/nginx/sites-available/demo.conf\n", lesson2Dir, containerName)
	fmt.Fprintf(&b, "  echo \"Дерево для %s скопировано.\"\n", s.Login)
	b.WriteString("fi\n")

	return b.String()
}

func startScript(blocks []string) string {
	header := "#!/bin/bash\nset -e\n"
	return header + strings.Join(blocks, "\n")
}

func addLesson2Content(preparedPath string) error {
	notePath := filepath.Join(lesson2Dir, "note_2.txt")
	if err := copyFile(notePath, filepath.Join(preparedPath, "note_2.txt")); err != nil {
		return fmt.Errorf("note_2.txt: %w", err)
	}

	seoMessSrc := filepath.Join(lesson2Dir, "seo_mess")
	seoMessDst := filepath.Join(preparedPath, "seo_mess")
	if err := copyDirRecursive(seoMessSrc, seoMessDst); err != nil {
		return fmt.Errorf("seo_mess: %w", err)
	}

	return nil
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

func copyDirRecursive(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDirRecursive(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func simpleScript(lines []string) string {
	header := "#!/usr/bin/env bash\n"
	return header + strings.Join(lines, "\n") + "\n"
}
