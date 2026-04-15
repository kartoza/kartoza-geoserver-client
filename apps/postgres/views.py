"""Views for PostgreSQL/PostGIS integration.

Provides endpoints for:
- Managing pg_service.conf entries
- Schema browsing
- Table data viewing
- Query execution
- Data import via ogr2ogr
"""

import subprocess
import tempfile
import uuid
from pathlib import Path

from rest_framework import status
from rest_framework.response import Response
from rest_framework.views import APIView

from apps.core.config import get_config, get_cache_dir

from . import schema
from .service import (
    PGService,
    delete_service,
    get_service,
    list_services,
    write_service,
)


# === PostgreSQL Services ===


class PGServiceListView(APIView):
    """List and create PostgreSQL services."""

    def get(self, request):
        """List all PostgreSQL services from pg_service.conf."""
        services = list_services()

        # Include state from config
        config_manager = get_config(request.user.id)
        config = config_manager.config
        state_map = {s.name: s.is_parsed for s in config.pg_services}

        result = []
        for name in services:
            svc = get_service(name)
            if svc:
                result.append({
                    "name": svc.name,
                    "host": svc.host,
                    "port": svc.port,
                    "dbname": svc.dbname,
                    "user": svc.user,
                    "isParsed": state_map.get(name, False),
                })

        return Response(result)

    def post(self, request):
        """Create a new PostgreSQL service.

        Expected body:
        {
            "name": "my_service",
            "host": "localhost",
            "port": 5432,
            "dbname": "mydb",
            "user": "myuser",
            "password": "mypassword",
            "sslmode": "prefer"
        }
        """
        name = request.data.get("name")
        if not name:
            return Response(
                {"error": "name is required"}, status=status.HTTP_400_BAD_REQUEST
            )

        # Check if service already exists
        if get_service(name):
            return Response(
                {"error": f"Service '{name}' already exists"},
                status=status.HTTP_409_CONFLICT,
            )

        service = PGService(
            name=name,
            host=request.data.get("host", "localhost"),
            port=int(request.data.get("port", 5432)),
            dbname=request.data.get("dbname", ""),
            user=request.data.get("user", ""),
            password=request.data.get("password", ""),
            sslmode=request.data.get("sslmode", ""),
        )

        write_service(service)

        return Response(
            {
                "name": service.name,
                "host": service.host,
                "port": service.port,
                "dbname": service.dbname,
                "user": service.user,
            },
            status=status.HTTP_201_CREATED,
        )


class PGServiceDetailView(APIView):
    """Get, update, or delete a PostgreSQL service."""

    def get(self, request, name):
        """Get service details."""
        service = get_service(name)
        if not service:
            return Response(
                {"error": "Service not found"}, status=status.HTTP_404_NOT_FOUND
            )

        return Response({
            "name": service.name,
            "host": service.host,
            "port": service.port,
            "dbname": service.dbname,
            "user": service.user,
            "sslmode": service.sslmode,
            "connectionString": service.connection_string(),
        })

    def put(self, request, name):
        """Update a service."""
        service = get_service(name)
        if not service:
            return Response(
                {"error": "Service not found"}, status=status.HTTP_404_NOT_FOUND
            )

        # Update fields
        if "host" in request.data:
            service.host = request.data["host"]
        if "port" in request.data:
            service.port = int(request.data["port"])
        if "dbname" in request.data:
            service.dbname = request.data["dbname"]
        if "user" in request.data:
            service.user = request.data["user"]
        if "password" in request.data:
            service.password = request.data["password"]
        if "sslmode" in request.data:
            service.sslmode = request.data["sslmode"]

        write_service(service)

        return Response({
            "name": service.name,
            "host": service.host,
            "port": service.port,
            "dbname": service.dbname,
            "user": service.user,
        })

    def delete(self, request, name):
        """Delete a service."""
        if not delete_service(name):
            return Response(
                {"error": "Service not found"}, status=status.HTTP_404_NOT_FOUND
            )

        return Response(status=status.HTTP_204_NO_CONTENT)


