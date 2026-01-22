"use client";

import { useState, useEffect, useCallback } from "react";
import { useRouter } from "next/navigation";
import { Upload } from "lucide-react";
import type { ToolAction } from "@rx-lab/dashboard-searching-ui";
import { SearchTrigger, SearchCommand } from "@rx-lab/dashboard-searching-ui";
import { SidebarTrigger } from "@/components/ui/sidebar";
import { Button } from "@/components/ui/button";
import {
  QuickSearchResult,
  DisplayFilesRenderer,
} from "@/components/search/renderers";
import { searchFilesAction } from "@/lib/actions/search-actions";
import type { SearchResult, SearchType } from "@/lib/api/types";
import "@rx-lab/dashboard-searching-ui/style.css";

interface SiteHeaderProps {
  onUploadClick?: () => void;
}

// Extended result type that includes the original SearchResult
interface FileSearchResult {
  id: number;
  title: string;
  snippet?: string;
  score?: number;
  metadata: SearchResult;
}

// Transform SearchResult to FileSearchResult
function transformResult(result: SearchResult): FileSearchResult {
  return {
    id: result.file.id,
    title: result.file.title,
    snippet: result.snippet,
    score: result.score,
    metadata: result,
  };
}

export function SiteHeader({ onUploadClick }: SiteHeaderProps) {
  const router = useRouter();
  const [searchOpen, setSearchOpen] = useState(false);

  // Handle Cmd+K keyboard shortcut
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === "k") {
        e.preventDefault();
        setSearchOpen(true);
      }
    };

    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, []);

  // Search handler
  const handleSearch = useCallback(
    async ({
      query,
      searchType,
      limit,
    }: {
      query: string;
      searchType?: string;
      limit?: number;
    }): Promise<FileSearchResult[]> => {
      const response = await searchFilesAction({
        q: query,
        type: (searchType as SearchType) || "hybrid",
        limit: limit || 10,
      });

      if (response.success && response.data) {
        return response.data.data.map(transformResult);
      }
      return [];
    },
    [],
  );

  // Result selection handler
  const handleResultSelect = useCallback(
    (result: FileSearchResult) => {
      const folderId = result.metadata.file.folder_id;
      if (folderId) {
        router.push(`/files/${folderId}?highlight=${result.id}`);
      } else {
        router.push(`/files?highlight=${result.id}`);
      }
      setSearchOpen(false);
    },
    [router],
  );

  // Tool action handler for closing dialog from within tool renderers
  const handleToolAction = useCallback((action: ToolAction) => {
    if (action.type === "close") {
      setSearchOpen(false);
    }
  }, []);

  // Render quick search result
  const renderResult = useCallback(
    (result: FileSearchResult, onSelect: () => void) => (
      <div
        onClick={onSelect}
        className="cursor-pointer px-4 py-4 hover:bg-accent/50 transition-colors"
      >
        <QuickSearchResult result={result.metadata} />
      </div>
    ),
    [],
  );

  return (
    <>
      <header className="flex h-14 shrink-0 items-center gap-2 border-b px-4">
        <SidebarTrigger className="-ml-1" />

        {/* Search bar - centered */}
        <div className="flex-1 flex justify-center px-4">
          <SearchTrigger
            onClick={() => setSearchOpen(true)}
            placeholder="Search files..."
            shortcut={{ key: "K", modifier: "⌘" }}
          />
        </div>

        {/* Right side actions */}
        <div className="flex items-center gap-2">
          {onUploadClick && (
            <Button onClick={onUploadClick} size="sm">
              <Upload className="h-4 w-4 mr-2" />
              Upload
            </Button>
          )}
        </div>
      </header>

      <SearchCommand
        open={searchOpen}
        onOpenChange={setSearchOpen}
        onSearch={handleSearch}
        onResultSelect={handleResultSelect}
        className="px-4"
        searchTypes={[
          { id: "hybrid", label: "Hybrid" },
          { id: "fulltext", label: "Fulltext" },
          { id: "semantic", label: "Semantic" },
        ]}
        defaultSearchType="hybrid"
        debounceMs={300}
        limit={10}
        renderResult={renderResult}
        placeholder="Search files... (Press Enter for AI search)"
        showSearchTypeSelector={true}
        enableAgentMode={true}
        chatHistoryStorageKey="search-chat-history"
        agentConfig={{
          apiEndpoint: "/api/search-agent",
          toolResultRenderers: {
            display_files: DisplayFilesRenderer,
          },
          onToolAction: handleToolAction,
          header: {
            title: "AI Search",
            showBackButton: true,
            showClearButton: true,
          },
          input: {
            placeholder: "Ask a follow-up question...",
            placeholderProcessing: "Generating...",
            streamingText: "Searching...",
          },
        }}
      />
    </>
  );
}
