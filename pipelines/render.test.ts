import { describe, expect, it } from "vitest";
import { render, type RenderContext } from "./render.js";

const baseCtx: RenderContext = {
  input: "hello",
  timestamp: "2026-05-09T16:10:23Z",
  steps: {
    research: { output_file: "/tmp/r.md", stdout: "ok" },
  },
};

describe("render", () => {
  it("substitutes {{ input }}", () => {
    expect(render("got: {{ input }}", baseCtx)).toBe("got: hello");
  });

  it("substitutes {{ timestamp }}", () => {
    expect(render("at {{ timestamp }}", baseCtx)).toBe(
      "at 2026-05-09T16:10:23Z",
    );
  });

  it("substitutes {{ steps.<id>.<field> }}", () => {
    expect(render("file={{ steps.research.output_file }}", baseCtx)).toBe(
      "file=/tmp/r.md",
    );
  });

  it("substitutes {{ output }} from optional ctx", () => {
    expect(render("o={{ output }}", { ...baseCtx, output: "X" })).toBe("o=X");
  });

  it("returns empty string when {{ output }} is unset", () => {
    expect(render("o={{ output }}", baseCtx)).toBe("o=");
  });

  it("tolerates extra whitespace inside braces", () => {
    expect(render("{{    input    }}", baseCtx)).toBe("hello");
  });

  it("returns empty string for unknown step", () => {
    expect(render("{{ steps.missing.x }}", baseCtx)).toBe("");
  });

  it("substitutes multiple tokens in one string", () => {
    expect(render("{{ input }}/{{ timestamp }}", baseCtx)).toBe(
      "hello/2026-05-09T16:10:23Z",
    );
  });
});