class PGServiceTestView(APIView):
    """Test a PostgreSQL service connection."""

    def post(self, request, name):
        """Test connection to a service."""
        service = get_service(name)
        if not service:
            return Response(
                {"error": "Service not found"}, status=status.HTTP_404_NOT_FOUND
            )

        success, message = schema.test_connection(name)

        return Response({
            "success": success,
            "message": message,
        })


# === Schema Browsing ===


class PGSchemaListView(APIView):
    """List schemas in a database."""

    def get(self, request, service_name):
        """List all schemas."""
        try:
            schemas = schema.list_schemas(service_name)
            return Response({"schemas": schemas})
        except ValueError as e:
            return Response({"error": str(e)}, status=status.HTTP_404_NOT_FOUND)
        except Exception as e:
            return Response(
                {"error": f"Database error: {str(e)}"},
                status=status.HTTP_502_BAD_GATEWAY,
            )


class PGTableListView(APIView):
    """List tables in a schema."""

    def get(self, request, service_name, schema_name):
        """List all tables in a schema."""
        try:
            tables = schema.list_tables(service_name, schema_name)
            return Response({"tables": tables})
        except ValueError as e:
            return Response({"error": str(e)}, status=status.HTTP_404_NOT_FOUND)
        except Exception as e:
            return Response(
                {"error": f"Database error: {str(e)}"},
                status=status.HTTP_502_BAD_GATEWAY,
            )


class PGTableDetailView(APIView):
    """Get table details including columns."""

    def get(self, request, service_name, schema_name, table_name):
        """Get table columns and metadata."""
        try:
            columns = schema.get_table_columns(service_name, schema_name, table_name)
            row_count = schema.get_table_row_count(service_name, schema_name, table_name)

            return Response({
                "schema": schema_name,
                "table": table_name,
                "columns": columns,
                "rowCount": row_count,
            })
        except ValueError as e:
            return Response({"error": str(e)}, status=status.HTTP_404_NOT_FOUND)
        except Exception as e:
            return Response(
                {"error": f"Database error: {str(e)}"},
                status=status.HTTP_502_BAD_GATEWAY,
            )


class PGTableDataView(APIView):
    """Get table data with pagination."""

    def get(self, request, service_name, schema_name, table_name):
        """Get paginated table data.

        Query params:
        - limit: Number of rows (default 100)
        - offset: Number of rows to skip (default 0)
        - orderBy: Column to order by
        """
        try:
            limit = int(request.query_params.get("limit", 100))
            offset = int(request.query_params.get("offset", 0))
            order_by = request.query_params.get("orderBy")

            data = schema.get_table_data(
                service_name, schema_name, table_name, limit, offset, order_by
            )

            return Response(data)
        except ValueError as e:
            return Response({"error": str(e)}, status=status.HTTP_404_NOT_FOUND)
        except Exception as e:
            return Response(
                {"error": f"Database error: {str(e)}"},
                status=status.HTTP_502_BAD_GATEWAY,
            )


# === Query Execution ===


class PGQueryView(APIView):
    """Execute SQL queries."""

    def post(self, request, service_name):
        """Execute a SQL query.

        Expected body:
        {
            "query": "SELECT * FROM my_table",
            "limit": 1000
        }
        """
        query = request.data.get("query")
        if not query:
            return Response(
                {"error": "query is required"}, status=status.HTTP_400_BAD_REQUEST
            )

        limit = request.data.get("limit", 1000)

        try:
            result = schema.execute_query(service_name, query, limit=limit)
            return Response(result)
        except ValueError as e:
            return Response({"error": str(e)}, status=status.HTTP_404_NOT_FOUND)
        except Exception as e:
            return Response(
                {"error": f"Query error: {str(e)}"},
                status=status.HTTP_400_BAD_REQUEST,
            )


# === Data Import ===

# Track import jobs
_import_jobs: dict[str, dict] = {}


