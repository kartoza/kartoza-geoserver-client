" ============================================================================
" Kartoza CloudBench - Neovim Project Configuration
" ============================================================================
" This file provides keybindings and commands for development tasks.
" Ensure you have 'set exrc' in your ~/.config/nvim/init.vim
"
" Prerequisites:
"   - Run inside 'nix develop' shell for all tools
"   - Python 3.12+ with Django and dependencies
"   - Node.js for frontend development
"   - delve (dlv) for Go debugging (legacy)
"
" Leader key is assumed to be <Space> or your configured leader
" ============================================================================

" Set local directory for this project
lcd %:p:h

" ============================================================================
" PYTHON/DJANGO WEB SERVER
" ============================================================================

" Start Django development server
command! CbStart :!cd %:p:h && fuser -k 8080/tcp 2>/dev/null; sleep 1; nix develop -c python manage.py runserver 0.0.0.0:8080 &
nnoremap <leader>ps :CbStart<CR>

" Stop web server (kill process on port 8080)
command! CbStop :!fuser -k 8080/tcp 2>/dev/null && echo "Server stopped" || echo "No server running"
nnoremap <leader>pq :CbStop<CR>

" Restart server (stop, build frontend, start)
command! CbRestart :!cd %:p:h && fuser -k 8080/tcp 2>/dev/null; sleep 1; cd web && npm run build && cd .. && nix develop -c python manage.py runserver 0.0.0.0:8080 &
nnoremap <leader>pr :CbRestart<CR>

" Check server status
command! CbStatus :!curl -s http://localhost:8080/health/ >/dev/null 2>&1 && echo "Server is RUNNING on http://localhost:8080" || echo "Server is NOT running"
nnoremap <leader>p? :CbStatus<CR>

" Open app in browser
command! CbOpen :!xdg-open http://localhost:8080 2>/dev/null || open http://localhost:8080
nnoremap <leader>po :CbOpen<CR>

" Run with uvicorn (ASGI)
command! CbUvicorn :terminal cd %:p:h && nix develop -c uvicorn cloudbench.asgi:application --host 0.0.0.0 --port 8080 --reload
nnoremap <leader>pU :CbUvicorn<CR>

" ============================================================================
" TUI (TEXTUAL)
" ============================================================================

" Run the Textual TUI
command! CbTui :terminal cd %:p:h && nix develop -c python -m tui
nnoremap <leader>pT :CbTui<CR>

" ============================================================================
" BUILD COMMANDS
" ============================================================================

" Build frontend only (React/Vite)
command! CbBuildFrontend :!cd %:p:h/web && npm run build
nnoremap <leader>bf :CbBuildFrontend<CR>

" Collect static files
command! CbCollectStatic :!cd %:p:h && nix develop -c python manage.py collectstatic --noinput
nnoremap <leader>bs :CbCollectStatic<CR>

" Build everything (frontend + collect static)
command! CbBuildAll :!cd %:p:h && make build-web
nnoremap <leader>ba :CbBuildAll<CR>

" Clean build artifacts
command! CbClean :!cd %:p:h && make clean
nnoremap <leader>bc :CbClean<CR>

" Full redeploy (clean, build, restart)
command! CbRedeploy :!cd %:p:h && make redeploy
nnoremap <leader>pR :CbRedeploy<CR>

" TypeScript type check (no emit)
command! CbTsc :!cd %:p:h/web && npx tsc --noEmit
nnoremap <leader>bT :CbTsc<CR>

" ============================================================================
" DJANGO MANAGEMENT
" ============================================================================

" Django shell
command! CbShell :terminal cd %:p:h && nix develop -c python manage.py shell
nnoremap <leader>ps :CbShell<CR>

" Run migrations
command! CbMigrate :!cd %:p:h && nix develop -c python manage.py migrate
nnoremap <leader>pm :CbMigrate<CR>

" Make migrations
command! CbMakeMigrations :!cd %:p:h && nix develop -c python manage.py makemigrations
nnoremap <leader>pM :CbMakeMigrations<CR>

" ============================================================================
" TESTING (PYTEST & VITEST)
" ============================================================================

" --- All Tests ---
command! CbTest :!cd %:p:h && nix develop -c pytest tests/ -v
nnoremap <leader>ta :CbTest<CR>

" Run tests for current file
command! CbTestFile :!cd %:p:h && nix develop -c pytest -v %
nnoremap <leader>tf :CbTestFile<CR>

