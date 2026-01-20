"use client";

import { useEffect, useRef, useState, useMemo } from "react";
import { useRouter } from "next/navigation";
import { Chat, useChat } from "@ai-sdk/react";
import { DefaultChatTransport, UIMessage as AIUIMessage } from "ai";
import { motion, AnimatePresence } from "framer-motion";
import {
  ArrowLeft,
  ArrowRight,
  ArrowUp,
  Loader2,
  Sparkles,
  Square,
  User,
  Bot,
  Trash2,
} from "lucide-react";
import Markdown from "react-markdown";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import {
  FileTypeIcon,
  getFileTypeColor,
} from "@/components/files/file-type-icon";
import type { FileType } from "@/lib/api/types";
import { cn } from "@/lib/utils";

// Type for the display_files tool output
interface DisplayFilesOutput {
  files: Array<{
    id: number;
    title: string;
    description?: string;
    file_type: string;
    mime_type?: string;
    folder_id?: number | null;
    folder_name?: string;
  }>;
  summary?: string;
}

// Type for tool parts with output
interface ToolPartWithOutput {
  type: string;
  toolCallId?: string;
  state?: string;
  output?: unknown;
}

interface SearchAgentProps {
  initialQuery: string;
  initialMessages?: AIUIMessage[];
  onMessagesChange?: (messages: AIUIMessage[]) => void;
  onClearHistory?: () => void;
  onBack: () => void;
  onClose: () => void;
}

export function SearchAgent({
  initialQuery,
  initialMessages,
  onMessagesChange,
  onClearHistory,
  onBack,
  onClose,
}: SearchAgentProps) {
  const router = useRouter();
  const scrollRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);
  const initialQuerySent = useRef(false);
  const [input, setInput] = useState("");

  // Create a stable Chat instance using useMemo, with initial messages if provided
  const chat = useMemo(
    () =>
      new Chat({
        transport: new DefaultChatTransport({ api: "/api/search-agent" }),
        messages: initialMessages,
      }),
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [] // Only create once on mount
  );

  const { messages, sendMessage, status, stop } = useChat({ chat });

  const isProcessing = status === "streaming" || status === "submitted";

  const handleCancel = () => {
    stop();
  };

  // Sync messages to parent for persistence
  useEffect(() => {
    if (onMessagesChange && messages.length > 0) {
      onMessagesChange(messages);
    }
  }, [messages, onMessagesChange]);

  // Send initial query on mount (only if no initial messages and query provided)
  useEffect(() => {
    if (initialQuery && !initialQuerySent.current && status === "ready" && (!initialMessages || initialMessages.length === 0)) {
      initialQuerySent.current = true;
      sendMessage({ text: initialQuery });
    }
  }, [initialQuery, sendMessage, status, initialMessages]);

  // Auto-scroll to bottom when messages change
  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [messages]);

  // Focus input after processing
  useEffect(() => {
    if (!isProcessing && inputRef.current) {
      inputRef.current.focus();
    }
  }, [isProcessing]);

  const handleFileClick = (file: { id: number; folder_id?: number | null }) => {
    const folderId = file.folder_id;
    if (folderId) {
      router.push(`/files/${folderId}?highlight=${file.id}`);
    } else {
      router.push(`/files?highlight=${file.id}`);
    }
    onClose();
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      if (input.trim() && !isProcessing) {
        sendMessage({ text: input });
        setInput("");
      }
    }
  };

  const handleSubmit = () => {
    if (input.trim() && !isProcessing) {
      sendMessage({ text: input });
      setInput("");
    }
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: 10 }}
      animate={{ opacity: 1, y: 0 }}
      exit={{ opacity: 0, y: -10 }}
      transition={{ duration: 0.2 }}
      className="flex flex-col h-full"
    >
      {/* Header */}
      <div className="flex items-center gap-2 px-3 py-2 border-b">
        <Button
          variant="ghost"
          size="sm"
          onClick={onBack}
          className="h-8 px-2"
        >
          <ArrowLeft className="h-4 w-4 mr-1" />
          Back
        </Button>
        <div className="flex items-center gap-2 text-sm font-medium flex-1">
          <Sparkles className="h-4 w-4 text-primary" />
          AI Search
        </div>
        {onClearHistory && messages.length > 0 && (
          <Button
            variant="ghost"
            size="sm"
            onClick={onClearHistory}
            className="h-8 px-2 text-muted-foreground hover:text-destructive"
          >
            <Trash2 className="h-4 w-4 mr-1" />
            Clear
          </Button>
        )}
      </div>

      {/* Messages */}
      <ScrollArea className="flex-1 min-h-0 px-3" viewportRef={scrollRef}>
        <div className="py-4 space-y-4">
          <AnimatePresence mode="popLayout">
            {messages.map((message) => {
              // Check if message has any content to display
              const hasTextContent = message.parts.some(
                (part) => part.type === "text" && (part as { type: "text"; text: string }).text.trim()
              );
              const hasToolCalls = message.parts.some((part) =>
                part.type.startsWith("tool-")
              );

              // Skip empty assistant messages
              if (message.role === "assistant" && !hasTextContent && !hasToolCalls) {
                return null;
              }

              return (
                <motion.div
                  key={message.id}
                  initial={{ opacity: 0, y: 20 }}
                  animate={{ opacity: 1, y: 0 }}
                  exit={{ opacity: 0, y: -20 }}
                  transition={{ duration: 0.3 }}
                >
                  <MessageBubble
                    message={message}
                    onFileClick={handleFileClick}
                  />
                </motion.div>
              );
            })}
          </AnimatePresence>

          {/* Loading indicator - only show when no assistant content yet */}
          <AnimatePresence>
            {isProcessing && !messages.some(
              (m) => m.role === "assistant" && m.parts.some(
                (p) => p.type === "text" || p.type.startsWith("tool-")
              )
            ) && (
                <motion.div
                  initial={{ opacity: 0, y: 20 }}
                  animate={{ opacity: 1, y: 0 }}
                  exit={{ opacity: 0, y: -20 }}
                  transition={{ duration: 0.2 }}
                >
                  <StreamingIndicator />
                </motion.div>
              )}
          </AnimatePresence>
        </div>
      </ScrollArea>

      {/* Input */}
      <div className="border-t p-3">
        <div className="flex items-center gap-2">
          <input
            ref={inputRef}
            type="text"
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder={
              isProcessing ? "Generating..." : "Ask a follow-up question..."
            }
            disabled={isProcessing}
            className="flex-1 bg-transparent text-sm outline-none placeholder:text-muted-foreground disabled:opacity-50"
          />
          <AnimatePresence mode="wait" initial={false}>
            {isProcessing ? (
              <motion.div
                key="stop"
                initial={{ scale: 0, rotate: -90 }}
                animate={{ scale: 1, rotate: 0 }}
                exit={{ scale: 0, rotate: 90 }}
                transition={{ duration: 0.15 }}
              >
                <Button
                  size="sm"
                  variant="ghost"
                  onClick={handleCancel}
                  className="h-6 w-6 p-0"
                >
                  <Square className="h-4 w-4 fill-black" />
                </Button>
              </motion.div>
            ) : (
              <motion.div
                key="send"
                initial={{ scale: 0, rotate: 90 }}
                animate={{ scale: 1, rotate: 0 }}
                exit={{ scale: 0, rotate: -90 }}
                transition={{ duration: 0.15 }}
              >
                <Button
                  size="sm"
                  variant="ghost"
                  onClick={handleSubmit}
                  disabled={!input.trim()}
                  className="h-6 w-6 p-0 rounded-full bg-black hover:bg-black"
                >
                  <ArrowUp className="h-4 w-4 text-white" />
                </Button>
              </motion.div>
            )}
          </AnimatePresence>
        </div>
      </div>
    </motion.div>
  );
}

