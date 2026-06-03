#!/usr/bin/env -S npx tsx
import { spawn } from "node:child_process";
import { render, type RenderContext } from "./render.js";

const argv2 = process.argv[2];
if (argv2 === undefined) {
  console.error("usage: tsx pipelines/hello.ts <input>");
  process.exit(2);
}

const ctx: RenderContext = {
  input: argv2,
  timestamp: new Date().toISOString(),
  steps: {},
};

const prompt = render("Echo this: {{ input }} (at {{ timestamp }})", ctx);
console.log("rendered prompt:", prompt);

// In a real pipeline, swap "echo" for ["claude-safe", "--no-firewall", "--", "claude", "-p", prompt]
// (mirrors pipeline.py:49). echo is used here so this runs in any environment.
const child = spawn("echo", [prompt], { stdio: "inherit" });

const code: number = await new Promise((resolve, reject) => {
  child.on("error", reject);
  child.on("exit", (c) => resolve(c ?? 1));
});

process.exit(code);
