import { computed, ref, shallowRef, triggerRef } from "vue";
import type { TrafficMessage } from "@/api/types";

const capacity = 30;

/** createTraffic retains the last thirty one-second traffic frames, oldest sample first. */
export function createTraffic() {
  const up = ref(0);
  const down = ref(0);
  const upTotal = ref(0);
  const downTotal = ref(0);
  const upSamples = shallowRef<number[]>([]);
  const downSamples = shallowRef<number[]>([]);
  let next = 0;

  // oldest first, padded with zeros to the window so the chart is a flat
  // line before the first frame and keeps one time scale as frames come in
  const ordered = (samples: number[]) => [
    ...new Array<number>(capacity - samples.length).fill(0),
    ...samples.map((_, i) => samples[(next + i) % samples.length]),
  ];
  const upSeries = computed(() => ordered(upSamples.value));
  const downSeries = computed(() => ordered(downSamples.value));

  function feed({ body }: TrafficMessage) {
    up.value = body.up;
    down.value = body.down;
    upTotal.value = body.upTotal;
    downTotal.value = body.downTotal;
    upSamples.value[next] = body.up;
    downSamples.value[next] = body.down;
    next = (next + 1) % capacity;
    triggerRef(upSamples);
    triggerRef(downSamples);
  }

  function reset() {
    next = 0;
    up.value = down.value = upTotal.value = downTotal.value = 0;
    upSamples.value = [];
    downSamples.value = [];
  }

  return { feed, reset, up, down, upTotal, downTotal, upSeries, downSeries };
}

// One stream for the page: the shell feeds it from the message socket
// and resets it with the session; the dashboard reads it.
const shared = createTraffic();

export function useTraffic() {
  return shared;
}
