package main

import (
	"encoding/csv"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	studentsCSVPath = "../container-generator/students.csv"
	templatesDir    = "templates"

	demoConfEnabledPath  = "/etc/nginx/sites-enabled/demo.conf"
	hiddenSecretDir      = "/tmp/.hidden-secret"
	hiddenSecretConfPath = hiddenSecretDir + "/nginx.conf"

	attackerIP     = "77.88.44.242"
	normalLogLines = 19800
	attackRepeats  = 30

	tmpLogDir = "tmp_logs"
)

var normalIPs = []string{
	"95.165.12.44", "188.243.56.71", "46.0.234.12",
	"213.87.145.33", "178.65.99.201", "109.252.44.18",
	"5.138.201.77", "31.173.82.119", "176.59.44.200",
}

var normalPaths = []string{
	"/", "/index.html", "/about", "/contact",
	"/static/style.css", "/static/logo.png",
	"/favicon.ico", "/robots.txt",
}

var attackPaths = []string{
	"/etc/nginx/sites-enabled/demo.conf",
	"/etc/nginx/nginx.conf",
	"/etc/nginx/sites-available/",
	"/tmp/.hidden-secret/nginx.conf",
	"/admin", "/.env", "/wp-login.php", "/config.php",
}

var normalUserAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15",
	"Mozilla/5.0 (X11; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/115.0",
}

const attackerUserAgent = "python-requests/2.28.0"

var normalMethods = []string{"GET", "GET", "GET", "GET", "GET", "GET", "GET", "GET", "POST", "POST"}
var normalStatuses = []int{200, 200, 200, 200, 301, 304, 404, 500}
var attackStatuses = []int{200, 403, 404}

func main() {
	logins, err := loadLogins(studentsCSVPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка чтения %s: %v\n", studentsCSVPath, err)
		os.Exit(1)
	}

	if err := os.MkdirAll(tmpLogDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка создания %s: %v\n", tmpLogDir, err)
		os.Exit(1)
	}

	for _, login := range logins {
		containerName := "sigma-" + login
		processStudent(containerName, login)
	}
}

func processStudent(containerName, login string) {
	if !isContainerRunning(containerName) {
		fmt.Printf("[%s] контейнер не запущен, пропускаю\n", login)
		return
	}

	if fileExistsInContainer(containerName, hiddenSecretConfPath) {
		fmt.Printf("[%s] кража уже была произведена ранее, пропускаю\n", login)
		return
	}

	if !fileExistsInContainer(containerName, demoConfEnabledPath) {
		fmt.Printf("[%s] ВНИМАНИЕ: %s не найден - похоже, занятие 2 не завершено. Пропускаю, разберитесь вручную.\n", login, demoConfEnabledPath)
		return
	}

	if err := runDockerExec(containerName, "mkdir", "-p", hiddenSecretDir); err != nil {
		fmt.Printf("[%s] ОШИБКА при создании %s: %v\n", login, hiddenSecretDir, err)
		return
	}
	if err := runDockerExec(containerName, "mv", demoConfEnabledPath, hiddenSecretConfPath); err != nil {
		fmt.Printf("[%s] ОШИБКА при краже конфига: %v\n", login, err)
		return
	}

	logPath := filepath.Join(tmpLogDir, login+"_access.log")
	if err := generateAccessLog(logPath); err != nil {
		fmt.Printf("[%s] ОШИБКА при генерации access.log: %v\n", login, err)
		return
	}

	sysadminDir := fmt.Sprintf("/home/%s/sysadmin", login)
	if err := runDockerExec(containerName, "mkdir", "-p", sysadminDir); err != nil {
		fmt.Printf("[%s] ОШИБКА при создании %s: %v\n", login, sysadminDir, err)
		return
	}

	if err := dockerCp(logPath, containerName, sysadminDir+"/access.log"); err != nil {
		fmt.Printf("[%s] ОШИБКА при копировании access.log: %v\n", login, err)
		return
	}

	notePath := filepath.Join(templatesDir, "note.txt")
	scammerPath := filepath.Join(templatesDir, "scammer.txt")

	if err := dockerCp(notePath, containerName, sysadminDir+"/note.txt"); err != nil {
		fmt.Printf("[%s] ОШИБКА при копировании note.txt: %v\n", login, err)
		return
	}
	if err := dockerCp(scammerPath, containerName, sysadminDir+"/scammer.txt"); err != nil {
		fmt.Printf("[%s] ОШИБКА при копировании scammer.txt: %v\n", login, err)
		return
	}

	if err := runDockerExec(containerName, "chown", "-R", login+":"+login, sysadminDir); err != nil {
		fmt.Printf("[%s] ПРЕДУПРЕЖДЕНИЕ: не удалось поправить права на %s: %v\n", login, sysadminDir, err)
	}
	if err := runDockerExec(containerName, "chown", "-R", login+":"+login, hiddenSecretDir); err != nil {
		fmt.Printf("[%s] ПРЕДУПРЕЖДЕНИЕ: не удалось поправить права на %s: %v\n", login, hiddenSecretDir, err)
	}

	fmt.Printf("[%s] готово: конфиг украден, access.log/note.txt/scammer.txt на месте\n", login)
}

