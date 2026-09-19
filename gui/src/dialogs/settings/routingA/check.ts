import { tokenize } from "./highlight";

export type ErrorKey =
  | "routingA.errors.noArrow"
  | "routingA.errors.brackets"
  | "routingA.errors.noOutbound";
export interface LineError {
  line: number;
  message: ErrorKey;
}

export function check(text: string): LineError[] {
  const errors: LineError[] = [];
  text.split("\n").forEach((line, index) => {
    const tokens = tokenize(line).filter((token) => token.type !== "comment");
    if (!tokens.some((token) => token.text.trim())) return;
    const code = tokens
      .map((token) => (token.type === "string" ? "value" : token.text))
      .join("");
    const stack: string[] = [];
    let unbalanced = false;
    for (const char of code) {
      if ("([{".includes(char)) stack.push(char);
      else if (")]}".includes(char)) {
        if (stack.pop() !== "([{"[")]}".indexOf(char)]) unbalanced = true;
      }
    }
    const arrow = code.indexOf("->");
    let message: ErrorKey | undefined;
    if (unbalanced || stack.length) message = "routingA.errors.brackets";
    else if (arrow >= 0 && !code.slice(arrow + 2).trim())
      message = "routingA.errors.noOutbound";
    else if (arrow < 0 && !/^\s*(?:default|outbound)\s*:/.test(code))
      message = "routingA.errors.noArrow";
    if (message) errors.push({ line: index + 1, message });
  });
  return errors;
}
