import { defineConfig } from "orval";

export default defineConfig({
  ddivination: {
    input: "./openapi.json",
    output: {
      target: "./src/api/generated/client.ts",
      schemas: "./src/api/generated/models",
      client: "fetch",
      clean: true,
      prettier: true,
    },
  },
});
