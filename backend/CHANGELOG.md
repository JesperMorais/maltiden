# Changelog

All notable changes to the Måltiden Go backend will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- _Nothing yet._

## [0.1.0] - 2026-05-07

Initial backend release.

### Added

- JWT-based authentication with bcrypt password hashing (registration, login, 7-day HS256 tokens).
- Household management: creation on registration, invite codes, member roles (`owner`, `member`, `guest`).
- Recipe CRUD with ingredients and instructions stored as JSON columns.
- AI-powered recipe parser using the Claude API to convert unstructured text into structured recipes (graceful degradation when `ANTHROPIC_API_KEY` is unset).
- Weekly menu generation and retrieval/update of the current household menu.
- Shopping list derived from the active menu, with per-item state updates.
- Tjek API integration for fetching Swedish grocery store offers and discounts (no auth required).
- User feedback submission endpoint.
- SQLite storage with sequential SQL migrations auto-applied on startup.
- HTTP middleware for auth, CORS, request IDs, and rate limiting.
