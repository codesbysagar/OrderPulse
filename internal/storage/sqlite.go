package storage

import (
	"database/sql"
	"fmt"

	"github.com/codesbysagar/OrderPulse/internal/domain"
	_ "modernc.org/sqlite"
)

type TicketRepository struct {
	db *sql.DB
}

func NewTicketRepository(dbPath string) (*TicketRepository, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite db: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping sqlite db: %w", err)
	}

	repo := &TicketRepository{db: db}
	if err := repo.migrate(); err != nil {
		return nil, fmt.Errorf("failed to migrate schema: %w", err)
	}

	return repo, nil
}

func (r *TicketRepository) Close() error {
	return r.db.Close()
}

func (r *TicketRepository) migrate() error {
	query := `
	CREATE TABLE IF NOT EXISTS tickets (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			created_by TEXT NOT NULL,
			status TEXT NOT NULL,
			created_at DATETIME NOT NULL,
			sla_deadline DATETIME NOT NULL,
			completed_at DATETIME,
			breach_alerted BOOLEAN NOT NULL DEFAULT 0
	);
	
	CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			created_at DATETIME NOT NULL
	);`
	_, err := r.db.Exec(query)
	return err
}

func (r *TicketRepository) Create(t *domain.Ticket) error {
	query := `
	INSERT INTO tickets (title, created_by, status, created_at, sla_deadline, completed_at, breach_alerted)
	VALUES (?, ?, ?, ?, ?, ?, ?)`

	res, err := r.db.Exec(
		query,
		t.Title,
		t.CreatedBy,
		string(t.Status),
		t.CreatedAt.UTC(),
		t.SLADeadline.UTC(),
		t.CompletedAt,
		t.BreachAlerted,
	)
	if err != nil {
		return fmt.Errorf("insert ticket: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("fetch last insert id: %w", err)
	}
	t.ID = id
	return nil
}

func (r *TicketRepository) ListActive() ([]*domain.Ticket, error) {
	query := `
	SELECT id, title, created_by, status, created_at, sla_deadline, completed_at, breach_alerted
	FROM tickets
	WHERE status != ?
	ORDER BY sla_deadline ASC`

	rows, err := r.db.Query(query, string(domain.StatusCompleted))
	if err != nil {
		return nil, fmt.Errorf("query active tickets: %w", err)
	}
	defer rows.Close()

	var tickets []*domain.Ticket
	for rows.Next() {
		var t domain.Ticket
		var statusStr string
		var completedAt sql.NullTime

		err := rows.Scan(
			&t.ID,
			&t.Title,
			&t.CreatedBy,
			&statusStr,
			&t.CreatedAt,
			&t.SLADeadline,
			&t.CompletedAt,
			&t.BreachAlerted,
		)
		if err != nil {
			return nil, fmt.Errorf("scan ticket row: %w", err)
		}

		t.Status = domain.TicketStatus(statusStr)
		if completedAt.Valid {
			t.CompletedAt = &completedAt.Time
		}
		tickets = append(tickets, &t)
	}
	return tickets, rows.Err()
}

func (r *TicketRepository) Update(t *domain.Ticket) error {
	query := `
	UPDATE tickets
	SET status = ?, completed_at = ?, breach_alerted = ?
	WHERE id = ?`

	_, err := r.db.Exec(query, string(t.Status), t.CompletedAt, t.BreachAlerted, t.ID)
	if err != nil {
		return fmt.Errorf("update ticket #%d: %w", t.ID, err)
	}

	return nil
}

func (r *TicketRepository) CreateUser(u *domain.User) error {
	query := `
	INSERT INTO users (username, password_hash, created_at)
	VALUES (?, ?, ?)`

	res, err := r.db.Exec(query, u.Username, u.PasswordHash, u.CreatedAt.UTC())
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	u.ID = id
	return nil
}

func (r *TicketRepository) GetUserByUsername(username string) (*domain.User, error) {
	query := `
	SELECT id, username, password_hash, created_at
	FROM users
	WHERE username = ?`
	row := r.db.QueryRow(query, username)

	var u domain.User
	err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}
