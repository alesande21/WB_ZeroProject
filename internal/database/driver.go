package database

import (
	"WB_ZeroProject/internal/colorAttribute"
	"context"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"time"
)

const (
	DefaultMaxConnAttemp     = 10
	DefaultConnTimeout       = time.Second
	DefaultConnBackoffFactor = 2
)

type DBConnection struct {
	Conn              *sql.DB
	connMax           int
	connTimeout       time.Duration
	connBackoffFactor int
}

func (db *DBConnection) Close() error {
	if db == nil {
		return nil
	}

	err := db.Conn.Close()
	if err != nil {
		return err
	}

	return nil
}

func (db *DBConnection) Ping() error {
	if db == nil {
		return nil
	}

	err := db.Conn.Ping()
	if err != nil {
		return fmt.Errorf("problems connecting to the database: %s", err.Error())
	}

	return nil
}

func Open(cfg *DBConfig) (*DBConnection, error) {
	psqlInfo := cfg.GetConfigInfo()

	log.Println(cfg.URL)
	log.Println(psqlInfo)
	db, err := sql.Open(cfg.Driver, cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("driver not found, %s", cfg.Driver)
	}

	log.Println(colorAttribute.ColorString(colorAttribute.FgYellow, "Успешное подключение к базе данных!"))

	return &DBConnection{Conn: db, connMax: DefaultMaxConnAttemp, connTimeout: DefaultConnTimeout,
		connBackoffFactor: DefaultConnBackoffFactor}, nil
}

// GetConn2 TODO: удалить если не нужен
func (db *DBConnection) GetConn2() (*sql.DB, error) {
	if db.Conn == nil {
		return nil, fmt.Errorf("драйвер подключения отсутсвует")
	}
	return db.Conn, nil
}

func (db *DBConnection) GetConn() *sql.DB {
	if db.Conn == nil {
		log.Printf("драйвер подключения отсутсвует")
		return nil
	}
	return db.Conn
}

func (db *DBConnection) CheckConn(ctx context.Context, cfg *DBConfig, updateCache chan interface{}) {
	var err error
	attempt := 0
	for attempt < db.connMax {
		select {
		case <-ctx.Done():
			log.Printf("Проверка соединения остановлена...")
			return
		default:
			err = db.Conn.Ping()
			if err != nil {
				log.Printf("Потеряно соединение с базой данных. Попытка восстановления (%d/%d)", attempt+1, db.connMax)

				var newDb *sql.DB
				newDb, err = sql.Open(cfg.Driver, cfg.URL)
				if err != nil {
					log.Printf("Не удалось подключиться к базе данных. Попытка %d/%d", attempt+1, db.connMax)
					attempt++
				} else {
					log.Println("Соединение с базой данных успешно восстановлено!")
					db.Conn = newDb
					updateCache <- struct{}{}
					attempt = 0
				}
			}

			backoff := db.connTimeout * time.Duration(attempt+1) * time.Duration(db.connBackoffFactor)
			sleepInterval := 10 * time.Microsecond
			elapsedTime := time.Duration(0)

			for elapsedTime < backoff {
				select {
				case <-ctx.Done():
					log.Printf("Проверка соединения остановлена...")
					return
				default:
					time.Sleep(sleepInterval)
					elapsedTime += sleepInterval * 100
				}
			}

		}

	}

	if attempt == db.connMax {
		log.Println("Все попытки подключения исчерпаны. Соединение не восстановлено.")
	}
}

func (db *DBConnection) InterapterConn() {
	for {
		time.Sleep(time.Second * 30)
		db.Conn.Close()
		log.Println("СОЕДИНЕНИЕ РАЗОРВАНО")
	}
}
