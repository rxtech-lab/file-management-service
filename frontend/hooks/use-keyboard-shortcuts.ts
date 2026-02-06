"use client";

import { useEffect, useRef } from "react";
import type { ToolbarMenuItem } from "@/lib/plugins/types";

const isMac =
  typeof navigator !== "undefined" &&
  /Mac|iPod|iPhone|iPad/.test(navigator.userAgent);

function parseShortcut(shortcut: string) {
  const parts = shortcut
    .split("+")
    .map((p) => p.trim().toLowerCase());
  return {
    ctrl: parts.includes("ctrl") || parts.includes("mod"),
    shift: parts.includes("shift"),
    alt: parts.includes("alt"),
    meta: parts.includes("meta") || parts.includes("mod"),
    key: parts.filter(
      (p) => !["ctrl", "shift", "alt", "meta", "mod"].includes(p),
    )[0],
  };
}

function matchesShortcut(e: KeyboardEvent, shortcut: string): boolean {
  const parsed = parseShortcut(shortcut);
  if (!parsed.key) return false;

  // On Mac, "Ctrl" maps to Cmd (metaKey); on others it's ctrlKey
  const modKey = isMac ? e.metaKey : e.ctrlKey;
  if ((parsed.ctrl || parsed.meta) && !modKey) return false;
  if (!(parsed.ctrl || parsed.meta) && modKey) return false;
  if (parsed.shift !== e.shiftKey) return false;
  if (parsed.alt !== e.altKey) return false;

  return e.key.toLowerCase() === parsed.key;
}

export function useKeyboardShortcuts(items: ToolbarMenuItem[]): void {
  const itemsRef = useRef(items);
  itemsRef.current = items;

  useEffect(() => {
    function handleKeyDown(e: KeyboardEvent) {
      // Skip if user is in a non-monaco input
      const target = e.target as HTMLElement;
      if (
        target.tagName === "INPUT" ||
        target.tagName === "TEXTAREA" ||
        target.tagName === "SELECT"
      ) {
        return;
      }

      for (const item of itemsRef.current) {
        if (item.disabled || !item.shortcut) continue;
        if (matchesShortcut(e, item.shortcut)) {
          e.preventDefault();
          e.stopPropagation();
          item.onAction();
          return;
        }
      }
    }

    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, []);
}
