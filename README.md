# mailms

## NAME

mailms - Mail Microservice

## SYNOPSIS

```shell
nix run github:andrieee44/mailms -- <ADDRESS>
```

```shell
mailms <ADDRESS>
```

### POST /mail

`POST /mail` sends an email and returns `200` on success. Returns `400` for
bad input, or `503` if the mail fails to send. This endpoint has no rate
limiting or timeout, no authentication of its own, and no request size limit.
It's the caller's responsibility to guard against abuse. Requests are also
sent over plain, unencrypted HTTP, so only run mailms on `localhost`, or put
it behind a reverse proxy (e.g. [Caddy](https://caddyserver.com/) or
[nginx](https://nginx.org/)) that terminates TLS.

```haskell
POST /mail :: {
  user :: String;
  pass :: String;
  host :: String;
  port :: String;
  from :: String;
  to :: [ String ];
  subject :: String;
  headers :: [ { key :: String; value :: String; } ];
  body :: String;
} -> Int
```

## DESCRIPTION

mailms is a minimal mail microservice: it just sends mail. There's no rate
limiting and no configuration beyond what you send per request. You decide the
headers and content yourself. It's designed as a companion to web databases
like [SurrealDB](https://surrealdb.com/), which can make outbound HTTP
requests but need something else to actually deliver mail. mailms fills that
gap.

Sending mail correctly is still the caller's responsibility, but mailms does
some basic validation to catch common mistakes using
[go-playground/validator](https://github.com/go-playground/validator).
For example, rejecting `\r\n` in both headers and subjects and checking that
`to` and `from` are valid email addresses. Authentication is handled with Go's
standard [`net/smtp`](https://pkg.go.dev/net/smtp) package, which supports
plain auth, making it easy to use with things like
[Gmail App Passwords](https://myaccount.google.com/apppasswords).

## EXAMPLES

```shell
nix run github:andrieee44/mailms -- localhost:8080
```

```shell
mailms localhost:8080
```

### Gmail App Password

1. Go to [Google App Passwords](https://myaccount.google.com/apppasswords).
1. Create an app password which is a 16-character code (remove the spaces
  once generated).
1. Assuming mailms is running on `localhost:8080`, your address is
  `me@example.com`, and your app password is `aaaa bbbb cccc dddd`, here's how
  to send yourself an email:

```shell
curl -X POST http://localhost:8080/mail --json '{
  "user": "me@example.com",
  "pass": "aaaabbbbccccdddd",
  "host": "smtp.gmail.com",
  "port": "587",
  "from": "me@example.com",
  "to": [ "me@example.com" ],
  "subject": "こんにちは from mailms",
  "headers": [
    {
      "key": "Content-Type",
      "value": "text/html; charset=UTF-8"
    }
  ],
  "body": "<h1>こんにちは</h1><p>This email was sent through mailms.</p>"
}'
```

## COPYRIGHT

See [`LICENSE`](./LICENSE). Uses
[AGPLv3 or later](https://www.gnu.org/licenses/agpl-3.0.html).

## SEE ALSO

- [GitHub repository](https://github.com/andrieee44/mailms)
- [Nix flakes](https://wiki.nixos.org/wiki/Flakes)
- [Nix](https://nixos.org/)
- [SurrealDB](https://surrealdb.com/)
- [go-playground/validator](https://github.com/go-playground/validator)
- [net/smtp documentation](https://pkg.go.dev/net/smtp)
