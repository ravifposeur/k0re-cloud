# k0re Daemon API Documentation

REST API for the **k0red** game server provisioning daemon.

- **Version:** 1.0.0
- **Base URL:** `http://127.0.0.1:3000`

---

## Endpoints

### 1. `POST /v1/provision`

Provision a new game server.

**Request Body** (`application/json`):

| Field | Type | Required | Example |
|-------|------|----------|---------|
| `name` | string | ✅ | `arena-ugm-01` |
| `game` | string | ✅ | `minecraft` |
| `flavor` | string | ✅ | `pvp-competitive` |
| `ram` | string | ✅ | `2G` |

**Example Request:**

```bash
curl -X POST http://127.0.0.1:3000/v1/provision \
  -H "Content-Type: application/json" \
  -d '{
    "name": "arena-ugm-01",
    "game": "minecraft",
    "flavor": "pvp-competitive",
    "ram": "2G"
  }'
```

**Responses:**

`200 OK` — Provisioning successful

```json
{
  "status": "success",
  "message": "Server provisioned"
}
```

`400 Bad Request` — Missing or invalid fields

```json
{
  "status": "error",
  "message": "Deskripsi error"
}
```

`500 Internal Server Error` — Server-side failure

```json
{
  "status": "error",
  "message": "Deskripsi error"
}
```

---

### 2. `GET /v1/status/{name}`

Get the status of a provisioned game server.

**Path Parameter:**

| Parameter | Type | Required | Example |
|-----------|------|----------|---------|
| `name` | string | ✅ | `arena-ugm-01` |

**Example Request:**

```bash
curl http://127.0.0.1:3000/v1/status/arena-ugm-01
```

**Responses:**

`200 OK` — Server status retrieved successfully

```json
{
  "name": "arena-ugm-01",
  "status": "running",
  "cpu_usage": "12%",
  "mem_usage": "512MB/2GB"
}
```

`404 Not Found` — Server does not exist

```json
{
  "status": "error",
  "message": "Deskripsi error"
}
```

---

## Schemas

### SuccessResponse

```json
{
  "status": "success",
  "message": "Server provisioned"
}
```

### ErrorResponse

```json
{
  "status": "error",
  "message": "Deskripsi error"
}
```

### StatusResponse

```json
{
  "name": "arena-ugm-01",
  "status": "running",
  "cpu_usage": "12%",
  "mem_usage": "512MB/2GB"
}
```
