package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

type Storage interface {
	CreateAccount(*Account) error
	DeleteAccount(id int) error
	GetAccountByID(id int) (*Account, error)
	GetAccounts() ([]*Account, error)
	GetAccountByNumber(number int) (*Account, error)
	UpdatePassword(id int, newPass string) error
}

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(user string, dbname string, pw string, ssl string) (*PostgresStore, error) {
	connStr := fmt.Sprintf("user=%s dbname=%s password=%s sslmode=%s", user, dbname, pw, ssl)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresStore{
		db: db,
	}, nil

}

func (s *PostgresStore) Init() error {
	return s.CreateAccountTable()
}

func (s *PostgresStore) CreateAccountTable() error {
	query := `create table if not exists account(
		id serial primary key,
		first_name varchar(50),
		last_name varchar(50),
		number serial,
		encrypted_password varchar(100),
		balance serial,
		created_at timestamp
	)`
	_, err := s.db.Exec(query)
	return err
}

func (s *PostgresStore) CreateAccount(acc *Account) error {
	query := (`insert into account
		(first_name, last_name, number, encrypted_password, balance, created_at)
		values ($1, $2, $3, $4, $5, $6)`)

	_, err := s.db.Query(
		query,
		acc.FirstName,
		acc.LastName,
		acc.Number,
		acc.EncryptedPassword,
		acc.Balance,
		time.Now())
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Printf("%+ v\n", resp)
	return nil
}

func (s *PostgresStore) GetAccountByNumber(number int) (*Account, error) {
	rows, err := s.db.Query("select * from account where number = $1", number)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		return ScanIntoAccount(rows)
	}
	return nil, fmt.Errorf("account %d not found", number)

}

func (s *PostgresStore) UpdatePassword(id int, newPass string) error {
	statement := `
	UPDATE account
	SET encrypted_password = $1
	WHERE id = $2 `

	rows, err := s.db.Exec(statement, newPass, id)
	if err != nil {
		return err
	}
	count, err := rows.RowsAffected()
	if err != nil {
		return err
	}
	if count > 0 {
		fmt.Println("password successfully changed")
	}

	return nil
}

func (s *PostgresStore) DeleteAccount(id int) error {
	sqlStatement := `
	DELETE FROM account
	WHERE id = $1;`

	rows, err := s.db.Exec(sqlStatement, id)
	if err != nil {
		return err
	}
	count, err := rows.RowsAffected()

	fmt.Println("rows successfully deleted: ", count)

	if err != nil {
		return err
	}
	return nil
}

func (s *PostgresStore) GetAccountByID(id int) (*Account, error) {
	query := `select * from account where id=$1`
	rows, err := s.db.Query(query, id)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		fmt.Print("rows queried with for id : ", rows)
		return ScanIntoAccount(rows)
	}

	return nil, fmt.Errorf("account %d not found", id)
}

func (s *PostgresStore) GetAccounts() ([]*Account, error) {
	rows, err := s.db.Query("select * from account")
	if err != nil {
		return nil, err
	}

	accounts := []*Account{}
	for rows.Next() {
		account, err := ScanIntoAccount(rows)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}

	return accounts, nil
}

func ScanIntoAccount(rows *sql.Rows) (*Account, error) {
	account := new(Account)
	err := rows.Scan(
		&account.ID,
		&account.FirstName,
		&account.LastName,
		&account.Number,
		&account.EncryptedPassword,
		&account.Balance,
		&account.CreatedAt)

	if err != nil {
		return nil, err
	}
	return account, nil
}
