# device-registry

REST API for managing devices. Go + AWS Lambda + DynamoDB, deployed via Serverless Framework.

## Stack

- **Go 1.26**
- **AWS Lambda** (arm64 / Graviton)
- **DynamoDB** (PAY_PER_REQUEST)
- **API Gateway HTTP API**
- **Serverless Framework v4**

## API

All endpoints require `Authorization: Bearer <API_SECRET>` header.

| Method | Path           | Description              |
|--------|----------------|--------------------------|
| GET    | /devices       | List devices (paginated) |
| GET    | /devices/{id}  | Get device by ID         |
| POST   | /devices       | Create device            |
| PUT    | /devices/{id}  | Update device            |
| DELETE | /devices/{id}  | Delete device            |

### Device model

```json
{
  "deviceId": "string (1–128 chars, alphanumeric / - / _)",
  "name":     "string (max 256 chars)",
  "type":     "string (max 128 chars)"
}
```

### GET /devices — pagination

Query parameters:

| Parameter    | Default | Max  | Description                       |
|--------------|---------|------|-----------------------------------|
| `limit`      | 100     | 1000 | Number of items per page          |
| `nextToken`  | —       | —    | Cursor from the previous response |

Response:

```json
{
  "devices": [{ "deviceId": "...", "name": "...", "type": "..." }],
  "nextToken": "base64-encoded-cursor"
}
```

`nextToken` is absent when there are no more pages.

### POST /devices — create

Request body: device object. Returns the created device with `201 Created`.

### PUT /devices/{id} — update

Request body: `{ "name": "...", "type": "..." }`. Returns the updated device with `200 OK`.

### Error responses

All errors return JSON:

```json
{ "error": "description" }
```

| Status | Meaning                        |
|--------|--------------------------------|
| 400    | Validation error / bad input   |
| 401    | Missing or invalid token       |
| 404    | Device not found               |
| 409    | Device already exists (POST)   |
| 500    | Internal error                 |

## Authentication

The API uses a Lambda authorizer with a Bearer token. Set the secret before deploying:

```bash
export API_SECRET=your-secret-here
serverless deploy
```

For `prod` stage, store the secret in AWS SSM Parameter Store:

```bash
aws ssm put-parameter \
  --name /device-registry/prod/apiSecret \
  --value "your-secret-here" \
  --type SecureString
```

Then deploy:

```bash
serverless deploy --stage prod
```

If `API_SECRET` is not set, the dev stage defaults to `dev-secret-change-me`.

## Requirements

- Go 1.26
- AWS CLI configured with appropriate credentials
- Node.js (for Serverless Framework)
- Serverless Framework v4: `npm install -g serverless`

## Build

```bash
make all        # clean + build + zip
make build      # compile only
make zip        # package only
make clean      # remove build/
```

## Deploy

```bash
# dev stage (uses API_SECRET env var or default)
export API_SECRET=my-dev-secret
serverless deploy

# prod stage (reads from SSM)
serverless deploy --stage prod
```

## Remove

```bash
serverless remove --stage dev
```

## Tests

```bash
go test ./...
```

## Project structure

```
.
├── cmd/lambda/          # Lambda function entry points
│   ├── authorizer/      # Bearer token authorizer
│   ├── deleteDevice/
│   ├── getAllDevices/
│   ├── getDeviceByID/
│   ├── saveDevice/
│   └── updateDevice/
├── internal/
│   └── lambdahelper/    # shared helpers (service interface, JSON responses)
├── pkg/
│   ├── model/           # data models and error types
│   ├── repository/      # DynamoDB layer
│   └── service/         # business logic and validation
├── Makefile
└── serverless.yml       # infrastructure (Lambda + API Gateway + DynamoDB)
```

## Environment variables

| Variable     | Default                  | Description                          |
|--------------|--------------------------|--------------------------------------|
| `TABLE_NAME` | `Devices`                | DynamoDB table name                  |
| `API_SECRET` | `dev-secret-change-me`   | Bearer token for the authorizer      |