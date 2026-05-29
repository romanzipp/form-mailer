# Form Mailer

> [!NOTE]
> This project is primarily developed on [Codeberg](https://codeberg.org/romanzipp/form-mailer) and only mirrored to GitHub. Please open issues and pull requests there.

A simple container you can point your **html contact-`<form>`** to that sends a mail to a given recipient. That's it.

## Quick Start

```bash
docker pull ghcr.io/romanzipp/form-mailer:latest

cp .env.example .env

docker run -d -p 8080:8080 --env-file .env ghcr.io/romanzipp/form-mailer:latest
```

Or with docker-compose:

```yaml
services:
  form-mailer:
    build: .
    ports:
      - "8080:8080"
    environment:
      - PORT=8080
      - SMTP_HOST=mail.example.com
      - SMTP_PORT=587
      - SMTP_USER=john@doe.com
      - SMTP_PASSWORD=secret
      - FROM_EMAIL=blog@doe.com
      - RECIPIENT_EMAIL=john@doe.com
      - SUCCESS_URL=https://example.com
    restart: unless-stopped
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
- `FROM_NAME` - From display name
- `BACK_URL` - URL used by the "Go back" link on validation and server error pages
- `VALIDATION_RULES` - Per-field validation rules (see below)

## Validation

Set `VALIDATION_RULES` to validate form fields before sending. Format:

```
VALIDATION_RULES=email:required,email|name:required,min.2|message:max.5000
```

- Fields are separated by `|`
- Field name and rules are separated by `:`
- Multiple rules per field are separated by `,`
- Rule arguments use `.` (e.g. `min.2`, `max.500`)

Supported rules:

- `required` - value must be present and non-empty
- `email` - value must be a valid email address
- `min.N` - value must be at least N characters
- `max.N` - value must be at most N characters

Failed validations respond with `400 Bad Request` and an error message.

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

Note: If form contains an `email` field, it will be set as Reply-To header.

## Endpoints

- `POST /submit` - Handle form submission
- `GET /health` - Health check

## Kubernetes

A Helm chart is published as an OCI artifact on each tag:

```bash
helm install form-mailer oci://ghcr.io/romanzipp/form-mailer-chart --version <tag>
```

Chart sources live in `deploy/helm/form-mailer/`.

## Build from Source

```bash
docker build -t form-mailer .
docker run -d -p 8080:8080 --env-file .env form-mailer
```

Or run locally:

```bash
go run main.go
```
