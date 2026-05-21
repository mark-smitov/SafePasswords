# SafePasswords

[![Go Version](https://img.shields.io/badge/Go-1.26.3-00ADD8?logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)
[![Platform](https://img.shields.io/badge/platform-windows%20%7C%20linux%20%7C%20macos-lightgrey)]()

Зашифрованный менеджер паролей с TUI-интерфейсом. Хранит пароли в AES-GCM, ключи выводятся через Argon2id.

## Возможности

- Хранение паролей в зашифрованном vault (AES-256-GCM + Argon2id)
- Интерактивный TUI (терминальный интерфейс на bubbletea)
- CLI-режим для скриптов
- Генератор надёжных паролей
- Поиск по названиям, логинам и URL
- Импорт/экспорт отсутствует (преднамеренно — vault привязан к файлу)

## Установка

```bash
git clone https://github.com/yourname/safepasswords.git
cd safepasswords
go build -o passman.exe .
```

## Использование

### TUI-режим

```bash
go run main.go
```

Открывает интерактивный интерфейс. Управление — клавиатура.

| Клавиша | Действие |
|---------|----------|
| `↑` / `↓` | Навигация |
| `Enter` | Выбрать / подтвердить |
| `Tab` / `Shift+Tab` | Следующее / предыдущее поле |
| `Esc` / `q` | Назад / выход |
| `y` / `n` | Подтверждение удаления |
| `e` | Редактировать запись |
| `d` | Удалить запись |
| `Space` | Вкл/выкл опцию генерации |
| `g` | Сгенерировать пароль |

### CLI-режим

```bash
# Инициализация хранилища
go run main.go init

# Добавить запись (интерактивный ввод)
go run main.go add

# Найти запись
go run main.go get <название>

# Список записей
go run main.go list

# Редактировать запись
go run main.go edit <название>

# Удалить запись
go run main.go delete <название>

# Сгенерировать пароль
go run main.go generate --len 24 --symbols

# Сбросить мастер-пароль
go run main.go reset
```

## Шифрование

- **Key derivation**: Argon2id (3 прохода, 64 MB памяти)
- **Encryption**: AES-256-GCM
- **Salt + Nonce**: случайные, 16 + 12 байт
- **Vault**: хранится в `passwords/vault.enc`, метаданные — в `passwords/vault.meta`

## Структура проекта

```
├── main.go              # Точка входа
├── internal/
│   ├── app.go           # CLI-обработчики
│   ├── tui.go           # TUI (bubbletea)
│   ├── crypto.go        # Argon2id + AES-GCM
│   ├── generator.go     # Генератор паролей
│   ├── logging.go       # Логирование (slog)
│   ├── models.go        # Entry, Vault, VaultMeta
│   └── storage.go       # Чтение/запись vault
```

## Зависимости

- [charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea) — TUI-фреймворк
- [charmbracelet/bubbles](https://github.com/charmbracelet/bubbles) — компоненты TUI
- [charmbracelet/lipgloss](https://github.com/charmbracelet/lipgloss) — стилизация
- [golang.org/x/crypto](https://pkg.go.dev/golang.org/x/crypto) — Argon2id
- [golang.org/x/term](https://pkg.go.dev/golang.org/x/term) — чтение пароля
