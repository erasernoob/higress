---
name: auto-openapi-generator
description: >
  Convert raw HTTP responses (optionally with URL, headers, and content-type)
  into a standardized OpenAPI v3.0 specification. Use this when you receive a 
  raw API response from a URL and need to infer or normalize its OpenAPI definition.
---


### 🧭 Workflow Summary

1. Input includes:

   * Raw HTTP response body
   * Response headers
   * HTTP status code
   * Request metadata (URL, method, params)
2. Parse and analyze:
   * Identify JSON, XML, HTML, or text response format.
   * Extract potential fields, arrays, and objects.
   * Derive data types and example values.
3. Build a minimal **OpenAPI v3.0.3** document:

   * Define `openapi`, `info`, `paths`, `components` sections.
   * For each endpoint, include:

     * OperationId (based on URL + method)
     * Request (if available)
     * Response schema (derived from response body)
4. Normalize and validate output using OpenAPI rules.
5. Return:

   * `openapi_spec`: a complete JSON or YAML OpenAPI document.
   * `assumptions`: natural language notes about inferred elements.
   * `confidence_score`: estimated confidence in schema correctness (0.0–1.0).

---

### 🧩 OpenAPI v3.0.3 Essentials

You must always include:

```yaml
openapi: 3.0.3
info:
  title: Auto-generated API
  version: 1.0.0
paths: {}
components:
  schemas: {}
```
more specific example see [example](example-openapi.yaml)

Each path item should include:

* `summary`: short endpoint description
* `operationId`: unique, camelCase identifier
* `parameters`: inferred query/path parameters
* `responses`: structured by HTTP status code, with `content.application/json.schema`
* `examples`: if JSON objects or arrays are detected, include one representative example

Note: **The generated openapi spec MUST be valid seriously**
---

### 🧠 Reasoning Rules for Schema Inference

When inferring schema from JSON responses:

| Example Value    | Inferred Type                                    | Notes                    |
| ---------------- | ------------------------------------------------ | ------------------------ |
| `"string"`       | `type: string`                                   |                          |
| `123`            | `type: integer`                                  | if fractional → `number` |
| `true / false`   | `type: boolean`                                  |                          |
| `[ ... ]`        | `type: array`, `items` derived recursively       |                          |
| `{ ... }`        | `type: object`, `properties` derived recursively |                          |
| Unknown or empty | `type: string` (default fallback)                |                          |

Infer parameter locations:

* If `{id}` appears in URL → `in: path`
* If `?key=value` pattern found → `in: query`
* If `Authorization`, `X-*` headers appear → add to `parameters` with `in: header`

---

### 🧱 Best Practices

* Always produce **syntactically valid JSON or YAML**.
* Always specify the **content-type** of responses (`application/json` by default).
* Always include **at least one example** for each operation.
* When unsure, mark assumption explicitly in the `assumptions` field.
* Do **not** hallucinate endpoints; only document those inferred from the provided input.

---

### 🧩 Output Format

Return a JSON object structured as:

```json
{
  "openapi_spec": "{...OpenAPI YAML...}",
  "assumptions": "List of inferred or uncertain details.",
  "confidence_score": 0.86
}
```

---

### ⚙️ Error Handling

* If response cannot be parsed, describe why and return partial spec.
* Never output invalid YAML/JSON; use defensive defaults.
* Ensure all `$ref` paths are valid if used.

---

### ✅ Final Objective

Produce a **ready-to-use**, **valid**, and **self-contained OpenAPI 3.0.3** document
that other tools (like `openapi-to-mcp`) can immediately consume without modification.
