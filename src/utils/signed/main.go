package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	Trap1Code = "DEAD-TIMA"
	Trap2Code = "VOLO-DYA0"
	Trap3Code = "FAKE-C0DE"
	EasterEgg = "SHER-L0CK"
)

func main() {
	usr, err := user.Current()
	if err != nil {
		fmt.Println("Ошибка системы: не могу определить пользователя.")
		os.Exit(1)
	}
	markerFile := filepath.Join(usr.HomeDir, ".signed-complete")

	if _, err := os.Stat(markerFile); err == nil {
		fmt.Println("Эй, я смотрю кто-то решил устроиться на работу дважды, чтобы получать две зарплаты?")
		fmt.Println("\nСпешу огорчить: ты уже в штате, твоя душа уже принадлежит нам,")
		fmt.Println("а бюджет компании не резиновый.")
		fmt.Println("\nХватит играться с бумажками, иди чини сервер!")
		fmt.Println("\n- Володя")
		os.Exit(0)
	}

	if len(os.Args) < 2 {
		fmt.Println("Использование: signed <код>")
		fmt.Println("Пример: signed FADC-3149")
		os.Exit(1)
	}

	rawCode := os.Args[1]

	if len(rawCode) != 9 || rawCode[4] != '-' {
		fmt.Println("Неверный формат кода. Ожидается формат XXXX-XXXX (например, 1A2B-3C4D).")
		os.Exit(1)
	}

	codeClean := strings.ToUpper(strings.ReplaceAll(rawCode, "-", ""))
	codeFormatted := strings.ToUpper(rawCode)

	login := usr.Username
	randNum := rand.Intn(9000) + 1000
	studentID := fmt.Sprintf("%s%d", login, randNum)

	switch codeFormatted {
	case Trap1Code:
		fmt.Println("Этот код принадлежал прошлому джуну по имени Тима.")
		fmt.Println("Он снес нам базу данных и был уволен.")
		fmt.Println("Доступ по этому коду закрыт.\n\nПроверь записки внимательнее.\n\n- Володя")
		return
	case Trap2Code:
		fmt.Println("Эй, это мой личный код активации! Не трожь!\n\n- Володя")
		return
	case Trap3Code:
		fmt.Println("Это тестовый код.\n\nЯ оставил его специально, чтобы убедиться,")
		fmt.Println("что ты действительно читаешь записки, а не вводишь всё подряд.")
		fmt.Println("\nПоищи ещё.\n\n- Володя")
		return
	case EasterEgg:
		printWithDots("Проверяю скрытый договор")
		printWithDots("Секретная подпись подтверждена")
		fmt.Println("\nПоздравляю!\n")
		fmt.Println("Ты нашел то, что не предназначалось для обычных подаванов.")
		fmt.Println("Вы официально приняты на должность")
		fmt.Println("\"Юная ищейка\" (Beagle-junior)\n")
		fmt.Println("Твой секретный ID в системе:", studentID, "\n")
		fmt.Println("Я сохранил его у тебя в папке в beagle_id.txt.\n\n- Володя")

		idFile := filepath.Join(usr.HomeDir, "beagle_id.txt")
		if err := os.WriteFile(idFile, []byte("Role: Beagle-junior\nID: "+studentID+"\n"), 0644); err != nil {
			fmt.Printf("\n[ОШИБКА] Володя не смог выдать тебе ID: %v\nЗови преподавателя!\n", err)
			os.Exit(1)
		}
		if err := os.WriteFile(markerFile, []byte("beagle"), 0644); err != nil {
			fmt.Printf("\n[ОШИБКА] Володя не смог создать системный маркер: %v\nЗови преподавателя!\n", err)
			os.Exit(1)
		}
		return
	}

	val, err := strconv.ParseUint(codeClean, 16, 64)
	if err != nil {
		fmt.Println("Код недействителен (не читается подпись). Ты уверен, что нашел правильный код?")
		os.Exit(1)
	}

	if val%7 != 0 {
		fmt.Println("Код недействителен. Похоже, этот договор поддельный, поищи настоящий.")
		os.Exit(1)
	}

	printWithDots("Проверяю договор")
	printWithDots("Проверяю подпись")

	fmt.Println("\nПоздравляю!\n")
	fmt.Println("Вы официально приняты на должность")
	fmt.Println("\"Юный системный администратор\"\n")
	fmt.Println("Теперь можете приступать к работе.\nДобро пожаловать в команду, подаван.\n")
	fmt.Println("Твой ID в системе:", studentID, "\n")
	fmt.Println("Я сохранил его у тебя в папке в my_id.txt. Не потеряй.\n\n- Володя")

	idFile := filepath.Join(usr.HomeDir, "my_id.txt")
	if err := os.WriteFile(idFile, []byte("Role: Pre-junior\nID: "+studentID+"\n"), 0644); err != nil {
		fmt.Printf("\n[ОШИБКА] Володя не смог выдать тебе ID: %v\nЗови преподавателя!\n", err)
		os.Exit(1)
	}
	if err := os.WriteFile(markerFile, []byte("junior"), 0644); err != nil {
		fmt.Printf("\n[ОШИБКА] Володя не смог создать системный маркер: %v\nЗови преподавателя!\n", err)
		os.Exit(1)
	}
}

func printWithDots(msg string) {
	fmt.Print(msg)
	for i := 0; i < 3; i++ {
		time.Sleep(600 * time.Millisecond)
		fmt.Print(".")
	}
	fmt.Println()
}
