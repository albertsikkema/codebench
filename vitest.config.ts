import { defineConfig } from "vitest/config";

export default defineConfig({
  test: {
    include: ["pipelines/**/*.test.ts"],
  },
});
