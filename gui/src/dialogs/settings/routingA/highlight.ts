export type TokenType =
  | "text"
  | "comment"
  | "keyword"
  | "function"
  | "argument"
  | "arrow"
  | "outbound"
  | "string"
  | "number"
  | "operator";

export interface Token {
  type: TokenType;
  text: string;
}

const syntax =
  /#[^\n]*|"(?:\\.|[^"\\])*"|'(?:\\.|[^'\\])*'|\b(?:default|outbound)\s*:|\b(?:domain|ip|port|network|protocol|source|app|extern)(?=\s*\()|\b(?:full|domain|contains|regexp|geosite|geoip|address|port|user|pass)\s*:|->|&&|\b\d+(?:\.\d+)*(?:\/\d+)?\b/g;

export function tokenize(line: string): Token[] {
  const tokens: Token[] = [];
  let offset = 0;
  for (const match of line.matchAll(syntax)) {
    if (match.index < offset) continue;
    if (match.index > offset)
      tokens.push({ type: "text", text: line.slice(offset, match.index) });
    const text = match[0];
    let type: TokenType;
    if (text.startsWith("#")) type = "comment";
    else if (/^["']/.test(text)) type = "string";
    else if (/^(default|outbound)\s*:/.test(text)) type = "keyword";
    else if (text.endsWith(":")) type = "argument";
    else if (text === "->") type = "arrow";
    else if (text === "&&") type = "operator";
    else if (/^\d/.test(text)) type = "number";
    else type = "function";
    tokens.push({ type, text });
    offset = match.index + text.length;
    if (type === "arrow") {
      const outbound = /^(\s*)([^\s#]+)/.exec(line.slice(offset));
      if (outbound) {
        if (outbound[1]) tokens.push({ type: "text", text: outbound[1] });
        tokens.push({ type: "outbound", text: outbound[2] });
        offset += outbound[0].length;
      }
    }
  }
  if (offset < line.length)
    tokens.push({ type: "text", text: line.slice(offset) });
  return tokens;
}
