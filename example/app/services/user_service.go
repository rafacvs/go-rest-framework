package services

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

type User struct {
	ID       int    `json:"id"`
	Nome     string `json:"nome"`
	Email    string `json:"email"`
	Telefone string `json:"telefone"`
}

type UserService struct {
	filePath string
	mu       sync.Mutex
}

func NewUserService() *UserService {
	// Caminho relativo ao diretório de execução (assume execução da raiz do projeto)
	filePath := filepath.Join("example", "app", "data", "users.txt")
	// Tentar tornar absoluto se possível
	if abs, err := filepath.Abs(filePath); err == nil {
		filePath = abs
	}
	return &UserService{
		filePath: filePath,
	}
}

func (s *UserService) LoadUsers() ([]User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	file, err := os.Open(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []User{}, nil
		}
		return nil, err
	}
	defer file.Close()

	var users []User
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) != 4 {
			continue
		}

		id, err := strconv.Atoi(strings.TrimSpace(parts[0]))
		if err != nil {
			continue
		}

		users = append(users, User{
			ID:       id,
			Nome:     strings.TrimSpace(parts[1]),
			Email:    strings.TrimSpace(parts[2]),
			Telefone: strings.TrimSpace(parts[3]),
		})
	}

	return users, scanner.Err()
}

func (s *UserService) SaveUsers(users []User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Criar diretório se não existir
	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	file, err := os.Create(s.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, user := range users {
		line := fmt.Sprintf("%d|%s|%s|%s\n", user.ID, user.Nome, user.Email, user.Telefone)
		if _, err := writer.WriteString(line); err != nil {
			return err
		}
	}

	return writer.Flush()
}

func (s *UserService) GetAllUsers() ([]User, error) {
	return s.LoadUsers()
}

func (s *UserService) GetUserByID(id int) (*User, error) {
	users, err := s.LoadUsers()
	if err != nil {
		return nil, err
	}

	for i := range users {
		if users[i].ID == id {
			return &users[i], nil
		}
	}

	return nil, fmt.Errorf("usuário não encontrado")
}

func (s *UserService) CreateUser(user User) (*User, error) {
	// Validação básica
	if user.Nome == "" {
		return nil, fmt.Errorf("nome é obrigatório")
	}
	if user.Email == "" {
		return nil, fmt.Errorf("email é obrigatório")
	}
	if !strings.Contains(user.Email, "@") {
		return nil, fmt.Errorf("email inválido")
	}

	users, err := s.LoadUsers()
	if err != nil {
		return nil, err
	}

	// Gerar novo ID
	newID := 1
	for _, u := range users {
		if u.ID >= newID {
			newID = u.ID + 1
		}
	}

	user.ID = newID
	users = append(users, user)

	if err := s.SaveUsers(users); err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *UserService) UpdateUser(id int, user User) (*User, error) {
	// Validação básica
	if user.Nome == "" {
		return nil, fmt.Errorf("nome é obrigatório")
	}
	if user.Email == "" {
		return nil, fmt.Errorf("email é obrigatório")
	}
	if !strings.Contains(user.Email, "@") {
		return nil, fmt.Errorf("email inválido")
	}

	users, err := s.LoadUsers()
	if err != nil {
		return nil, err
	}

	found := false
	for i := range users {
		if users[i].ID == id {
			users[i].Nome = user.Nome
			users[i].Email = user.Email
			users[i].Telefone = user.Telefone
			found = true
			user = users[i]
			break
		}
	}

	if !found {
		return nil, fmt.Errorf("usuário não encontrado")
	}

	if err := s.SaveUsers(users); err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *UserService) DeleteUser(id int) error {
	users, err := s.LoadUsers()
	if err != nil {
		return err
	}

	found := false
	newUsers := make([]User, 0, len(users))
	for _, u := range users {
		if u.ID != id {
			newUsers = append(newUsers, u)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("usuário não encontrado")
	}

	return s.SaveUsers(newUsers)
}

