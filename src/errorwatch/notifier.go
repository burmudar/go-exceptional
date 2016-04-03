package errorwatch

import (
	"fmt"
	"log"
	"net/smtp"
	"time"
)

type Notifier interface {
	Fire(n *ErrorNotification) error
}

type ErrorNotification struct {
	ErrorEvent *ErrorEvent
	Summary    *Summary
	Stats      *StatItem
}

type EmailNotifier struct {
	from     string
	password string
	to       string
	store    NotifyStore
}

type ConsoleNotifier struct {
	store NotifyStore
}

func NewEmailNotifier(from, pass, to string, store NotifyStore) Notifier {
	n := new(EmailNotifier)
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
	msg := fmt.Sprintf("From: %v\nTo: %v\nSubject: %v\n\n%v", n.from, n.to, subject, body)
	if err := smtp.SendMail("smtp.gmail.com:587", smtp.PlainAuth("", n.from, n.password, "smtp.gmail.com"), n.from, []string{n.to}, []byte(msg)); err != nil {
		return err
	} else {
		n.store.UpdateNotificationSent(notification.ErrorEvent)
		return nil
	}
}

func (n *ErrorNotification) isNewError() bool {
	return n.Summary == nil && n.Stats == nil
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
		body = fmt.Sprintf("Error Event: [%v] - [%v] : [%v]\nCaused by: [%v] - [%v]\nSeen today = %v\nMax = %v", err.Timestamp, err.Source, err.Description, err.Exception, err.Detail, n.Summary.Total, n.Stats.StdDevMax())
	}
	return subject, body
}
