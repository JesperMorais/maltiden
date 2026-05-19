# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) version 1.1.0.

## [Unreleased]

## [0.3.0] - 2026-05-11

### Added
- Content-sanitization regex set and emoji-strip helper in backend (`internal/interceptors`)
- Same validation and emoji-strip logic wired into the `RecipeService` parser path

### Changed
- Recipe content is now sanitized on both the direct-save and parse-and-save paths

## [0.2.0] - 2026-05-12

### Added
- Sentry `CaptureException` wired into 14 `ERROR`-level sites across backend services
- Sentry environment and release tags sourced from `FLY_APP_NAME` and `FLY_MACHINE_VERSION`

## [0.1.0] - 2026-05-04

### Added
- Shopping list category map expanded to 200+ Swedish ingredients
- Categories reorganized to match Swedish supermarket aisle layout
- `PATCH /shopping-list/items/{id}` endpoint for toggling item checked state
- `SetChecked` / `IsChecked` storage methods with full test coverage
- Frontend mock data updated to reflect new category structure
