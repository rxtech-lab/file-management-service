"use client";

import { toast } from "sonner";
import { organizeFileAction, organizeFolderAction } from "@/lib/actions/agent-actions";
import type { AgentEvent } from "@/lib/api/types";
import {
  Sparkles,
  Wrench,
  Brain,
  Check,
  AlertCircle,
} from "lucide-react";
import ReactMarkdown from "react-markdown";

/**
 * Renders markdown content with proper styling
 */
function MarkdownContent({ content }: { content: string }) {
  return (
    <div className="text-foreground prose prose-sm max-w-none">
      <ReactMarkdown
        components={{
          p: ({ children }) => <p className="my-1">{children}</p>,
          strong: ({ children }) => <strong className="font-semibold">{children}</strong>,
          em: ({ children }) => <em className="italic">{children}</em>,
          code: ({ children }) => (
            <code className="px-1 py-0.5 rounded bg-muted text-foreground font-mono text-xs">
              {children}
            </code>
          ),
          ul: ({ children }) => <ul className="list-disc pl-4 my-1">{children}</ul>,
          ol: ({ children }) => <ol className="list-decimal pl-4 my-1">{children}</ol>,
          li: ({ children }) => <li className="my-0.5">{children}</li>,
        }}
      >
        {content}
      </ReactMarkdown>
    </div>
  );
}

/**
 * Renders description with proper styling
 */
function Description({ content }: { content: string }) {
  return <div className="text-foreground">{content}</div>;
}

/**
 * Triggers AI file organization with real-time toast updates
 * @param fileId - The ID of the file to organize
 * @param fileName - Optional file name for better feedback
 * @param onComplete - Optional callback when organization completes
 */
export async function triggerAIOrganize(
  fileId: number,
  fileName?: string,
  onComplete?: () => void
): Promise<void> {
  let toastId: string | number | undefined;

  try {
    // Show initial loading toast
    toastId = toast.loading(
      <MarkdownContent content="Starting AI organization..." />,
      {
        description: fileName ? <Description content={`Organizing "${fileName}"`} /> : undefined,
        icon: <Sparkles className="h-4 w-4 text-purple-500" />,
        classNames: {
          title: "text-foreground",
          description: "text-foreground",
        },
      }
    );

    // Trigger the organize endpoint
    const result = await organizeFileAction(fileId);
    if (!result.success) {
      toast.error(<MarkdownContent content="Failed to start AI agent" />, {
        id: toastId,
        description: <Description content={result.error || "Unknown error"} />,
        icon: <AlertCircle className="h-4 w-4" />,
        classNames: {
          title: "text-foreground",
          description: "text-foreground",
        },
      });
      return;
    }

    // Connect to SSE stream
    const eventSource = new EventSource(`/api/agent-stream/${fileId}`);
    let lastMessage = "Starting AI organization...";

    eventSource.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data) as AgentEvent;

        // Update the toast based on event type
        switch (data.type) {
          case "connected":
            lastMessage = "Connected to AI agent";
            toast.loading(<MarkdownContent content={lastMessage} />, {
              id: toastId,
              icon: <Sparkles className="h-4 w-4 text-blue-500" />,
              classNames: {
                title: "text-foreground",
                description: "text-foreground",
              },
            });
            break;

          case "status":
            lastMessage = data.message;
            toast.loading(<MarkdownContent content={lastMessage} />, {
              id: toastId,
              description: fileName ? <Description content={fileName} /> : undefined,
              icon: <Sparkles className="h-4 w-4 text-blue-500" />,
              classNames: {
                title: "text-foreground",
                description: "text-foreground",
              },
            });
            break;

          case "tool_call":
            lastMessage = data.message;
            toast.loading(<MarkdownContent content={lastMessage} />, {
              id: toastId,
              description: data.tool ? (
                <Description content={`Using: ${data.tool}`} />
              ) : fileName ? (
                <Description content={fileName} />
              ) : undefined,
              icon: <Wrench className="h-4 w-4 text-orange-500" />,
              classNames: {
                title: "text-foreground",
                description: "text-foreground",
              },
            });
            break;

          case "tool_result":
            // Don't show tool results, just keep the loading state
            break;

          case "result":
          case "done":
            eventSource.close();
            toast.success(
              <MarkdownContent content={data.message || "File organized successfully!"} />,
              {
                id: toastId,
                description: fileName ? <Description content={fileName} /> : undefined,
                icon: <Check className="h-4 w-4" />,
                duration: 4000,
                classNames: {
                  title: "text-foreground",
                  description: "text-foreground",
                },
              }
            );
            if (onComplete) {
              onComplete();
            }
            break;

          case "error":
            eventSource.close();
            toast.error(
              <MarkdownContent content={data.message || "AI organization failed"} />,
              {
                id: toastId,
                description: fileName ? <Description content={fileName} /> : undefined,
                icon: <AlertCircle className="h-4 w-4" />,
                classNames: {
                  title: "text-foreground",
                  description: "text-foreground",
                },
              }
            );
            break;
        }
      } catch (error) {
        console.error("Failed to parse agent event:", error);
      }
    };

    eventSource.onerror = () => {
      eventSource.close();
      toast.error(<MarkdownContent content="Connection to AI agent lost" />, {
        id: toastId,
        description: fileName ? <Description content={fileName} /> : undefined,
        icon: <AlertCircle className="h-4 w-4" />,
        classNames: {
          title: "text-foreground",
          description: "text-foreground",
        },
      });
    };
  } catch (error) {
    toast.error(<MarkdownContent content="Failed to start AI organization" />, {
      id: toastId,
      description: (
        <Description content={error instanceof Error ? error.message : "Unknown error"} />
      ),
      icon: <AlertCircle className="h-4 w-4" />,
      classNames: {
        title: "text-foreground",
        description: "text-foreground",
      },
    });
  }
}

