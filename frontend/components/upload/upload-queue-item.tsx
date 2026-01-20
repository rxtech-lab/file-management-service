"use client";

import { X, RotateCcw, Check, Loader2, AlertCircle, Sparkles, Receipt } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Progress } from "@/components/ui/progress";
import {
  FileTypeIcon,
  getFileTypeColor,
} from "@/components/files/file-type-icon";
import type { UploadQueueItem } from "@/hooks/use-upload-queue";
import { cn } from "@/lib/utils";

interface UploadQueueItemComponentProps {
  item: UploadQueueItem;
  onRemove: () => void;
  onRetry: () => void;
}

function formatFileSize(bytes: number): string {
  const units = ["B", "KB", "MB", "GB"];
  let size = bytes;
  let unitIndex = 0;
  while (size >= 1024 && unitIndex < units.length - 1) {
    size /= 1024;
    unitIndex++;
  }
  return `${size.toFixed(1)} ${units[unitIndex]}`;
}

function getStatusText(item: UploadQueueItem): string {
  switch (item.status) {
    case "queued":
      return "Queued...";
    case "uploading":
      return "Uploading...";
    case "creating":
      return "Creating file record...";
    case "processing":
      // Show processing message if available
      if (item.processingMessage) {
        return item.processingMessage;
      }
      if (item.processingStatus === "processing") {
        return "Processing...";
      }
      return "Waiting for processing...";
    case "completed":
      return item.aiSummary
        ? `AI: ${item.aiSummary.slice(0, 50)}...`
        : "Complete";
    case "failed":
      return item.error || "Failed";
    default:
      return "";
  }
}

export function UploadQueueItemComponent({
  item,
  onRemove,
  onRetry,
}: UploadQueueItemComponentProps) {
  const isInProgress =
    item.status === "queued" ||
    item.status === "uploading" ||
    item.status === "creating" ||
    item.status === "processing";

  return (
    <div className="p-3 border-b last:border-b-0 w-full overflow-hidden">
      <div className="flex items-start gap-3 w-full">
        {/* File icon */}
        <div
          className={cn("mt-0.5", getFileTypeColor(undefined, item.file.type))}
        >
          <FileTypeIcon mimeType={item.file.type} className="h-5 w-5" />
        </div>

        {/* Content */}
        <div className="flex-1 min-w-0 overflow-hidden">
          <div className="flex items-center justify-between gap-2">
            <p className="font-medium text-sm truncate">{item.file.name}</p>

            {/* Status icon / Actions */}
            <div className="flex items-center gap-1 shrink-0">
              {item.status === "completed" && (
                <Check className="h-4 w-4 text-green-500" />
              )}
              {item.status === "failed" && (
                <>
                  <AlertCircle className="h-4 w-4 text-destructive" />
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-6 w-6"
                    onClick={onRetry}
                  >
                    <RotateCcw className="h-3 w-3" />
                  </Button>
                </>
              )}
              {isInProgress && (
                <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />
              )}
              <Button
                variant="ghost"
                size="icon"
                className="h-6 w-6"
                onClick={onRemove}
              >
                <X className="h-3 w-3" />
              </Button>
            </div>
          </div>

          {/* Progress bar */}
          {isInProgress && (
            <Progress value={item.progress} className="h-1 mt-2" />
          )}

          {/* Status text */}
          <p
            className={cn(
              "text-xs mt-1 max-w-[400px]",
              item.status === "failed"
                ? "text-destructive"
                : "text-muted-foreground",
            )}
            style={{ overflowWrap: "break-word", wordBreak: "break-word" }}
          >
            {item.status === "processing" && item.processingMessage && (
              item.processingSource === "invoice" ? (
                <Receipt className="inline h-3 w-3 mr-1 text-blue-500 animate-pulse" />
              ) : item.processingSource === "agent" ? (
                <Sparkles className="inline h-3 w-3 mr-1 text-purple-500 animate-pulse" />
              ) : null
            )}
            {getStatusText(item)}
          </p>

          {/* File size */}
          <p className="text-xs text-muted-foreground">
            {formatFileSize(item.file.size)}
          </p>
        </div>
      </div>
    </div>
  );
}
