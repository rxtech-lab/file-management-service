"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import {
  MoreHorizontal,
  Loader2,
  Pencil,
  FolderInput,
  Download,
  Tags,
  Info,
  Trash2,
  Sparkles,
} from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import {
  ContextMenu,
  ContextMenuContent,
  ContextMenuItem,
  ContextMenuSeparator,
  ContextMenuTrigger,
} from "@/components/ui/context-menu";
import { toast } from "sonner";
import { FileTypeIcon, getFileTypeColor } from "./file-type-icon";
import { TagBadge } from "@/components/tags/tag-badge";
import {
  deleteFileAction,
  getFileDownloadAction,
} from "@/lib/actions/file-actions";
import type { FileItem } from "@/lib/api/types";
import { cn } from "@/lib/utils";

interface FileCardProps {
  file: FileItem;
  selected?: boolean;
  onSelect?: (file: FileItem, multi?: boolean) => void;
  onClick?: (file: FileItem) => void;
  onRename?: (file: FileItem) => void;
  onMove?: (file: FileItem) => void;
  onManageTags?: (file: FileItem) => void;
  onViewMetadata?: (file: FileItem) => void;
  onAIOrganize?: (file: FileItem) => void;
}

function formatFileSize(bytes?: number): string {
  if (!bytes) return "Unknown size";
  const units = ["B", "KB", "MB", "GB"];
  let size = bytes;
  let unitIndex = 0;
  while (size >= 1024 && unitIndex < units.length - 1) {
    size /= 1024;
    unitIndex++;
  }
  return `${size.toFixed(1)} ${units[unitIndex]}`;
}

export function FileCard({
  file,
  selected,
  onSelect,
  onClick,
  onRename,
  onMove,
  onManageTags,
  onViewMetadata,
  onAIOrganize,
}: FileCardProps) {
  const router = useRouter();
  const [isHovered, setIsHovered] = useState(false);

  const handleClick = (e: React.MouseEvent) => {
    if (onSelect) {
      onSelect(file, e.metaKey || e.ctrlKey);
    }
    if (onClick) {
      onClick(file);
    }
  };

  const handleDownload = async () => {
    try {
      const result = await getFileDownloadAction(file.id);
      if (result.success && result.data) {
        window.open(result.data.download_url, "_blank");
      } else {
        toast.error(result.error || "Failed to get download URL");
      }
    } catch {
      toast.error("Failed to download file");
    }
  };

  const handleDelete = async () => {
    if (
      !confirm(
        `Are you sure you want to delete "${file.title}"? This action cannot be undone.`,
      )
    ) {
      return;
    }

    try {
      const result = await deleteFileAction(file.id);
      if (result.success) {
        toast.success("File deleted successfully");
        router.refresh();
      } else {
        toast.error(result.error || "Failed to delete file");
      }
    } catch {
      toast.error("Failed to delete file");
    }
  };

  const isProcessing =
    file.processing_status === "pending" ||
    file.processing_status === "processing";

  return (
    <ContextMenu>
      <Tooltip>
        <TooltipTrigger asChild>
          <ContextMenuTrigger asChild>
            <Card
              className={cn(
                "h-full cursor-pointer transition-all hover:shadow-md",
                selected && "ring-2 ring-primary",
                isProcessing && "opacity-70",
              )}
              onClick={handleClick}
              onMouseEnter={() => setIsHovered(true)}
              onMouseLeave={() => setIsHovered(false)}
            >
              <CardContent className="p-4">
                <div className="flex flex-col items-center gap-2">
                  {/* File icon */}
                  <div
                    className={cn(
                      "relative",
                      getFileTypeColor(file.file_type, file.mime_type),
                    )}
                  >
                    <FileTypeIcon
                      fileType={file.file_type}
                      mimeType={file.mime_type}
                      className="h-12 w-12"
                    />
                    {isProcessing && (
                      <div className="absolute inset-0 flex items-center justify-center bg-background/80 rounded">
                        <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
                      </div>
                    )}
                  </div>

                  {/* File name */}
                  <div className="w-full text-center">
                    <p
                      className="font-medium text-sm truncate"
                      title={file.title}
                    >
                      {file.title}
                    </p>
                    <p className="text-xs text-muted-foreground truncate">
                      {file.original_filename}
                    </p>
                  </div>

                  {/* File size */}
                  <p className="text-xs text-muted-foreground">
                    {formatFileSize(file.size)}
                  </p>

                  {/* Processing status indicator */}
                  {file.processing_status === "failed" && (
                    <span className="text-xs text-destructive">
                      Processing failed
                    </span>
                  )}
                </div>

                {/* Hover actions */}
                {isHovered && (
                  <Button
                    variant="ghost"
                    size="icon"
                    className="absolute top-2 right-2 h-6 w-6"
                    onClick={(e) => {
                      e.stopPropagation();
                    }}
                  >
                    <MoreHorizontal className="h-4 w-4" />
                  </Button>
                )}
              </CardContent>
            </Card>
          </ContextMenuTrigger>
        </TooltipTrigger>

        {/* Tooltip with tags and AI summary */}
        {(file.summary || (file.tags && file.tags.length > 0)) && (
          <TooltipContent side="bottom" className="max-w-xs">
            <div className="space-y-2">
              {file.tags && file.tags.length > 0 && (
                <div className="space-y-1">
                  <p className="font-medium text-xs">Tags</p>
                  <div className="flex flex-wrap gap-1">
                    {file.tags.map((tag) => (
                      <TagBadge key={tag.id} tag={tag} className="text-xs" />
                    ))}
                  </div>
                </div>
              )}
              {file.summary && (
                <div className="space-y-1">
                  <p className="font-medium text-xs">AI Summary</p>
                  <p className="text-xs text-muted-foreground break-all">
                    {file.summary}
                  </p>
                </div>
              )}
            </div>
          </TooltipContent>
        )}
      </Tooltip>

      <ContextMenuContent className="w-48">
        <ContextMenuItem onClick={() => onRename?.(file)}>
          <Pencil className="mr-2 h-4 w-4" />
          Rename
        </ContextMenuItem>
        <ContextMenuItem onClick={() => onMove?.(file)}>
          <FolderInput className="mr-2 h-4 w-4" />
          Move to...
        </ContextMenuItem>
        <ContextMenuItem onClick={handleDownload}>
          <Download className="mr-2 h-4 w-4" />
          Download
        </ContextMenuItem>

        <ContextMenuSeparator />

        <ContextMenuItem onClick={() => onManageTags?.(file)}>
          <Tags className="mr-2 h-4 w-4" />
          Add tags...
        </ContextMenuItem>
        <ContextMenuItem onClick={() => onViewMetadata?.(file)}>
          <Info className="mr-2 h-4 w-4" />
          View metadata
        </ContextMenuItem>

        <ContextMenuSeparator />

        <ContextMenuItem onClick={() => onAIOrganize?.(file)}>
          <Sparkles className="mr-2 h-4 w-4" />
          AI Organize
        </ContextMenuItem>

        <ContextMenuSeparator />

        <ContextMenuItem onClick={handleDelete} variant="destructive">
          <Trash2 className="mr-2 h-4 w-4" />
          Delete
        </ContextMenuItem>
      </ContextMenuContent>
    </ContextMenu>
  );
}
