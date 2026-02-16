# Plan: API Consolidation Review

## Goal
Go igenom alla frontend API-anrop och verifiera att de matchar Go API-endpoints.

## Arkitekturbeslut

**Endast Go API** — Denna applikation har ett enda backend-API i Go. Inga Nuxt server routes.

## Scope
- Granska Go API i `cmd/api/main.go`
- Granska alla frontend API-anrop i `frontend/pages/*.vue`
- Jämför och identifiera luckor/mismatch
- Verifiera att `frontend/server/api/` är tom

## Output
- Dokumentation av alla endpoints i `agent-os/specs/2026-02-03-1430-api-consolidation/`

## Spec Status: ✅ DONE

**Completed:** 2026-02-03

**Summary:**
All API standards now implemented:
- Centralized error handling via `internal/api/errors.go`
- Input validation on all POST/PUT endpoints
- Field-specific error messages in validation responses
