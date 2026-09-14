package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"mime"
	"net"
	"net/http"
	"net/smtp"
	"os"
	"strings"

	"github.com/go-playground/validator/v10"
)

type Header struct {
	Key   string `json:"key"   validate:"required,excludesall=\r\n"`
	Value string `json:"value" validate:"required,excludesall=\r\n"`
}

var validate *validator.Validate = validator.New(
	validator.WithRequiredStructEnabled(),
	validator.WithTagNameFuncBlankOmit(),
)

func mailHandler(w http.ResponseWriter, r *http.Request) {
	type payload struct {
		User    string   `json:"user"    validate:"required"`
		Pass    string   `json:"pass"    validate:"required"`
		Host    string   `json:"host"    validate:"required"`
		Port    string   `json:"port"    validate:"required"`
		From    string   `json:"from"    validate:"required,email"`
		To      []string `json:"to"      validate:"min=1,dive,email"`
		Subject string   `json:"subject" validate:"required,excludesall=\r\n"`
		Headers []Header `json:"headers" validate:"min=1,dive"`
		Body    string   `json:"body"    validate:"required"`
	}

	var (
		decoder *json.Decoder
		data    payload
		builder strings.Builder
		header  Header
		err     error
	)

	decoder = json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	err = decoder.Decode(&data)
	if err != nil {
		slog.Error(r.URL.String(), "method", r.Method, "error", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)

		return
	}

	err = validate.Struct(data)
	if err != nil {
		slog.Error(r.URL.String(), "method", r.Method, "error", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)

		return
	}

	fmt.Fprintf(
		&builder,
		"Subject: %s\r\n",
		mime.BEncoding.Encode("UTF-8", data.Subject),
	)

	builder.WriteString("MIME-version: 1.0\r\n")

	for _, header = range data.Headers {
		builder.WriteString(header.Key)
		builder.WriteString(": ")
		builder.WriteString(header.Value)
		builder.WriteString("\r\n")
	}

	builder.WriteString("\r\n")
	builder.WriteString(data.Body)
	builder.WriteString("\r\n")

	err = smtp.SendMail(
		net.JoinHostPort(data.Host, data.Port),
		smtp.PlainAuth("", data.User, data.Pass, data.Host),
		data.From,
		data.To,
		[]byte(builder.String()),
	)
	if err != nil {
		slog.Error(r.URL.String(), "method", r.Method, "error", err)
		http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)

		return
	}

	w.WriteHeader(http.StatusOK)
}

func run() error {
	var (
		address string
		mux     *http.ServeMux
		srv     *http.Server
		err     error
	)

	if len(os.Args) < 2 {
		return errors.New("missing argument <ADDRESS>")
	}

	address = os.Args[1]

	mux = http.NewServeMux()
	mux.HandleFunc("POST /mail", mailHandler)

	srv = &http.Server{
		Addr:    address,
		Handler: mux,
	}

	srv.Protocols = new(http.Protocols)
	srv.Protocols.SetHTTP1(true)
	srv.Protocols.SetHTTP2(true)
	srv.Protocols.SetUnencryptedHTTP2(true)

	slog.Info("Mail Microservice Starting", "address", address)

	err = srv.ListenAndServe()
	if err != nil {
		return err
	}

	return nil
}

func main() {
	var err error

	err = run()
	if err != nil {
		fmt.Fprintf(os.Stderr, `mailms: %v

Usage:   mailms <ADDRESS>
Example: mailms localhost:8080
`, err)

		os.Exit(1)
	}
}
