---
name: api-tools
description: Create and manage Bruno API collections. Use when building API requests, test suites, environments, or working with .bru files in api_tools/. Bruno is a Git-native, offline-first API client (Postman alternative).
---

# API Tools (Bruno)

Skill for creating and managing [Bruno](https://www.usebruno.com/) API collections. Bruno stores API requests as plain-text `.bru` files on disk — Git-friendly, no cloud sync, no account required.

## When to Use

- Creating new API requests or test collections
- Adding tests/assertions to existing requests
- Setting up environments (dev, staging, production)
- Writing pre-request or post-response scripts
- Reviewing or modifying `.bru` files
- Setting up CI test runners with `bru run`

**Not for:** Unit testing, frontend testing, or non-HTTP protocols (use other tools).

## Collection Location

By convention, Bruno collections live in `api_tools/` at the project root.

## Collection Structure

```
api_tools/
├── bruno.json              # Collection manifest (required)
├── collection.bru          # Collection-level scripts, vars, auth, headers
├── .env                    # Secrets (gitignored)
├── environments/
│   ├── Development.bru
│   ├── Staging.bru
│   └── Production.bru
├── auth/
│   ├── folder.bru          # Folder-level config
│   ├── login.bru
│   └── register.bru
├── users/
│   ├── folder.bru
│   ├── list-users.bru
│   ├── get-user.bru
│   ├── create-user.bru
│   └── delete-user.bru
└── health/
    └── healthcheck.bru
```

Filesystem structure maps directly to the Bruno UI sidebar. The `seq` field in `meta` controls sort order within a folder.

## bruno.json

Required at the collection root. Minimal example:

```json
{
  "version": "1",
  "name": "My API",
  "type": "collection",
  "ignore": ["node_modules", ".git"]
}
```

Full example with scripts and proxy:

```json
{
  "version": "1",
  "name": "My API",
  "type": "collection",
  "ignore": ["node_modules", ".git"],
  "scripts": {
    "flow": "sandwich",
    "filesystemAccess": {
      "allow": true
    }
  },
  "proxy": {
    "enabled": true,
    "protocol": "https",
    "hostname": "proxy.example.com",
    "port": 8080
  }
}
```

Script flow modes:
- `"sandwich"` (default): collection pre → folder pre → request pre → **request** → request post → folder post → collection post
- `"sequential"`: collection pre → folder pre → request pre → **request** → collection post → folder post → request post

## .bru File Format

The Bru language uses three block types: **dictionary** (key-value), **text** (free-form content), and **list** (items in brackets).

### Request File

```bru
meta {
  name: Get User
  type: http
  seq: 1
}

post {
  url: {{baseUrl}}/users
  body: json
}

params:query {
  filter: active
  ~debug: true
}

headers {
  Authorization: Bearer {{token}}
  Content-Type: application/json
  ~X-Debug: enabled
}

auth:bearer {
  token: {{apiToken}}
}

body:json {
  {
    "name": "{{userName}}",
    "email": "user@example.com"
  }
}

assert {
  res.status: eq 201
  res.body.id: isDefined
}

tests {
  test("should create user", function() {
    expect(res.status).to.equal(201);
    expect(res.body).to.have.property("id");
  });
}
```

### Key Syntax

| Prefix | Meaning |
|--------|---------|
| `~` | Disabled entry (skipped during execution) |
| `@` | Local/runtime variable (not persisted to file) |
| `~@` | Disabled + local |

### HTTP Methods

Use the method name as the block: `get`, `post`, `put`, `patch`, `delete`, `head`, `options`. When using a request body, declare the body type inside the method block with `body: json` (or `text`, `xml`, `form-urlencoded`, `multipart-form`, `graphql`). Without this declaration, the `body:*` block is ignored.

```bru
get {
  url: {{baseUrl}}/users/{{userId}}
}
```

```bru
post {
  url: {{baseUrl}}/users
  body: json
}
```

### Body Types

**Important:** You must declare `body: <type>` inside the HTTP method block for the `body:*` block to take effect. Without it, Bruno ignores the body.

| Block | Method block declaration |
|-------|------------------------|
| `body:json` | `body: json` |
| `body:text` | `body: text` |
| `body:xml` | `body: xml` |
| `body:form-urlencoded` | `body: form-urlencoded` |
| `body:multipart-form` | `body: multipart-form` |
| `body:graphql` | `body: graphql` |
| `body:graphql:vars` | (paired with `body:graphql`) |

### Auth Types

| Block | Fields |
|-------|--------|
| `auth:bearer` | `token` |
| `auth:basic` | `username`, `password` |
| `auth:oauth2` | `grant_type`, `access_token_url`, etc. |
| `auth:api-key` | `add_to`, `key`, `value` |
| `auth:digest` | `username`, `password` |

## collection.bru

Collection-level config that applies to ALL requests:

```bru
headers {
  Accept: application/json
  X-Api-Version: v2
}

auth {
  mode: bearer
}

auth:bearer {
  token: {{authToken}}
}

script:pre-request {
  // Runs before every request in the collection
  const token = bru.getEnvVar("authToken");
  if (!token) {
    throw new Error("No auth token set");
  }
}

script:post-response {
  // Runs after every request
  if (res.status === 401) {
    bru.setVar("authExpired", true);
  }
}
```

## folder.bru

Same as `collection.bru` but scoped to a folder:

```bru
meta {
  name: Users API
}

headers {
  X-Resource-Type: user
}

auth:bearer {
  token: {{userServiceToken}}
}
```

## Environments

Environment files live in `environments/` as `.bru` files:

```bru
vars {
  host: https://dev-api.example.com
  apiKey: dev-key-12345
  timeout: 5000
}
```

For secrets, reference process env variables:

```bru
vars {
  host: https://api.example.com
  apiKey: {{process.env.API_KEY}}
}
```

## Variables

### Precedence (highest to lowest)

1. **Runtime** — `bru.setVar()`, in-memory only
2. **Request** — `vars:pre-request` / `vars:post-response` in request `.bru`
3. **Folder** — in `folder.bru`
4. **Collection** — in `collection.bru`
5. **Environment** — in `environments/<name>.bru`
6. **Process env** — from `.env` file

### Declarative Variables

```bru
vars:pre-request {
  baseUrl: https://api.example.com
  @tempId: {{$randomUUID}}
}

vars:post-response {
  userId: $res.body.id
  @token: $res.body.token
}
```

### Script Variable Access

```javascript
// Runtime
bru.setVar("key", "value");
bru.getVar("key");

// Environment
bru.getEnvVar("key");
bru.setEnvVar("key", "value");

// Process env (.env file)
bru.getProcessEnv("API_KEY");
```

All variables are interpolated with `{{variableName}}`.

## Scripts

### Pre-Request

```bru
script:pre-request {
  req.setHeader("X-Request-Id", Date.now().toString());
  req.setUrl(bru.getVar("baseUrl") + "/users");
}
```

**`req` methods:** `getUrl()`, `setUrl()`, `getMethod()`, `setMethod()`, `getHeader(name)`, `setHeader(name, value)`, `getBody()`, `setBody(body)`, `setTimeout(ms)`

### Post-Response

```bru
script:post-response {
  let data = res.getBody();
  bru.setVar("userId", data.id);
}
```

**`res` methods:** `getStatus()`, `getHeader(name)`, `getHeaders()`, `getBody()`, `getResponseTime()`

**`res` properties:** `res.status`, `res.statusText`, `res.headers`, `res.body`, `res.responseTime`

### bru Utilities

| Method | Purpose |
|--------|---------|
| `bru.setNextRequest(name)` | Chain to another request |
| `bru.runner.skipRequest()` | Skip current request |
| `bru.runner.stopExecution()` | Stop collection run |
| `bru.sleep(ms)` | Async delay |
| `bru.getProcessEnv(key)` | Read `.env` variable |

## Assertions

Declarative (no-code) response validation:

```bru
assert {
  res.status: eq 200
  res.body.success: eq true
  res.body.data.id: isDefined
  res.body.data.name: isString
  res.body.data.items: isArray
  res.body.data.items: length 5
  res.body.data.email: contains @
  res.body.data.age: gt 18
  res.body.data.age: lt 100
  res.body.data.age: between 18 100
  res.headers.content-type: contains application/json
  res.responseTime: lt 2000
}
```

### Operators

| Category | Operators |
|----------|-----------|
| Comparison | `eq`, `neq`, `gt`, `gte`, `lt`, `lte` |
| String | `contains`, `notContains`, `startsWith`, `endsWith`, `matches` |
| Type | `isNumber`, `isString`, `isBoolean`, `isArray`, `isJson` |
| Existence | `isDefined`, `isUndefined`, `isNull`, `isEmpty`, `isNotEmpty` |
| Truthiness | `isTruthy`, `isFalsy` |
| Collection | `in`, `notIn`, `between`, `length` |

## Tests

Chai-style JavaScript tests for complex validation:

```bru
tests {
  test("status is 200", function() {
    expect(res.status).to.equal(200);
  });

  test("response has required fields", function() {
    const body = res.getBody();
    expect(body).to.have.property("id");
    expect(body.name).to.be.a("string");
  });

  test("response time acceptable", function() {
    expect(res.getResponseTime()).to.be.below(2000);
  });
}
```

## CLI Usage

Install: `npm install -g @usebruno/cli`

```bash
# Run all requests (from collection root)
bru run

# Run a single request or folder
bru run users/get-user.bru
bru run auth/

# With environment
bru run --env Development

# Override env vars (for CI)
bru run --env CI --env-var "apiKey=$API_KEY"

# Reports
bru run --reporter-json results.json
bru run --reporter-junit results.xml
bru run --reporter-html results.html

# Control execution
bru run --bail              # Stop on first failure
bru run --delay 500         # Delay between requests (ms)
bru run --tests-only        # Only requests with tests
bru run --tags smoke        # Filter by tag

# Developer mode (filesystem access, require())
bru run --sandbox=developer
```

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Test/assertion failure |
| 4 | Not run from collection root |
| 5 | Input file missing |
| 6 | Environment not found |

## CI/CD Integration

```yaml
- name: Run API tests
  run: |
    npm install -g @usebruno/cli
    cd api_tools
    bru run --env CI \
      --env-var "baseUrl=${{ secrets.API_URL }}" \
      --env-var "apiKey=${{ secrets.API_KEY }}" \
      --reporter-junit results.xml \
      --bail
```

## Git Best Practices

- Commit all `.bru` files, `bruno.json`, `collection.bru`, `folder.bru`, and environment files
- Add `.env` to `.gitignore` (contains secrets)
- Use placeholder values in environment files; override via CLI `--env-var` in CI
- `.bru` files produce clean, readable diffs — review them in PRs

## Workflow Tests (Chaining Requests)

Use a dedicated `workflows/` folder with sequenced requests that pass data between steps via `bru.setVar()`. The `seq` field controls execution order.

### Structure

```
api_tools/workflows/
├── folder.bru
├── 1-extract-location.bru     # Step 1: extract data, save to variable
└── 2-get-weather.bru          # Step 2: use extracted data
```

### Step 1: Extract and Store

```bru
meta {
  name: 1. Extract Location
  type: http
  seq: 1
}

post {
  url: {{baseUrl}}/api/v1/locations/extract
  body: json
}

body:json {
  {
    "text": "The weather in Amsterdam is nice today"
  }
}

assert {
  res.status: eq 200
  res.body.locations: isArray
}

script:post-response {
  const locations = res.getBody().locations;
  if (locations && locations.length > 0) {
    bru.setVar("extractedPlace", locations[0].name);
  }
}

tests {
  test("should extract at least one location", function() {
    expect(res.body.locations.length).to.be.greaterThan(0);
  });
}
```

### Step 2: Use Stored Variable

```bru
meta {
  name: 2. Get Weather for Extracted Location
  type: http
  seq: 2
}

get {
  url: {{baseUrl}}/api/v1/weather/{{extractedPlace}}
}

assert {
  res.status: eq 200
  res.body.location: isDefined
}

tests {
  test("should return weather for extracted location", function() {
    expect(res.status).to.equal(200);
    expect(res.body.location).to.be.a("string");
  });
}
```

### Running Workflows

**CLI:**
```bash
cd api_tools && bru run workflows/ --env Development
```

**GUI:**
1. Open the collection (point Bruno to `api_tools/`)
2. Select environment from the top-right dropdown
3. Click the `workflows` folder in the sidebar
4. Click the **Run** button (play icon) on the folder — runs all requests in `seq` order

Variables set via `bru.setVar()` in step 1 carry over to step 2 within the same folder run.

### Chaining Patterns

| Pattern | How |
|---------|-----|
| Pass data between requests | `bru.setVar("key", value)` in post-response, `{{key}}` in next request |
| Conditional skip | `bru.runner.skipRequest()` in pre-request script |
| Stop on failure | `bru.runner.stopExecution()` in post-response script |
| Jump to specific request | `bru.setNextRequest("Request Name")` |

## Workflow: Creating a New Collection

1. Create `api_tools/bruno.json` with collection name
2. Create `api_tools/environments/` with at least Development and Production
3. Create `api_tools/collection.bru` with shared auth/headers
4. Organize requests into folders matching API resource groups
5. Add `folder.bru` to each folder for shared folder config
6. Add assertions and tests to each request
7. Verify with `cd api_tools && bru run --env Development`

## Documentation

- [Bruno Documentation](https://docs.usebruno.com/)
- [Bru Language Overview](https://docs.usebruno.com/bru-lang/overview)
- [Bru Language Tag Reference](https://docs.usebruno.com/bru-lang/tag-reference)
- [JavaScript Reference (Scripts)](https://docs.usebruno.com/testing/script/javascript-reference)
- [Assertions Reference](https://docs.usebruno.com/testing/tests/assertions)
- [CLI Overview](https://docs.usebruno.com/bru-cli/overview)
- [CLI Command Options](https://docs.usebruno.com/bru-cli/commandOptions)
- [Bruno GitHub](https://github.com/usebruno/bruno)
- [Bru Language Spec](https://github.com/brulang/bru-lang)
