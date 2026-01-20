"use client";

import { useState, useRef, useEffect } from "react";
import { useRouter } from "next/navigation";
import { readStreamableValue } from "@ai-sdk/rsc";
import { motion, AnimatePresence } from "framer-motion";
import {
  ArrowLeft,
  ArrowRight,
  Check,
  Loader2,
  Sparkles,
  User,
  Bot,
  AlertCircle,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import {
  FileTypeIcon,
  getFileTypeColor,
} from "@/components/files/file-type-icon";
import { searchWithAgentAction } from "@/lib/actions/search-agent-actions";
import type {
  AgentSearchMessage,
  AgentSearchToolCall,
  AgentSearchProgress,
  FileItem,
} from "@/lib/api/types";
import { cn } from "@/lib/utils";

interface SearchAgentProps {
  initialQuery: string;
  onBack: () => void;
  onClose: () => void;
}

export function SearchAgent({
  initialQuery,
  onBack,
  onClose,
}: SearchAgentProps) {
  const router = useRouter();
  const [messages, setMessages] = useState<AgentSearchMessage[]>([]);
  const [inputValue, setInputValue] = useState("");
  const [isProcessing, setIsProcessing] = useState(false);
  const [currentProgress, setCurrentProgress] =
    useState<AgentSearchProgress | null>(null);
  const [currentText, setCurrentText] = useState("");
  const scrollRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  // Process initial query on mount
  useEffect(() => {
    if (initialQuery && messages.length === 0) {
      handleSubmit(initialQuery);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [initialQuery]);

  // Auto-scroll to bottom when messages change
  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [messages, currentProgress, currentText]);

  // Focus input after processing
  useEffect(() => {
    if (!isProcessing && inputRef.current) {
      inputRef.current.focus();
    }
  }, [isProcessing]);

  const handleSubmit = async (query: string) => {
    if (!query.trim() || isProcessing) return;

    const userMessage: AgentSearchMessage = {
      id: crypto.randomUUID(),
      role: "user",
      content: query,
      createdAt: new Date(),
    };

    const newMessages = [...messages, userMessage];
    setMessages(newMessages);
    setInputValue("");
    setIsProcessing(true);
    setCurrentText("");

    try {
      const { progress, textStream } =
        await searchWithAgentAction(newMessages);

      // Read progress stream
      if (progress) {
        (async () => {
          for await (const progressUpdate of readStreamableValue(progress)) {
            if (progressUpdate) {
              setCurrentProgress(progressUpdate);
            }
          }
        })();
      }

      // Read text stream
      let fullText = "";
      if (textStream) {
        for await (const textPart of readStreamableValue(textStream)) {
          if (textPart) {
            fullText += textPart;
            setCurrentText(fullText);
          }
        }
      }

      // Add assistant response
      const assistantMessage: AgentSearchMessage = {
        id: crypto.randomUUID(),
        role: "assistant",
        content: fullText || "I couldn't find any results for your search.",
        createdAt: new Date(),
      };

      setMessages((prev) => [...prev, assistantMessage]);
    } catch (error) {
      console.error("Search error:", error);
      const errorMessage: AgentSearchMessage = {
        id: crypto.randomUUID(),
        role: "assistant",
        content:
          error instanceof Error
            ? error.message
            : "An error occurred while searching.",
        createdAt: new Date(),
      };
      setMessages((prev) => [...prev, errorMessage]);
    } finally {
      setCurrentProgress(null);
      setCurrentText("");
      setIsProcessing(false);
    }
  };

  const handleFileClick = (file: FileItem) => {
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
      handleSubmit(inputValue);
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
        <div className="flex items-center gap-2 text-sm font-medium">
          <Sparkles className="h-4 w-4 text-primary" />
          AI Search
        </div>
      </div>

      {/* Messages */}
      <ScrollArea className="flex-1 px-3" ref={scrollRef}>
        <div className="py-4 space-y-4">
          <AnimatePresence mode="popLayout">
            {messages.map((message) => (
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
            ))}
          </AnimatePresence>

          {/* Current progress/streaming text */}
          <AnimatePresence>
            {isProcessing && (
              <motion.div
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, y: -20 }}
                transition={{ duration: 0.2 }}
              >
                <StreamingResponse
                  progress={currentProgress}
                  text={currentText}
                />
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
            value={inputValue}
            onChange={(e) => setInputValue(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder={
              isProcessing ? "Searching..." : "Ask a follow-up question..."
            }
            disabled={isProcessing}
            className="flex-1 bg-transparent text-sm outline-none placeholder:text-muted-foreground disabled:opacity-50"
          />
          <Button
            size="sm"
            variant="ghost"
            onClick={() => handleSubmit(inputValue)}
            disabled={isProcessing || !inputValue.trim()}
          >
            {isProcessing ? (
              <Loader2 className="h-4 w-4 animate-spin" />
            ) : (
              <span className="text-xs text-muted-foreground">Enter</span>
            )}
          </Button>
        </div>
      </div>
    </motion.div>
  );
}

function MessageBubble({
  message,
  onFileClick,
}: {
  message: AgentSearchMessage;
  onFileClick: (file: FileItem) => void;
}) {
  const isUser = message.role === "user";

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
        <div
          className={cn(
            "inline-block px-4 py-2 rounded-2xl text-sm max-w-[85%]",
            isUser
              ? "bg-primary text-primary-foreground rounded-tr-md"
              : "bg-muted rounded-tl-md text-left"
          )}
        >
          <div className="whitespace-pre-wrap">{message.content}</div>
        </div>

        {/* File results from message */}
        {message.files && message.files.length > 0 && (
          <div className="space-y-2">
            <AnimatePresence>
              {message.files.map((result, index) => (
                <motion.div
                  key={result.file.id}
                  initial={{ opacity: 0, x: -20 }}
                  animate={{ opacity: 1, x: 0 }}
                  transition={{ delay: index * 0.1, duration: 0.3 }}
                >
                  <FileResultCard
                    file={result.file}
                    description={result.description}
                    onClick={() => onFileClick(result.file)}
                  />
                </motion.div>
              ))}
            </AnimatePresence>
          </div>
        )}
      </div>
    </div>
  );
}

function StreamingResponse({
  progress,
  text,
}: {
  progress: AgentSearchProgress | null;
  text: string;
}) {
  const isError = progress?.status === "error";
  const isCalling = progress?.status === "calling";
  const isThinking = progress?.status === "thinking";

  return (
    <div className="flex gap-3">
      <div className="flex-shrink-0 w-8 h-8 rounded-full bg-muted flex items-center justify-center">
        <Bot className="h-4 w-4" />
      </div>
      <div className="flex-1 space-y-2">
        {/* Progress indicator */}
        {progress && (isThinking || isCalling) && (
          <motion.div
            initial={{ scale: 0.95, opacity: 0 }}
            animate={{ scale: 1, opacity: 1 }}
            className="flex items-center gap-2 px-4 py-2 rounded-2xl rounded-tl-md bg-muted text-sm w-fit"
          >
            {isError ? (
              <AlertCircle className="h-4 w-4 text-destructive" />
            ) : (
              <Loader2 className="h-4 w-4 animate-spin text-primary" />
            )}
            <span className="text-muted-foreground">
              {progress.message || "Processing..."}
            </span>
            {isCalling && progress.toolName && (
              <span className="text-xs bg-primary/10 text-primary px-2 py-0.5 rounded-full">
                {progress.toolName}
              </span>
            )}
          </motion.div>
        )}

        {/* Streaming text */}
        {text && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            className="px-4 py-2 rounded-2xl rounded-tl-md bg-muted text-sm"
          >
            <div className="whitespace-pre-wrap">{text}</div>
            <motion.span
              animate={{ opacity: [1, 0] }}
              transition={{ repeat: Infinity, duration: 0.8 }}
              className="inline-block w-2 h-4 bg-primary ml-1 align-middle"
            />
          </motion.div>
        )}
      </div>
    </div>
  );
}

function FileResultCard({
  file,
  description,
  onClick,
}: {
  file: FileItem;
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
          <p className="text-xs text-muted-foreground line-clamp-2 mt-1">
            {description}
          </p>
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