function MessageBubble({
  message,
  onFileClick,
}: {
  message: AIUIMessage;
  onFileClick: (file: { id: number; folder_id?: number | null }) => void;
}) {
  const isUser = message.role === "user";

  // Extract text content from parts
  const textContent = message.parts
    .filter((part) => part.type === "text")
    .map((part) => (part as { type: "text"; text: string }).text)
    .join("");

  // Extract tool calls for display (tool parts have type starting with "tool-")
  const toolCalls = message.parts
    .filter((part) => part.type.startsWith("tool-"))
    .map((part) => part as unknown as ToolPartWithOutput);

  // Extract display_files tool for rendering
  const displayFilesTool = toolCalls.find(
    (tool) => tool.type === "tool-display_files"
  );
  const displayFilesOutput = displayFilesTool?.state === "output-available"
    ? displayFilesTool.output as DisplayFilesOutput
    : undefined;
  const isDisplayFilesLoading = displayFilesTool && displayFilesTool.state !== "output-available";

  // Filter out display_files from tool indicators (we render it differently)
  const otherToolCalls = toolCalls.filter((tool) => tool.type !== "tool-display_files");

  return (
    <div className={cn("flex gap-3", isUser && "flex-row-reverse")}>
      <div
        className={cn(
          "flex-shrink-0 w-8 h-8 rounded-full flex items-center justify-center",
          isUser ? "bg-primary text-primary-foreground" : "bg-muted"
        )}
      >
        {isUser ? <User className="h-4 w-4" /> : <Bot className="h-4 w-4" />}
      </div>

      <div className={cn("flex-1 space-y-3", isUser && "text-right")}>
        {/* Tool calls indicator (excluding display_files) */}
        {!isUser && otherToolCalls.length > 0 && (
          <div className="space-y-1">
            {otherToolCalls.map((tool, idx) => {
              // Extract tool name from type (e.g., "tool-search_files" -> "search_files")
              const toolName = tool.type.replace("tool-", "");
              return (
                <motion.div
                  key={`${tool.toolCallId || idx}-${idx}`}
                  initial={{ scale: 0.95, opacity: 0 }}
                  animate={{ scale: 1, opacity: 1 }}
                  className="flex items-center gap-2 px-3 py-1.5 rounded-lg bg-muted/50 text-xs w-fit"
                >
                  {tool.state === "output-available" ? (
                    <span className="text-green-500">✓</span>
                  ) : (
                    <Loader2 className="h-3 w-3 animate-spin text-primary" />
                  )}
                  <span className="text-muted-foreground">
                    {formatToolName(toolName)}
                  </span>
                </motion.div>
              );
            })}
          </div>
        )}

        {/* Display files loading state */}
        {!isUser && isDisplayFilesLoading && (
          <motion.div
            initial={{ scale: 0.95, opacity: 0 }}
            animate={{ scale: 1, opacity: 1 }}
            className="flex items-center gap-2 px-3 py-1.5 rounded-lg bg-muted/50 text-xs w-fit"
          >
            <Loader2 className="h-3 w-3 animate-spin text-primary" />
            <span className="text-muted-foreground">Preparing results...</span>
          </motion.div>
        )}

        {/* Display files result as clickable cards */}
        {!isUser && displayFilesOutput && (
          <div className="space-y-2">
            {displayFilesOutput.summary && (
              <p className="text-sm text-muted-foreground px-1">
                {displayFilesOutput.summary}
              </p>
            )}
            <div className="space-y-2">
              {displayFilesOutput.files.map((file) => (
                <FileResultCard
                  key={file.id}
                  file={{
                    id: file.id,
                    title: file.title,
                    file_type: file.file_type as FileType,
                    mime_type: file.mime_type,
                    folder_id: file.folder_id ?? null,
                    folder: file.folder_name ? { name: file.folder_name } : undefined,
                  }}
                  description={file.description || ""}
                  onClick={() => onFileClick({ id: file.id, folder_id: file.folder_id })}
                />
              ))}
            </div>
          </div>
        )}

        {/* Message content */}
        {textContent && (
          <div
            className={cn(
              "inline-block px-4 py-2 rounded-2xl text-sm max-w-[85%]",
              isUser
                ? "bg-primary text-primary-foreground rounded-tr-md"
                : "bg-muted rounded-tl-md text-left"
            )}
          >
            {isUser ? (
              <div className="whitespace-pre-wrap">{textContent}</div>
            ) : (
              <div className="prose prose-sm dark:prose-invert max-w-none break-all prose-p:my-1 prose-ul:my-1 prose-ol:my-1 prose-li:my-0.5">
                <Markdown>{textContent}</Markdown>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}

function StreamingIndicator() {
  return (
    <div className="flex gap-3">
      <div className="flex-shrink-0 w-8 h-8 rounded-full bg-muted flex items-center justify-center">
        <Bot className="h-4 w-4" />
      </div>
      <div className="flex-1">
        <motion.div
          initial={{ scale: 0.95, opacity: 0 }}
          animate={{ scale: 1, opacity: 1 }}
          className="flex items-center gap-2 px-4 py-2 rounded-2xl rounded-tl-md bg-muted text-sm w-fit"
        >
          <Loader2 className="h-4 w-4 animate-spin text-primary" />
          <span className="text-muted-foreground">Searching...</span>
        </motion.div>
      </div>
    </div>
  );
}

// Simplified file type for display_files tool output
interface DisplayFile {
  id: number;
  title: string;
  file_type: FileType;
  mime_type?: string;
  folder_id: number | null;
  folder?: { name: string };
}

function FileResultCard({
  file,
  description,
  onClick,
}: {
  file: DisplayFile;
  description: string;
  onClick: () => void;
}) {
  return (
    <motion.button
      whileHover={{ scale: 1.02, x: 4 }}
      whileTap={{ scale: 0.98 }}
      onClick={onClick}
      className="w-full text-left p-3 rounded-xl border bg-card hover:bg-accent/50 transition-colors group"
    >
      <div className="flex items-start gap-3">
        <div
          className={cn(
            "flex-shrink-0 mt-0.5",
            getFileTypeColor(file.file_type, file.mime_type)
          )}
        >
          <FileTypeIcon
            fileType={file.file_type}
            mimeType={file.mime_type}
            className="h-5 w-5"
          />
        </div>
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <p className="font-medium text-sm truncate">{file.title}</p>
            <ArrowRight className="h-4 w-4 opacity-0 group-hover:opacity-100 transition-opacity text-muted-foreground" />
          </div>
          {description && (
            <p className="text-xs text-muted-foreground line-clamp-2 mt-1">
              {description}
            </p>
          )}
          {file.folder && (
            <p className="text-xs text-muted-foreground/70 mt-1">
              in {file.folder.name}
            </p>
          )}
        </div>
      </div>
    </motion.button>
  );
}

function formatToolName(toolName: string): string {
  return toolName
    .split("_")
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(" ");
}
