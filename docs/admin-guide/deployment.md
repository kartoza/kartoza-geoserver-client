# Deployment

This guide covers deploying CloudBench in production environments.

## Requirements

- Python 3.12+
- PostgreSQL 14+ (recommended for production)
- Reverse proxy (nginx, Caddy, etc.)

## Production Setup

### 1. Clone and Install

```bash
git clone https://github.com/kartoza/kartoza-cloudbench.git
cd kartoza-cloudbench
pip install -e .
```

### 2. Configure Environment

Create `/etc/cloudbench/env`:

```bash
DEBUG=false
SECRET_KEY=your-secure-random-key-here
ALLOWED_HOSTS=cloudbench.example.com
DATABASE_URL=postgres://user:pass@localhost:5432/cloudbench
```

### 3. Database Setup

```bash
# Create database
createdb cloudbench

# Run migrations
python manage.py migrate

# Create admin user
python manage.py createsuperuser
```

### 4. Collect Static Files

```bash
python manage.py collectstatic --noinput
```

### 5. Run with Gunicorn

```bash
gunicorn cloudbench.wsgi:application \
    --bind 0.0.0.0:8000 \
    --workers 4 \
    --threads 2
```

### 6. Nginx Configuration

```nginx
server {
    listen 80;
    server_name cloudbench.example.com;

    location /static/ {
        alias /path/to/cloudbench/staticfiles/;
    }

    location / {
        proxy_pass http://127.0.0.1:8000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## Systemd Service

Create `/etc/systemd/system/cloudbench.service`:

```ini
[Unit]
Description=CloudBench Web Application
After=network.target postgresql.service

[Service]
User=cloudbench
Group=cloudbench
WorkingDirectory=/opt/cloudbench
EnvironmentFile=/etc/cloudbench/env
ExecStart=/opt/cloudbench/venv/bin/gunicorn cloudbench.wsgi:application --bind 127.0.0.1:8000 --workers 4
Restart=always

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
systemctl enable cloudbench
systemctl start cloudbench
```

## Security Checklist

- [ ] Set `DEBUG=false`
- [ ] Use strong `SECRET_KEY`
- [ ] Configure HTTPS
- [ ] Set proper `ALLOWED_HOSTS`
- [ ] Use PostgreSQL (not SQLite)
- [ ] Configure firewall
- [ ] Set up log rotation
