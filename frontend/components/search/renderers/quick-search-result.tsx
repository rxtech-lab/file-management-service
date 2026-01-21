"use client";

import {
  FileTypeIcon,
  getFileTypeColor,
} from "@/components/files/file-type-icon";
import { Badge } from "@/components/ui/badge";
import type { SearchResult } from "@/lib/api/types";
import { cn } from "@/lib/utils";

interface QuickSearchResultProps {
  result: SearchResult;
}

export function QuickSearchResult({ result }: QuickSearchResultProps) {
  return (
    <div className="flex items-start gap-3 w-full">
      <div
        className={cn(
          "mt-0.5",
          getFileTypeColor(result.file.file_type, result.file.mime_type)
        )}
      >
        <FileTypeIcon
          fileType={result.file.file_type}
          mimeType={result.file.mime_type}
          className="h-5 w-5"
        />
      </div>
      <div className="flex-1 min-w-0">
        <p className="font-medium truncate">{result.file.title}</p>
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
  );
}

export function renderQuickSearchResult(result: SearchResult) {
  return <QuickSearchResult result={result} />;
}
