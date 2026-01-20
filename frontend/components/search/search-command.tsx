"use client";

import { useState, useEffect, useCallback } from "react";
import { useRouter } from "next/navigation";
import { Loader2 } from "lucide-react";
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
import { searchFilesAction } from "@/lib/actions/search-actions";
import type { SearchResult, SearchType } from "@/lib/api/types";
import { cn } from "@/lib/utils";

interface SearchCommandProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function SearchCommand({ open, onOpenChange }: SearchCommandProps) {
  const router = useRouter();
  const [query, setQuery] = useState("");
  const [results, setResults] = useState<SearchResult[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [searchType, setSearchType] = useState<SearchType>("fulltext");

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
    [searchType],
  );

  // Debounce effect
  useEffect(() => {
    const timeoutId = setTimeout(() => {
      performSearch(query);
    }, 300);

    return () => clearTimeout(timeoutId);
  }, [query, performSearch]);

  // Reset when dialog closes
  useEffect(() => {
    if (!open) {
      setQuery("");
      setResults([]);
    }
  }, [open]);

  const handleSelect = (result: SearchResult) => {
    const folderId = result.file.folder_id;
    if (folderId) {
      router.push(`/files/${folderId}`);
    } else {
      router.push("/files");
    }
    onOpenChange(false);
  };

  return (
    <CommandDialog open={open} onOpenChange={onOpenChange}>
      <CommandInput
        placeholder="Search files..."
        value={query}
        onValueChange={setQuery}
      />
      <CommandList>
        {isLoading ? (
          <div className="flex items-center justify-center py-6">
            <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
          </div>
        ) : (
          <>
            <CommandEmpty>
              {query ? "No files found." : "Type to search files..."}
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
                            result.file.mime_type,
                          ),
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
                  : "bg-muted text-muted-foreground hover:bg-muted/80",
              )}
            >
              {type.charAt(0).toUpperCase() + type.slice(1)}
            </button>
          ))}
        </div>
        <span className="text-xs text-muted-foreground">ESC to close</span>
      </div>
    </CommandDialog>
  );
}
