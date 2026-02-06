"use client";

import { useState, useRef, useCallback, useEffect } from "react";

export type SaveStatus = "idle" | "saving" | "saved" | "error";

export interface UseAutoSaveOptions {
  onSave: (content: string) => Promise<void>;
  debounceMs?: number;
  enabled?: boolean;
}

export interface UseAutoSaveReturn {
  saveStatus: SaveStatus;
  triggerSave: (content: string) => void;
  forceSave: () => Promise<void>;
  lastSavedAt: Date | null;
  error: string | null;
}

export function useAutoSave({
  onSave,
  debounceMs = 1500,
  enabled = true,
}: UseAutoSaveOptions): UseAutoSaveReturn {
  const [saveStatus, setSaveStatus] = useState<SaveStatus>("idle");
  const [lastSavedAt, setLastSavedAt] = useState<Date | null>(null);
  const [error, setError] = useState<string | null>(null);

  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const pendingContentRef = useRef<string | null>(null);
  const onSaveRef = useRef(onSave);
  onSaveRef.current = onSave;

  const doSave = useCallback(async (content: string) => {
    setSaveStatus("saving");
    setError(null);
    try {
      await onSaveRef.current(content);
      setSaveStatus("saved");
      setLastSavedAt(new Date());
      pendingContentRef.current = null;
    } catch (err) {
      const msg =
        err instanceof Error ? err.message : "Save failed";
      setSaveStatus("error");
      setError(msg);
    }
  }, []);

  const triggerSave = useCallback(
    (content: string) => {
      if (!enabled) return;
      pendingContentRef.current = content;

      if (timerRef.current) {
        clearTimeout(timerRef.current);
      }

      timerRef.current = setTimeout(() => {
        timerRef.current = null;
        doSave(content);
      }, debounceMs);
    },
    [enabled, debounceMs, doSave],
  );

  const forceSave = useCallback(async () => {
    if (timerRef.current) {
      clearTimeout(timerRef.current);
      timerRef.current = null;
    }
    if (pendingContentRef.current !== null) {
      await doSave(pendingContentRef.current);
    }
  }, [doSave]);

  // Flush on unmount
  useEffect(() => {
    return () => {
      if (timerRef.current) {
        clearTimeout(timerRef.current);
      }
    };
  }, []);

  // Reset saved status after 3 seconds
  useEffect(() => {
    if (saveStatus === "saved") {
      const timeout = setTimeout(() => setSaveStatus("idle"), 3000);
      return () => clearTimeout(timeout);
    }
  }, [saveStatus]);

  return { saveStatus, triggerSave, forceSave, lastSavedAt, error };
}
