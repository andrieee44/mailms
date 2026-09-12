package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
	Key   string `json:"key"   validate:"required,no_crlf"`
	Value string `json:"value" validate:"required,no_crlf"`
}

var validate *validator.Validate

func decodeAndValidate(r io.Reader, v any) error {
	var (
		decoder *json.Decoder
		err     error
	)

	decoder = json.NewDecoder(r)
	decoder.DisallowUnknownFields()

	err = decoder.Decode(v)
	if err != nil {
		return err
	}

	err = validate.Struct(v)
	if err != nil {
		return err
	}

	return nil
}

func mailHandler(w http.ResponseWriter, r *http.Request) {
	type payload struct {
		User    string   `json:"user"    validate:"required"`
		Pass    string   `json:"pass"    validate:"required"`
		Host    string   `json:"host"    validate:"required"`
		Port    string   `json:"port"    validate:"required"`
		From    string   `json:"from"    validate:"required,email"`
		To      []string `json:"to"      validate:"min=1,dive,email"`
		Subject string   `json:"subject" validate:"required,no_crlf"`
		Headers []Header `json:"headers" validate:"min=1,dive"`
		Body    string   `json:"body"    validate:"required"`
	}

	var (
		data    payload
		builder strings.Builder
		header  Header
		err     error
	)

	err = decodeAndValidate(r.Body, &data)
	if err != nil {
		slog.Error("POST /mail", "error", err)
		http.Error(w, "invalid JSON", http.StatusBadRequest)

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
		slog.Error("POST /mail", "error", err)
		http.Error(w, "failed to send mail", http.StatusServiceUnavailable)

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

	if len(os.Args) == 1 {
		return errors.New("missing argument <ADDRESS>")
	}

	address = os.Args[1]

	validate = validator.New(
		validator.WithRequiredStructEnabled(),
		validator.WithTagNameFuncBlankOmit(),
	)

	err = validate.RegisterValidation(
		"no_crlf",
		func(fl validator.FieldLevel) bool {
			return !strings.ContainsAny(fl.Field().String(), "\r\n")
		},
	)
	if err != nil {
		return err
	}

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
