# API Go - VideoJuegos

Servidor HTTP en Go que actúa como proxy a RAWG y guarda una colección personal en MongoDB.

## Ejecutar

```bash
cd /Users/felipe/Desktop/cursoGo/APP-VideoJuegos
go mod tidy
go run .
```

## Configuración

El proyecto carga variables de entorno desde `conn.env` si existe. Las variables disponibles son:

- `DB_NAME` - nombre de la base de datos MongoDB (por defecto `goVideoJuegos`)
- `DB_USER` - usuario MongoDB
- `DB_PASSWORD` - contraseña MongoDB
- `DB_HOST` - host MongoDB (por defecto `localhost`)
- `DB_PORT` - puerto MongoDB (por defecto `27017`)

También se mantiene compatibilidad con los nombres antiguos en caso de que se usen variables de entorno ya existentes:

- `userDB`
- `passUserDB..`
- `MONGO_HOST`

## Endpoints

### RAWG proxy

- `GET /api/search?q=zelda`
  - Busca juegos en RAWG y devuelve resultados con campos mínimos.
- `GET /api/games/{rawg_id}`
  - Devuelve el detalle completo del juego en RAWG.
- `GET /api/db/health`
  - Verifica la conexión a la base de datos MongoDB.

### Biblioteca personal

- `GET /api/library`
  - Lista todos los juegos guardados.
  - Parámetro opcional: `?status=pendiente|jugando|completado|abandonado`
- `POST /api/library`
  - Agrega un juego a la biblioteca.
  - Payload mínimo:

```json
{
  "rawg_id": 1,
  "title": "Nombre del juego"
}
```

- `PUT /api/library/{id}`
  - Actualiza un documento existente en la biblioteca.
  - `id` es el ObjectID MongoDB del documento.
  - Campos permitidos en el body:

```json
{
  "personal_note": "Comentario personal",
  "personal_score": 8,
  "status": "jugando"
}
```

- `DELETE /api/library/{id}`
  - Elimina un juego por su ObjectID.
- `GET /api/library/stats`
  - Devuelve estadísticas de la colección, incluyendo conteo por estado y puntaje promedio.

## Reglas de validación

- `status` solo acepta los valores:
  - `pendiente`
  - `jugando`
  - `completado`
  - `abandonado`
- `personal_score` solo acepta valores enteros entre `1` y `10`.

## Notas

- Los IDs de la colección son ObjectID hex y se usan en rutas `PUT /api/library/{id}` y `DELETE /api/library/{id}`.
- El servidor corre en `:8080` por defecto.
