"use client";

import { useEffect, useState, useCallback, useRef } from "react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import {
  Loader2,
  Check,
  AlertCircle,
  Sparkles,
  Wrench,
  MessageSquare,
  Brain,
} from "lucide-react";
import { cn } from "@/lib/utils";
import type { AgentEvent } from "@/lib/api/types";
import { organizeFileAction } from "@/lib/actions/agent-actions";

interface AIOrganizeDialogProps {
  fileId: number | null;
  fileName?: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onComplete?: () => void;
}

export function AIOrganizeDialog({
  fileId,
  fileName,
  open,
  onOpenChange,
  onComplete,
}: AIOrganizeDialogProps) {
  const [events, setEvents] = useState<AgentEvent[]>([]);
  const [isComplete, setIsComplete] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const eventSourceRef = useRef<EventSource | null>(null);
  const scrollRef = useRef<HTMLDivElement>(null);

  const cleanup = useCallback(() => {
    if (eventSourceRef.current) {
      eventSourceRef.current.close();
      eventSourceRef.current = null;
    }
  }, []);

  useEffect(() => {
    if (!open || !fileId) {
      cleanup();
      return;
    }

    // Reset state when opening
    setEvents([]);
    setIsComplete(false);
    setError(null);

    const startAgent = async () => {
      try {
        // First, trigger the organize endpoint to start the agent
        const result = await organizeFileAction(fileId);
        if (!result.success) {
          setError(result.error || "Failed to start AI agent");
          return;
        }

        // Use Next.js API route to proxy the SSE stream with authentication
        const eventSource = new EventSource(`/api/agent-stream/${fileId}`);
        eventSourceRef.current = eventSource;

        eventSource.onmessage = (event) => {
          try {
            const data = JSON.parse(event.data) as AgentEvent;
            setEvents((prev) => [...prev, data]);

            if (data.type === "result" || data.type === "done") {
              setIsComplete(true);
              cleanup();
              if (data.type === "done" && onComplete) {
                onComplete();
              }
            }

            if (data.type === "error") {
              setError(data.message);
              cleanup();
            }
          } catch {
            console.error("Failed to parse agent event");
          }
        };

        eventSource.onerror = () => {
          setError("Connection to agent lost");
          cleanup();
        };
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to start AI agent");
      }
    };

    startAgent();

    return cleanup;
  }, [open, fileId, cleanup, onComplete]);

  // Auto-scroll to bottom when new events arrive
  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [events]);

  const handleClose = () => {
    cleanup();
    onOpenChange(false);
  };

  const getEventIcon = (type: AgentEvent["type"]) => {
    switch (type) {
      case "connected":
        return <Sparkles className="h-4 w-4 text-blue-500" />;
      case "thinking":
        return <Brain className="h-4 w-4 text-purple-500 animate-pulse" />;
      case "tool_call":
        return <Wrench className="h-4 w-4 text-orange-500" />;
      case "tool_result":
        return <MessageSquare className="h-4 w-4 text-gray-500" />;
      case "status":
        return <Loader2 className="h-4 w-4 text-blue-500 animate-spin" />;
      case "result":
        return <Check className="h-4 w-4 text-green-500" />;
      case "done":
        return <Check className="h-4 w-4 text-green-500" />;
      case "error":
        return <AlertCircle className="h-4 w-4 text-red-500" />;
      default:
        return <Loader2 className="h-4 w-4 animate-spin" />;
    }
  };

  const getEventStyle = (type: AgentEvent["type"]) => {
    switch (type) {
      case "error":
        return "bg-red-50 border-red-200 text-red-800";
      case "result":
      case "done":
        return "bg-green-50 border-green-200 text-green-800";
      case "tool_call":
        return "bg-orange-50 border-orange-200 text-orange-800";
      case "tool_result":
        return "bg-gray-50 border-gray-200 text-gray-600 text-sm";
      default:
        return "bg-white border-gray-200";
    }
  };

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Sparkles className="h-5 w-5 text-purple-500" />
            AI Organizing File
          </DialogTitle>
          <DialogDescription>
            {fileName
              ? `Organizing "${fileName}"...`
              : "AI is analyzing and organizing your file..."}
          </DialogDescription>
        </DialogHeader>

        <ScrollArea className="h-[300px] pr-4" ref={scrollRef}>
          <div className="space-y-2">
            {events.length === 0 && !error && (
              <div className="flex items-center justify-center py-8 text-muted-foreground">
                <Loader2 className="h-6 w-6 animate-spin mr-2" />
                Connecting to AI agent...
              </div>
            )}

            {events.map((event, index) => (
              <div
                key={index}
                className={cn(
                  "flex items-start gap-2 p-2 rounded-md border",
                  getEventStyle(event.type)
                )}
              >
                <div className="mt-0.5">{getEventIcon(event.type)}</div>
                <div className="flex-1 min-w-0">
                  <p className="break-words">{event.message}</p>
                  {event.tool && (
                    <p className="text-xs text-muted-foreground mt-1">
                      Tool: {event.tool}
                    </p>
                  )}
                </div>
              </div>
            ))}

            {error && (
              <div className="flex items-start gap-2 p-2 rounded-md border bg-red-50 border-red-200 text-red-800">
                <AlertCircle className="h-4 w-4 mt-0.5" />
                <p>{error}</p>
              </div>
            )}
          </div>
        </ScrollArea>

        <div className="flex justify-end gap-2 pt-4 border-t">
          {isComplete || error ? (
            <Button onClick={handleClose}>
              {isComplete ? "Done" : "Close"}
            </Button>
          ) : (
            <Button variant="outline" onClick={handleClose}>
              Cancel
            </Button>
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}
