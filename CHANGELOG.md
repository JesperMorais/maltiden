# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog 1.1.0](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

### Changed

### Deprecated

### Removed

### Fixed

### Security

## [0.1.0] - 2026-05-07

### Added
- Authentication system with email/password registration, bcrypt hashing, and JWT-based sessions (7-day expiry).
- Household management with owner/member/guest roles, invite codes, and member administration.
- Recipe management: create, edit, delete, and browse recipes with ingredients and instructions.
- AI-powered recipe parser using the Claude API to convert unstructured recipe text into structured data (graceful degradation when API key absent).
- Weekly menu generation with current-menu retrieval and update endpoints.
- Smart shopping list derived from the active menu, with item check-off support.
- Grocery offers integration via the Tjek API, including offer search, discounts, and store listings.
- Feedback submission endpoint.
- Vue 3 + TypeScript frontend with Pinia state management, mobile-first design, and warm cream/coral theme with full dark mode support.
- Go 1.26 backend with layered architecture (handlers → services → storage → domain) and SQLite persistence.
- Fly.io deployment configuration (Stockholm region).

[Unreleased]: https://github.com/davdd1/maltiden/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/davdd1/maltiden/releases/tag/v0.1.0
