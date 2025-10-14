# Form Mailer

Go service that receives form submissions and sends emails.

## Quick Start

```bash
docker pull ghcr.io/romanzipp/form-mailer:latest
cp .env.example .env
# Edit .env
docker run -d -p 8080:8080 --env-file .env ghcr.io/romanzipp/form-mailer:latest
```

Or with docker-compose:

```bash
cp .env.example .env
# Edit .env
docker-compose up -d
```

## Environment Variables

Required:
- `SMTP_HOST` - SMTP server
- `SMTP_USER` - SMTP username
- `SMTP_PASSWORD` - SMTP password
- `FROM_EMAIL` - From address
- `RECIPIENT_EMAIL` - Where to send submissions
- `SUCCESS_URL` - Redirect URL after submission

Optional:
- `PORT` - Server port (default: 8080)
- `SMTP_PORT` - SMTP port (default: 587)

## Gmail Setup

1. Enable 2FA
2. Generate [App Password](https://myaccount.google.com/apppasswords)
3. Use app password for `SMTP_PASSWORD`

## HTML Form

```html
<form action="http://localhost:8080/submit" method="POST">
  <input type="text" name="name" required>
  <input type="email" name="email" required>
  <textarea name="message" required></textarea>
  <button type="submit">Submit</button>
</form>
```

See `example.html` for styled version.

## Endpoints

- `POST /submit` - Handle form submission
- `GET /health` - Health check

## Build from Source

```bash
docker build -t form-mailer .
docker run -d -p 8080:8080 --env-file .env form-mailer
```

Or run locally:

```bash
go run main.go
```
