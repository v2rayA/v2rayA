import ConfirmDialog from "@/components/hosts/ConfirmDialog.vue";
import { openDialog } from "./useDialog";

export interface ConfirmOptions {
  title?: string;
  message: string;
  confirmText?: string;
  cancelText?: string;
  destructive?: boolean;
}

/** useConfirm resolves true on confirm, false on cancel, ESC or the scrim. */
export function useConfirm() {
  return (options: ConfirmOptions): Promise<boolean> =>
    openDialog<boolean>(
      ConfirmDialog,
      { ...options },
      { width: 440 },
    ).result.then((r) => r === true);
}

export interface PromptOptions extends Omit<ConfirmOptions, "destructive"> {
  input: {
    label?: string;
    placeholder?: string;
    value?: string;
    maxlength?: number;
    validate?: (v: string) => string | true;
  };
}

/** usePrompt resolves the entered text, or null on cancel. */
export function usePrompt() {
  return (options: PromptOptions): Promise<string | null> =>
    openDialog<string | null>(
      ConfirmDialog,
      { ...options },
      { width: 440 },
    ).result.then((r) => (typeof r === "string" ? r : null));
}
