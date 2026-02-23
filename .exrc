" ============================================================================
" Kartoza CloudBench - Neovim Project Configuration
" ============================================================================
" This file provides keybindings and commands for development tasks.
" Ensure you have 'set exrc' in your ~/.config/nvim/init.vim
"
" Prerequisites:
"   - Run inside 'nix develop' shell for all tools
"   - delve (dlv) for Go debugging
"   - DAP (nvim-dap) plugin for full debugging UI (optional)
"
" Leader key is assumed to be <Space> or your configured leader
" ============================================================================

" Set local directory for this project
lcd %:p:h

" ============================================================================
" WHICH-KEY MENU SETUP (if you have which-key.nvim)
" ============================================================================
" These commands can be displayed nicely if you have which-key
" All project commands are under <leader>p (project)

" ============================================================================
" SERVER MANAGEMENT
" ============================================================================

" Start web server in background (with nix develop environment)
command! CbStart :!cd %:p:h && fuser -k 8080/tcp 2>/dev/null; sleep 1; nix develop -c go run ./cmd/web/ serve &
nnoremap <leader>ps :CbStart<CR>

" Stop web server (kill process on port 8080)
command! CbStop :!fuser -k 8080/tcp 2>/dev/null && echo "Server stopped" || echo "No server running"
nnoremap <leader>pq :CbStop<CR>

" Restart server (stop, build frontend, start)
command! CbRestart :!cd %:p:h && fuser -k 8080/tcp 2>/dev/null; sleep 1; cd web && npm run build && cd .. && nix develop -c go run ./cmd/web/ serve &
nnoremap <leader>pr :CbRestart<CR>

" Check server status
command! CbStatus :!curl -s http://localhost:8080/api/health >/dev/null 2>&1 && echo "Server is RUNNING on http://localhost:8080" || echo "Server is NOT running"
nnoremap <leader>p? :CbStatus<CR>

" Open app in browser
command! CbOpen :!xdg-open http://localhost:8080 2>/dev/null || open http://localhost:8080
nnoremap <leader>po :CbOpen<CR>

" ============================================================================
" BUILD COMMANDS
" ============================================================================

" Build frontend only (React/Vite)
command! CbBuildFrontend :!cd %:p:h/web && npm run build
nnoremap <leader>bf :CbBuildFrontend<CR>

" Build Go web server binary
command! CbBuildWeb :!cd %:p:h && nix develop -c go build -o bin/kartoza-cloudbench ./cmd/web
nnoremap <leader>bw :CbBuildWeb<CR>

" Build TUI binary
command! CbBuildTui :!cd %:p:h && nix develop -c go build -o bin/kartoza-cloudbench-client .
nnoremap <leader>bt :CbBuildTui<CR>

" Build everything (frontend + both binaries)
command! CbBuildAll :!cd %:p:h && make build
nnoremap <leader>ba :CbBuildAll<CR>

" Clean build artifacts
command! CbClean :!cd %:p:h && make clean
nnoremap <leader>bc :CbClean<CR>

" Full redeploy (clean, build, restart)
command! CbRedeploy :!cd %:p:h && make kill-server && make clean && cd web && npm install && npm run build && cd .. && nix develop -c go run ./cmd/web/ serve &
nnoremap <leader>pR :CbRedeploy<CR>

" TypeScript type check (no emit)
command! CbTsc :!cd %:p:h/web && npx tsc --noEmit
nnoremap <leader>bT :CbTsc<CR>

" ============================================================================
" TESTING
" ============================================================================

" Run all Go tests
command! CbTest :!cd %:p:h && nix develop -c go test -v ./...
nnoremap <leader>ta :CbTest<CR>

" Run tests for current package (Go)
command! CbTestPkg :!cd %:p:h && nix develop -c go test -v ./%:h
nnoremap <leader>tp :CbTestPkg<CR>