" Run test under cursor
command! CbTestCursor :!cd %:p:h && nix develop -c pytest -v % -k <cword>
nnoremap <leader>tc :CbTestCursor<CR>

" Run tests with coverage (Python + Frontend)
command! CbTestCover :!cd %:p:h && nix develop -c pytest tests/ --cov=apps --cov=cloudbench --cov=tui --cov-report=html && cd web && npm run test:coverage && cd .. && xdg-open htmlcov/index.html
nnoremap <leader>tC :CbTestCover<CR>

" --- Test by Category ---

" Run unit tests only (fast)
command! CbTestUnit :!cd %:p:h && nix develop -c pytest tests/unit -v --tb=short
nnoremap <leader>tu :CbTestUnit<CR>

" Run API tests
command! CbTestApi :!cd %:p:h && nix develop -c pytest tests/api -v --tb=short
nnoremap <leader>tA :CbTestApi<CR>

" Run integration tests (requires services)
command! CbTestIntegration :!cd %:p:h && nix develop -c pytest tests/integration -v --tb=short
nnoremap <leader>ti :CbTestIntegration<CR>

" Run E2E tests (requires browser)
command! CbTestE2E :!cd %:p:h && nix develop -c pytest tests/e2e -v --tb=short
nnoremap <leader>te :CbTestE2E<CR>

" Run TUI tests
command! CbTestTui :!cd %:p:h && nix develop -c pytest tests/tui -v --tb=short
nnoremap <leader>tT :CbTestTui<CR>

" Run frontend tests (Vitest)
command! CbTestFrontend :!cd %:p:h/web && npm run test
nnoremap <leader>tF :CbTestFrontend<CR>

" Run frontend tests in watch mode
command! CbTestFrontendWatch :terminal cd %:p:h/web && npm run test:watch
nnoremap <leader>tW :CbTestFrontendWatch<CR>

" --- Quick Tests ---

" Run quick tests (unit only, for pre-commit)
command! CbTestQuick :!cd %:p:h && nix develop -c pytest tests/unit -v --tb=short -q && cd web && npm run test -- --run --reporter=dot
nnoremap <leader>tq :CbTestQuick<CR>

" --- Pre-commit & CI ---

" Run pre-commit hooks
command! CbPreCommit :!cd %:p:h && nix develop -c pre-commit run --all-files
nnoremap <leader>tp :CbPreCommit<CR>

" Install pre-commit hooks
command! CbPreCommitInstall :!cd %:p:h && nix develop -c pre-commit install
nnoremap <leader>tP :CbPreCommitInstall<CR>

" --- Playwright ---

" Install Playwright browsers
command! CbPlaywrightInstall :!cd %:p:h && nix develop -c playwright install --with-deps chromium
nnoremap <leader>tI :CbPlaywrightInstall<CR>

" --- Coverage Reports ---

" Open Python coverage report
command! CbCoverPython :!xdg-open %:p:h/htmlcov/index.html
nnoremap <leader>tR :CbCoverPython<CR>

" Open frontend coverage report
command! CbCoverFrontend :!xdg-open %:p:h/web/coverage/index.html
nnoremap <leader>tr :CbCoverFrontend<CR>

" ============================================================================
" LINTING & FORMATTING (PYTHON)
" ============================================================================

" Run ruff
command! CbRuff :!cd %:p:h && nix develop -c ruff check apps cloudbench tui
nnoremap <leader>lr :CbRuff<CR>

" Run mypy
command! CbMypy :!cd %:p:h && nix develop -c mypy apps cloudbench tui
nnoremap <leader>lm :CbMypy<CR>

" Run all linters
command! CbLint :!cd %:p:h && make lint
nnoremap <leader>la :CbLint<CR>

" Format with black
command! CbBlack :!cd %:p:h && nix develop -c black apps cloudbench tui
nnoremap <leader>fb :CbBlack<CR>

" Format current file
command! CbBlackFile :!nix develop -c black %
nnoremap <leader>ff :CbBlackFile<CR>

" Run eslint on frontend
command! CbLintFrontend :!cd %:p:h/web && npm run lint 2>/dev/null || npx eslint src/
nnoremap <leader>lf :CbLintFrontend<CR>

" ============================================================================
" DOCKER CONTAINERS (Test Environment)
" ============================================================================

" --- GeoServer ---
command! CbGeoStart :!docker run -d --name kartoza-geoserver-test -p 8600:8080 -e GEOSERVER_ADMIN_USER=admin -e GEOSERVER_ADMIN_PASSWORD=geoserver -e STABLE_EXTENSIONS=wps-plugin kartoza/geoserver:2.26.0 && echo "GeoServer starting on http://localhost:8600/geoserver"
nnoremap <leader>dgs :CbGeoStart<CR>

