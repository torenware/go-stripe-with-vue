package models

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// DBModel is the type for database connection values
type DBModel struct {
	DB *sql.DB
}

// Models is the wrapper for all models
type Models struct {
	DB DBModel
}

// NewModels returns a model type with database connection pool
func NewModels(db *sql.DB) Models {
	return Models{
		DB: DBModel{DB: db},
	}
}

// Widget is the type for all widgets
type Widget struct {
	ID             int       `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	InventoryLevel int       `json:"inventory_level"`
	Price          int       `json:"price"`
	IsRecurring    bool      `json:"is_recurring"`
	PlanID         string    `json:"plan_id"`
	Image          string    `json:"image"`
	CreatedAt      time.Time `json:"-"`
	UpdatedAt      time.Time `json:"-"`
}

// Order is the type for all orders
type Order struct {
	ID            int       `json:"id"`
	WidgetID      int       `json:"widget_id"`
	TransactionID int       `json:"transaction_id"`
	CustomerID    int       `json:"customer_id"`
	StatusID      int       `json:"status_id"`
	Quantity      int       `json:"quantity"`
	Amount        int       `json:"amount"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"-"`
	Widget	Widget `json:"widget"`
	Transaction Transaction `json:"transaction"`
	Customer Customer  `json:"customer"`
}

// Status is the type for order statuses
type Status struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

// TransactionStatus is the type for transaction statuses
type TransactionStatus struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

// Transaction is the type for transactions
type Transaction struct {
	ID                  int       `json:"id"`
	Amount              int       `json:"amount"`
	Currency            string    `json:"currency"`
	LastFour            string    `json:"last_four"`
	ExpiryMonth         int       `json:"expiry_month"`
	ExpiryYear          int       `json:"expiry_year"`
	BankReturnCode      string    `json:"bank_return_code"`
	PaymentIntent       string    `json:"payment_intent"`
	PaymentMethod       string    `json:"payment_method"`
	TransactionStatusID int       `json:"transaction_status_id"`
	CreatedAt           time.Time `json:"-"`
	UpdatedAt           time.Time `json:"-"`
}

// User is the type for users
type User struct {
	ID        int       `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

// Customer is the type for users
type Customer struct {
	ID        int       `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

// GetWidget gets one widget by id
func (m *DBModel) GetWidget(id int) (Widget, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var widget Widget

	row := m.DB.QueryRowContext(ctx, `
		select
			id, name, description, inventory_level, price, coalesce(image, ''),
			is_recurring, plan_id,
			created_at, updated_at
		from
			widgets
		where id = ?`, id)
	err := row.Scan(
		&widget.ID,
		&widget.Name,
		&widget.Description,
		&widget.InventoryLevel,
		&widget.Price,
		&widget.Image,
		&widget.IsRecurring,
		&widget.PlanID,
		&widget.CreatedAt,
		&widget.UpdatedAt,
	)
	if err != nil {
		return widget, err
	}

	return widget, nil
}

// InsertTransaction inserts a new txn, and returns its id
func (m *DBModel) InsertTransaction(txn Transaction) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
		insert into transactions
			(amount, currency, last_four, bank_return_code,
			payment_intent, payment_method,
			transaction_status_id,
			expiry_month, expiry_year,
			created_at, updated_at)
		values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := m.DB.ExecContext(ctx, stmt,
		txn.Amount,
		txn.Currency,
		txn.LastFour,
		txn.BankReturnCode,
		txn.PaymentIntent,
		txn.PaymentMethod,
		txn.TransactionStatusID,
		txn.ExpiryMonth,
		txn.ExpiryYear,
		time.Now(),
		time.Now(),
	)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

// InsertOrder inserts a new order, and returns its id
func (m *DBModel) InsertOrder(order Order) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
		insert into orders
			(amount, quantity, widget_id, transaction_id,
			 status_id, customer_id, created_at, updated_at)
		values (?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := m.DB.ExecContext(ctx, stmt,
		order.Amount,
		order.Quantity,
		order.WidgetID,
		order.TransactionID,
		order.StatusID,
		order.CustomerID,
		time.Now(),
		time.Now(),
	)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

// InsertCustomer inserts a new customer, and returns its id
func (m *DBModel) InsertCustomer(customer Customer) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
		insert into customers
			(first_name, last_name, email,
			 created_at, updated_at)
		values (?, ?, ?, ?, ?)
	`

	result, err := m.DB.ExecContext(ctx, stmt,
		customer.FirstName,
		customer.LastName,
		customer.Email,
		time.Now(),
		time.Now(),
	)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (m *DBModel) GetUserByEmail(email string) (User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var u User
	row := m.DB.QueryRowContext(ctx, `
		select
			id, first_name, last_name, email, password
		from users
		where email = ?
	`, strings.ToLower(email))
	err := row.Scan(
		&u.ID,
		&u.FirstName,
		&u.LastName,
		&u.Email,
		&u.Password,
	)
	if err != nil {
		return u, err
	}
	return u, nil
}

func (m *DBModel) Authenticate(email, password string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var u User
	row := m.DB.QueryRowContext(ctx, `
		select
			id, email, password
		from users
		where email = ?
	`, strings.ToLower(email))
	err := row.Scan(
		&u.ID,
		&u.Email,
		&u.Password,
	)
	if err != nil {
		return 0, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	if err != nil {
		return 0, err
	}

	return u.ID, nil

}

func (m *DBModel) UpdatePasswordForUser(u User, hash string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `update users set password = ?, updated_at=now() where id = ?`
	_, err := m.DB.ExecContext(ctx, stmt, hash, u.ID)
	if err != nil {
		return err
	}

	return nil
}


func (m *DBModel) GetAllSales() ([]*Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
select
    o.amount, o.quantity, w.price, t.currency,
    o.id as order_id, o.widget_id, o.transaction_id,o.customer_id,
    o.created_at,  o.status_id,
    w.name as item, w.description,
    t.last_four, t.expiry_month, t.expiry_year,
    t.payment_intent, t.bank_return_code,
    c.first_name, c.last_name, c.email

from orders o
         left join widgets w on (o.widget_id = w.id)
         left join transactions t on (o.transaction_id = t.id)
         left join customers c on (o.customer_id= c.id)

where
        w.is_recurring = 0
order by
		o.created_at desc
`
	var rslt []*Order
	rows, err := m.DB.QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)
	for rows.Next() {
		var o Order
		err = rows.Scan(
			&o.Amount,
			&o.Quantity,
			&o.Widget.Price,
			&o.Transaction.Currency,
			&o.ID,
			&o.WidgetID,
			&o.TransactionID,
			&o.CustomerID,
			&o.CreatedAt,
			&o.StatusID,
			&o.Widget.Name,
			&o.Widget.Description,
			&o.Transaction.LastFour,
			&o.Transaction.ExpiryMonth,
			&o.Transaction.ExpiryYear,
			&o.Transaction.PaymentIntent,
			&o.Transaction.BankReturnCode,
			&o.Customer.FirstName,
			&o.Customer.LastName,
			&o.Customer.Email,
		)
		if err != nil {
			return nil, err
		}
		rslt = append(rslt, &o)
	}

	return rslt, nil
}

func (m *DBModel) GetAllSubscriptions() ([]*Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
select
    o.amount, o.quantity, w.price, t.currency,
    o.id as order_id, o.widget_id, o.transaction_id,o.customer_id,
    o.created_at,  o.status_id,
    w.name as item, w.description,
    t.last_four, t.expiry_month, t.expiry_year,
    t.payment_intent, t.bank_return_code,
    c.first_name, c.last_name, c.email

from orders o
         left join widgets w on (o.widget_id = w.id)
         left join transactions t on (o.transaction_id = t.id)
         left join customers c on (o.customer_id= c.id)

where
        w.is_recurring = 1
order by
		o.created_at desc
`
	var rslt []*Order
	rows, err := m.DB.QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)
	for rows.Next() {
		var o Order
		err = rows.Scan(
			&o.Amount,
			&o.Quantity,
			&o.Widget.Price,
			&o.Transaction.Currency,
			&o.ID,
			&o.WidgetID,
			&o.TransactionID,
			&o.CustomerID,
			&o.CreatedAt,
			&o.StatusID,
			&o.Widget.Name,
			&o.Widget.Description,
			&o.Transaction.LastFour,
			&o.Transaction.ExpiryMonth,
			&o.Transaction.ExpiryYear,
			&o.Transaction.PaymentIntent,
			&o.Transaction.BankReturnCode,
			&o.Customer.FirstName,
			&o.Customer.LastName,
			&o.Customer.Email,
		)
		if err != nil {
			return nil, err
		}
		rslt = append(rslt, &o)
	}

	return rslt, nil
}
