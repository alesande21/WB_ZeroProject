package database

import (
	"WB_ZeroProject/internal/colorAttribute"
	"context"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	log2 "github.com/sirupsen/logrus"
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
		return fmt.Errorf("-> db.Conn.Close: ошибка при закрытии подключения к базе данных: %w", err)
	}

	return nil
}

func (db *DBConnection) Ping() error {
	if db == nil {
		return nil
	}

	err := db.Conn.Ping()
	if err != nil {
		return fmt.Errorf("-> db.Conn.Ping: проблемы с подключением к базе данных: %w", err)
	}

	return nil
}

func Open(cfg *DBConfig) (*DBConnection, error) {
	db, err := sql.Open(cfg.Driver, cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("-> sql.Open: ошибка при открытии базы данных: %w", err)
	}

	log2.Info(colorAttribute.ColorString(colorAttribute.FgYellow, "Успешное подключение к базе данных!"))

	return &DBConnection{Conn: db, connMax: DefaultMaxConnAttemp, connTimeout: DefaultConnTimeout,
		connBackoffFactor: DefaultConnBackoffFactor}, nil
}

// GetConn2 TODO: удалить если не нужен
func (db *DBConnection) GetConn2() (*sql.DB, error) {
	if db.Conn == nil {
		return nil, fmt.Errorf(": подключение к базе данных отсуствует")
	}
	return db.Conn, nil
}

func (db *DBConnection) GetConn() *sql.DB {
	if db.Conn == nil {
		log2.Warn("GetConn: подключение к базе данных отсуствует")
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
			log2.Info("CheckConn: Проверка соединения остановлена...")
			return

		default:
			err = db.Conn.Ping()
			if err != nil {
				log2.Warnf("CheckConn-> db.Conn.Ping: потеряно соединение с базой данных. "+
					"Попытка восстановления (%d/%d): %s", attempt+1, db.connMax, err.Error())

				var newDb *sql.DB
				newDb, err = sql.Open(cfg.Driver, cfg.URL)
				if err != nil {
					log2.Warnf("CheckConn-> sql.Open: не удалось подключиться к базе данных. "+
						"Попытка %d/%d: %s", attempt+1, db.connMax, err.Error())
					attempt++
				} else {
					log2.Info("Соединение с базой данных успешно восстановлена!")
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
					log2.Info("Проверка соединения остановлена...")
					return
				default:
					time.Sleep(sleepInterval)
					elapsedTime += sleepInterval * 100
				}
			}
		}
	}

	if attempt == db.connMax {
		log2.Warn("Все попытки подключения к базе данных исчерпаны. Соединение не восстановлено.")
	}
}

func (db *DBConnection) InterapterConn() {
	for {
		time.Sleep(time.Second * 30)
		db.Conn.Close()
		log2.Warnf("СОЕДИНЕНИЕ РАЗОРВАНО")
	}
}
