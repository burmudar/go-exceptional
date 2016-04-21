package errord

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net/smtp"
	"time"
)

type Notifier interface {
	Fire(n *ErrorNotification) error
}

type ErrorNotification struct {
	ErrorEvent *ErrorEvent
	DaySummary *DaySummary
	Stats      *StatItem
}

type EmailNotifier struct {
	host     string
	from     string
	password string
	to       string
	store    NotifyStore
}

type ConsoleNotifier struct {
	store NotifyStore
}

type DatabaseNotifier struct {
	host     string
	username string
	password string
	to       string
	store    NotifyStore
}

func NewDatabaseNotifier(host, username, password, to string, store NotifyStore) Notifier {
	n := new(DatabaseNotifier)
	n.host = host
	n.username = username
	n.password = password
	n.to = to
	n.store = store
	return n
}

func NewEmailNotifier(host, from, pass, to string, store NotifyStore) Notifier {
	n := new(EmailNotifier)
	n.host = host
	n.from = from
	n.to = to
	n.password = pass
	n.store = store
	return n
}

func NewConsoleNotifier(store NotifyStore) Notifier {
	c := new(ConsoleNotifier)
	c.store = store
	return c
}

func (c *ConsoleNotifier) Fire(n *ErrorNotification) error {
	if c.store.HasNotification(n.ErrorEvent) {
		log.Printf("Notification already sent for %v\n", n.ErrorEvent)
		return nil
	}

	subject, body := n.describe()
	fmt.Printf("\n*** NOTIFICATION ***\nTime: %v\nSubject: %v\nBody: %v\n\n*** END OF NOTIFICATION ***\n", time.Now(), subject, body)
	c.store.UpdateNotificationSent(n.ErrorEvent)
	return nil
}

func (n *EmailNotifier) Fire(notification *ErrorNotification) error {
	if n.store.HasNotification(notification.ErrorEvent) {
		log.Printf("Notification already sent for %v\n", notification.ErrorEvent)
		return nil
	}

	subject, body := notification.describe()
	msg := fmt.Sprintf("From: %v\r\nTo: %v\r\nSubject: %v\r\n\r\n%v\r\n", n.from, n.to, subject, body)
	if err := smtp.SendMail(n.host+":587", smtp.PlainAuth("", n.from, n.password, n.host), n.from, []string{n.to}, []byte(msg)); err != nil {
		return err
	} else {
		n.store.UpdateNotificationSent(notification.ErrorEvent)
		return nil
	}
}

func (d *DatabaseNotifier) Fire(n *ErrorNotification) error {
	db, err := sql.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v:3306)/", d.username, d.password, d.host))
	if err != nil {
		return err
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		return err
	}
	subject, body := n.describe()

	_, err = db.Query("call common.sp_create_email_request(?, ?, ?, ?, ?, ?, ?, ?)", d.to, "", "", subject, body, "N", "test", "")
	if err != nil {
		return err
	}
	d.store.UpdateNotificationSent(n.ErrorEvent)
	return nil
}

func (n *ErrorNotification) isNewError() bool {
	return n.DaySummary == nil && n.Stats == nil
}

func (n *ErrorNotification) describe() (title string, description string) {
	subject := ""
	body := ""
	if n.isNewError() {
		err := n.ErrorEvent
		subject = fmt.Sprintf("New Error: %v", err.Exception)
		body = fmt.Sprintf("New Error Event: [%v] - [%v] : [%v]\nCaused by: [%v] - [%v]\n", err.Timestamp, err.Source, err.Description, err.Exception, err.Detail)
	} else {
		err := n.ErrorEvent
		subject = fmt.Sprintf("[%v] exceeds Statistical Limit: %v", err.Exception, n.Stats.StdDevMax())
		body = fmt.Sprintf("Error Event: [%v] - [%v] : [%v]\nCaused by: [%v] - [%v]\nSeen today = %v\nMax = %v", err.Timestamp, err.Source, err.Description, err.Exception, err.Detail, n.DaySummary.Total, n.Stats.StdDevMax())
	}
	return subject, body
}
