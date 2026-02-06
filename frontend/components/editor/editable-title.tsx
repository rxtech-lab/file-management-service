"use client";

import { useState, useRef, useCallback, useEffect } from "react";
import { Loader2 } from "lucide-react";
import { Input } from "@/components/ui/input";
import { cn } from "@/lib/utils";

interface EditableTitleProps {
  value: string;
  onSave: (value: string) => Promise<void>;
  className?: string;
}

export function EditableTitle({ value, onSave, className }: EditableTitleProps) {
  const [isEditing, setIsEditing] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [editValue, setEditValue] = useState(value);
  const [width, setWidth] = useState<number | undefined>(undefined);
  const inputRef = useRef<HTMLInputElement>(null);
  const measureRef = useRef<HTMLSpanElement>(null);

  const displayText = isEditing ? editValue : value;

  // Measure text width whenever displayText changes
  useEffect(() => {
    if (measureRef.current) {
      const measured = measureRef.current.scrollWidth;
      // Add extra space for the spinner when saving
      const spinnerSpace = isSaving ? 24 : 0;
      setWidth(Math.min(measured + 4 + spinnerSpace, 300));
    }
  }, [displayText, isSaving]);

  useEffect(() => {
    setEditValue(value);
  }, [value]);

  useEffect(() => {
    if (isEditing && inputRef.current) {
      inputRef.current.focus();
      inputRef.current.select();
    }
  }, [isEditing]);

  const handleSave = useCallback(async () => {
    const trimmed = editValue.trim();
    if (trimmed && trimmed !== value) {
      setIsSaving(true);
      try {
        await onSave(trimmed);
      } finally {
        setIsSaving(false);
      }
    } else {
      setEditValue(value);
    }
    setIsEditing(false);
  }, [editValue, value, onSave]);

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.key === "Enter") {
        e.preventDefault();
        handleSave();
      } else if (e.key === "Escape") {
        setEditValue(value);
        setIsEditing(false);
      }
    },
    [handleSave, value],
  );

  return (
    <div className={cn("relative inline-flex items-center", className)}>
      {/* Hidden span to measure text width */}
      <span
        ref={measureRef}
        className="invisible absolute whitespace-pre text-base font-semibold px-1.5 pointer-events-none"
        aria-hidden="true"
      >
        {displayText || "\u00A0"}
      </span>

      {isEditing ? (
        <div className="relative inline-flex items-center" style={{ width: width ? `${width}px` : "auto" }}>
          <Input
            ref={inputRef}
            value={editValue}
            onChange={(e) => setEditValue(e.target.value)}
            onBlur={handleSave}
            onKeyDown={handleKeyDown}
            disabled={isSaving}
            className={cn(
              "h-8 text-base font-semibold border-none shadow-none focus-visible:ring-1 px-1.5",
              isSaving && "pr-7",
            )}
          />
          {isSaving && (
            <Loader2 className="absolute right-1.5 h-4 w-4 animate-spin text-muted-foreground" />
          )}
        </div>
      ) : (
        <button
          onClick={() => setIsEditing(true)}
          style={{ width: width ? `${width}px` : "auto" }}
          className="text-base font-semibold px-1.5 py-1 rounded-md hover:bg-accent transition-colors text-left truncate cursor-text text-foreground"
        >
          {value}
        </button>
      )}
    </div>
  );
}
