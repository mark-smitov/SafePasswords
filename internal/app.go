package internal

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"

	"golang.org/x/term"
)

const banner = `
 ____         __      ____                                     _ 
/ ___|  __ _ / _| ___|  _ \ __ _ ___ _____      _____  _ __ __| |
\___ \ / _\` + "`" + ` | |_ / _ \ |_) / _\` + "`" + ` / __/ __\ \ /\ / / _ \| '__/ _\` + "`" + ` |
 ___) | (_| |  _|  __/  __/ (_| \__ \__ \ V  V / (_) | | | (_| |
|____/ \__,_|_|  \___|_|   \__,_|___/___/ \_/\_/ \___/|_|  \__,_|
`

func Run() {
	initLogger(true)

	if len(os.Args) < 2 {
		fmt.Print(banner)
		startTUI()
		return
	}

	fmt.Print(banner)

	cmd := os.Args[1]
	switch cmd {
	case "init":
		handleInit()
	case "add":
		handleAdd()
	case "get":
		handleGet()
	case "list":
		handleList()
	case "delete":
		handleDelete()
	case "edit":
		handleEdit()
	case "generate":
		handleGenerate()
	case "reset":
		handleReset()
	default:
		fmt.Printf("Неизвестная команда: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Использование: passman <команда>")
	fmt.Println()
	fmt.Println("Команды:")
	fmt.Println("  init       Инициализировать новое хранилище")
	fmt.Println("  add        Добавить запись")
	fmt.Println("  get        Найти пароль по названию")
	fmt.Println("  list       Список всех записей")
	fmt.Println("  delete     Удалить запись")
	fmt.Println("  edit       Изменить запись")
	fmt.Println("  generate   Создать надёжный пароль")
	fmt.Println("  reset      Сбросить мастер-пароль")
}

func readPassword(prompt string) string {
	fmt.Print(prompt)
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		logger.Error("failed to read password", "error", err)
		os.Exit(1)
	}
	return string(bytePassword)
}

func handleInit() {
	if err := ensurePasswordsDir(); err != nil {
		logger.Error("failed to create passwords dir", "error", err)
		os.Exit(1)
	}
	if vaultExists() {
		fmt.Println("Хранилище уже существует. Операция отменена.")
		os.Exit(1)
	}

	pass := readPassword("Придумайте мастер-пароль: ")
	confirm := readPassword("Подтвердите мастер-пароль: ")
	if pass != confirm {
		fmt.Println("Пароли не совпадают.")
		os.Exit(1)
	}
	if len(pass) < 8 {
		fmt.Println("Мастер-пароль должен быть не менее 8 символов.")
		os.Exit(1)
	}

	vault := &Vault{
		Version:  vaultVersion,
		Created:  time.Now(),
		Modified: time.Now(),
		Entries:  []Entry{},
	}
	if err := saveVault(vault, pass); err != nil {
		logger.Error("failed to create vault", "error", err)
		os.Exit(1)
	}
	fmt.Println("Хранилище успешно создано.")
	logger.Info("vault initialized")
}

func handleAdd() {
	if !vaultExists() {
		fmt.Println("Хранилище не найдено. Сначала выполните 'passman init'.")
		os.Exit(1)
	}
	pass := readPassword("Мастер-пароль: ")

	vault, err := loadVault(pass)
	if err != nil {
		logger.Error("failed to load vault", "error", err)
		os.Exit(1)
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Название: ")
	title, _ := reader.ReadString('\n')
	fmt.Print("Логин: ")
	username, _ := reader.ReadString('\n')
	fmt.Print("Пароль (оставьте пустым, чтобы сгенерировать): ")
	pwd, _ := reader.ReadString('\n')
	fmt.Print("URL: ")
	url, _ := reader.ReadString('\n')
	fmt.Print("Заметки: ")
	notes, _ := reader.ReadString('\n')
	fmt.Print("Теги (через запятую): ")
	tagsStr, _ := reader.ReadString('\n')

	entry := Entry{
		ID:        generateID(),
		Title:     strings.TrimSpace(title),
		Username:  strings.TrimSpace(username),
		Password:  strings.TrimSpace(pwd),
		URL:       strings.TrimSpace(url),
		Notes:     strings.TrimSpace(notes),
		Tags:      splitTags(tagsStr),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if entry.Password == "" {
		entry.Password = generatePassword(20, true, true, true, true)
		fmt.Printf("Создан пароль: %s\n", entry.Password)
	}

	vault.Entries = append(vault.Entries, entry)
	vault.Modified = time.Now()

	if err := saveVault(vault, pass); err != nil {
		logger.Error("failed to save vault", "error", err)
		os.Exit(1)
	}
	fmt.Println("Запись добавлена.")
	logger.Info("entry added", "title", entry.Title)
}

func handleGet() {
	if !vaultExists() {
		fmt.Println("Хранилище не найдено.")
		os.Exit(1)
	}
	pass := readPassword("Мастер-пароль: ")
	vault, err := loadVault(pass)
	if err != nil {
		logger.Error("failed to load vault", "error", err)
		os.Exit(1)
	}

	if len(os.Args) < 3 {
		fmt.Println("Использование: passman get <название>")
		os.Exit(1)
	}
	query := strings.ToLower(os.Args[2])

	for _, e := range vault.Entries {
		if strings.Contains(strings.ToLower(e.Title), query) {
			fmt.Println()
			fmt.Printf("ID:       %s\n", e.ID)
			fmt.Printf("Название: %s\n", e.Title)
			fmt.Printf("Логин:    %s\n", e.Username)
			fmt.Printf("Пароль:   %s\n", e.Password)
			fmt.Printf("URL:      %s\n", e.URL)
			fmt.Printf("Заметки:  %s\n", e.Notes)
			fmt.Printf("Теги:     %s\n", strings.Join(e.Tags, ", "))
			fmt.Printf("Обновлено: %s\n", e.UpdatedAt.Format(time.RFC3339))
			fmt.Println()
			return
		}
	}
	fmt.Println("Запись не найдена.")
}

func handleList() {
	if !vaultExists() {
		fmt.Println("Хранилище не найдено.")
		os.Exit(1)
	}
	pass := readPassword("Мастер-пароль: ")
	vault, err := loadVault(pass)
	if err != nil {
		logger.Error("failed to load vault", "error", err)
		os.Exit(1)
	}

	if len(vault.Entries) == 0 {
		fmt.Println("В хранилище нет записей.")
		return
	}
	fmt.Printf("%-4s %-20s %-20s %-30s\n", "#", "Название", "Логин", "URL")
	fmt.Println(strings.Repeat("-", 80))
	for i, e := range vault.Entries {
		fmt.Printf("%-4d %-20s %-20s %-30s\n", i+1, truncate(e.Title, 20), truncate(e.Username, 20), truncate(e.URL, 30))
	}
}

func handleDelete() {
	if !vaultExists() {
		fmt.Println("Хранилище не найдено.")
		os.Exit(1)
	}
	pass := readPassword("Мастер-пароль: ")
	vault, err := loadVault(pass)
	if err != nil {
		logger.Error("failed to load vault", "error", err)
		os.Exit(1)
	}

	if len(os.Args) < 3 {
		fmt.Println("Использование: passman delete <название>")
		os.Exit(1)
	}
	query := strings.ToLower(os.Args[2])

	found := false
	newEntries := vault.Entries[:0]
	for _, e := range vault.Entries {
		if strings.Contains(strings.ToLower(e.Title), query) {
			found = true
			logger.Info("deleting entry", "title", e.Title)
			continue
		}
		newEntries = append(newEntries, e)
	}
	if !found {
		fmt.Println("Запись не найдена.")
		os.Exit(1)
	}
	vault.Entries = newEntries
	vault.Modified = time.Now()

	if err := saveVault(vault, pass); err != nil {
		logger.Error("failed to save vault", "error", err)
		os.Exit(1)
	}
	fmt.Println("Запись удалена.")
}

func handleEdit() {
	if !vaultExists() {
		fmt.Println("Хранилище не найдено.")
		os.Exit(1)
	}
	pass := readPassword("Мастер-пароль: ")
	vault, err := loadVault(pass)
	if err != nil {
		logger.Error("failed to load vault", "error", err)
		os.Exit(1)
	}

	if len(os.Args) < 3 {
		fmt.Println("Использование: passman edit <название>")
		os.Exit(1)
	}
	query := strings.ToLower(os.Args[2])

	reader := bufio.NewReader(os.Stdin)
	for i := range vault.Entries {
		e := &vault.Entries[i]
		if strings.Contains(strings.ToLower(e.Title), query) {
			fmt.Printf("Редактирование «%s» (Enter — оставить без изменений)\n", e.Title)
			fmt.Printf("Логин [%s]: ", e.Username)
			if v, _ := reader.ReadString('\n'); strings.TrimSpace(v) != "" {
				e.Username = strings.TrimSpace(v)
			}
			fmt.Printf("Пароль [%s]: ", strings.Repeat("*", len(e.Password)))
			if v, _ := reader.ReadString('\n'); strings.TrimSpace(v) != "" {
				e.Password = strings.TrimSpace(v)
			}
			fmt.Printf("URL [%s]: ", e.URL)
			if v, _ := reader.ReadString('\n'); strings.TrimSpace(v) != "" {
				e.URL = strings.TrimSpace(v)
			}
			fmt.Printf("Заметки [%s]: ", e.Notes)
			if v, _ := reader.ReadString('\n'); strings.TrimSpace(v) != "" {
				e.Notes = strings.TrimSpace(v)
			}
			e.UpdatedAt = time.Now()
			vault.Modified = time.Now()
			if err := saveVault(vault, pass); err != nil {
				logger.Error("failed to save vault", "error", err)
				os.Exit(1)
			}
			fmt.Println("Запись обновлена.")
			return
		}
	}
	fmt.Println("Запись не найдена.")
}

func handleGenerate() {
	length := 20
	useLower := true
	useUpper := true
	useDigits := true
	useSymbols := true

	fs := flag.NewFlagSet("generate", flag.ExitOnError)
	fs.IntVar(&length, "len", 20, "Длина пароля")
	fs.BoolVar(&useLower, "lower", true, "Строчные буквы")
	fs.BoolVar(&useUpper, "upper", true, "Заглавные буквы")
	fs.BoolVar(&useDigits, "digits", true, "Цифры")
	fs.BoolVar(&useSymbols, "symbols", true, "Символы")
	fs.Parse(os.Args[2:])

	pwd := generatePassword(length, useLower, useUpper, useDigits, useSymbols)
	fmt.Println(pwd)
}

func handleReset() {
	if !vaultExists() {
		fmt.Println("Хранилище не найдено.")
		os.Exit(1)
	}

	oldPass := readPassword("Текущий мастер-пароль: ")
	vault, err := loadVault(oldPass)
	if err != nil {
		logger.Error("failed to load vault", "error", err)
		fmt.Println("Неверный мастер-пароль.")
		os.Exit(1)
	}

	newPass := readPassword("Новый мастер-пароль: ")
	confirm := readPassword("Подтвердите новый мастер-пароль: ")
	if newPass != confirm {
		fmt.Println("Пароли не совпадают.")
		os.Exit(1)
	}
	if len(newPass) < 8 {
		fmt.Println("Мастер-пароль должен быть не менее 8 символов.")
		os.Exit(1)
	}

	if err := saveVault(vault, newPass); err != nil {
		logger.Error("failed to reset password", "error", err)
		fmt.Println("Ошибка при смене пароля.")
		os.Exit(1)
	}
	fmt.Println("Мастер-пароль изменён.")
	logger.Info("master password reset")
}

func splitTags(s string) []string {
	parts := strings.Split(s, ",")
	var out []string
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-3] + "..."
}
