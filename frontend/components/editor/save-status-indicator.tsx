"use client";

import { Loader2, Check, AlertCircle } from "lucide-react";
import type { SaveStatus } from "@/hooks/use-auto-save";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";

interface SaveStatusIndicatorProps {
  status: SaveStatus;
  lastSavedAt: Date | null;
  error: string | null;
}

export function SaveStatusIndicator({
  status,
  lastSavedAt,
  error,
}: SaveStatusIndicatorProps) {
  if (status === "idle" && !lastSavedAt) return null;

  if (status === "saving") {
    return (
      <div className="flex items-center gap-1.5 text-muted-foreground text-sm">
        <Loader2 className="h-3.5 w-3.5 animate-spin" />
        <span>Saving...</span>
      </div>
    );
  }

  if (status === "saved") {
    return (
      <div className="flex items-center gap-1.5 text-muted-foreground text-sm">
        <Check className="h-3.5 w-3.5 text-green-600" />
        <span>Saved</span>
      </div>
    );
  }

  if (status === "error") {
    return (
      <Tooltip>
        <TooltipTrigger asChild>
          <div className="flex items-center gap-1.5 text-destructive text-sm cursor-help">
            <AlertCircle className="h-3.5 w-3.5" />
            <span>Save failed</span>
          </div>
        </TooltipTrigger>
        <TooltipContent>{error || "Unknown error"}</TooltipContent>
      </Tooltip>
    );
  }

  return null;
}
