# Project Distribute

Backend (Go) for a decentralized music platform paired with an offline-first Flutter client (local SQLite).

## Backend Features
- **Core**: Serves audio; syncs with trusted servers & clients for library sharing (push/pull).
- **Search**: Meilisearch for Songs, Albums, Artists, Playlists.
- **Management**: JWT Auth, User/Playlist/Folder data, Admin tools (`/doctor`, re-indexing, content management).
- **API**: RESTful (Swagger/OpenAPI).
- **Uploading**: Only admins can manage songs, albums, artists. Other users can create playlists and folders. 

## Tech Stack
- **Go 1.25** | **Echo v4** | **SQLite (GORM)** | **Meilisearch** | **Docker**

## Architecture
- **Flow**: Router -> Middleware -> Handler -> Service -> Store -> DB
- **Structure**: `main.go` entry. `router/` setup. `handler/routes.go` endpoints. `model/` GORM structs.
- **Data**: `data/` volume persists `/album_covers`, `/songs`, `/db`.
- **Config**: Uses development version of `docker-compose.yml`
- **Logging**: `log.Printf` to log.

## Database
- **Identifiers** Special entities only for songs, albums, artists. Used for cross-referencing.
- **Songs** May have several files attached of different quality. Songs can have multiple artists, but only one album.

## Development
- **Docker**: Docker container is currently running with Air hot-reloading.
- **Convention**: Auth `Bearer <token>`, Validation `go-playground/validator`.
