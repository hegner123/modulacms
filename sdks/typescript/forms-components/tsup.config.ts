import { defineConfig } from "tsup";

export default defineConfig({
  entry: ["src/index.ts"],
  format: ["esm", "cjs", "iife"],
  globalName: "ModulaCMSForms",
  dts: true,
  clean: true,
  sourcemap: true,
});
