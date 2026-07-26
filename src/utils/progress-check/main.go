package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"os/exec"
)

const studentsCSVPath = "../container-generator/students.csv"

type lessonCheck struct {
	label     string
	checkPath func(login string) string
}

var lessonChecks = []lessonCheck{
	{
		label:     "занятие 1 (signed)",
		checkPath: func(login string) string { return fmt.Sprintf("/home/%s/.signed-complete", login) },
	},
	{
		label:     "занятие 2 (nginx demo.conf)",
		checkPath: func(login string) string { return "/etc/nginx/sites-enabled/demo.conf" },
	},
	{
		label:     "занятие 3 (report.txt)",
		checkPath: func(login string) string { return fmt.Sprintf("/home/%s/sysadmin/report.txt", login) },
	},
}

func main() {
	logins, err := loadLogins(studentsCSVPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка чтения %s: %v\n", studentsCSVPath, err)
		os.Exit(1)
	}

	for _, login := range logins {
		containerName := "sigma-" + login

		if !isContainerRunning(containerName) {
			fmt.Printf("%s: off\n", login)
			continue
		}

		bits := buildProgressBits(containerName, login)
		fmt.Printf("%s: %s\n", login, bits)
	}
}

func isContainerRunning(containerName string) bool {
	cmd := exec.Command("docker", "inspect", "-f", "{{.State.Running}}", containerName)
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	return string(out) == "true\n"
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

func buildProgressBits(containerName, login string) string {
	bits := make([]byte, 0, 4)

	for _, check := range lessonChecks {
		if fileExistsInContainer(containerName, check.checkPath(login)) {
			bits = append(bits, '1')
		} else {
			bits = append(bits, '0')
		}
	}

	backupPath := fmt.Sprintf("/home/%s/backup.sh", login)
	cleanupPath := fmt.Sprintf("/home/%s/cleanup.sh", login)
	if fileExistsInContainer(containerName, backupPath) && fileExistsInContainer(containerName, cleanupPath) {
		bits = append(bits, '1')
	} else {
		bits = append(bits, '0')
	}

	return string(bits)
}

func fileExistsInContainer(containerName, path string) bool {
	cmd := exec.Command("docker", "exec", containerName, "test", "-f", path)
	err := cmd.Run()
	return err == nil
}
