declare module "js-base64" {
  export const Base64: {
    encode(s: string, urlsafe?: boolean): string;
    encodeURI(s: string): string;
    decode(s: string): string;
  };
}

declare module "urijs" {
  interface URI {
    protocol(value: string): URI;
    username(value: string): URI;
    password(value: string): URI;
    host(value: string): URI;
    port(value: string | number): URI;
    path(value: string): URI;
    query(value: Record<string, unknown>): URI;
    hash(value: string): URI;
    toString(): string;
  }
  export default function URI(): URI;
}

// qrcode ships no types; the two calls the sharing dialog makes.
declare module "qrcode" {
  const QRCode: {
    toCanvas(
      canvas: HTMLCanvasElement,
      text: string,
      options: { errorCorrectionLevel?: string; width?: number },
      callback: (error: Error | null | undefined) => void,
    ): void;
  };
  export default QRCode;
}

// @nuintun/qrcode has types its package exports do not expose.
declare module "@nuintun/qrcode" {
  export class Decoder {
    scan(dataUrl: string): Promise<{ data: string }>;
  }
}

// highlight.js types its core but not the per-language modules
declare module "highlight.js/lib/languages/accesslog" {
  import type { LanguageFn } from "highlight.js";
  const language: LanguageFn;
  export default language;
}

// Markdown under src/docs is compiled to an HTML string by the vite plugin.
declare module "*.md" {
  const html: string;
  export default html;
}
