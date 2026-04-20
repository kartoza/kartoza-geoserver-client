"""PostgreSQL service client.

Provides an object-oriented client for PostgreSQL operations,
wrapping psycopg2 connections via pg_service.conf entries.
"""

from typing import Any

import psycopg2

from apps.core.config import get_config
from apps.core.models import PGService


class PGServiceClient:
    """Client for a PostgreSQL service entry."""

    def __init__(self, service: PGService):
        self.service = service

    def _connect(self) -> Any:
        return psycopg2.connect(
            host=self.service.host,
            port=self.service.port,
            dbname=self.service.dbname,
            user=self.service.user,
            password=self.service.password,
        )

    def test_connection(self) -> tuple[bool, str]:
        try:
            with self._connect() as conn:
                with conn.cursor() as cur:
                    cur.execute("SELECT version()")
                    version = cur.fetchone()[0]
                    return True, f"Connected: {version}"
        except Exception as e:
            return False, str(e)

    def list_schemas(self) -> list[str]:
        with self._connect() as conn:
            with conn.cursor() as cur:
                cur.execute("""
                    SELECT schema_name
                    FROM information_schema.schemata
                    WHERE schema_name NOT IN (
                        'pg_catalog', 'information_schema', 'pg_toast'
                    )
                    ORDER BY schema_name
                """)
                return [row[0] for row in cur.fetchall()]

    def list_tables(self, schema: str = "public") -> list[dict[str, Any]]:
        with self._connect() as conn:
            with conn.cursor() as cur:
                cur.execute("""
                    SELECT
                        t.table_name,
                        t.table_type,
                        gc.f_geometry_column,
                        gc.type,
                        gc.srid
                    FROM information_schema.tables t
                    LEFT JOIN geometry_columns gc
                        ON gc.f_table_schema = t.table_schema
                        AND gc.f_table_name = t.table_name
                    WHERE t.table_schema = %s
                        AND t.table_type IN ('BASE TABLE', 'VIEW')
                    ORDER BY t.table_name
                """, (schema,))
                return [
                    {
                        "name": row[0],
                        "type": row[1],
                        "geometryColumn": row[2],
                        "geometryType": row[3],
                        "srid": row[4],
                        "schema": schema,
                    }
                    for row in cur.fetchall()
                ]

    def get_table_columns(self, schema: str, table: str) -> list[dict[str, Any]]:
        with self._connect() as conn:
            with conn.cursor() as cur:
                cur.execute("""
                    SELECT
                        c.column_name,
                        c.data_type,
                        c.is_nullable,
                        c.column_default,
                        CASE WHEN pk.column_name IS NOT NULL THEN true ELSE false END
                    FROM information_schema.columns c
                    LEFT JOIN (
                        SELECT kcu.column_name
                        FROM information_schema.table_constraints tc
                        JOIN information_schema.key_column_usage kcu
                            ON tc.constraint_name = kcu.constraint_name
                        WHERE tc.table_schema = %s
                            AND tc.table_name = %s
                            AND tc.constraint_type = 'PRIMARY KEY'
                    ) pk ON pk.column_name = c.column_name
                    WHERE c.table_schema = %s AND c.table_name = %s
                    ORDER BY c.ordinal_position
                """, (schema, table, schema, table))

                columns = [
                    {
                        "name": row[0],
                        "dataType": row[1],
                        "isNullable": row[2] == "YES",
                        "default": row[3],
                        "isPrimaryKey": row[4],
                    }
                    for row in cur.fetchall()
                ]

                cur.execute("""
                    SELECT f_geometry_column, type, srid
                    FROM geometry_columns
                    WHERE f_table_schema = %s AND f_table_name = %s
                """, (schema, table))
                geom = cur.fetchone()
                if geom:
                    for col in columns:
                        if col["name"] == geom[0]:
                            col["isGeometry"] = True
                            col["geometryType"] = geom[1]
                            col["srid"] = geom[2]

                return columns

    def get_table_row_count(self, schema: str, table: str) -> int:
        with self._connect() as conn:
            with conn.cursor() as cur:
                cur.execute("""
                    SELECT reltuples::bigint
                    FROM pg_class c
                    JOIN pg_namespace n ON n.oid = c.relnamespace
                    WHERE n.nspname = %s AND c.relname = %s
                """, (schema, table))
                result = cur.fetchone()
                return result[0] if result else 0

    def get_table_data(
        self,
        schema: str,
        table: str,
        limit: int = 100,
        offset: int = 0,
        order_by: str | None = None,
    ) -> dict[str, Any]:
        with self._connect() as conn:
            with conn.cursor() as cur:
                cur.execute(f'SELECT COUNT(*) FROM "{schema}"."{table}"')
                total = cur.fetchone()[0]

                query = f'SELECT * FROM "{schema}"."{table}"'
                if order_by:
                    query += f' ORDER BY "{order_by}"'
                query += f" LIMIT {limit} OFFSET {offset}"
                cur.execute(query)

                columns = [desc[0] for desc in cur.description] if cur.description else []
                rows = []
                for row in cur.fetchall():
                    row_values = []
                    for value in row:
                        if hasattr(value, "isoformat"):
                            value = value.isoformat()
                        elif isinstance(value, bytes):
                            value = f"<binary {len(value)} bytes>"
                        elif value is not None:
                            value = str(value)
                        row_values.append(value)
                    rows.append(row_values)

                return {
                    "columns": columns,
                    "rows": rows,
                    "total": total,
                    "limit": limit,
                    "offset": offset,
                }

    def execute_query(
        self,
        query: str,
        params: tuple | None = None,
        limit: int = 1000,
    ) -> dict[str, Any]:
        with self._connect() as conn:
            with conn.cursor() as cur:
                query_lower = query.lower().strip()
                if "limit" not in query_lower and query_lower.startswith("select"):
                    query = f"{query.rstrip(';')} LIMIT {limit}"

                cur.execute(query, params)
                columns = [desc[0] for desc in cur.description] if cur.description else []
                rows = []
                for row in cur.fetchall():
                    row_dict = {}
                    for i, col in enumerate(columns):
                        value = row[i]
                        if hasattr(value, "isoformat"):
                            value = value.isoformat()
                        elif isinstance(value, bytes):
                            value = value.hex()
                        row_dict[col] = value
                    rows.append(row_dict)

                return {
                    "columns": columns,
                    "rows": rows,
                    "rowCount": len(rows),
                }


def get_pg_client(service_name: str, user_id: str = "default") -> PGServiceClient:
    """Get a PGServiceClient for a named service."""
    service = get_config(user_id).get_pg_service(service_name)
    if not service:
        raise ValueError(f"PostgreSQL service not found: {service_name}")
    return PGServiceClient(service)