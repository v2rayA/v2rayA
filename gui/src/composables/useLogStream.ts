// The core's log as a growing list of lines: GET /logger from a byte
// offset every few seconds, a half line kept until its rest arrives,
// the level and source of each line read off its text. Scope-bound:
// the timer stops with the component.
import { computed, onScopeDispose, ref, watch } from "vue";
import { getLogger } from "@/api";

export interface LogLine {
  id: number;
  text: string;
  level: string;
  source: string;
}

export function detectLevel(text: string): string {
  const lower = text.toLowerCase();
  if (lower.includes("[e]") || lower.includes(" error ")) return "error";
  if (lower.includes("[w]") || lower.includes(" warn")) return "warn";
  if (lower.includes("[d]") || lower.includes(" debug")) return "debug";
  if (lower.includes("[t]") || lower.includes(" trace")) return "trace";
  if (lower.includes("[i]") || lower.includes(" info")) return "info";
  return "other";
}

/** detectSource reads the [file.go:123] or [SomeService] tag. */
export function detectSource(text: string): string {
  const m =
    /\[([^\]]+\.go|[A-Za-z]+Service|[A-Za-z]+\.[A-Za-z]+)(?::\d+)?\]/.exec(
      text,
    );
  return m ? m[1] : "other";
}

export function useLogStream(intervalSeconds = 5) {
  const maxLines = 20000;
  const lines = ref<LogLine[]>([]);
  const sources = ref<string[]>([]);
  const interval = ref(intervalSeconds);
  let skip = 0;
  let endOfLine = true;
  let fetching = false;
  let timer: ReturnType<typeof setInterval> | null = null;

  function noteSource(source: string) {
    if (source && !sources.value.includes(source)) sources.value.push(source);
  }

  /** append folds a chunk into the lines; the chunk may end mid-line. */
  function append(chunk: string) {
    if (!chunk) return;
    const parts = chunk.split("\n");
    const complete = chunk.endsWith("\n");
    if (complete) parts.pop();
    if (!endOfLine && lines.value.length) {
      const last = lines.value[lines.value.length - 1];
      last.text += parts.shift() ?? "";
      last.level = detectLevel(last.text);
      last.source = detectSource(last.text);
      noteSource(last.source);
    }
    const base = lines.value.length;
    for (const [i, text] of parts.entries()) {
      const source = detectSource(text);
      noteSource(source);
      lines.value.push({
        id: base + i,
        text,
        level: detectLevel(text),
        source,
      });
    }
    // a page left open for days must not keep every line ever streamed
    if (lines.value.length > maxLines) {
      lines.value.splice(0, lines.value.length - maxLines);
    }
    endOfLine = complete;
    // the offset is in bytes, the text in characters
    skip += new Blob([chunk]).size;
  }

  async function fetch(): Promise<void> {
    if (fetching) return;
    fetching = true;
    try {
      append(await getLogger({ skip }));
    } catch {
      // the next tick tries again
    } finally {
      fetching = false;
    }
  }

  function start() {
    stop();
    timer = setInterval(fetch, interval.value * 1000);
  }
  function stop() {
    if (timer) clearInterval(timer);
    timer = null;
  }
  watch(interval, start);
  onScopeDispose(stop);

  return {
    lines,
    sources,
    interval,
    fetch,
    start,
    stop,
    append,
    count: computed(() => lines.value.length),
  };
}
