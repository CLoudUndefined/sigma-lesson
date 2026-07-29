// ghost-check - финальная проверка для побочного квеста "Призрак R."
//
//	ghost-check <код>       - проверить собранный код
//	ghost-check hint        - получить очередную подсказку (по одной за раз)
//
// Код собирается из трёх фрагментов, разбросанных по уликам:
//  1. день месяца в поддельной метке времени /opt/.legacy-diag (акт 1, diag.log)
//  2. интервал respawn-проверки процесса-призрака в секундах (акт 3, наблюдение)
//  3. число, спрятанное в расшифрованном черновике R. (акт 2, resignation_draft.txt)
//
// Итоговый код - сумма трёх фрагментов, оформленная как RD-XXXX.
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
	expectedCode = "RD-0031"

	hintStateFile = ".ghost-hints"
)

var hints = []string{
	"Начни с того, что старше самой системы. ls -la --time-style=long-iso /opt - что-то там датировано подозрительно давно.",
	"У настоящих процессов ядра (kworker) нет исполняемого файла на диске. Найди подозрительный процесс через ps aux, и проверь ls -l /proc/<pid>/exe - если резолвится в реальный путь, это не ядро.",
	"Если убить найденный процесс (kill -9 <pid>) - подожди немного и проверь ps aux снова. Заметил, что он вернулся? Значит, он не один. Их двое, и они следят друг за другом.",
	"В /opt/.legacy-diag есть файл-дневник. Раз в несколько строк среди автоматических записей проскакивают человеческие комментарии - читай внимательно, там же спрятана метка времени, которая старше самого сервера. День месяца в этой метке - первый фрагмент кода.",
	"Один процесс респавнится не мгновенно. Замерь секундомером (или просто date), сколько проходит между kill -9 видимого процесса и его повторным появлением в ps aux. Это число секунд - второй фрагмент кода.",
	"Где-то рядом с дневником лежит папка, которую нельзя ни удалить, ни переместить обычным способом - rm скажет 'Operation not permitted' без объяснений. Это не баг и не права доступа. Погугли, какой ещё бывает 'immutable' флаг у файлов в Linux, и как его посмотреть (lsattr) и снять (у тебя есть sudo).",
	"Внутри защищённой папки - черновик, зашифрованный в два слоя: сначала ROT13, потом base64. Расшифруй: ... | base64 -d | tr 'A-Za-z' 'N-ZA-Mn-za-m'. В тексте R. трижды переписывал одно и то же - число повторов и есть третий фрагмент.",
	"Три фрагмента: день месяца из метки времени, интервал респавна в секундах, число повторов из черновика. Сложи их. Код: RD-<сумма, дополненная нулями слева до 4 цифр>.",
}

func main() {
	usr, err := user.Current()
	if err != nil {
		fmt.Println("Ошибка системы: не могу определить пользователя.")
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		fmt.Println("Использование: ghost-check <код>")
		fmt.Println("Или: ghost-check hint  - получить подсказку")
		os.Exit(1)
	}

	if os.Args[1] == "hint" {
		giveHint(usr.HomeDir)
		return
	}

	code := strings.ToUpper(strings.TrimSpace(os.Args[1]))

	markerFile := filepath.Join(usr.HomeDir, ".ghost-complete")
	if _, err := os.Stat(markerFile); err == nil {
		fmt.Println("Ты уже нашёл его. Призрак не любит повторений.")
		fmt.Println("\n- R. (наверное)")
		return
	}

	if code != expectedCode {
		fmt.Println("Не сходится.")
		fmt.Println("\nЛибо один из трёх фрагментов неверен, либо ты собрал их не в том порядке.")
		fmt.Println("Формат ответа: RD-XXXX, где XXXX - сумма трёх найденных чисел.")
		fmt.Println("\nЗастрял - ghost-check hint подскажет по одному шагу за раз.")
		os.Exit(1)
	}

	printWithDots("Сверяю с архивом")
	fmt.Println()
	fmt.Println("Совпадает.")
	fmt.Println()
	fmt.Println("Ты только что нашёл то, что не нашёл даже Володя.")
	fmt.Println("R. давно ушёл. Его самого здесь больше нет, его тихо заменили,")
	fmt.Println("даже не особо объяснив зачем. Но часть системы всё ещё работает")
	fmt.Println("ровно так, как её и оставили. Без присмотра целыми годами.")
	fmt.Println()
	fmt.Println("Может, это была месть. А может просто способ")
	fmt.Println("не быть до конца забытым.")
	fmt.Println()
	fmt.Println("Ты был внимательнее, чем большинство, что в наше время редкость.")
	fmt.Println()
	fmt.Println("- R.")

	if err := os.WriteFile(markerFile, []byte("found\n"), 0644); err != nil {
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
