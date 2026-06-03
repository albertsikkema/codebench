export type RenderContext = {
  input: string;
  timestamp: string;
  output?: string;
  steps: Record<string, Record<string, string>>;
};

const TOKEN_RE = /\{\{\s*([^}]+?)\s*\}\}/g;

export function render(template: string, ctx: RenderContext): string {
  return template.replace(TOKEN_RE, (_match, expr: string) => {
    const path = expr.split(".");
    const head = path[0];
    if (head === undefined) return "";

    if (head === "input" && path.length === 1) return ctx.input;
    if (head === "timestamp" && path.length === 1) return ctx.timestamp;
    if (head === "output" && path.length === 1) return ctx.output ?? "";

    if (head === "steps" && path.length === 3) {
      const stepId = path[1];
      const field = path[2];
      if (stepId === undefined || field === undefined) return "";
      const step = ctx.steps[stepId];
      if (step === undefined) return "";
      return step[field] ?? "";
    }

    return "";
  });
}
