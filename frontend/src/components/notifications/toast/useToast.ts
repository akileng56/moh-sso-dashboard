import { useToastContext } from "./ToastProvider";

export function useToast() {
  const { push } = useToastContext();

  return {
    success: (title: string, subtitle?: string) =>
      push({ kind: "success", title, subtitle }),

    error: (title: string, subtitle?: string) =>
      push({ kind: "error", title, subtitle }),

    info: (title: string, subtitle?: string) =>
      push({ kind: "info", title, subtitle }),

    warning: (title: string, subtitle?: string) =>
      push({ kind: "warning", title, subtitle }),
  };
}
