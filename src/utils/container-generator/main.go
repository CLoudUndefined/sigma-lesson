package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
)

const (
	adjectivesPath = "./wordlists/adjectives.txt"
	nounsPath      = "./wordlists/nouns.txt"
	outputPath     = "students.csv"
	domainSuffix   = ".student.teiwi.art"

	portMin = 32000
	portMax = 49151

	passwordWordsMin = 3
	passwordWordsMax = 4
)

func main() {
	countStr := os.Getenv("STUDENT_COUNT")
	if countStr == "" {
		fmt.Fprintln(os.Stderr, "Ошибка: переменная окружения STUDENT_COUNT не задана.")
		fmt.Fprintln(os.Stderr, "Пример запуска: STUDENT_COUNT=22 ./container-generator")
		os.Exit(1)
	}

	count, err := strconv.Atoi(countStr)
	if err != nil || count <= 0 {
		fmt.Fprintf(os.Stderr, "Ошибка: STUDENT_COUNT должен быть положительным целым числом, получено: %q\n", countStr)
		os.Exit(1)
	}

	adjectives, err := loadWordlist(adjectivesPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка чтения словаря прилагательных (%s): %v\n", adjectivesPath, err)
		os.Exit(1)
	}
	nouns, err := loadWordlist(nounsPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка чтения словаря существительных (%s): %v\n", nounsPath, err)
		os.Exit(1)
	}

	allWords := make([]string, 0, len(adjectives)+len(nouns))
	allWords = append(allWords, adjectives...)
	allWords = append(allWords, nouns...)

	maxLogins := len(adjectives) * len(nouns)
	if count > maxLogins {
		fmt.Fprintf(os.Stderr, "Ошибка: запрошено %d логинов, но словари дают максимум %d уникальных комбинаций (adjectives=%d x nouns=%d).\n", count, maxLogins, len(adjectives), len(nouns))
		fmt.Fprintln(os.Stderr, "Пополни словари или уменьши STUDENT_COUNT.")
		os.Exit(1)
	}

	usedLogins := make(map[string]bool)
	usedPorts := make(map[int]bool)

	type studentRow struct {
		login    string
		password string
		port     int
		domain   string
	}

	rows := make([]studentRow, 0, count)

	for i := 0; i < count; i++ {
		login := generateUniqueLogin(adjectives, nouns, usedLogins)
		usedLogins[login] = true

		password := generatePassword(allWords)
		port := generateUniquePort(usedPorts)
		usedPorts[port] = true

		domain := login + domainSuffix

		rows = append(rows, studentRow{
			login:    login,
			password: password,
			port:     port,
			domain:   domain,
		})
	}

	f, err := os.Create(outputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка создания файла %s: %v\n", outputPath, err)
		os.Exit(1)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	if err := w.Write([]string{"login", "password", "port", "domain"}); err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка записи заголовка CSV: %v\n", err)
		os.Exit(1)
	}

	for _, row := range rows {
		record := []string{row.login, row.password, strconv.Itoa(row.port), row.domain}
		if err := w.Write(record); err != nil {
			fmt.Fprintf(os.Stderr, "Ошибка записи строки CSV: %v\n", err)
			os.Exit(1)
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка при финальной записи CSV: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Готово. Сгенерировано %d записей в %s\n", count, outputPath)
}

func loadWordlist(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var words []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		words = append(words, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(words) == 0 {
		return nil, fmt.Errorf("файл пуст или не содержит валидных строк")
	}
	return words, nil
}

func generateUniqueLogin(adjectives, nouns []string, usedLogins map[string]bool) string {
	for {
		adj := adjectives[rand.Intn(len(adjectives))]
		noun := nouns[rand.Intn(len(nouns))]
		login := adj + "-" + noun
		if !usedLogins[login] {
			return login
		}
	}
}

func generatePassword(allWords []string) string {
	wordCount := passwordWordsMin + rand.Intn(passwordWordsMax-passwordWordsMin+1)

	parts := make([]string, wordCount)
	for i := 0; i < wordCount; i++ {
		parts[i] = allWords[rand.Intn(len(allWords))]
	}

	digitCount := 1 + rand.Intn(2)
	digits := ""
	for i := 0; i < digitCount; i++ {
		digits += strconv.Itoa(rand.Intn(10))
	}

	return strings.Join(parts, "-") + digits
}

func generateUniquePort(usedPorts map[int]bool) int {
	for {
		port := portMin + rand.Intn(portMax-portMin+1)
		if !usedPorts[port] {
			return port
		}
	}
}
