"""Integration tests for PostgreSQL workflows.

Tests complex interactions with PostgreSQL services.
These tests are marked as skipped until the underlying functionality is implemented.
"""

import pytest
from rest_framework.test import APIClient


@pytest.mark.integration
class TestPostgresServiceWorkflow:
    """Test PostgreSQL service management workflows."""

    @pytest.mark.skip(reason="PostgreSQL service API not yet implemented")
    def test_list_pg_services(self, api_client: APIClient) -> None:
        """Test listing PostgreSQL services from pg_service.conf."""
        pass

    @pytest.mark.skip(reason="PostgreSQL service test API not yet implemented")
    def test_pg_service_connection_test(self, api_client: APIClient) -> None:
        """Test testing a PostgreSQL service connection."""
        pass


@pytest.mark.integration
class TestPostgresSchemaWorkflow:
    """Test PostgreSQL schema browsing workflows."""

    @pytest.mark.skip(reason="PostgreSQL schema API not yet implemented")
    def test_list_schemas(self, api_client: APIClient) -> None:
        """Test listing schemas in a database."""
        pass

    @pytest.mark.skip(reason="PostgreSQL tables API not yet implemented")
    def test_list_tables(self, api_client: APIClient) -> None:
        """Test listing tables in a schema."""
        pass


@pytest.mark.integration
class TestPostgresQueryWorkflow:
    """Test PostgreSQL query execution workflows."""

    @pytest.mark.skip(reason="PostgreSQL query API not yet implemented")
    def test_execute_query(self, api_client: APIClient) -> None:
        """Test executing a SQL query."""
        pass

    @pytest.mark.skip(reason="PostgreSQL query API not yet implemented")
    def test_execute_invalid_query(self, api_client: APIClient) -> None:
        """Test executing an invalid SQL query."""
        pass


@pytest.mark.integration
class TestPostgresImportWorkflow:
    """Test data import to PostgreSQL workflows."""

    @pytest.mark.skip(reason="GeoJSON import not yet implemented")
    def test_import_geojson(self, api_client: APIClient) -> None:
        """Test importing GeoJSON to PostgreSQL."""
        pass

    @pytest.mark.skip(reason="Shapefile import not yet implemented")
    def test_import_shapefile(self, api_client: APIClient) -> None:
        """Test importing Shapefile to PostgreSQL via ogr2ogr."""
        pass


@pytest.mark.integration
class TestPostgresBridgeWorkflow:
    """Test PostgreSQL to GeoServer bridge workflows."""

    @pytest.mark.skip(reason="Bridge API not yet implemented")
    def test_create_postgis_store(self, api_client: APIClient) -> None:
        """Test creating a PostGIS store in GeoServer from PG service."""
        pass

    @pytest.mark.skip(reason="Layer publishing not yet implemented")
    def test_publish_layer_from_table(self, api_client: APIClient) -> None:
        """Test publishing a GeoServer layer from a PostgreSQL table."""
        pass
