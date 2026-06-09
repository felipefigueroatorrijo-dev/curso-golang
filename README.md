# API Go - VideoJuegos

Servidor HTTP en Go que actúa como proxy a RAWG y guarda una colección personal en MongoDB.

## Requisitos previos

### MongoDB

MongoDB debe estar corriendo localmente en `localhost:27017`.

#### macOS (Homebrew)

```bash
# Instalar MongoDB Community Edition
brew tap mongodb/brew
brew install mongodb-community

# Iniciar el servicio (se ejecutará en segundo plano)
brew services start mongodb-community

# Verificar que está corriendo
mongosh

# Crear usuario de prueba (opcional, si usas autenticación)
# En mongosh:
use admin
db.createUser({
  user: "userDB",
  pwd: "passUserDB..",
  roles: ["root"]
})
```

#### Linux (apt)

```bash
wget -qO - https://www.mongodb.org/static/pgp/server-6.0.asc | sudo apt-key add -
sudo apt-get install -y mongodb-org
sudo systemctl start mongod
sudo systemctl enable mongod
```

#### Docker (alternativa multiplataforma)

```bash
docker run -d \
  -p 27017:27017 \
  --name mongodb \
  mongo:latest
```

#### Windows

Descargar el instalador desde [mongodb.com/try/download/community](https://www.mongodb.com/try/download/community) y seguir el asistente de instalación.

## Ejecutar

```bash
cd /Users/felipe/Desktop/cursoGo/APP-VideoJuegos
go mod tidy
go run .
```

## Configuración

El proyecto carga variables de entorno desde `conn.env` si existe. Las variables disponibles son:

- `DB_NAME` - nombre de la base de datos MongoDB (por defecto `goVideoJuegos`)
- `DB_USER` - usuario MongoDB (opcional)
- `DB_PASSWORD` - contraseña MongoDB (opcional)
- `DB_HOST` - host MongoDB (por defecto `localhost`)
- `DB_PORT` - puerto MongoDB (por defecto `27017`)
- `DB_AUTH_SOURCE` - base de datos para autenticación (por defecto: el valor de `DB_NAME`)
- `DB_AUTH_MECHANISM` - mecanismo de autenticación (ej: `SCRAM-SHA-1`, `SCRAM-SHA-256`; por defecto: automático)

### Ejemplo conn.env (sin autenticación)

```
DB_NAME=goVideoJuegos
DB_HOST=localhost
DB_PORT=27017
```

### Ejemplo conn.env (con autenticación)

```
DB_NAME=goVideoJuegos
DB_USER=userDB
DB_PASSWORD=passUserDB..
DB_HOST=localhost
DB_PORT=27017
DB_AUTH_SOURCE=admin
DB_AUTH_MECHANISM=SCRAM-SHA-256
```

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