class PGImportView(APIView):
    """Import data to PostgreSQL via ogr2ogr."""

    def post(self, request):
        """Start a data import job.

        Expected body:
        {
            "serviceName": "my_service",
            "filePath": "/path/to/file.gpkg",
            "schema": "public",
            "tableName": "my_table",
            "srid": 4326,
            "overwrite": false
        }
        """
        service_name = request.data.get("serviceName")
        file_path = request.data.get("filePath")
        target_schema = request.data.get("schema", "public")
        table_name = request.data.get("tableName")
        srid = request.data.get("srid")
        overwrite = request.data.get("overwrite", False)

        if not service_name or not file_path:
            return Response(
                {"error": "serviceName and filePath are required"},
                status=status.HTTP_400_BAD_REQUEST,
            )

        # Get service info
        service = get_service(service_name)
        if not service:
            return Response(
                {"error": f"Service not found: {service_name}"},
                status=status.HTTP_404_NOT_FOUND,
            )

        # Check file exists
        if not Path(file_path).exists():
            return Response(
                {"error": f"File not found: {file_path}"},
                status=status.HTTP_400_BAD_REQUEST,
            )

        # Generate job ID
        job_id = str(uuid.uuid4())

        # Build ogr2ogr command
        pg_conn = f"PG:host={service.host} port={service.port} dbname={service.dbname} user={service.user} password={service.password}"

        cmd = [
            "ogr2ogr",
            "-f", "PostgreSQL",
            pg_conn,
            file_path,
        ]

        if overwrite:
            cmd.append("-overwrite")
        else:
            cmd.append("-append")

        if table_name:
            cmd.extend(["-nln", f"{target_schema}.{table_name}"])

        if srid:
            cmd.extend(["-t_srs", f"EPSG:{srid}"])

        # Start job in background
        _import_jobs[job_id] = {
            "id": job_id,
            "status": "running",
            "file": file_path,
            "service": service_name,
            "output": "",
            "error": "",
        }

        try:
            result = subprocess.run(
                cmd,
                capture_output=True,
                text=True,
                timeout=600,  # 10 minute timeout
            )

            if result.returncode == 0:
                _import_jobs[job_id]["status"] = "completed"
                _import_jobs[job_id]["output"] = result.stdout
            else:
                _import_jobs[job_id]["status"] = "failed"
                _import_jobs[job_id]["error"] = result.stderr

        except subprocess.TimeoutExpired:
            _import_jobs[job_id]["status"] = "timeout"
            _import_jobs[job_id]["error"] = "Import timed out after 10 minutes"
        except FileNotFoundError:
            _import_jobs[job_id]["status"] = "failed"
            _import_jobs[job_id]["error"] = "ogr2ogr not found. Please install GDAL."
        except Exception as e:
            _import_jobs[job_id]["status"] = "failed"
            _import_jobs[job_id]["error"] = str(e)

        return Response(
            {"jobId": job_id, "status": _import_jobs[job_id]["status"]},
            status=status.HTTP_202_ACCEPTED,
        )


class PGImportStatusView(APIView):
    """Get import job status."""

    def get(self, request, job_id):
        """Get status of an import job."""
        if job_id not in _import_jobs:
            return Response(
                {"error": "Job not found"}, status=status.HTTP_404_NOT_FOUND
            )

        return Response(_import_jobs[job_id])


