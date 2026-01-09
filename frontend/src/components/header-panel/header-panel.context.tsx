// src/components/header-panel/header-panel.context.tsx
import React, { createContext, useContext, useState } from "react";
import { ReusableHeaderPanel } from "./ReusableHeaderPanel";

type HeaderPanelState = {
  isOpen: boolean;
  title?: string;
  content?: React.ReactNode;
};

type HeaderPanelContextType = {
  openPanel: (opts: { title?: string; content: React.ReactNode }) => void;
  closePanel: () => void;
};

const HeaderPanelContext = createContext<HeaderPanelContextType | null>(null);

export function HeaderPanelProvider({
  children,
}: {
  children: React.ReactNode;
}) {
  const [state, setState] = useState<HeaderPanelState>({
    isOpen: false,
  });

  const openPanel = ({
    title,
    content,
  }: {
    title?: string;
    content: React.ReactNode;
  }) => {
    setState({ isOpen: true, title, content });
  };

  const closePanel = () => {
    setState({ isOpen: false });
  };

  return (
    <HeaderPanelContext.Provider value={{ openPanel, closePanel }}>
      {children}
      <ReusableHeaderPanel {...state} onClose={closePanel} />
    </HeaderPanelContext.Provider>
  );
}

export function useHeaderPanel() {
  const ctx = useContext(HeaderPanelContext);
  if (!ctx) {
    throw new Error("useHeaderPanel must be used within HeaderPanelProvider");
  }
  return ctx;
}