" Run test under cursor (requires vim-test or similar, fallback to go test)
command! CbTestCursor :!cd %:p:h && nix develop -c go test -v -run <cword> ./%:h
nnoremap <leader>tc :CbTestCursor<CR>

" Run tests with coverage
command! CbTestCover :!cd %:p:h && nix develop -c go test -v -cover ./... && nix develop -c go tool cover -html=coverage.out
nnoremap <leader>tC :CbTestCover<CR>

" ============================================================================
" LINTING
" ============================================================================

" Run golangci-lint
command! CbLint :!cd %:p:h && nix develop -c golangci-lint run
nnoremap <leader>lg :CbLint<CR>

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
" DEBUGGING (Go with Delve)
" ============================================================================

" Debug web server (headless, connect with DAP or dlv connect)
command! CbDebugWeb :terminal cd %:p:h && nix develop -c dlv debug ./cmd/web/ -- serve
nnoremap <leader>Dw :CbDebugWeb<CR>

" Debug TUI
command! CbDebugTui :terminal cd %:p:h && nix develop -c dlv debug . --
nnoremap <leader>Dt :CbDebugTui<CR>

" Debug current test file
command! CbDebugTest :terminal cd %:p:h && nix develop -c dlv test ./%:h
nnoremap <leader>DT :CbDebugTest<CR>

" Debug with DAP (headless mode for nvim-dap connection)
command! CbDebugWebDap :!cd %:p:h && nix develop -c dlv debug ./cmd/web/ --headless --listen=:2345 --api-version=2 -- serve &
nnoremap <leader>DD :CbDebugWebDap<CR>

" Attach to running headless delve session
command! CbDebugAttach :terminal dlv connect :2345
nnoremap <leader>Da :CbDebugAttach<CR>

" ============================================================================
" NVIM-DAP CONFIGURATION (if you have nvim-dap installed)
" ============================================================================
" Add this to your nvim-dap config for Go debugging:
"
" lua << EOF
" local dap = require('dap')
" dap.adapters.go = {
"   type = 'server',
"   port = '${port}',
"   executable = {
"     command = 'dlv',
"     args = {'dap', '-l', '127.0.0.1:${port}'},
"   },
" }
" dap.configurations.go = {
"   {
"     type = 'go',
"     name = 'Debug Web Server',
"     request = 'launch',
"     program = '${workspaceFolder}/cmd/web/',
"     args = {'serve'},
"   },
"   {
"     type = 'go',
"     name = 'Debug TUI',
"     request = 'launch',
"     program = '${workspaceFolder}',
"   },
"   {
"     type = 'go',
"     name = 'Debug Test',
"     request = 'launch',
"     mode = 'test',
"     program = '${file}',
"   },
"   {
"     type = 'go',
"     name = 'Attach to Headless',
"     request = 'attach',
"     mode = 'remote',
"     port = 2345,
"   },
" }
" EOF

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

" Open key directories
command! CbWeb :edit %:p:h/web/src/
nnoremap <leader>nw :CbWeb<CR>

command! CbComponents :edit %:p:h/web/src/components/
nnoremap <leader>nc :CbComponents<CR>

command! CbStores :edit %:p:h/web/src/stores/
nnoremap <leader>ns :CbStores<CR>

command! CbApi :edit %:p:h/web/src/api/
nnoremap <leader>na :CbApi<CR>

command! CbCmd :edit %:p:h/cmd/
nnoremap <leader>nC :CbCmd<CR>

command! CbInternal :edit %:p:h/internal/
nnoremap <leader>ni :CbInternal<CR>

command! CbCore :edit %:p:h/core/
nnoremap <leader>nO :CbCore<CR>

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
" QUICKFIX / LOCATION LIST HELPERS
" ============================================================================

" Run build and capture errors to quickfix
command! CbMake :cd %:p:h | make 2>&1 | cexpr system('cat') | copen
nnoremap <leader>qm :CbMake<CR>

