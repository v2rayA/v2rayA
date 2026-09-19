<script setup lang="ts">
// Renders the dialog stack from useDialog: each entry is a v-dialog around
// the caller's component; the component ends the dialog by emitting close
// with its result. Mounted once inside v-app by the shell.
import { closeDialog, dialogState } from "@/composables/useDialog";

function onModel(id: number, open: boolean) {
  if (!open) closeDialog(id);
}
</script>

<template>
  <v-dialog
    v-for="entry in dialogState.stack"
    :key="entry.id"
    :model-value="entry.open"
    :persistent="entry.persistent"
    :max-width="entry.width"
    scrollable
    @update:model-value="onModel(entry.id, $event)"
  >
    <component
      :is="entry.component"
      v-bind="entry.props"
      @close="(result: unknown) => closeDialog(entry.id, result)"
    />
  </v-dialog>
</template>
