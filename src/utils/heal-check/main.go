package main

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	expectedCode = "SH-0047"

	hintStateFile = ".heal-hints"
)

var hints = []string{
	"Сначала останови. ps aux - найди процесс, который крутится в фоне без остановки (родитель - bash, внутри длинный sleep). kill <pid> его остановит - без sudo, это твой собственный процесс.",
	"Убитый процесс не чинит систему, только останавливает деградацию. Дальше - читай сам скрипт: он лежал в фоне через nohup, найди его (сохранился путь в выводе ps aux до того, как ты убил процесс, или поищи через find ~ -name 'self_heal.sh').",
	"В скрипте почти везде 2>/dev/null. Это и прячет от тебя, что на самом деле идёт не так. Скопируй скрипт, убери 2>/dev/null везде и прогони один раз руками - что реально выводится в терминал?",
	"chattr -R +i применяется рекурсивно на всю папку сайта - /var/www/demo. Проверь: lsattr -R /var/www/demo. Заметь, что это накрывает не только сам сайт, но и папку backups внутри него, и лог тоже.",
	"Раз лог замолкает после первой строки - это не значит, что дальше всё шло гладко. Это значит, что дальше echo уже не мог писать в файл. Одна строка в логе - это не 'всё стабильно', а 'сломалось сразу после старта'.",
	"Посмотри на переменную ARCHIVE в скрипте - куда именно складываются архивы. Это внутри той же папки, которую tar же и архивирует. Что случится, если архивировать папку, где уже лежат архивы?",
	"Собери первый фрагмент кода: сколько секунд стоит в sleep в конце цикла self_heal.sh?",
	"Второй фрагмент - Володя пронумеровал этот инцидент в комментарии прямо в скрипте. Какой у него номер?",
	"Третий фрагмент - у Володи есть словесная привычка, которую он явно не замечает за собой. Пересчитай, сколько раз фраза \"на коленке\" встречается в его записке и в комментариях self_heal.sh вместе.",
	"Три фрагмента: интервал sleep, номер инцидента, количество повторов фразы. Сложи их. Код: SH-XXXX (сумма, дополненная нулями слева до 4 цифр).",
	"Когда во всём разберёшься - напиши свою версию скрипта на основе backup.sh и cleanup.sh: архив вне сайта (например, в ~/backups, как и было изначально), chattr точечно на саму папку backups, а не на весь сайт, ошибки не прячутся в /dev/null, а nginx проверяется по-настоящему перед reload.",
}

func main() {
	usr, err := user.Current()
	if err != nil {
		fmt.Println("Ошибка системы: не могу определить пользователя.")
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		fmt.Println("Использование: heal-check <код>")
		fmt.Println("Или: heal-check hint  - получить подсказку")
		os.Exit(1)
	}

	if os.Args[1] == "hint" {
		giveHint(usr.HomeDir)
		return
	}

	code := strings.ToUpper(strings.TrimSpace(os.Args[1]))

	markerFile := filepath.Join(usr.HomeDir, ".heal-complete")
	if _, err := os.Stat(markerFile); err == nil {
		fmt.Println("Уже сходится, ты уже это проверял.")
		fmt.Println("\n- Володя (наверное, всё ещё в самолёте)")
		return
	}

	if code != expectedCode {
		fmt.Println("Не сходится.")
		fmt.Println("\nЛибо один из трёх фрагментов неверен, либо ты собрал их не в том порядке.")
		fmt.Println("Формат ответа: SH-XXXX, где XXXX - сумма трёх найденных чисел.")
		fmt.Println("\nЗастрял - heal-check hint подскажет по одному шагу за раз.")
		os.Exit(1)
	}

	printWithDots("Проверяю по журналу")
	fmt.Println()
	fmt.Println("Совпадает.")
	fmt.Println()
	fmt.Println("Диск больше не тает, сайт стоит ровно, а лог наконец")
	fmt.Println("снова пишется - потому что ты писал его версию, а не Володину.")
	fmt.Println()
	fmt.Println("Он был прав в одном: ошибаются все, даже те, кто учит.")
	fmt.Println("Разница между ним и тобой сегодня - в том, что после тебя")
	fmt.Println("не нужно разбираться, что именно сломалось и почему.")
	fmt.Println()
	fmt.Println("- Володя (когда вернётся, обязательно спроси его,")
	fmt.Println("  почему он вообще решил, что sudo нужен буквально везде)")

	if err := os.WriteFile(markerFile, []byte("passed\n"), 0644); err != nil {
		fmt.Printf("\n[ОШИБКА] Не удалось сохранить маркер: %v\n", err)
	}
}

func giveHint(homeDir string) {
	stateFile := filepath.Join(homeDir, hintStateFile)

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

func printWithDots(msg string) {
	fmt.Print(msg)
	for i := 0; i < 3; i++ {
		time.Sleep(500 * time.Millisecond)
		fmt.Print(".")
	}
	fmt.Println()
}
