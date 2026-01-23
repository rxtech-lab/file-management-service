"use client";

import { useState, useCallback, useEffect, useRef } from "react";
import { useRouter } from "next/navigation";
import { getPresignedURLAction } from "@/lib/actions/upload-actions";
import {
  createFileAction,
  getFileAction,
} from "@/lib/actions/file-actions";
import type { ProcessingStatus, ProcessingEvent } from "@/lib/api/types";

export interface UploadQueueItem {
  id: string;
  file: File;
  status:
  | "queued"
  | "uploading"
  | "creating"
  | "processing"
  | "completed"
  | "failed";
  progress: number;
  s3Key?: string;
  fileId?: number;
  processingStatus?: ProcessingStatus;
  aiSummary?: string;
  error?: string;
  // Processing status for real-time updates
  processingMessage?: string;
  processingSource?: string;
  processingStep?: ProcessingEvent["type"];
}

export function useUploadQueue(folderId: number | null) {
  const router = useRouter();
  const [items, setItems] = useState<UploadQueueItem[]>([]);
  const [isExpanded, setIsExpanded] = useState(false);
  const pollingIntervals = useRef<Map<string, NodeJS.Timeout>>(new Map());
  const agentEventSources = useRef<Map<string, EventSource>>(new Map());

  // Subscribe to unified process-stream for real-time status updates
  const subscribeToProcessStream = useCallback(
    (itemId: string, fileId: number) => {
      // Use Next.js API route to proxy the SSE stream with authentication
      const eventSource = new EventSource(`/api/process-stream/${fileId}`);

      agentEventSources.current.set(itemId, eventSource);

      eventSource.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data) as ProcessingEvent;

          // Update processing status in the queue item
          setItems((prev) =>
            prev.map((item) =>
              item.id === itemId
                ? {
                  ...item,
                  processingMessage: data.message,
                  processingSource: data.source,
                  processingStep: data.type,
                }
                : item
            )
          );

          // Close connection when done
          if (data.type === "done" || data.type === "error") {
            eventSource.close();
            agentEventSources.current.delete(itemId);
            // Refresh file data after processing completes
            router.refresh();
          }
        } catch {
          // Ignore parse errors
        }
      };

      eventSource.onerror = () => {
        eventSource.close();
        agentEventSources.current.delete(itemId);
      };
    },
    [router]
  );

  const updateItem = useCallback(
    (id: string, updates: Partial<UploadQueueItem>) => {
      setItems((prev) =>
        prev.map((item) => (item.id === id ? { ...item, ...updates } : item)),
      );
    },
    [],
  );

  const processUpload = useCallback(
    async (item: UploadQueueItem) => {
      try {
        // Step 1: Get presigned URL from backend
        updateItem(item.id, { status: "uploading", progress: 5 });

        const presignedResult = await getPresignedURLAction(
          item.file.name,
          item.file.type || "application/octet-stream"
        );

        if (!presignedResult.success || !presignedResult.data) {
          throw new Error(presignedResult.error || "Failed to get upload URL");
        }

        const { upload_url, key: s3Key } = presignedResult.data;
        updateItem(item.id, { progress: 15 });

        // Step 2: Upload directly to S3
        const uploadResponse = await fetch(upload_url, {
          method: "PUT",
          body: item.file,
          headers: {
            "Content-Type": item.file.type || "application/octet-stream",
          },
        });

        if (!uploadResponse.ok) {
          throw new Error(`S3 upload failed: ${uploadResponse.status}`);
        }

        updateItem(item.id, {
          status: "creating",
          progress: 40,
          s3Key,
        });

        // Step 2: Create file record
        const createResult = await createFileAction({
          title: item.file.name,
          s3_key: s3Key,
          original_filename: item.file.name,
          mime_type: item.file.type,
          size: item.file.size,
          folder_id: folderId,
        });

        if (!createResult.success || !createResult.data) {
          throw new Error(createResult.error || "Failed to create file record");
        }

        const fileId = createResult.data.id;
        updateItem(item.id, {
          status: "processing",
          progress: 60,
          fileId,
          processingStatus: "pending",
        });

        // Step 3: Subscribe to unified process-stream (triggers processing AND streams events)
        subscribeToProcessStream(item.id, fileId);

        // Step 4: Start polling for processing status
        const pollInterval = setInterval(async () => {
          const fileResult = await getFileAction(fileId);

          if (fileResult.success && fileResult.data) {
            const file = fileResult.data;

            if (file.processing_status === "completed") {
              clearInterval(pollInterval);
              pollingIntervals.current.delete(item.id);
              updateItem(item.id, {
                status: "completed",
                progress: 100,
                processingStatus: "completed",
                aiSummary: file.summary,
              });
              router.refresh();
            } else if (file.processing_status === "failed") {
              clearInterval(pollInterval);
              pollingIntervals.current.delete(item.id);
              updateItem(item.id, {
                status: "failed",
                processingStatus: "failed",
                error: file.processing_error || "Processing failed",
              });
            } else {
              // Still processing
              updateItem(item.id, {
                processingStatus: file.processing_status,
                progress: file.processing_status === "processing" ? 80 : 70,
              });
            }
          }
        }, 2000); // Poll every 2 seconds

        pollingIntervals.current.set(item.id, pollInterval);
      } catch (error) {
        updateItem(item.id, {
          status: "failed",
          error: error instanceof Error ? error.message : "Upload failed",
        });
      }
    },
    [folderId, updateItem, router, subscribeToProcessStream],
  );

  const addFiles = useCallback(
    (files: File[]) => {
      const newItems: UploadQueueItem[] = files.map((file) => ({
        id: `${Date.now()}-${Math.random().toString(36).substr(2, 9)}`,
        file,
        status: "queued" as const,
        progress: 0,
      }));

      setItems((prev) => [...prev, ...newItems]);
      setIsExpanded(true);

      // Start processing each file
      newItems.forEach((item) => {
        processUpload(item);
      });
    },
    [processUpload],
  );

  const removeItem = useCallback((id: string) => {
    // Clear any polling interval
    const interval = pollingIntervals.current.get(id);
    if (interval) {
      clearInterval(interval);
      pollingIntervals.current.delete(id);
    }

    // Close any agent event source
    const eventSource = agentEventSources.current.get(id);
    if (eventSource) {
      eventSource.close();
      agentEventSources.current.delete(id);
    }

    setItems((prev) => prev.filter((item) => item.id !== id));
    setTimeout(() => {
      console.log("refreshing");
      router.refresh();
    }, 1000);
  }, [router]);

  const retryItem = useCallback(
    (id: string) => {
      const item = items.find((i) => i.id === id);
      if (item && item.status === "failed") {
        updateItem(id, { status: "queued", progress: 0, error: undefined });
        processUpload(item);
      }
    },
    [items, updateItem, processUpload],
  );

  const clearCompleted = useCallback(() => {
    setItems((prev) => prev.filter((item) => item.status !== "completed"));
  }, []);

  const toggleExpanded = useCallback(() => {
    setIsExpanded((prev) => !prev);
  }, []);

  // Cleanup polling intervals and agent event sources on unmount
  useEffect(() => {
    return () => {
      pollingIntervals.current.forEach((interval) => {
        clearInterval(interval);
      });
      agentEventSources.current.forEach((eventSource) => {
        eventSource.close();
      });
    };
  }, []);

  // Auto-expand when items are added
  useEffect(() => {
    if (items.length > 0) {
      setIsExpanded(true);
    }
  }, [items.length]);

  return {
    items,
    isExpanded,
    addFiles,
    removeItem,
    retryItem,
    clearCompleted,
    toggleExpanded,
  };
}