command! CbGeoStop :!docker stop kartoza-geoserver-test 2>/dev/null; docker rm kartoza-geoserver-test 2>/dev/null && echo "GeoServer stopped"
nnoremap <leader>dgq :CbGeoStop<CR>

command! CbGeoLogs :!docker logs -f kartoza-geoserver-test
nnoremap <leader>dgl :CbGeoLogs<CR>

command! CbGeoStatus :!docker ps --filter name=kartoza-geoserver-test --format "{{.Status}}" | grep -q . && echo "GeoServer: RUNNING (http://localhost:8600/geoserver)" || echo "GeoServer: NOT RUNNING"
nnoremap <leader>dg? :CbGeoStatus<CR>

" --- PostGIS ---
command! CbPgStart :!docker run -d --name kartoza-postgis-test -p 5433:5432 -e POSTGRES_USER=docker -e POSTGRES_PASS=docker -e POSTGRES_DBNAME=gis -e ALLOW_IP_RANGE=0.0.0.0/0 -v kartoza-postgis-data:/var/lib/postgresql kartoza/postgis:16-3.4 && echo "PostGIS starting on localhost:5433"
nnoremap <leader>dps :CbPgStart<CR>

command! CbPgStop :!docker stop kartoza-postgis-test 2>/dev/null; docker rm kartoza-postgis-test 2>/dev/null && echo "PostGIS stopped (data preserved)"
nnoremap <leader>dpq :CbPgStop<CR>

command! CbPgClean :!docker volume rm kartoza-postgis-data 2>/dev/null && echo "PostGIS data volume removed"
nnoremap <leader>dpc :CbPgClean<CR>

command! CbPgLogs :!docker logs -f kartoza-postgis-test
nnoremap <leader>dpl :CbPgLogs<CR>

command! CbPgPsql :terminal docker exec -e PGPASSWORD=docker -it kartoza-postgis-test psql -h localhost -U docker -d gis
nnoremap <leader>dpP :CbPgPsql<CR>

command! CbPgStatus :!docker ps --filter name=kartoza-postgis-test --format "{{.Status}}" | grep -q . && echo "PostGIS: RUNNING (localhost:5433)" || echo "PostGIS: NOT RUNNING"
nnoremap <leader>dp? :CbPgStatus<CR>

" --- MinIO (S3) ---
command! CbMinioStart :!docker run -d --name kartoza-minio-test -p 9000:9000 -p 9001:9001 -e MINIO_ROOT_USER=minioadmin -e MINIO_ROOT_PASSWORD=minioadmin -v kartoza-minio-data:/data minio/minio:latest server /data --console-address ":9001" && echo "MinIO starting - API: http://localhost:9000, Console: http://localhost:9001"
nnoremap <leader>dms :CbMinioStart<CR>

command! CbMinioStop :!docker stop kartoza-minio-test 2>/dev/null; docker rm kartoza-minio-test 2>/dev/null && echo "MinIO stopped (data preserved)"
nnoremap <leader>dmq :CbMinioStop<CR>

command! CbMinioClean :!docker volume rm kartoza-minio-data 2>/dev/null && echo "MinIO data volume removed"
nnoremap <leader>dmc :CbMinioClean<CR>

command! CbMinioLogs :!docker logs -f kartoza-minio-test
nnoremap <leader>dml :CbMinioLogs<CR>

command! CbMinioConsole :!xdg-open http://localhost:9001 2>/dev/null || open http://localhost:9001
nnoremap <leader>dmC :CbMinioConsole<CR>

command! CbMinioStatus :!docker ps --filter name=kartoza-minio-test --format "{{.Status}}" | grep -q . && echo "MinIO: RUNNING (API: localhost:9000, Console: localhost:9001)" || echo "MinIO: NOT RUNNING"
nnoremap <leader>dm? :CbMinioStatus<CR>

" --- All containers ---
command! CbDockerStartAll :CbGeoStart | CbPgStart | CbMinioStart
nnoremap <leader>das :CbDockerStartAll<CR>

command! CbDockerStopAll :CbGeoStop | CbPgStop | CbMinioStop
nnoremap <leader>daq :CbDockerStopAll<CR>

command! CbDockerStatus :CbGeoStatus | CbPgStatus | CbMinioStatus
nnoremap <leader>da? :CbDockerStatus<CR>

