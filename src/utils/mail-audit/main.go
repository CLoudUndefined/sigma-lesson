// mail-audit - проверка для побочного квеста "Тень в расписании" (занятие 5).
// Использование:
//
//	mail-audit test-cron         - реально проверить, отработает ли текущая
//	                                строка disk-watch в расписании именно так,
//	                                как её видит настоящий cron (минимальный
//	                                PATH, без .bashrc), а не в интерактивном shell
//	mail-audit check <код>       - проверить собранный код (три фрагмента)
//	mail-audit hint              - получить очередную подсказку
package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const expectedCode = "MV-0298"

var hints = []string{
	"Загляни в своё расписание: crontab -l. Есть там незнакомая строка про disk-watch?",
	"Попробуй запустить disk-watch --threshold 80 прямо в терминале руками. Работает? А теперь спроси себя: точно ли cron видит твоё окружение так же, как твой интерактивный shell?",
	"Проверь разницу напрямую: добавь на минуту в свой crontab -e строку * * * * * env > ~/cron_env_test.log, подожди минуту, и сравни её с выводом env > ~/shell_env_test.log, который ты снял руками. Ищи различия в PATH.",
	"В Unix-системах у каждого пользователя есть личный почтовый ящик, даже если никто им не пользуется. Загляни в /var/mail/, есть там что-то с твоим именем?",
	"Файл в /var/mail/<login> - обычный текст, читается через cat/less/grep. Десятки писем похожи друг на друга - но одно из них (самое раннее) отличается по теме и содержанию. Найди его.",
	"В том необычном письме Володя сам себе что-то пишет - там есть число. Это первый фрагмент кода.",
	"Второй и третий фрагмент - в самой строке crontab с disk-watch: значение после --threshold, и число в интервале */N. Сложи все три числа.",
	"Код собирается так: MV-<сумма трёх чисел, дополненная нулями слева до 4 цифр>. И не забудь на самом деле починить строку в crontab - mail-audit test-cron проверит, годится ли она для настоящего cron, а не только для твоего интерактивного shell.",
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "hint":
		giveHint()
	case "check":
		if len(os.Args) < 3 {
			fmt.Println("Использование: mail-audit check <код>")
			os.Exit(1)
		}
		runCheck(os.Args[2])
	case "test-cron":
		runTestCron()
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Использование:")
	fmt.Println("  mail-audit test-cron         - проверить строку disk-watch в реальном cron-окружении")
	fmt.Println("  mail-audit check <код>       - проверить собранный код")
	fmt.Println("  mail-audit hint              - получить подсказку")
}

func giveHint() {
	usr, err := user.Current()
	if err != nil {
		fmt.Println("Ошибка системы: не могу определить пользователя.")
		os.Exit(1)
	}
	stateFile := filepath.Join(usr.HomeDir, ".mail-hints")

	idx := 0
	if data, err := os.ReadFile(stateFile); err == nil {
		if n, err := strconv.Atoi(strings.TrimSpace(string(data))); err == nil {
			idx = n
		}
	}
	if idx >= len(hints) {
		fmt.Println("Подсказки закончились. Дальше - сам.")
		return
	}
	fmt.Printf("Подсказка %d/%d:\n\n%s\n", idx+1, len(hints), hints[idx])
	_ = os.WriteFile(stateFile, []byte(strconv.Itoa(idx+1)), 0644)
}

func runCheck(code string) {
	usr, err := user.Current()
	if err != nil {
		fmt.Println("Ошибка системы: не могу определить пользователя.")
		os.Exit(1)
	}

	markerFile := filepath.Join(usr.HomeDir, ".mail-complete")
	if _, err := os.Stat(markerFile); err == nil {
		fmt.Println("Ты уже разобрался с этим. Возвращаться незачем - разве что почта опять начнёт копиться.")
		return
	}

	code = strings.ToUpper(strings.TrimSpace(code))
	if code != expectedCode {
		fmt.Println("Не сходится.")
		fmt.Println("\nПроверь три фрагмента: число из письма-канарейки, значение --threshold,")
		fmt.Println("и интервал */N из crontab. Код - их сумма, формат MV-XXXX.")
		fmt.Println("\nЗастрял - mail-audit hint подскажет по шагу за раз.")
		os.Exit(1)
	}

	printWithDots("Сверяю")
	fmt.Println()
	fmt.Println("Совпадает.")
	fmt.Println()
	fmt.Println("Тихая переписка Володи с самим собой длиной в пять недель -")
	fmt.Println("и всё это время он даже не подозревал, что она существует.")
	fmt.Println()
	fmt.Println("Не забудь ещё и саму строку в crontab на самом деле починить -")
	fmt.Println("mail-audit test-cron покажет, держит она удар или нет.")
	fmt.Println()
	fmt.Println("- Тень в расписании (ну, или просто забывчивый Володя)")

	_ = os.WriteFile(markerFile, []byte("found\n"), 0644)
}

