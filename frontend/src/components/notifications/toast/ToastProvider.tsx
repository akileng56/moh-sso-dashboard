import { createContext, useContext, useState, type ReactNode } from "react";
import { ToastNotification } from "@carbon/react";

type ToastKind = "success" | "error" | "info" | "warning";

type Toast = {
  id: string;
  kind: ToastKind;
  title: string;
  subtitle?: string;
};

type ToastContextType = {
  push: (toast: Omit<Toast, "id">) => void;
};

const ToastContext = createContext<ToastContextType | null>(null);

export function ToastProvider({ children }: { children: ReactNode }) {
  const [toasts, setToasts] = useState<Toast[]>([]);

  const push = (toast: Omit<Toast, "id">) => {
    const id = crypto.randomUUID();
    setToasts((t) => [...t, { ...toast, id }]);

    setTimeout(() => {
      setToasts((t) => t.filter((x) => x.id !== id));
    }, 5000);
  };

  return (
    <ToastContext.Provider value={{ push }}>
      {children}

      <div
        style={{
          position: "fixed",
          top: 16,
          right: 16,
          zIndex: 9999,
          display: "flex",
          flexDirection: "column",
          gap: 8,
        }}
      >
        {toasts.map((t) => (
          <ToastNotification
            key={t.id}
            kind={t.kind}
            title={t.title}
            subtitle={t.subtitle}
            lowContrast
          />
        ))}
      </div>
    </ToastContext.Provider>
  );
}

export function useToastContext() {
  const ctx = useContext(ToastContext);
  if (!ctx) {
    throw new Error("useToastContext must be used inside ToastProvider");
  }
  return ctx;
}
