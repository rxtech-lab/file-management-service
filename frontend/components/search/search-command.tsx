"use client";

import { useState, useEffect, useCallback, useRef } from "react";
import { useRouter } from "next/navigation";
import { motion, AnimatePresence } from "framer-motion";
import { Loader2, Sparkles } from "lucide-react";
import {
  CommandDialog,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@/components/ui/command";
import { Badge } from "@/components/ui/badge";
import {
  FileTypeIcon,
  getFileTypeColor,
} from "@/components/files/file-type-icon";
import { SearchAgent } from "@/components/search/search-agent";
import { searchFilesAction } from "@/lib/actions/search-actions";
import type { SearchResult, SearchType } from "@/lib/api/types";
import { cn } from "@/lib/utils";

interface SearchCommandProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

type SearchMode = "quick" | "agent";

export function SearchCommand({ open, onOpenChange }: SearchCommandProps) {
  const router = useRouter();
  const [query, setQuery] = useState("");
  const [results, setResults] = useState<SearchResult[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [searchType, setSearchType] = useState<SearchType>("hybrid");
  const [mode, setMode] = useState<SearchMode>("quick");
  const [agentQuery, setAgentQuery] = useState("");
  const inputRef = useRef<HTMLInputElement>(null);

  // Debounced search
  const performSearch = useCallback(
    async (searchQuery: string) => {
      if (!searchQuery.trim()) {
        setResults([]);
        return;
      }

      setIsLoading(true);
      try {
        const response = await searchFilesAction({
          q: searchQuery,
          type: searchType,
          limit: 10,
        });

        if (response.success && response.data) {
          setResults(response.data.data);
        } else {
          setResults([]);
        }
      } catch (error) {
        console.error("Search error:", error);
        setResults([]);
      } finally {
        setIsLoading(false);
      }
    },
    [searchType]
  );

  // Debounce effect - only in quick mode
  useEffect(() => {
    if (mode !== "quick") return;

    const timeoutId = setTimeout(() => {
      performSearch(query);
    }, 300);

    return () => clearTimeout(timeoutId);
  }, [query, performSearch, mode]);

  // Reset when dialog closes
  useEffect(() => {
    if (!open) {
      setQuery("");
      setResults([]);
      setMode("quick");
      setAgentQuery("");
    }
  }, [open]);

  const handleSelect = (result: SearchResult) => {
    const folderId = result.file.folder_id;
    if (folderId) {
      router.push(`/files/${folderId}?highlight=${result.file.id}`);
    } else {
      router.push(`/files?highlight=${result.file.id}`);
    }
    onOpenChange(false);
  };

  const handleEnterPress = () => {
    if (query.trim() && mode === "quick") {
      setAgentQuery(query);
      setMode("agent");
    }
  };

  const handleBackToQuick = () => {
    setMode("quick");
    setAgentQuery("");
    // Re-focus input after animation
    setTimeout(() => {
      inputRef.current?.focus();
    }, 100);
  };

  const handleClose = () => {
    onOpenChange(false);
  };

  return (
    <CommandDialog open={open} onOpenChange={onOpenChange}>
      <AnimatePresence mode="wait">
        {mode === "quick" ? (
          <motion.div
            key="quick-search"
            initial={{ opacity: 0, x: -20 }}
            animate={{ opacity: 1, x: 0 }}
            exit={{ opacity: 0, x: -20 }}
            transition={{ duration: 0.2 }}
            className="flex flex-col"
          >
            <CommandInput
              ref={inputRef}
              placeholder="Search files... (Press Enter for AI search)"
              value={query}
              onValueChange={setQuery}
              onKeyDown={(e) => {
                if (e.key === "Enter" && query.trim()) {
                  e.preventDefault();
                  handleEnterPress();
                }
              }}
            />
            <CommandList>
              {isLoading ? (
                <div className="flex items-center justify-center py-6">
                  <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
                </div>
              ) : (
                <>
                  <CommandEmpty>
                    {query ? (
                      <div className="space-y-2 py-4">
                        <p>No files found.</p>
                        <button
                          onClick={handleEnterPress}
                          className="inline-flex items-center gap-2 text-sm text-primary hover:underline"
                        >
                          <Sparkles className="h-4 w-4" />
                          Try AI search
                        </button>
                      </div>
                    ) : (
                      <div className="space-y-2 py-4">
                        <p>Type to search files...</p>
                        <p className="text-xs text-muted-foreground">
                          Press <kbd className="px-1.5 py-0.5 bg-muted rounded text-[10px]">Enter</kbd> for AI-powered search
                        </p>
                      </div>
                    )}
                  </CommandEmpty>

                  {results.length > 0 && (
                    <CommandGroup heading="Results">
                      {results.map((result) => (
                        <CommandItem
                          key={result.file.id}
                          onSelect={() => handleSelect(result)}
                          className="cursor-pointer"
                        >
                          <div className="flex items-start gap-3 w-full">
                            <div
                              className={cn(
                                "mt-0.5",
                                getFileTypeColor(
                                  result.file.file_type,
                                  result.file.mime_type
                                )
                              )}
                            >
                              <FileTypeIcon
                                fileType={result.file.file_type}
                                mimeType={result.file.mime_type}
                                className="h-5 w-5"
                              />
                            </div>
                            <div className="flex-1 min-w-0">
                              <p className="font-medium truncate">
                                {result.file.title}
                              </p>
                              {result.snippet && (
                                <p className="text-xs text-muted-foreground truncate">
                                  {result.snippet}
                                </p>
                              )}
                              {result.file.folder && (
                                <p className="text-xs text-muted-foreground">
                                  {result.file.folder.name}
                                </p>
                              )}
                            </div>
                            {result.score > 0 && (
                              <Badge variant="secondary" className="text-xs">
                                {(result.score * 100).toFixed(0)}%
                              </Badge>
                            )}
                          </div>
                        </CommandItem>
                      ))}
                    </CommandGroup>
                  )}
                </>
              )}
            </CommandList>

            {/* Search type selector */}
            <div className="border-t p-2 flex items-center justify-between">
              <div className="flex items-center gap-2">
                {(["fulltext", "semantic", "hybrid"] as const).map((type) => (
                  <button
                    key={type}
                    onClick={() => setSearchType(type)}
                    className={cn(
                      "px-2 py-1 text-xs rounded-md transition-colors",
                      searchType === type
                        ? "bg-primary text-primary-foreground"
                        : "bg-muted text-muted-foreground hover:bg-muted/80"
                    )}
                  >
                    {type.charAt(0).toUpperCase() + type.slice(1)}
                  </button>
                ))}
              </div>
              <div className="flex items-center gap-3 text-xs text-muted-foreground">
                <span className="flex items-center gap-1">
                  <Sparkles className="h-3 w-3" />
                  <kbd className="px-1.5 py-0.5 bg-muted rounded text-[10px]">Enter</kbd>
                  AI
                </span>
                <span>ESC to close</span>
              </div>
            </div>
          </motion.div>
        ) : (
          <motion.div
            key="agent-search"
            initial={{ opacity: 0, x: 20 }}
            animate={{ opacity: 1, x: 0 }}
            exit={{ opacity: 0, x: 20 }}
            transition={{ duration: 0.2 }}
            className="h-[400px]"
          >
            <SearchAgent
              initialQuery={agentQuery}
              onBack={handleBackToQuick}
              onClose={handleClose}
            />
          </motion.div>
        )}
      </AnimatePresence>
    </CommandDialog>
  );
}
