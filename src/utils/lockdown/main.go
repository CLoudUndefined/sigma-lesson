package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const (
	studentsCSVPath = "../container-generator/students.csv"

	memLimit = "128m"
	cpuLimit = "0.5"
)

var binariesToBlock = []string{
	"/usr/bin/curl",
	"/usr/bin/wget",
	"/usr/bin/apt",
	"/usr/bin/apt-get",
	"/usr/bin/aptitude",
}

var packagesToRemove = []string{
	"curl",
	"wget",
}

func main() {
	logins, err := loadLogins(studentsCSVPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка чтения %s: %v\n", studentsCSVPath, err)
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

	if err := removePackages(containerName); err != nil {
		fmt.Printf("[%s] ПРЕДУПРЕЖДЕНИЕ: apt remove: %v\n", login, err)
	}

	for _, path := range binariesToBlock {
		if err := blockBinary(containerName, path); err != nil {
			fmt.Printf("[%s] ПРЕДУПРЕЖДЕНИЕ: блокировка %s: %v\n", login, path, err)
		}
	}

	if err := applyResourceLimits(containerName); err != nil {
		fmt.Printf("[%s] ОШИБКА при установке лимитов ресурсов: %v\n", login, err)
		return
	}

	fmt.Printf("[%s] готово: пакеты удалены, бинарники заблокированы, лимиты применены\n", login)
}

func removePackages(containerName string) error {
	args := append([]string{"remove", "-y", "--purge"}, packagesToRemove...)
	return runDockerExec(containerName, append([]string{"apt-get"}, args...)...)
}

func blockBinary(containerName, path string) error {
	if !fileExistsInContainer(containerName, path) {
		return nil
	}
	return runDockerExec(containerName, "chmod", "000", path)
}

func applyResourceLimits(containerName string) error {
	cmd := exec.Command("docker", "update",
		"--memory", memLimit,
		"--memory-swap", memLimit,
		"--cpus", cpuLimit,
		containerName,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
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
