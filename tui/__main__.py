#!/usr/bin/env python
"""Entry point for the Kartoza CloudBench TUI."""

import click

from .app import CloudBenchApp


@click.command()
@click.option(
    "--config",
    "-c",
    type=click.Path(),
    help="Path to config file (default: ~/.config/kartoza-cloudbench/config.json)",
)
@click.option(
    "--debug",
    "-d",
    is_flag=True,
    help="Enable debug mode with verbose logging",
)
@click.version_option(version="0.3.0", prog_name="kartoza-cloudbench-tui")
def main(config: str | None, debug: bool):
    """Kartoza CloudBench - Geospatial Infrastructure Management TUI.

    A beautiful terminal interface for managing GeoServer, PostgreSQL/PostGIS,
    S3 storage, and cloud-native geospatial data.

    Made with love by Kartoza | https://kartoza.com
    """
    app = CloudBenchApp(config_path=config, debug=debug)
    app.run()


if __name__ == "__main__":
    main()
