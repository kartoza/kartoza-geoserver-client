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

from apps.core.config import get_cache_dir, get_config
from apps.core.models import PGService

from .client import get_pg_client


# === PostgreSQL Services ===


class PGServiceListView(APIView):
    """List and create PostgreSQL services."""

    def get(self, request):
        """List all PostgreSQL services."""
        config = get_config(request.user.id)
        return Response([
            {
                "name": svc.name,
                "host": svc.host,
                "port": svc.port,
                "dbname": svc.dbname,
                "user": svc.user,
                "sslmode": svc.sslmode,
            }
            for svc in config.list_pg_services()
        ])

    def post(self, request):
        """Create a new PostgreSQL service."""
        name = request.data.get("name")
        if not name:
            return Response(
                {"error": "name is required"}, status=status.HTTP_400_BAD_REQUEST
            )

        config = get_config(request.user.id)
        if config.get_pg_service(name):
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
        config.add_pg_service(service)

        return Response(
            {"name": service.name, "host": service.host, "port": service.port,
             "dbname": service.dbname, "user": service.user},
            status=status.HTTP_201_CREATED,
        )


class PGServiceDetailView(APIView):
    """Get, update, or delete a PostgreSQL service."""

    def get(self, request, name):
        """Get service details."""
        service = get_config(request.user.id).get_pg_service(name)
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
        config = get_config(request.user.id)
        service = config.get_pg_service(name)
        if not service:
            return Response(
                {"error": "Service not found"}, status=status.HTTP_404_NOT_FOUND
            )
        data = request.data
        updated = service.model_copy(update={
            k: (int(data[k]) if k == "port" else data[k])
            for k in ("host", "port", "dbname", "user", "password", "sslmode")
            if k in data
        })
        config.update_pg_service(updated)
        return Response({
            "name": updated.name,
            "host": updated.host,
            "port": updated.port,
            "dbname": updated.dbname,
            "user": updated.user,
        })

    def delete(self, request, name):
        """Delete a service."""
        if not get_config(request.user.id).delete_pg_service(name):
            return Response(
                {"error": "Service not found"}, status=status.HTTP_404_NOT_FOUND
            )
        return Response(status=status.HTTP_204_NO_CONTENT)


class PGServiceTestView(APIView):
    """Test a PostgreSQL service connection."""

    def post(self, request, name):
        """Test connection to a service."""
        if not get_config(request.user.id).get_pg_service(name):
            return Response(
                {"error": "Service not found"}, status=status.HTTP_404_NOT_FOUND
            )
        success, message = get_pg_client(name, str(request.user.id)).test_connection()
        return Response({"success": success, "message": message})


# === Schema Browsing ===


class PGSchemaListView(APIView):
    """List schemas in a database."""

    def get(self, request, service_name):
        """List all schemas."""
        try:
            schemas = get_pg_client(service_name).list_schemas()
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
            tables = get_pg_client(service_name).list_tables(schema_name)
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
            client = get_pg_client(service_name)
            columns = client.get_table_columns(schema_name, table_name)
            row_count = client.get_table_row_count(schema_name, table_name)
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

            data = get_pg_client(service_name).get_table_data(
                schema_name, table_name, limit, offset, order_by
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
            result = get_pg_client(service_name).execute_query(query, limit=limit)
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
        service = get_config(request.user.id).get_pg_service(service_name)
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
        service = get_config(request.user.id).get_pg_service(service_name)
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