" Run TypeScript check and capture errors
command! CbTscQf :cd %:p:h/web | cexpr system('npx tsc --noEmit 2>&1') | copen
nnoremap <leader>qt :CbTscQf<CR>

" ============================================================================
" HELP / INFO
" ============================================================================

command! CbHelp :echo "
\Kartoza CloudBench Keybindings (Leader = <Space>)
\
\SERVER:
\  <leader>ps  Start server
\  <leader>pq  Stop server
\  <leader>pr  Restart server (rebuild frontend)
\  <leader>p?  Check server status
\  <leader>po  Open in browser
\  <leader>pR  Full redeploy (clean + build + start)
\
\BUILD:
\  <leader>bf  Build frontend only
\  <leader>bw  Build web server binary
\  <leader>bt  Build TUI binary
\  <leader>ba  Build all
\  <leader>bc  Clean build artifacts
\  <leader>bT  TypeScript type check
\
\TEST:
\  <leader>ta  Run all tests
\  <leader>tp  Run tests in current package
\  <leader>tc  Run test under cursor
\  <leader>tC  Run tests with coverage
\
\LINT:
\  <leader>lg  Run golangci-lint
\  <leader>lf  Run eslint on frontend
\
\DOCKER (Test Containers):
\  <leader>dgs Start GeoServer    <leader>dgq Stop GeoServer
\  <leader>dps Start PostGIS      <leader>dpq Stop PostGIS
\  <leader>dms Start MinIO        <leader>dmq Stop MinIO
\  <leader>das Start all          <leader>daq Stop all
\  <leader>da? Status all         <leader>dgl/dpl/dml Logs
\  <leader>dpP PostGIS psql       <leader>dmC MinIO console
\
\DEBUG:
\  <leader>Dw  Debug web server (terminal)
\  <leader>Dt  Debug TUI (terminal)
\  <leader>DT  Debug current test file
\  <leader>DD  Start DAP server (headless)
\  <leader>Da  Attach to headless debugger
\
\GIT:
\  <leader>gs  Git status
\  <leader>gd  Git diff
\  <leader>gl  Git log
\
\DOCS:
\  <leader>xd  Serve docs locally
\  <leader>xD  Build docs
\  <leader>xs  Open SPECIFICATION.md
\  <leader>xc  Open CLAUDE.md
\
\NAVIGATE:
\  <leader>nw  web/src/
\  <leader>nc  web/src/components/
\  <leader>ns  web/src/stores/
\  <leader>na  web/src/api/
\  <leader>nC  cmd/
\  <leader>ni  internal/
\  <leader>nO  core/
\
\TERMINAL:
\  <leader>tt  Terminal in project root (nix develop)
\  <leader>tw  Terminal in web/ (nix develop)
\"
nnoremap <leader>ph :CbHelp<CR>

" ============================================================================
" AUTO COMMANDS
" ============================================================================

" Auto-format Go files on save (if you have gofmt or goimports)
augroup cloudbench_go
  autocmd!
  autocmd BufWritePre *.go silent! !goimports -w %
augroup END

" Auto-format TypeScript/TSX on save (if you have prettier)
augroup cloudbench_ts
  autocmd!
  autocmd BufWritePre *.ts,*.tsx silent! !npx prettier --write % 2>/dev/null
augroup END

" Set filetype specific settings
augroup cloudbench_filetypes
  autocmd!
  autocmd FileType go setlocal tabstop=4 shiftwidth=4 noexpandtab
  autocmd FileType typescript,typescriptreact setlocal tabstop=2 shiftwidth=2 expandtab
  autocmd FileType json setlocal tabstop=2 shiftwidth=2 expandtab
augroup END

" ============================================================================
" STATUS LINE INFO (optional)
" ============================================================================
" If you want project info in your statusline, you can use:
" let g:cloudbench_project = 'Kartoza CloudBench'

echo "Kartoza CloudBench config loaded. Type :CbHelp or <leader>ph for keybindings."
