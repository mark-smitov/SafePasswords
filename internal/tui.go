package internal

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type screen int

const (
	screenMenu screen = iota
	screenInitPass
	screenInitConfirm
	screenUnlock
	screenList
	screenShow
	screenAdd
	screenFillPass
	screenEdit
	screenDelete
	screenGenerate
	screenResetOld
	screenResetNew
	screenResetConfirm
	screenMsg
)

var (
	baseStyle    = lipgloss.NewStyle().Padding(0, 1)
	accentStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7C3AED")).Padding(0, 1)
	dimStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280"))
	okStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#10B981"))
	failStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444"))
	selStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#7C3AED")).Bold(true)
	promptStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#9CA3AF"))
)

func checkMark(on bool) string {
	if on {
		return "[x]"
	}
	return "[ ]"
}

type model struct {
	screen    screen
	menuIdx   int
	width     int
	height    int

	vault     *Vault
	vaultPass string

	entries   []Entry
	entryIdx  int

	passInput textinput.Model
	searchInp textinput.Model

	addTitle    textinput.Model
	addUsername textinput.Model
	addPassword textinput.Model
	addURL      textinput.Model
	addNotes    textinput.Model
	addTags     textinput.Model
	addField    int

	editIdx     int
	editField   int
	editInputs  []textinput.Model

	genLength   int
	genLower    bool
	genUpper    bool
	genDigits   bool
	genSymbols  bool

	msgText     string
	msgIsError  bool

	initPass1   string
}

func initialModel() model {
	return model{
		screen:    screenMenu,
		genLength: 20,
		genLower:  true,
		genUpper:  true,
		genDigits: true,
		genSymbols: true,
	}
}

