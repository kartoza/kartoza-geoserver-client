# Configuration

CloudBench can be configured through environment variables and configuration files.

## Environment Variables

### Django Settings

| Variable | Description | Default |
|----------|-------------|---------|
| `SECRET_KEY` | Django secret key | Auto-generated |
| `DEBUG` | Enable debug mode | `true` |
| `ALLOWED_HOSTS` | Comma-separated allowed hosts | `localhost,127.0.0.1` |

### Database

| Variable | Description | Default |
|----------|-------------|---------|
| `DATABASE_URL` | Database connection URL | SQLite |

Example:
```bash
export DATABASE_URL=postgres://user:password@localhost:5432/cloudbench
```

### Upload Settings

| Variable | Description | Default |
|----------|-------------|---------|
| `UPLOAD_MAX_FILE_SIZE` | Maximum upload size in bytes | `10737418240` (10GB) |

## PostgreSQL Services

PostgreSQL connections are configured via the standard `~/.pg_service.conf` file:

```ini
[production]
host=db.example.com
port=5432
dbname=geodata
user=geouser
password=secret

[development]
host=localhost
port=5432
dbname=geodata_dev
user=postgres
password=postgres
```

These services appear in the CloudBench sidebar under "PostgreSQL Services".

## GeoServer Connections

GeoServer connections are stored in the database and can be managed through:

1. **Web UI**: Click the + button next to "GeoServer Connections"
2. **Admin Panel**: Visit `/admin/` to manage connections
3. **API**: POST to `/api/connections`

## S3 Storage

S3/MinIO connections can be added through the Web UI:

1. Click + next to "S3 Storage"
2. Enter:
   - **Name**: Connection name
   - **Endpoint**: S3 endpoint URL
   - **Access Key**: AWS access key
   - **Secret Key**: AWS secret key
   - **Region**: AWS region (optional)

## Development Settings

For development, create a `.env` file in the project root:

```bash
DEBUG=true
SECRET_KEY=dev-secret-key-change-in-production
ALLOWED_HOSTS=localhost,127.0.0.1
```

## Production Settings

For production deployments:

```bash
DEBUG=false
SECRET_KEY=your-secure-random-key
ALLOWED_HOSTS=your-domain.com
DATABASE_URL=postgres://user:pass@db:5432/cloudbench
```

See [Deployment Guide](../admin-guide/deployment.md) for more details.