func loadLogins(path string) ([]string, error) {
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

	var logins []string
	for _, rec := range records[1:] {
		if len(rec) < 1 {
			continue
		}
		logins = append(logins, rec[0])
	}
	return logins, nil
}

func isContainerRunning(containerName string) bool {
	cmd := exec.Command("docker", "inspect", "-f", "{{.State.Running}}", containerName)
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	return string(out) == "true\n"
}

func fileExistsInContainer(containerName, path string) bool {
	cmd := exec.Command("docker", "exec", containerName, "test", "-e", path)
	err := cmd.Run()
	return err == nil
}

func runDockerExec(containerName string, args ...string) error {
	fullArgs := append([]string{"exec", containerName}, args...)
	cmd := exec.Command("docker", fullArgs...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func dockerCp(hostPath, containerName, containerPath string) error {
	cmd := exec.Command("docker", "cp", hostPath, containerName+":"+containerPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

type logEntry struct {
	timestamp time.Time
	line      string
}

func generateAccessLog(outPath string) error {
	now := time.Now()
	var entries []logEntry

	normalStart := now.Add(-24 * time.Hour)
	for i := range normalLogLines {
		t := normalStart.Add(time.Duration(i*4+rand.Intn(4)) * time.Second)
		ip := normalIPs[rand.Intn(len(normalIPs))]
		method := normalMethods[rand.Intn(len(normalMethods))]
		path := normalPaths[rand.Intn(len(normalPaths))]
		status := normalStatuses[rand.Intn(len(normalStatuses))]
		size := 200 + rand.Intn(7800)
		ua := normalUserAgents[rand.Intn(len(normalUserAgents))]

		line := formatLogLine(ip, t, method, path, status, size, ua)
		entries = append(entries, logEntry{timestamp: t, line: line})
	}

	attackStart := now.Add(-2 * time.Hour)
	i := 0
	for range attackRepeats {
		for _, path := range attackPaths {
			t := attackStart.Add(time.Duration(i*22+rand.Intn(11)) * time.Second)
			status := attackStatuses[rand.Intn(len(attackStatuses))]
			size := 100 + rand.Intn(1900)

			line := formatLogLine(attackerIP, t, "GET", path, status, size, attackerUserAgent)
			entries = append(entries, logEntry{timestamp: t, line: line})
			i++
		}
	}

	sort.Slice(entries, func(a, b int) bool {
		return entries[a].timestamp.Before(entries[b].timestamp)
	})

	var sb strings.Builder
	for _, e := range entries {
		sb.WriteString(e.line)
		sb.WriteString("\n")
	}

	return os.WriteFile(outPath, []byte(sb.String()), 0644)
}

func formatLogLine(ip string, t time.Time, method, path string, status, size int, ua string) string {
	ts := t.Format("02/Jan/2006:15:04:05 -0700")
	return fmt.Sprintf(`%s - - [%s] "%s %s HTTP/1.1" %d %d "-" "%s"`, ip, ts, method, path, status, size, ua)
}