var cronVarLine = regexp.MustCompile(^([A-Za-z_][A-Za-z0-9_]*)=(.*)$)

var cronScheduleLine = regexp.MustCompile(^\S+\s+\S+\s+\S+\s+\S+\s+\S+\s+(.+)$)

func runTestCron() {
	usr, err := user.Current()
	if err != nil {
		fmt.Println("Ошибка системы: не могу определить пользователя.")
		os.Exit(1)
	}

	out, err := exec.Command("crontab", "-l").Output()
	if err != nil {
		fmt.Println("Не смог прочитать crontab. Он у тебя вообще настроен? (crontab -l)")
		os.Exit(1)
	}

	envVars := map[string]string{}
	var targetCmd string

	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if m := cronVarLine.FindStringSubmatch(line); m != nil && !looksLikeSchedule(line) {
			envVars[m[1]] = strings.Trim(m[2], "')
			continue
		}
		if strings.Contains(line, "disk-watch") {
			if m := cronScheduleLine.FindStringSubmatch(line); m != nil {
				targetCmd = m[1]
			}
		}
	}

	if targetCmd == "" {
		fmt.Println("Не нашёл в crontab строку с disk-watch. Она закомментирована, удалена")
		fmt.Println("или названа как-то иначе? Проверь crontab -l глазами.")
		os.Exit(1)
	}

	fmt.Printf("Нашёл команду: %s\n", targetCmd)
	fmt.Println("Прогоняю её в окружении, максимально похожем на настоящий cron...")
	fmt.Println("(минимальный PATH, без чтения .bashrc, /bin/sh вместо интерактивного bash)")
	fmt.Println()

	env := []string{
		"HOME=" + usr.HomeDir,
		"LOGNAME=" + usr.Username,
		"USER=" + usr.Username,
	}
	path := "/usr/bin:/bin"
	if p, ok := envVars["PATH"]; ok {
		path = p
	}
	env = append(env, "PATH="+path)
	for k, v := range envVars {
		if k == "PATH" {
			continue
		}
		env = append(env, k+"="+v)
	}

	cmd := exec.Command("/bin/sh", "-c", targetCmd)
	cmd.Env = env
	outBytes, runErr := cmd.CombinedOutput()
	outputStr := strings.TrimSpace(string(outBytes))

	fmt.Println("--- вывод команды ---")
	if outputStr == "" {
		fmt.Println("(пусто)")
	} else {
		fmt.Println(outputStr)
	}
	fmt.Println("---------------------")
	fmt.Println()

	failed := runErr != nil || strings.Contains(outputStr, "not found") || strings.Contains(outputStr, "No such file")

	if failed {
		fmt.Println("ПРОВАЛ: команда не отрабатывает в условиях, максимально приближенных к")
		fmt.Println("реальному cron. Скорее всего дело в PATH - интерактивный shell видит твою")
		fmt.Println("личную папку с бинарниками через .bashrc, а cron - нет.")
		fmt.Println()
		fmt.Println("Варианты честного исправления: абсолютный путь к disk-watch в самой")
		fmt.Println("строке crontab, либо PATH=... в шапке файла crontab, либо перенос")
		fmt.Println("disk-watch в стандартный каталог.")
		os.Exit(1)
	}

	fmt.Println("OK: команда отрабатывает корректно даже в условиях реального cron.")
}

func looksLikeSchedule(line string) bool {
	fields := strings.Fields(line)
	return len(fields) >= 6 && cronScheduleLine.MatchString(line)
}

func printWithDots(msg string) {
	fmt.Print(msg)
	for i := 0; i < 3; i++ {
		time.Sleep(400 * time.Millisecond)
		fmt.Print(".")
	}
	fmt.Println()
}