" ============================================================================
" DEBUGGING (Python with debugpy)
" ============================================================================

" Debug Django server
command! CbDebugWeb :terminal cd %:p:h && nix develop -c python -m debugpy --listen 5678 manage.py runserver 0.0.0.0:8080 --noreload
nnoremap <leader>Dw :CbDebugWeb<CR>

" Debug TUI
command! CbDebugTui :terminal cd %:p:h && nix develop -c python -m debugpy --listen 5678 -m tui
nnoremap <leader>Dt :CbDebugTui<CR>

" Debug current test file
command! CbDebugTest :terminal cd %:p:h && nix develop -c python -m debugpy --listen 5678 -m pytest -v %
nnoremap <leader>DT :CbDebugTest<CR>

" ============================================================================
" LEGACY GO COMMANDS (for backward compatibility)
" ============================================================================

" Start Go web server
command! CbStartGo :!cd %:p:h && fuser -k 8080/tcp 2>/dev/null; sleep 1; nix develop -c go run ./cmd/web/ serve &
nnoremap <leader>Ps :CbStartGo<CR>

" Build Go web server binary
command! CbBuildWebGo :!cd %:p:h && nix develop -c go build -o bin/kartoza-cloudbench-go ./cmd/web
nnoremap <leader>Bw :CbBuildWebGo<CR>

" Build Go TUI binary
command! CbBuildTuiGo :!cd %:p:h && nix develop -c go build -o bin/kartoza-cloudbench-tui-go .
nnoremap <leader>Bt :CbBuildTuiGo<CR>

" Run Go tests
command! CbTestGo :!cd %:p:h && nix develop -c go test -v ./...
nnoremap <leader>Tg :CbTestGo<CR>

" Run golangci-lint
command! CbLintGo :!cd %:p:h && nix develop -c golangci-lint run
nnoremap <leader>Lg :CbLintGo<CR>

" Debug Go web server
command! CbDebugWebGo :terminal cd %:p:h && nix develop -c dlv debug ./cmd/web/ -- serve
nnoremap <leader>DW :CbDebugWebGo<CR>

" ============================================================================
" GIT SHORTCUTS
" ============================================================================

command! CbGitStatus :!cd %:p:h && git status
nnoremap <leader>gs :CbGitStatus<CR>

command! CbGitDiff :!cd %:p:h && git diff
nnoremap <leader>gd :CbGitDiff<CR>

command! CbGitLog :!cd %:p:h && git log --oneline -20
nnoremap <leader>gl :CbGitLog<CR>

" ============================================================================
" DOCUMENTATION
" ============================================================================

" Serve docs locally
command! CbDocs :terminal cd %:p:h && nix develop -c mkdocs serve
nnoremap <leader>xd :CbDocs<CR>

" Build docs
command! CbDocsBuild :!cd %:p:h && nix develop -c mkdocs build
nnoremap <leader>xD :CbDocsBuild<CR>

" Open SPECIFICATION.md
command! CbSpec :edit %:p:h/SPECIFICATION.md
nnoremap <leader>xs :CbSpec<CR>

" Open CLAUDE.md
command! CbClaude :edit %:p:h/CLAUDE.md
nnoremap <leader>xc :CbClaude<CR>

" ============================================================================
" NAVIGATION SHORTCUTS
" ============================================================================

" Open key directories (Python)
command! CbApps :edit %:p:h/apps/
nnoremap <leader>na :CbApps<CR>

command! CbCloud :edit %:p:h/cloudbench/
nnoremap <leader>nC :CbCloud<CR>

command! CbTuiDir :edit %:p:h/tui/
nnoremap <leader>nt :CbTuiDir<CR>

" Open key directories (Web)
command! CbWeb :edit %:p:h/web/src/
nnoremap <leader>nw :CbWeb<CR>

command! CbComponents :edit %:p:h/web/src/components/
nnoremap <leader>nc :CbComponents<CR>

command! CbStores :edit %:p:h/web/src/stores/
nnoremap <leader>ns :CbStores<CR>

command! CbApiTs :edit %:p:h/web/src/api/
nnoremap <leader>nA :CbApiTs<CR>

" Legacy Go directories
command! CbInternal :edit %:p:h/internal/
nnoremap <leader>ni :CbInternal<CR>

" ============================================================================
" TERMINAL SHORTCUTS
" ============================================================================

" Open terminal in project root
command! CbTerm :terminal cd %:p:h && nix develop
nnoremap <leader>tt :CbTerm<CR>

