{
  description = "Kartoza CloudBench - Unified management for GeoServer and PostgreSQL/PostGIS";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
    }:
    let
      version = "0.3.0";

      # Overlay that can be imported by other flakes
      overlay = final: prev: {
        # Web frontend built with Nix
        kartoza-cloudbench-web-frontend = final.buildNpmPackage {
          pname = "kartoza-cloudbench-web-frontend";
          inherit version;
          src = "${self}/web";

          npmDepsHash = "sha256-eLUsdm63gn13DF6oP1HloDQUwOh1OuY7wYfGn1QCh0Y=";

          buildPhase = ''
            npm run build
          '';

          installPhase = ''
            mkdir -p $out
            cp -r dist/* $out/ 2>/dev/null || cp -r ../internal/webserver/static/* $out/
          '';

          meta = with final.lib; {
            description = "Web frontend for Kartoza CloudBench";
            license = licenses.mit;
          };
        };

        # Python web server package
        kartoza-cloudbench-web = final.python312Packages.buildPythonApplication {
          pname = "kartoza-cloudbench-web";
          inherit version;
          format = "pyproject";
          src = self;

          nativeBuildInputs = with final.python312Packages; [
            hatchling
          ];

          propagatedBuildInputs = with final.python312Packages; [
            django
            djangorestframework
            django-cors-headers
            whitenoise
            httpx
            psycopg2
            boto3
            duckdb
            lxml
            pydantic
            uvicorn
            gunicorn
          ];

          postInstall = ''
            mkdir -p $out/share/cloudbench/static
            cp -r ${final.kartoza-cloudbench-web-frontend}/* $out/share/cloudbench/static/ || true
          '';

          meta = with final.lib; {
            description = "Web interface for Kartoza CloudBench (Django)";
            homepage = "https://github.com/kartoza/kartoza-cloudbench";
            license = licenses.mit;
            mainProgram = "cloudbench";
          };
        };

        # Python TUI package
        kartoza-cloudbench-tui = final.python312Packages.buildPythonApplication {
          pname = "kartoza-cloudbench-tui";
          inherit version;
          format = "pyproject";
          src = self;

          nativeBuildInputs = with final.python312Packages; [
            hatchling
          ];

          propagatedBuildInputs = with final.python312Packages; [
            textual
            rich
            click
            httpx
            pydantic
          ];

          meta = with final.lib; {
            description = "Terminal UI for Kartoza CloudBench (Textual)";
            homepage = "https://github.com/kartoza/kartoza-cloudbench";
            license = licenses.mit;
            mainProgram = "cloudbench-tui";
          };
        };

        # Legacy Go TUI application (kept for backward compatibility)
        kartoza-cloudbench-go = final.buildGoModule {
          pname = "kartoza-cloudbench-go";
          inherit version;
          src = self;

          vendorHash = null;

          # Skip tests that require a running GeoServer
          doCheck = false;

          nativeBuildInputs = [ final.makeWrapper ];

          ldflags = [
            "-s"
            "-w"
            "-X main.version=${version}"
          ];

          # Wrap binary to include gdal, postgis, and duckdb in PATH
          postInstall = ''
            wrapProgram $out/bin/kartoza-cloudbench \
              --prefix PATH : ${
                final.lib.makeBinPath [
                  final.gdal
                  final.postgresqlPackages.postgis
                  final.duckdb
                ]
              }
          '';

          meta = with final.lib; {
            description = "Unified management TUI for GeoServer and PostgreSQL/PostGIS (Go - Legacy)";
            homepage = "https://github.com/kartoza/kartoza-cloudbench";
            license = licenses.mit;
            mainProgram = "kartoza-cloudbench";
          };
        };

        # Legacy Go Web server (kept for backward compatibility)
        kartoza-cloudbench-web-go = final.buildGoModule {
          pname = "kartoza-cloudbench-web-go";
          inherit version;
          src = self;

          vendorHash = null;

          doCheck = false;

          nativeBuildInputs = [ final.makeWrapper ];

          preBuild = ''
            mkdir -p internal/webserver/static
            cp -r ${final.kartoza-cloudbench-web-frontend}/* internal/webserver/static/ || true
          '';

          subPackages = [ "cmd/web" ];

          ldflags = [
            "-s"
            "-w"
            "-X main.version=${version}"
          ];

          postInstall = ''
            wrapProgram $out/bin/web \
              --prefix PATH : ${
                final.lib.makeBinPath [
                  final.gdal
                  final.postgresqlPackages.postgis
                  final.duckdb
                ]
              }
          '';

          meta = with final.lib; {
            description = "Web interface for Kartoza CloudBench (Go - Legacy)";
            homepage = "https://github.com/kartoza/kartoza-cloudbench";
            license = licenses.mit;
            mainProgram = "web";
          };
        };
      };
    in
    {
      # Export overlay for use in other flakes
      overlays.default = overlay;
      overlays.kartoza-cloudbench = overlay;

    }
    // flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs {
          inherit system;
          overlays = [ overlay ];
        };

        # Python with all dependencies for development
        pythonEnv = pkgs.python312.withPackages (
          ps: with ps; [
            # Django and web
            django
            djangorestframework
            django-cors-headers
            whitenoise
            httpx
            psycopg2
            boto3
            duckdb
            lxml
            pydantic
            uvicorn
            gunicorn
            # TUI
            textual
            rich
            click
            # Development tools
            pytest
            pytest-django
            pytest-asyncio
            pytest-cov
            black
            mypy
            # Type stubs
            types-requests
            django-stubs
          ]
        );

        # MkDocs with Material theme for documentation
        mkdocsEnv = pkgs.python3.withPackages (
          ps: with ps; [
            mkdocs
            mkdocs-material
            mkdocs-minify-plugin
            pygments
            pymdown-extensions
          ]
        );

        # Test GeoServer configuration
        geoserverContainer = "kartoza-geoserver-test";
        geoserverPort = "8600";
        geoserverUser = "admin";
        geoserverPass = "geoserver";

        # Test PostGIS configuration
        postgisContainer = "kartoza-postgis-test";
        postgisPort = "5433";
        postgisUser = "docker";
        postgisPass = "docker";
        postgisDb = "gis";

        # Test MinIO (S3-compatible) configuration
        minioContainer = "kartoza-minio-test";
        minioPort = "9000";
        minioConsolePort = "9001";
        minioRootUser = "minioadmin";
        minioRootPassword = "minioadmin";
        minioDefaultBucket = "geospatial-data";

      in
      {
        packages = {
          default = pkgs.kartoza-cloudbench-web;
          web = pkgs.kartoza-cloudbench-web;
          tui = pkgs.kartoza-cloudbench-tui;
          frontend = pkgs.kartoza-cloudbench-web-frontend;
          # Legacy Go packages
          go-tui = pkgs.kartoza-cloudbench-go;
          go-web = pkgs.kartoza-cloudbench-web-go;
        };

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            # Python environment with all dependencies
            pythonEnv
            ruff  # Fast Python linter

            # Node.js for web frontend (from Nix)
            nodejs_22
            nodePackages.typescript
            nodePackages.typescript-language-server

            # Build tools
            gnumake
            gcc
            pkg-config

            # CLI utilities
            ripgrep
            fd
            eza
            bat
            fzf
            tree
            jq
            yq

            # GIS tools for data import and cloud-native conversion
            gdal # Provides ogr2ogr for vector data import and COG conversion
            pdal # Provides COPC (Cloud Optimized Point Cloud) conversion
            postgresqlPackages.postgis # Provides raster2pgsql for raster data import
            duckdb # Analytical database for querying Parquet/GeoParquet files

            # Documentation
            mkdocsEnv

            # Nix tools
            nil
            nixpkgs-fmt
            nixfmt-classic

            # Git
            git
            gh

            # Docker for test environment
            docker

            # Legacy Go toolchain (for migration period)
            go
            gopls
            golangci-lint
            delve
          ];

          shellHook = ''
            export EDITOR=nvim
            export PYTHONPATH="$PWD:$PYTHONPATH"
            export DJANGO_SETTINGS_MODULE=cloudbench.settings.development

            # Test GeoServer configuration
            export GEOSERVER_CONTAINER="${geoserverContainer}"
            export GEOSERVER_PORT="${geoserverPort}"
            export GEOSERVER_USER="${geoserverUser}"
            export GEOSERVER_PASS="${geoserverPass}"
            export GEOSERVER_URL="http://localhost:${geoserverPort}/geoserver"

            # Test PostGIS configuration
            export POSTGIS_CONTAINER="${postgisContainer}"
            export POSTGIS_PORT="${postgisPort}"
            export POSTGIS_USER="${postgisUser}"
            export POSTGIS_PASS="${postgisPass}"
            export POSTGIS_DB="${postgisDb}"

            # Test MinIO (S3-compatible) configuration
            export MINIO_CONTAINER="${minioContainer}"
            export MINIO_PORT="${minioPort}"
            export MINIO_CONSOLE_PORT="${minioConsolePort}"
            export MINIO_ROOT_USER="${minioRootUser}"
            export MINIO_ROOT_PASSWORD="${minioRootPassword}"
            export MINIO_DEFAULT_BUCKET="${minioDefaultBucket}"
            export MINIO_ENDPOINT="http://localhost:${minioPort}"

            # Legacy Go paths (for migration period)
            export GOPATH="$PWD/.go"
            export GOCACHE="$PWD/.go/cache"
            export GOMODCACHE="$PWD/.go/pkg/mod"
            export PATH="$GOPATH/bin:$PATH"

            # Helpful aliases
            alias runserver='python manage.py runserver 0.0.0.0:8080'
            alias runtui='python -m tui'
            alias makemigrations='python manage.py makemigrations'
            alias migrate='python manage.py migrate'
            alias shell='python manage.py shell'

            # Documentation aliases
            alias docs='mkdocs serve'
            alias docs-build='mkdocs build'

            # GeoServer test environment commands
            geoserver-start() {
              echo "Starting test GeoServer on port ${geoserverPort}..."
              docker run -d \
                --name ${geoserverContainer} \
                -p ${geoserverPort}:8080 \
                -e GEOSERVER_ADMIN_USER=${geoserverUser} \
                -e GEOSERVER_ADMIN_PASSWORD=${geoserverPass} \
                -e STABLE_EXTENSIONS=wps-plugin \
                kartoza/geoserver:2.26.0
              echo ""
              echo "GeoServer starting at: http://localhost:${geoserverPort}/geoserver"
              echo "Credentials: ${geoserverUser} / ${geoserverPass}"
            }
            export -f geoserver-start

            geoserver-stop() {
              echo "Stopping test GeoServer..."
              docker stop ${geoserverContainer} 2>/dev/null || true
              docker rm ${geoserverContainer} 2>/dev/null || true
              echo "GeoServer stopped."
            }
            export -f geoserver-stop

            geoserver-status() {
              if docker ps --format '{{.Names}}' | grep -q "^${geoserverContainer}$"; then
                echo "GeoServer is running"
                echo "URL: http://localhost:${geoserverPort}/geoserver"
              else
                echo "GeoServer is not running"
              fi
            }
            export -f geoserver-status

            # PostGIS test environment commands
            postgis-start() {
              echo "Starting test PostGIS on port ${postgisPort}..."
              docker run -d \
                --name ${postgisContainer} \
                -p ${postgisPort}:5432 \
                -e POSTGRES_USER=${postgisUser} \
                -e POSTGRES_PASS=${postgisPass} \
                -e POSTGRES_DBNAME=${postgisDb} \
                -e ALLOW_IP_RANGE=0.0.0.0/0 \
                -v kartoza-postgis-data:/var/lib/postgresql \
                kartoza/postgis:16-3.4
              echo "PostGIS starting at: localhost:${postgisPort}"
            }
            export -f postgis-start

            postgis-stop() {
              echo "Stopping test PostGIS..."
              docker stop ${postgisContainer} 2>/dev/null || true
              docker rm ${postgisContainer} 2>/dev/null || true
              echo "PostGIS stopped."
            }
            export -f postgis-stop

            # MinIO test environment commands
            minio-start() {
              echo "Starting test MinIO..."
              docker run -d \
                --name ${minioContainer} \
                -p ${minioPort}:9000 \
                -p ${minioConsolePort}:9001 \
                -e MINIO_ROOT_USER=${minioRootUser} \
                -e MINIO_ROOT_PASSWORD=${minioRootPassword} \
                -v kartoza-minio-data:/data \
                minio/minio:latest server /data --console-address ":9001"
              echo "MinIO starting - API: http://localhost:${minioPort}, Console: http://localhost:${minioConsolePort}"
            }
            export -f minio-start

            minio-stop() {
              echo "Stopping test MinIO..."
              docker stop ${minioContainer} 2>/dev/null || true
              docker rm ${minioContainer} 2>/dev/null || true
              echo "MinIO stopped."
            }
            export -f minio-stop

            # Install npm dependencies if needed
            if [ -d "web" ] && [ ! -d "web/node_modules" ]; then
              echo "📦 Installing npm dependencies..."
              (cd web && npm install --silent)
            fi

            echo ""
            echo "🌍 Kartoza CloudBench Development Environment (Python/Django)"
            echo ""
            echo "Quick commands:"
            echo "  runserver      - Run Django development server (port 8080)"
            echo "  runtui         - Run Textual TUI"
            echo "  make dev-web   - Run Django development server"
            echo "  make dev-tui   - Run Textual TUI"
            echo "  make test      - Run pytest"
            echo "  make lint      - Run ruff and mypy"
            echo ""
            echo "Test containers:"
            echo "  geoserver-start/stop/status"
            echo "  postgis-start/stop"
            echo "  minio-start/stop"
            echo ""
            echo "Documentation:"
            echo "  docs           - Serve docs locally (http://localhost:8000)"
            echo "  docs-build     - Build static docs site"
            echo ""
          '';
        };

        apps = {
          default = {
            type = "app";
            program = "${self.packages.${system}.web}/bin/cloudbench";
          };

          web = {
            type = "app";
            program = "${self.packages.${system}.web}/bin/cloudbench";
          };

          tui = {
            type = "app";
            program = "${self.packages.${system}.tui}/bin/cloudbench-tui";
          };

          # Legacy Go apps
          go-web = {
            type = "app";
            program = "${self.packages.${system}.go-web}/bin/web";
          };

          go-tui = {
            type = "app";
            program = "${self.packages.${system}.go-tui}/bin/kartoza-cloudbench";
          };
        };
      }
    );
}
