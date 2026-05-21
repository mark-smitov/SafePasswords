# SafePasswords

![Go](https://img.shields.io/badge/Go-1.26-blue?logo=go)
![License](https://img.shields.io/badge/license-MIT-green)
![Windows](https://img.shields.io/badge/Windows-0078D6?logo=windows)
![Linux](https://img.shields.io/badge/Linux-FCC624?logo=linux&logoColor=black)
![macOS](https://img.shields.io/badge/macOS-000000?logo=apple)

Менеджер паролей в терминале. Всё шифруется, ничего не улетает в интернет.

## Что внутри

- AES-256-GCM + Argon2id — ключи выводит, данные шифрует
- TUI на bubbletea — мышка не нужна, всё с клавиатуры
- CLI-режим — для скриптов и быстрых команд
- Генератор паролей — длина, регистр, цифры, символы

## Как собрать

```bash
git clone https://github.com/vksmitov/safepasswords.git
cd safepasswords
go build -o passman.exe .
```

## Как пользоваться

```bash
# TUI — интерактивный режим
go run main.go

# CLI — команды
go run main.go init          # создать хранилище
go run main.go add           # добавить запись
go run main.go list          # список записей
go run main.go get <имя>     # найти запись
go run main.go edit <имя>    # изменить запись
go run main.go delete <имя>  # удалить запись
go run main.go generate      # сгенерировать пароль
go run main.go reset          # сменить мастер-пароль
```

## Горячие клавиши в TUI

| Клавиша | Что делает |
|---------|-----------|
| `↑` `↓` | Выбрать пункт |
| `Enter` | Подтвердить |
| `Tab` | Следующее поле |
| `e` | Редактировать |
| `d` | Удалить |
| `q` / `Esc` | Назад / выход |

## Куда сохраняется

`passwords/vault.enc` — зашифрованные данные
`passwords/vault.meta` — соль, nonce, параметры Argon2

Никаких облаков, никаких серверов. Всё локально.

## Автор

[@vksmitov](https://github.com/vksmitov)