" Open terminal in web directory
command! CbTermWeb :terminal cd %:p:h/web && nix develop
nnoremap <leader>tw :CbTermWeb<CR>

" ============================================================================
" HELP / INFO
" ============================================================================

command! CbHelp :echo "
\Kartoza CloudBench Keybindings (Leader = <Space>)
\
\PYTHON/DJANGO SERVER:
\  <leader>ps  Start Django server
\  <leader>pq  Stop server
\  <leader>pr  Restart server (rebuild frontend)
\  <leader>p?  Check server status
\  <leader>po  Open in browser
\  <leader>pU  Run with uvicorn (ASGI)
\  <leader>pR  Full redeploy
\  <leader>pT  Run Textual TUI
\
\DJANGO:
\  <leader>pm  Run migrations
\  <leader>pM  Make migrations
\
\BUILD:
\  <leader>bf  Build frontend only
\  <leader>bs  Collect static files
\  <leader>ba  Build all
\  <leader>bc  Clean build artifacts
\  <leader>bT  TypeScript type check
\
\TEST:
\  <leader>ta  Run all tests          <leader>tu  Unit tests only
\  <leader>tA  API tests              <leader>ti  Integration tests
\  <leader>te  E2E tests (browser)    <leader>tT  TUI tests
\  <leader>tF  Frontend tests         <leader>tW  Frontend watch mode
\  <leader>tf  Current file           <leader>tc  Test under cursor
\  <leader>tq  Quick tests            <leader>tC  Tests with coverage
\  <leader>tp  Pre-commit hooks       <leader>tP  Install pre-commit
\  <leader>tI  Install Playwright     <leader>tR  Python coverage
\  <leader>tr  Frontend coverage
\
\LINT/FORMAT:
\  <leader>lr  Run ruff
\  <leader>lm  Run mypy
\  <leader>la  Run all linters
\  <leader>fb  Format with black
\  <leader>ff  Format current file
\  <leader>lf  Run eslint on frontend
\
\DOCKER (Test Containers):
\  <leader>dgs Start GeoServer    <leader>dgq Stop GeoServer
\  <leader>dps Start PostGIS      <leader>dpq Stop PostGIS
\  <leader>dms Start MinIO        <leader>dmq Stop MinIO
\  <leader>das Start all          <leader>daq Stop all
\  <leader>da? Status all
\
\DEBUG (debugpy):
\  <leader>Dw  Debug Django server
\  <leader>Dt  Debug TUI
\  <leader>DT  Debug current test
\
\LEGACY GO:
\  <leader>Ps  Start Go server
\  <leader>Bw  Build Go web
\  <leader>Bt  Build Go TUI
\  <leader>Tg  Run Go tests
\  <leader>Lg  Run golangci-lint
\
\GIT:
\  <leader>gs  Git status
\  <leader>gd  Git diff
\  <leader>gl  Git log
\
\DOCS:
\  <leader>xd  Serve docs
\  <leader>xD  Build docs
\  <leader>xs  Open SPECIFICATION.md
\  <leader>xc  Open CLAUDE.md
\
\NAVIGATE:
\  <leader>na  apps/
\  <leader>nC  cloudbench/
\  <leader>nt  tui/
\  <leader>nw  web/src/
\  <leader>nc  web/src/components/
\  <leader>ni  internal/ (Go)
\
\TERMINAL:
\  <leader>tt  Terminal in project root
\  <leader>tw  Terminal in web/
\"
nnoremap <leader>ph :CbHelp<CR>

" ============================================================================
" AUTO COMMANDS
" ============================================================================

" Auto-format Python files on save
augroup cloudbench_python
  autocmd!
  autocmd BufWritePre *.py silent! !black % 2>/dev/null
augroup END

" Auto-format TypeScript/TSX on save
augroup cloudbench_ts
  autocmd!
  autocmd BufWritePre *.ts,*.tsx silent! !npx prettier --write % 2>/dev/null
augroup END

" Set filetype specific settings
augroup cloudbench_filetypes
  autocmd!
  autocmd FileType python setlocal tabstop=4 shiftwidth=4 expandtab
  autocmd FileType typescript,typescriptreact setlocal tabstop=2 shiftwidth=2 expandtab
  autocmd FileType json setlocal tabstop=2 shiftwidth=2 expandtab
  autocmd FileType go setlocal tabstop=4 shiftwidth=4 noexpandtab
augroup END

echo "Kartoza CloudBench config loaded. Type :CbHelp or <leader>ph for keybindings."
