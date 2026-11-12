package store

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"libratrack-security/internal/model"
)

var (
	ErrNotFound       = errors.New("not found")
	ErrDuplicateEmail = errors.New("email already registered")
	ErrAlreadyLoaned  = errors.New("book already on loan")
	ErrNotOnLoan      = errors.New("loan not active")
)

type Store struct {
	mu     sync.RWMutex
	users  map[int]*model.User
	books  map[int]*model.Book
	loans  map[int]*model.Loan
	nextU  int
	nextB  int
	nextL  int
}

func New() *Store {
	s := &Store{
		users: make(map[int]*model.User),
		books: make(map[int]*model.Book),
		loans: make(map[int]*model.Loan),
		nextU: 1,
		nextB: 1,
		nextL: 1,
	}

	// seed some books
	s.books[1] = &model.Book{ID: 1, Title: "The Go Programming Language", Author: "Donovan & Kernighan", ISBN: "978-0134190440", CreatedAt: time.Now()}
	s.books[2] = &model.Book{ID: 2, Title: "Designing Data-Intensive Applications", Author: "Martin Kleppmann", ISBN: "978-1449373320", CreatedAt: time.Now()}
	s.books[3] = &model.Book{ID: 3, Title: "Clean Code", Author: "Robert C. Martin", ISBN: "978-0132350884", CreatedAt: time.Now()}
	s.nextB = 4
	return s
}

func (s *Store) CreateUser(email, passwordHash, role string) (*model.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, u := range s.users {
		if u.Email == email {
			return nil, ErrDuplicateEmail
		}
	}

	u := &model.User{
		ID:           s.nextU,
		Email:        email,
		PasswordHash: passwordHash,
		Role:         role,
		CreatedAt:    time.Now(),
	}
	s.users[s.nextU] = u
	s.nextU++
	return u, nil
}

func (s *Store) GetUserByEmail(email string) (*model.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, u := range s.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, ErrNotFound
}

func (s *Store) GetUserByID(id int) (*model.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	u, ok := s.users[id]
	if !ok {
		return nil, ErrNotFound
	}
	return u, nil
}

func (s *Store) ListUsers() []*model.User {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]*model.User, 0, len(s.users))
	for _, u := range s.users {
		out = append(out, u)
	}
	return out
}

func (s *Store) UpdateUserRole(id int, role string) (*model.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	u, ok := s.users[id]
	if !ok {
		return nil, ErrNotFound
	}
	u.Role = role
	return u, nil
}

func (s *Store) ListBooks() []*model.Book {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]*model.Book, 0, len(s.books))
	for _, b := range s.books {
		out = append(out, b)
	}
	return out
}

func (s *Store) GetBook(id int) (*model.Book, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	b, ok := s.books[id]
	if !ok {
		return nil, ErrNotFound
	}
	return b, nil
}

func (s *Store) CreateBook(title, author, isbn string) *model.Book {
	s.mu.Lock()
	defer s.mu.Unlock()

	b := &model.Book{
		ID:        s.nextB,
		Title:     title,
		Author:    author,
		ISBN:      isbn,
		CreatedAt: time.Now(),
	}
	s.books[s.nextB] = b
	s.nextB++
	return b
}

func (s *Store) UpdateBook(id int, title, author, isbn string) (*model.Book, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	b, ok := s.books[id]
	if !ok {
		return nil, ErrNotFound
	}
	if title != "" {
		b.Title = title
	}
	if author != "" {
		b.Author = author
	}
	if isbn != "" {
		b.ISBN = isbn
	}
	return b, nil
}

func (s *Store) CreateLoan(userID, bookID int) (*model.Loan, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.users[userID]; !ok {
		return nil, fmt.Errorf("user %d: %w", userID, ErrNotFound)
	}
	if _, ok := s.books[bookID]; !ok {
		return nil, fmt.Errorf("book %d: %w", bookID, ErrNotFound)
	}

	for _, l := range s.loans {
		if l.BookID == bookID && l.ReturnedAt == nil {
			return nil, ErrAlreadyLoaned
		}
	}

	l := &model.Loan{
		ID:         s.nextL,
		UserID:     userID,
		BookID:     bookID,
		BorrowedAt: time.Now(),
	}
	s.loans[s.nextL] = l
	s.nextL++
	return l, nil
}

func (s *Store) ReturnLoan(loanID int) (*model.Loan, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	l, ok := s.loans[loanID]
	if !ok {
		return nil, ErrNotFound
	}
	if l.ReturnedAt != nil {
		return nil, ErrNotOnLoan
	}
	now := time.Now()
	l.ReturnedAt = &now
	return l, nil
}

func (s *Store) UserLoans(userID int) []*model.Loan {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var out []*model.Loan
	for _, l := range s.loans {
		if l.UserID == userID {
			out = append(out, l)
		}
	}
	return out
}