/**
 * Triggers AI folder organization with real-time toast updates
 * @param folderId - The ID of the folder to organize
 * @param folderName - Optional folder name for better feedback
 * @param onComplete - Optional callback when organization completes
 */
export async function triggerAIOrganizeFolder(
  folderId: number,
  folderName?: string,
  onComplete?: () => void
): Promise<void> {
  let toastId: string | number | undefined;

  try {
    // Show initial loading toast
    toastId = toast.loading(
      <MarkdownContent content="Starting AI folder organization..." />,
      {
        description: folderName ? <Description content={`Organizing "${folderName}"`} /> : undefined,
        icon: <Sparkles className="h-4 w-4 text-purple-500" />,
        classNames: {
          title: "text-foreground",
          description: "text-foreground",
        },
      }
    );

    // Trigger the organize endpoint
    const result = await organizeFolderAction(folderId);
    if (!result.success) {
      toast.error(<MarkdownContent content="Failed to start AI agent" />, {
        id: toastId,
        description: <Description content={result.error || "Unknown error"} />,
        icon: <AlertCircle className="h-4 w-4" />,
        classNames: {
          title: "text-foreground",
          description: "text-foreground",
        },
      });
      return;
    }

    // Connect to SSE stream
    const eventSource = new EventSource(`/api/folder-agent-stream/${folderId}`);
    let lastMessage = "Starting AI folder organization...";

    eventSource.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data) as AgentEvent;

        // Update the toast based on event type
        switch (data.type) {
          case "connected":
            lastMessage = "Connected to AI agent";
            toast.loading(<MarkdownContent content={lastMessage} />, {
              id: toastId,
              icon: <Sparkles className="h-4 w-4 text-blue-500" />,
              classNames: {
                title: "text-foreground",
                description: "text-foreground",
              },
            });
            break;

          case "status":
            lastMessage = data.message;
            toast.loading(<MarkdownContent content={lastMessage} />, {
              id: toastId,
              description: folderName ? <Description content={folderName} /> : undefined,
              icon: <Sparkles className="h-4 w-4 text-blue-500" />,
              classNames: {
                title: "text-foreground",
                description: "text-foreground",
              },
            });
            break;

          case "tool_call":
            lastMessage = data.message;
            toast.loading(<MarkdownContent content={lastMessage} />, {
              id: toastId,
              description: data.tool ? (
                <Description content={`Using: ${data.tool}`} />
              ) : folderName ? (
                <Description content={folderName} />
              ) : undefined,
              icon: <Wrench className="h-4 w-4 text-orange-500" />,
              classNames: {
                title: "text-foreground",
                description: "text-foreground",
              },
            });
            break;

          case "tool_result":
            // Don't show tool results, just keep the loading state
            break;

          case "result":
          case "done":
            eventSource.close();
            toast.success(
              <MarkdownContent content={data.message || "Folder organized successfully!"} />,
              {
                id: toastId,
                description: folderName ? <Description content={folderName} /> : undefined,
                icon: <Check className="h-4 w-4" />,
                duration: 4000,
                classNames: {
                  title: "text-foreground",
                  description: "text-foreground",
                },
              }
            );
            if (onComplete) {
              onComplete();
            }
            break;

          case "error":
            eventSource.close();
            toast.error(
              <MarkdownContent content={data.message || "AI folder organization failed"} />,
              {
                id: toastId,
                description: folderName ? <Description content={folderName} /> : undefined,
                icon: <AlertCircle className="h-4 w-4" />,
                classNames: {
                  title: "text-foreground",
                  description: "text-foreground",
                },
              }
            );
            break;
        }
      } catch (error) {
        console.error("Failed to parse agent event:", error);
      }
    };

    eventSource.onerror = () => {
      eventSource.close();
      toast.error(<MarkdownContent content="Connection to AI agent lost" />, {
        id: toastId,
        description: folderName ? <Description content={folderName} /> : undefined,
        icon: <AlertCircle className="h-4 w-4" />,
        classNames: {
          title: "text-foreground",
          description: "text-foreground",
        },
      });
    };
  } catch (error) {
    toast.error(<MarkdownContent content="Failed to start AI folder organization" />, {
      id: toastId,
      description: (
        <Description content={error instanceof Error ? error.message : "Unknown error"} />
      ),
      icon: <AlertCircle className="h-4 w-4" />,
      classNames: {
        title: "text-foreground",
        description: "text-foreground",
      },
    });
  }
}