const tuiBanner = `
 ____         __      ____                                     _ 
/ ___|  __ _ / _| ___|  _ \ __ _ ___ _____      _____  _ __ __| |
\___ \ / _\` + "`" + ` | |_ / _ \ |_) / _\` + "`" + ` / __/ __\ \ /\ / / _ \| '__/ _\` + "`" + ` |
 ___) | (_| |  _|  __/  __/ (_| \__ \__ \ V  V / (_) | | | (_| |
|____/ \__,_|_|  \___|_|   \__,_|___/___/ \_/\_/ \___/|_|  \__,_|
`

func startTUI() {
	initLogger(false)
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}

		switch m.screen {
		case screenMenu:
			return m.updateMenu(msg)
		case screenInitPass:
			return m.updateInitPass(msg)
		case screenInitConfirm:
			return m.updateInitConfirm(msg)
		case screenUnlock:
			return m.updateUnlock(msg)
		case screenList:
			return m.updateList(msg)
		case screenShow:
			return m.updateShow(msg)
		case screenAdd:
			return m.updateAdd(msg)
		case screenFillPass:
			return m.updateFillPass(msg)
		case screenEdit:
			return m.updateEdit(msg)
		case screenDelete:
			return m.updateDelete(msg)
		case screenGenerate:
			return m.updateGenerate(msg)
		case screenResetOld:
			return m.updateResetOld(msg)
		case screenResetNew:
			return m.updateResetNew(msg)
		case screenResetConfirm:
			return m.updateResetConfirm(msg)
		case screenMsg:
			return m.updateMsg(msg)
		}
	}

	return m, nil
}

var menuItems = []string{
	"Список записей",
	"Добавить запись",
	"Найти запись",
	"Редактировать запись",
	"Удалить запись",
	"Сгенерировать пароль",
	"Сбросить мастер-пароль",
	"Выход",
}

var initItems = []string{
	"Инициализировать хранилище",
	"Выход",
}

func (m model) menuItems() []string {
	if vaultExists() {
		return menuItems
	}
	return initItems
}

func (m model) updateMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	items := m.menuItems()
	switch msg.String() {
	case "up", "k":
		if m.menuIdx > 0 {
			m.menuIdx--
		}
	case "down", "j":
		if m.menuIdx < len(items)-1 {
			m.menuIdx++
		}
	case "q", "esc":
		return m, tea.Quit
	case "enter":
		if !vaultExists() {
			switch m.menuIdx {
			case 0:
				m.screen = screenInitPass
				m.passInput = textinput.New()
		m.passInput.Placeholder = "Мастер-пароль"
		m.passInput.Width = 40
		m.passInput.EchoMode = textinput.EchoPassword
		m.passInput.Focus()
		return m, textinput.Blink
	case 1:
		return m, tea.Quit
			}
		} else {
			switch m.menuIdx {
			case 0:
				return m.openList()
			case 1:
				return m.openAdd()
			case 2:
				return m.openSearch()
			case 3:
				return m.openList()
			case 4:
				return m.openList()
			case 5:
				m.screen = screenGenerate
				return m, nil
			case 6:
				if m.vaultPass == "" {
					return m.openUnlock()
				}
				m.screen = screenResetOld
				m.passInput = textinput.New()
				m.passInput.Placeholder = "Текущий мастер-пароль"
				m.passInput.Width = 40
				m.passInput.EchoMode = textinput.EchoPassword
				m.passInput.Focus()
				return m, textinput.Blink
			case 7:
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m model) updateInitPass(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if len(m.passInput.Value()) < 8 {
			m.msgText = "Мастер-пароль должен быть не менее 8 символов"
			m.msgIsError = true
			m.screen = screenMsg
			return m, nil
		}
		m.initPass1 = m.passInput.Value()
		m.screen = screenInitConfirm
		m.passInput.SetValue("")
		m.passInput.Focus()
		return m, textinput.Blink
	case "esc":
		m.screen = screenMenu
		return m, nil
	}
	var cmd tea.Cmd
	m.passInput, cmd = m.passInput.Update(msg)
	return m, cmd
}

func (m model) updateInitConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if m.passInput.Value() != m.initPass1 {
			m.msgText = "Пароли не совпадают"
			m.msgIsError = true
			m.screen = screenMsg
			return m, nil
		}
		if err := ensurePasswordsDir(); err != nil {
			m.msgText = "Ошибка: " + err.Error()
			m.msgIsError = true
			m.screen = screenMsg
			return m, nil
		}
		vault := &Vault{
			Version:  vaultVersion,
			Created:  time.Now(),
			Modified: time.Now(),
			Entries:  []Entry{},
		}
		if err := saveVault(vault, m.initPass1); err != nil {
			m.msgText = "Ошибка: " + err.Error()
			m.msgIsError = true
			m.screen = screenMsg
			return m, nil
		}
		m.vaultPass = m.initPass1
		m.vault = vault
		m.msgText = "Хранилище создано"
		m.msgIsError = false
		m.screen = screenMsg
		return m, nil
	case "esc":
		m.screen = screenMenu
		return m, nil
	}
	var cmd tea.Cmd
	m.passInput, cmd = m.passInput.Update(msg)
	return m, cmd
}

func (m model) openUnlock() (tea.Model, tea.Cmd) {
	m.screen = screenUnlock
	m.passInput = textinput.New()
		m.passInput.Placeholder = "Мастер-пароль"
		m.passInput.Width = 40
		m.passInput.EchoMode = textinput.EchoPassword
		m.passInput.Focus()
		return m, textinput.Blink
}

func (m model) updateUnlock(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		pass := m.passInput.Value()
		vault, err := loadVault(pass)
		if err != nil {
			m.msgText = "Неверный мастер-пароль"
			m.msgIsError = true
			m.screen = screenMsg
			return m, nil
		}
		m.vault = vault
		m.vaultPass = pass
		m.entries = vault.Entries
		m.entryIdx = 0
		m.screen = screenMenu
		return m, nil
	case "esc":
		m.screen = screenMenu
		return m, nil
	}
	var cmd tea.Cmd
	m.passInput, cmd = m.passInput.Update(msg)
	return m, cmd
}

func (m model) openList() (tea.Model, tea.Cmd) {
	if !vaultExists() {
		m.msgText = "Хранилище не найдено"
		m.msgIsError = true
		m.screen = screenMsg
		return m, nil
	}
	if m.vaultPass == "" {
		return m.openUnlock()
	}
	vault, err := loadVault(m.vaultPass)
	if err != nil {
		m.msgText = "Ошибка загрузки: " + err.Error()
		m.msgIsError = true
		m.screen = screenMsg
		return m, nil
	}
	m.vault = vault
	m.entries = vault.Entries
	m.entryIdx = 0
	m.screen = screenList
	return m, nil
}

func (m model) openListFor(next screen) (tea.Model, tea.Cmd) {
	if !vaultExists() {
		m.msgText = "Хранилище не найдено"
		m.msgIsError = true
		m.screen = screenMsg
		return m, nil
	}
	if m.vaultPass == "" {
		m.menuIdx = int(next)
		return m.openUnlock()
	}
	vault, err := loadVault(m.vaultPass)
	if err != nil {
		m.msgText = "Ошибка загрузки: " + err.Error()
		m.msgIsError = true
		m.screen = screenMsg
		return m, nil
	}
	m.vault = vault
	m.entries = vault.Entries
	m.entryIdx = 0
	m.screen = next
	return m, nil
}

func (m model) openSearch() (tea.Model, tea.Cmd) {
	if !vaultExists() {
		m.msgText = "Хранилище не найдено"
		m.msgIsError = true
		m.screen = screenMsg
		return m, nil
	}
	if m.vaultPass == "" {
		return m.openUnlock()
	}
	vault, err := loadVault(m.vaultPass)
	if err != nil {
		m.msgText = "Ошибка загрузки: " + err.Error()
		m.msgIsError = true
		m.screen = screenMsg
		return m, nil
	}
	m.vault = vault
	m.entries = vault.Entries
	m.searchInp = textinput.New()
	m.searchInp.Placeholder = "Поиск..."
	m.searchInp.Width = 40
	m.searchInp.Focus()
	m.screen = screenShow
	return m, textinput.Blink
}

func (m model) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.entryIdx > 0 {
			m.entryIdx--
		}
	case "down", "j":
		if m.entryIdx < len(m.entries)-1 {
			m.entryIdx++
		}
	case "enter":
		m.screen = screenShow
		return m, nil
	case "d":
		if len(m.entries) > 0 {
			m.screen = screenDelete
		}
		return m, nil
	case "e":
		if len(m.entries) > 0 {
			return m.openEditEntry(m.entryIdx)
		}
	case "esc", "q":
		m.screen = screenMenu
		return m, nil
	}
	return m, nil
}

func (m model) updateShow(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.searchInp = textinput.New()
		m.screen = screenMenu
		return m, nil
	case "e":
		return m.openEditEntry(m.entryIdx)
	case "d":
		m.screen = screenDelete
		return m, nil
	case "left", "h":
		if m.entryIdx > 0 {
			m.entryIdx--
		}
		return m, nil
	case "right", "l":
		if m.entryIdx < len(m.entries)-1 {
			m.entryIdx++
		}
		return m, nil
	}

	if m.searchInp.Placeholder == "Поиск..." {
		var cmd tea.Cmd
		m.searchInp, cmd = m.searchInp.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m model) openAdd() (tea.Model, tea.Cmd) {
	if m.vaultPass == "" {
		return m.openUnlock()
	}

	m.addTitle = textinput.New()
	m.addTitle.Placeholder = "Название"
	m.addTitle.Width = 40
	m.addTitle.Focus()

	m.addUsername = textinput.New()
	m.addUsername.Placeholder = "Логин"
	m.addUsername.Width = 40

	m.addPassword = textinput.New()
	m.addPassword.Placeholder = "Пароль (пусто — сгенерировать)"
	m.addPassword.Width = 40

	m.addURL = textinput.New()
	m.addURL.Placeholder = "URL"
	m.addURL.Width = 40

	m.addNotes = textinput.New()
	m.addNotes.Placeholder = "Заметки"
	m.addNotes.Width = 40

	m.addTags = textinput.New()
	m.addTags.Placeholder = "Теги (через запятую)"
	m.addTags.Width = 40

	m.addField = 0
	m.screen = screenAdd
	return m, textinput.Blink
}

func (m model) updateAdd(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	fields := []*textinput.Model{
		&m.addTitle, &m.addUsername, &m.addPassword,
		&m.addURL, &m.addNotes, &m.addTags,
	}

	switch msg.String() {
	case "tab", "down":
		fields[m.addField].Blur()
		m.addField = (m.addField + 1) % len(fields)
		fields[m.addField].Focus()
		return m, textinput.Blink
	case "shift+tab", "up":
		fields[m.addField].Blur()
		m.addField = (m.addField - 1 + len(fields)) % len(fields)
		fields[m.addField].Focus()
		return m, textinput.Blink
	case "enter":
		if m.addField < len(fields)-1 {
			fields[m.addField].Blur()
			m.addField++
			fields[m.addField].Focus()
			return m, textinput.Blink
		}
		return m.submitAdd()
	case "esc":
		m.screen = screenMenu
		return m, nil
	}

	var cmd tea.Cmd
	*fields[m.addField], cmd = fields[m.addField].Update(msg)
	return m, cmd
}

func (m model) submitAdd() (tea.Model, tea.Cmd) {
	title := strings.TrimSpace(m.addTitle.Value())
	if title == "" {
		m.msgText = "Название обязательно"
		m.msgIsError = true
		m.screen = screenMsg
		return m, nil
	}

	password := strings.TrimSpace(m.addPassword.Value())
	if password == "" {
		password = generatePassword(20, true, true, true, true)
		m.msgText = "Создан пароль: " + password
		m.msgIsError = false
	}

	entry := Entry{
		ID:        generateID(),
		Title:     title,
		Username:  strings.TrimSpace(m.addUsername.Value()),
		Password:  password,
		URL:       strings.TrimSpace(m.addURL.Value()),
		Notes:     strings.TrimSpace(m.addNotes.Value()),
		Tags:      splitTags(m.addTags.Value()),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	m.vault.Entries = append(m.vault.Entries, entry)
	m.vault.Modified = time.Now()

	if err := saveVault(m.vault, m.vaultPass); err != nil {
		m.msgText = "Ошибка: " + err.Error()
		m.msgIsError = true
		m.screen = screenMsg
		return m, nil
	}
	m.entries = m.vault.Entries
	m.msgText = "Запись \"" + title + "\" добавлена"
	m.msgIsError = false
	m.screen = screenMsg
	return m, nil
}

func (m model) updateFillPass(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m model) openEditEntry(idx int) (tea.Model, tea.Cmd) {
	e := m.vault.Entries[idx]
	m.editIdx = idx

	inputs := make([]textinput.Model, 5)
	labels := []string{"Название", "Логин", "Пароль", "URL", "Заметки"}
	values := []string{e.Title, e.Username, e.Password, e.URL, e.Notes}
	for i := range inputs {
		inputs[i] = textinput.New()
		inputs[i].Placeholder = labels[i]
		inputs[i].SetValue(values[i])
		inputs[i].Width = 40
		inputs[i].Prompt = ""
	}
	inputs[0].Focus()
	m.editInputs = inputs
	m.editField = 0
	m.screen = screenEdit
	return m, textinput.Blink
}

func (m model) updateEdit(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if len(m.editInputs) == 0 {
		m.screen = screenMenu
		return m, nil
	}
	switch msg.String() {
	case "tab", "down":
		m.editInputs[m.editField].Blur()
		m.editField = (m.editField + 1) % len(m.editInputs)
		m.editInputs[m.editField].Focus()
		return m, textinput.Blink
	case "shift+tab", "up":
		m.editInputs[m.editField].Blur()
		m.editField = (m.editField - 1 + len(m.editInputs)) % len(m.editInputs)
		m.editInputs[m.editField].Focus()
		return m, textinput.Blink
	case "enter":
		if m.editField < len(m.editInputs)-1 {
			m.editInputs[m.editField].Blur()
			m.editField++
			m.editInputs[m.editField].Focus()
			return m, textinput.Blink
		}
		return m.submitEdit()
	case "esc":
		m.screen = screenShow
		return m, nil
	}

	var cmd tea.Cmd
	m.editInputs[m.editField], cmd = m.editInputs[m.editField].Update(msg)
	return m, cmd
}

func (m model) submitEdit() (tea.Model, tea.Cmd) {
	title := strings.TrimSpace(m.editInputs[0].Value())
	if title == "" {
		m.msgText = "Название обязательно"
		m.msgIsError = true
		m.screen = screenMsg
		return m, nil
	}

	e := &m.vault.Entries[m.editIdx]
	e.Title = title
	e.Username = strings.TrimSpace(m.editInputs[1].Value())
	e.Password = strings.TrimSpace(m.editInputs[2].Value())
	e.URL = strings.TrimSpace(m.editInputs[3].Value())
	e.Notes = strings.TrimSpace(m.editInputs[4].Value())
	e.UpdatedAt = time.Now()
	m.vault.Modified = time.Now()

	if err := saveVault(m.vault, m.vaultPass); err != nil {
		m.msgText = "Ошибка: " + err.Error()
		m.msgIsError = true
		m.screen = screenMsg
		return m, nil
	}
	m.entries = m.vault.Entries
	m.msgText = "Запись \"" + title + "\" обновлена"
	m.msgIsError = false
	m.screen = screenMsg
	return m, nil
}

func (m model) updateDelete(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "д":
		if m.entryIdx < len(m.vault.Entries) {
			title := m.vault.Entries[m.entryIdx].Title
			m.vault.Entries = append(m.vault.Entries[:m.entryIdx], m.vault.Entries[m.entryIdx+1:]...)
			m.vault.Modified = time.Now()
			if err := saveVault(m.vault, m.vaultPass); err != nil {
				m.msgText = "Ошибка: " + err.Error()
				m.msgIsError = true
				m.screen = screenMsg
				return m, nil
			}
			m.entries = m.vault.Entries
			if m.entryIdx >= len(m.entries) {
				m.entryIdx = max(0, len(m.entries)-1)
			}
			m.msgText = "Запись \"" + title + "\" удалена"
			m.msgIsError = false
			m.screen = screenMsg
			return m, nil
		}
	case "n", "т", "esc":
		m.screen = screenList
		return m, nil
	}
	return m, nil
}

func (m model) updateGenerate(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.menuIdx == 0 && m.genLength > 1 {
			m.genLength--
		}
	case "down", "j":
		if m.menuIdx == 0 && m.genLength < 128 {
			m.genLength++
		}
	case "left", "h":
		m.menuIdx = (m.menuIdx - 1 + 5) % 5
	case "right", "l":
		m.menuIdx = (m.menuIdx + 1) % 5
	case " ", "enter":
		switch m.menuIdx {
		case 1:
			m.genLower = !m.genLower
		case 2:
			m.genUpper = !m.genUpper
		case 3:
			m.genDigits = !m.genDigits
		case 4:
			m.genSymbols = !m.genSymbols
		}
	case "g":
		pwd := generatePassword(m.genLength, m.genLower, m.genUpper, m.genDigits, m.genSymbols)
		m.msgText = "Пароль: " + pwd
		m.msgIsError = false
		m.screen = screenMsg
		return m, nil
	case "esc", "q":
		m.screen = screenMenu
		return m, nil
	}
	return m, nil
}

func (m model) updateMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", "esc", " ", "q":
		m.screen = screenMenu
		return m, nil
	}
	return m, nil
}

func (m model) updateResetOld(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		pass := m.passInput.Value()
		_, err := loadVault(pass)
		if err != nil {
			m.msgText = "Неверный мастер-пароль"
			m.msgIsError = true
			m.screen = screenMsg
			return m, nil
		}
		m.initPass1 = pass
		m.screen = screenResetNew
		m.passInput.SetValue("")
		m.passInput.Placeholder = "Новый мастер-пароль"
		m.passInput.Focus()
		return m, textinput.Blink
	case "esc":
		m.screen = screenMenu
		return m, nil
	}
	var cmd tea.Cmd
	m.passInput, cmd = m.passInput.Update(msg)
	return m, cmd
}

func (m model) updateResetNew(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if len(m.passInput.Value()) < 8 {
			m.msgText = "Мастер-пароль должен быть не менее 8 символов"
			m.msgIsError = true
			m.screen = screenMsg
			return m, nil
		}
		m.initPass1 = m.passInput.Value()
		m.screen = screenResetConfirm
		m.passInput.SetValue("")
		m.passInput.Placeholder = "Подтвердите новый пароль"
		m.passInput.Focus()
		return m, textinput.Blink
	case "esc":
		m.screen = screenMenu
		return m, nil
	}
	var cmd tea.Cmd
	m.passInput, cmd = m.passInput.Update(msg)
	return m, cmd
}

func (m model) updateResetConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if m.passInput.Value() != m.initPass1 {
			m.msgText = "Пароли не совпадают"
			m.msgIsError = true
			m.screen = screenMsg
			return m, nil
		}
		vault, err := loadVault(m.vaultPass)
		if err != nil {
			m.msgText = "Ошибка: " + err.Error()
			m.msgIsError = true
			m.screen = screenMsg
			return m, nil
		}
		newPass := m.passInput.Value()
		if err := saveVault(vault, newPass); err != nil {
			m.msgText = "Ошибка: " + err.Error()
			m.msgIsError = true
			m.screen = screenMsg
			return m, nil
		}
		m.vaultPass = newPass
		m.msgText = "Мастер-пароль изменён"
		m.msgIsError = false
		m.screen = screenMsg
		return m, nil
	case "esc":
		m.screen = screenMenu
		return m, nil
	}
	var cmd tea.Cmd
	m.passInput, cmd = m.passInput.Update(msg)
	return m, cmd
}

func (m model) View() string {
	var b strings.Builder
	b.WriteString("\n")

	switch m.screen {
	case screenMenu:
		b.WriteString(tuiBanner)
		b.WriteString("\n")
		items := m.menuItems()
		for i, item := range items {
			if i == m.menuIdx {
				b.WriteString("  > " + selStyle.Render(item) + "\n")
			} else {
				b.WriteString("    " + item + "\n")
			}
		}
		b.WriteString("\n  " + dimStyle.Render("↑↓ — навигация  Enter — выбрать  q — выход"))

	case screenInitPass:
		b.WriteString(accentStyle.Render("Инициализация хранилища") + "\n\n")
		b.WriteString("  Придумайте мастер-пароль (мин. 8 символов):\n\n")
		b.WriteString("  " + m.passInput.View() + "\n\n")
		b.WriteString("  " + dimStyle.Render("Enter — подтвердить  Esc — назад"))

	case screenInitConfirm:
		b.WriteString(accentStyle.Render("Инициализация хранилища") + "\n\n")
		b.WriteString("  Подтвердите мастер-пароль:\n\n")
		b.WriteString("  " + m.passInput.View() + "\n\n")
		b.WriteString("  " + dimStyle.Render("Enter — подтвердить  Esc — назад"))

	case screenUnlock:
		b.WriteString(accentStyle.Render("Разблокировка") + "\n\n")
		b.WriteString("  Мастер-пароль:\n\n")
		b.WriteString("  " + m.passInput.View() + "\n\n")
		b.WriteString("  " + dimStyle.Render("Enter — разблокировать  Esc — назад"))

	case screenList:
		header := fmt.Sprintf("Записи (%d)", len(m.entries))
		b.WriteString(accentStyle.Render(header) + "\n\n")
		if len(m.entries) == 0 {
			b.WriteString("  В хранилище нет записей\n\n")
			b.WriteString("  " + dimStyle.Render("q — назад"))
			return b.String()
		}
		b.WriteString(fmt.Sprintf("  %-3s %-25s %-20s %-30s\n", "#", "Название", "Логин", "URL"))
		b.WriteString("  " + strings.Repeat("─", 78) + "\n")

		maxShow := len(m.entries)
		if m.height > 8 {
			maxShow = min(maxShow, m.height-8)
		}
		start := 0
		if len(m.entries) > maxShow && m.entryIdx > maxShow-3 {
			start = m.entryIdx - maxShow + 3
		}
		end := min(start+maxShow, len(m.entries))

		for i := start; i < end; i++ {
			e := m.entries[i]
			cursor := "  "
			if i == m.entryIdx {
				cursor = "> "
			}
			b.WriteString(fmt.Sprintf("  %s%-3d %-25s %-20s %-30s\n",
				cursor, i+1,
				truncate(e.Title, 22),
				truncate(e.Username, 17),
				truncate(e.URL, 27)))
		}
		b.WriteString("\n  " + dimStyle.Render("↑↓ — навигация  Enter — открыть  d — удалить  q — назад"))

	case screenShow:
		if m.searchInp.Value() != "" {
			q := strings.ToLower(m.searchInp.Value())
			found := false
			for _, e := range m.entries {
				if strings.Contains(strings.ToLower(e.Title), q) ||
					strings.Contains(strings.ToLower(e.Username), q) ||
					strings.Contains(strings.ToLower(e.URL), q) {
					b.WriteString(accentStyle.Render("Поиск") + "\n\n")
					b.WriteString(fmt.Sprintf("    Название: %s\n", e.Title))
					b.WriteString(fmt.Sprintf("      Логин: %s\n", e.Username))
					b.WriteString(fmt.Sprintf("     Пароль: %s\n", e.Password))
					b.WriteString(fmt.Sprintf("         URL: %s\n", e.URL))
					if e.Notes != "" {
						b.WriteString(fmt.Sprintf("     Заметки: %s\n", e.Notes))
					}
					if len(e.Tags) > 0 {
						b.WriteString(fmt.Sprintf("        Теги: %s\n", strings.Join(e.Tags, ", ")))
					}
					b.WriteString("\n  " + dimStyle.Render("Esc — назад"))
					found = true
					break
				}
			}
			if !found {
				b.WriteString(accentStyle.Render("Поиск") + "\n\n")
				b.WriteString("  " + m.searchInp.View() + "\n\n")
				b.WriteString("  Запись не найдена\n\n")
				b.WriteString("  " + dimStyle.Render("Esc — назад"))
			}
			return b.String()
		}

		if len(m.entries) == 0 {
			b.WriteString(accentStyle.Render("Поиск") + "\n\n")
			b.WriteString("  " + m.searchInp.View() + "\n\n")
			b.WriteString("  " + dimStyle.Render("Введите текст для поиска  Esc — назад"))
			return b.String()
		}

		if m.entryIdx >= len(m.entries) {
			m.entryIdx = max(0, len(m.entries)-1)
		}
		e := m.entries[m.entryIdx]
		b.WriteString(accentStyle.Render(e.Title) + "\n\n")
		b.WriteString(fmt.Sprintf("      ID: %s\n", e.ID))
		b.WriteString(fmt.Sprintf("   Логин: %s\n", e.Username))
		b.WriteString(fmt.Sprintf("  Пароль: %s\n", e.Password))
		b.WriteString(fmt.Sprintf("     URL: %s\n", e.URL))
		if e.Notes != "" {
			b.WriteString(fmt.Sprintf("  Заметки: %s\n", e.Notes))
		}
		if len(e.Tags) > 0 {
			b.WriteString(fmt.Sprintf("     Теги: %s\n", strings.Join(e.Tags, ", ")))
		}
		b.WriteString(fmt.Sprintf("Обновлено: %s\n", e.UpdatedAt.Format("02.01.2006 15:04")))
		b.WriteString("\n  " + dimStyle.Render("← → — навигация  e — править  d — удалить  q — назад"))

	case screenAdd:
		b.WriteString(accentStyle.Render("Добавить запись") + "\n\n")
		fields := []*textinput.Model{
			&m.addTitle, &m.addUsername, &m.addPassword,
			&m.addURL, &m.addNotes, &m.addTags,
		}
		labels := []string{
			"Название", "Логин", "Пароль",
			"URL", "Заметки", "Теги (через запятую)",
		}
		for i, f := range fields {
			cursor := " "
			if i == m.addField {
				cursor = ">"
			}
			b.WriteString(fmt.Sprintf("  %s %s\n", cursor, labels[i]+":"))
			b.WriteString(fmt.Sprintf("     %s\n", f.View()))
		}
		b.WriteString("\n  " + dimStyle.Render("Tab — след. поле  Enter — сохранить  Esc — отмена"))

	case screenEdit:
		b.WriteString(accentStyle.Render("Редактировать") + "\n\n")
		labels := []string{"Название", "Логин", "Пароль", "URL", "Заметки"}
		for i, inp := range m.editInputs {
			cursor := " "
			if i == m.editField {
				cursor = ">"
			}
			b.WriteString(fmt.Sprintf("  %s %s\n", cursor, labels[i]+":"))
			b.WriteString(fmt.Sprintf("     %s\n", inp.View()))
		}
		b.WriteString("\n  " + dimStyle.Render("Tab — след. поле  Enter — сохранить  Esc — назад"))

	case screenDelete:
		b.WriteString(accentStyle.Render("Удалить запись") + "\n\n")
		if m.entryIdx < len(m.entries) {
			e := m.entries[m.entryIdx]
			b.WriteString(fmt.Sprintf("  Удалить \"%s\"?\n\n", e.Title))
			b.WriteString("  " + dimStyle.Render("y — да  n — нет"))
		}

	case screenGenerate:
		b.WriteString(accentStyle.Render("Генератор паролей") + "\n\n")
		items := []struct {
			label string
			val   string
		}{
			{"Длина", fmt.Sprintf("%d", m.genLength)},
			{"Строчные (a-z)", checkMark(m.genLower)},
			{"Заглавные (A-Z)", checkMark(m.genUpper)},
			{"Цифры (0-9)", checkMark(m.genDigits)},
			{"Символы (!@#...)", checkMark(m.genSymbols)},
		}
		for i, item := range items {
			cursor := " "
			if i == m.menuIdx {
				cursor = ">"
			}
			b.WriteString(fmt.Sprintf("  %s %-20s %s\n", cursor, item.label+":", item.val))
		}
		b.WriteString("\n  " + dimStyle.Render("↑↓ — длина  ← → — выбор  Space — вкл/выкл  g — создать  q — назад"))

	case screenResetOld:
		b.WriteString(accentStyle.Render("Сброс мастер-пароля") + "\n\n")
		b.WriteString("  Введите текущий мастер-пароль:\n\n")
		b.WriteString("  " + m.passInput.View() + "\n\n")
		b.WriteString("  " + dimStyle.Render("Enter — далее  Esc — назад"))

	case screenResetNew:
		b.WriteString(accentStyle.Render("Сброс мастер-пароля") + "\n\n")
		b.WriteString("  Придумайте новый мастер-пароль (мин. 8 символов):\n\n")
		b.WriteString("  " + m.passInput.View() + "\n\n")
		b.WriteString("  " + dimStyle.Render("Enter — далее  Esc — назад"))

	case screenResetConfirm:
		b.WriteString(accentStyle.Render("Сброс мастер-пароля") + "\n\n")
		b.WriteString("  Подтвердите новый мастер-пароль:\n\n")
		b.WriteString("  " + m.passInput.View() + "\n\n")
		b.WriteString("  " + dimStyle.Render("Enter — подтвердить  Esc — назад"))

	case screenMsg:
		if m.msgIsError {
			b.WriteString(failStyle.Render(m.msgText))
		} else {
			b.WriteString(okStyle.Render(m.msgText))
		}
		b.WriteString("\n\n  " + dimStyle.Render("Enter — продолжить"))
	}

	return b.String()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