class PGImportRasterView(APIView):
    """Import raster data to PostgreSQL via raster2pgsql."""

    def post(self, request):
        """Start a raster import job.

        Expected body:
        {
            "serviceName": "my_service",
            "filePath": "/path/to/raster.tif",
            "schema": "public",
            "tableName": "my_raster",
            "srid": 4326,
            "tileSize": "100x100"
        }
        """
        service_name = request.data.get("serviceName")
        file_path = request.data.get("filePath")
        target_schema = request.data.get("schema", "public")
        table_name = request.data.get("tableName")
        srid = request.data.get("srid", 4326)
        tile_size = request.data.get("tileSize", "100x100")

        if not service_name or not file_path or not table_name:
            return Response(
                {"error": "serviceName, filePath, and tableName are required"},
                status=status.HTTP_400_BAD_REQUEST,
            )

        # Get service info
        service = get_service(service_name)
        if not service:
            return Response(
                {"error": f"Service not found: {service_name}"},
                status=status.HTTP_404_NOT_FOUND,
            )

        # Check file exists
        if not Path(file_path).exists():
            return Response(
                {"error": f"File not found: {file_path}"},
                status=status.HTTP_400_BAD_REQUEST,
            )

        # Generate job ID
        job_id = str(uuid.uuid4())

        # Build raster2pgsql command
        full_table = f"{target_schema}.{table_name}"

        cmd = [
            "raster2pgsql",
            "-s", str(srid),
            "-I",  # Create spatial index
            "-C",  # Apply constraints
            "-M",  # Vacuum analyze
            "-t", tile_size,
            file_path,
            full_table,
        ]

        _import_jobs[job_id] = {
            "id": job_id,
            "status": "running",
            "file": file_path,
            "service": service_name,
            "output": "",
            "error": "",
        }

        try:
            # raster2pgsql outputs SQL, pipe to psql
            raster_result = subprocess.run(
                cmd,
                capture_output=True,
                text=True,
                timeout=600,
            )

            if raster_result.returncode != 0:
                _import_jobs[job_id]["status"] = "failed"
                _import_jobs[job_id]["error"] = raster_result.stderr
                return Response(
                    {"jobId": job_id, "status": "failed"},
                    status=status.HTTP_202_ACCEPTED,
                )

            # Pipe SQL to psql
            psql_cmd = [
                "psql",
                "-h", service.host,
                "-p", str(service.port),
                "-U", service.user,
                "-d", service.dbname,
            ]

            psql_result = subprocess.run(
                psql_cmd,
                input=raster_result.stdout,
                capture_output=True,
                text=True,
                timeout=600,
                env={**subprocess.os.environ, "PGPASSWORD": service.password},
            )

            if psql_result.returncode == 0:
                _import_jobs[job_id]["status"] = "completed"
                _import_jobs[job_id]["output"] = psql_result.stdout
            else:
                _import_jobs[job_id]["status"] = "failed"
                _import_jobs[job_id]["error"] = psql_result.stderr

        except subprocess.TimeoutExpired:
            _import_jobs[job_id]["status"] = "timeout"
            _import_jobs[job_id]["error"] = "Import timed out"
        except FileNotFoundError as e:
            _import_jobs[job_id]["status"] = "failed"
            _import_jobs[job_id]["error"] = f"Tool not found: {e.filename}"
        except Exception as e:
            _import_jobs[job_id]["status"] = "failed"
            _import_jobs[job_id]["error"] = str(e)

        return Response(
            {"jobId": job_id, "status": _import_jobs[job_id]["status"]},
            status=status.HTTP_202_ACCEPTED,
        )


class PGDetectLayersView(APIView):
    """Detect layers in a file for import."""

    def post(self, request):
        """Detect layers in a geospatial file.

        Expected body:
        {
            "filePath": "/path/to/file.gpkg"
        }
        """
        file_path = request.data.get("filePath")
        if not file_path:
            return Response(
                {"error": "filePath is required"}, status=status.HTTP_400_BAD_REQUEST
            )

        if not Path(file_path).exists():
            return Response(
                {"error": f"File not found: {file_path}"},
                status=status.HTTP_400_BAD_REQUEST,
            )

        try:
            # Use ogrinfo to list layers
            result = subprocess.run(
                ["ogrinfo", "-so", "-al", file_path],
                capture_output=True,
                text=True,
                timeout=30,
            )

            if result.returncode != 0:
                return Response(
                    {"error": f"ogrinfo failed: {result.stderr}"},
                    status=status.HTTP_400_BAD_REQUEST,
                )

            # Parse output to extract layer names
            layers = []
            for line in result.stdout.split("\n"):
                if line.startswith("Layer name:"):
                    layer_name = line.split(":", 1)[1].strip()
                    layers.append(layer_name)

            return Response({"layers": layers, "file": file_path})

        except FileNotFoundError:
            return Response(
                {"error": "ogrinfo not found. Please install GDAL."},
                status=status.HTTP_500_INTERNAL_SERVER_ERROR,
            )
        except Exception as e:
            return Response(
                {"error": str(e)}, status=status.HTTP_500_INTERNAL_SERVER_ERROR
            )
