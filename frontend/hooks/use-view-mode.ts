"use client";

import { useState, useEffect } from "react";

type ViewMode = "grid" | "list";

const VIEW_MODE_KEY = "file-manager-view-mode";

export function useViewMode() {
  const [viewMode, setViewMode] = useState<ViewMode>("grid");

  // Load from localStorage on mount
  useEffect(() => {
    const saved = localStorage.getItem(VIEW_MODE_KEY);
    if (saved === "grid" || saved === "list") {
      setViewMode(saved);
    }
  }, []);

  // Save to localStorage when changed
  const updateViewMode = (mode: ViewMode) => {
    setViewMode(mode);
    localStorage.setItem(VIEW_MODE_KEY, mode);
  };

  return {
    viewMode,
    setViewMode: updateViewMode,
    isGrid: viewMode === "grid",
    isList: viewMode === "list",
  };
}
