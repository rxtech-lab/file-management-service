"use client";

import { ChevronDown, ChevronUp, X } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { ScrollArea } from "@/components/ui/scroll-area";
import { UploadQueueItemComponent } from "./upload-queue-item";
import type { UploadQueueItem } from "@/hooks/use-upload-queue";
import { cn } from "@/lib/utils";

interface UploadQueueProps {
  items: UploadQueueItem[];
  isExpanded: boolean;
  onToggleExpanded: () => void;
  onRemoveItem: (id: string) => void;
  onRetryItem: (id: string) => void;
  onClearCompleted: () => void;
}

export function UploadQueue({
  items,
  isExpanded,
  onToggleExpanded,
  onRemoveItem,
  onRetryItem,
  onClearCompleted,
}: UploadQueueProps) {
  if (items.length === 0) {
    return null;
  }

  const completedCount = items.filter((i) => i.status === "completed").length;
  const inProgressCount = items.filter(
    (i) =>
      i.status === "queued" ||
      i.status === "uploading" ||
      i.status === "creating" ||
      i.status === "processing",
  ).length;

  return (
    <Card
      className={cn(
        "fixed bottom-4 right-4 w-96 shadow-lg z-50 transition-all",
        !isExpanded && "h-auto",
      )}
    >
      <CardHeader
        className="py-2 px-3 border-b cursor-pointer"
        onClick={onToggleExpanded}
      >
        <div className="flex items-center justify-between">
          <CardTitle className="text-sm font-medium">
            Uploads ({items.length})
            {inProgressCount > 0 && (
              <span className="text-muted-foreground ml-1">
                - {inProgressCount} in progress
              </span>
            )}
          </CardTitle>
          <div className="flex items-center gap-1">
            <Button variant="ghost" size="icon" className="h-6 w-6">
              {isExpanded ? (
                <ChevronDown className="h-4 w-4" />
              ) : (
                <ChevronUp className="h-4 w-4" />
              )}
            </Button>
          </div>
        </div>
      </CardHeader>

      {isExpanded && (
        <CardContent className="p-0">
          <ScrollArea className="max-h-64">
            {items.map((item) => (
              <UploadQueueItemComponent
                key={item.id}
                item={item}
                onRemove={() => onRemoveItem(item.id)}
                onRetry={() => onRetryItem(item.id)}
              />
            ))}
          </ScrollArea>

          {completedCount > 0 && (
            <div className="p-2 border-t">
              <Button
                variant="ghost"
                size="sm"
                className="w-full text-xs"
                onClick={onClearCompleted}
              >
                Clear completed ({completedCount})
              </Button>
            </div>
          )}
        </CardContent>
      )}
    </Card>
  );
}
