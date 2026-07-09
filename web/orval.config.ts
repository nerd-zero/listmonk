import { defineConfig } from "orval";

// Generates a typed react-query client from the Go backend's swaggo-
// produced OpenAPI spec (internal/apidocs/swagger.json -- run
// `swag init -g cmd/api/main.go --parseInternal -o internal/apidocs`
// from the repo root after changing any handler's doc comments, then
// rerun `npx orval` here).
export default defineConfig({
  listnun: {
    input: {
      target: "../internal/apidocs/swagger.json",
    },
    output: {
      mode: "tags-split",
      target: "src/api/generated/endpoints",
      schemas: "src/api/generated/model",
      client: "react-query",
      httpClient: "fetch",
      override: {
        mutator: {
          path: "src/api/mutator.ts",
          name: "customFetch",
        },
        query: {
          useQuery: true,
        },
      },
    },
  },
});
