"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import {
  Folder as FolderIcon,
  MoreHorizontal,
  Pencil,
  FolderInput,
  Tags,
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
import { TagBadge } from "@/components/tags/tag-badge";
import { toast } from "sonner";
import { deleteFolderAction } from "@/lib/actions/folder-actions";
import type { Folder } from "@/lib/api/types";
import { cn } from "@/lib/utils";

interface FolderCardProps {
  folder: Folder;
  selected?: boolean;
  onSelect?: (folder: Folder, multi?: boolean) => void;
  onDoubleClick?: (folder: Folder) => void;
  onRename?: (folder: Folder) => void;
  onMove?: (folder: Folder) => void;
  onManageTags?: (folder: Folder) => void;
  onAIOrganize?: (folder: Folder) => void;
}

export function FolderCard({
  folder,
  selected,
  onSelect,
  onDoubleClick,
  onRename,
  onMove,
  onManageTags,
  onAIOrganize,
}: FolderCardProps) {
  const router = useRouter();
  const [isHovered, setIsHovered] = useState(false);

  const handleClick = (e: React.MouseEvent) => {
    if (onSelect) {
      onSelect(folder, e.metaKey || e.ctrlKey);
    }
  };

  const handleDoubleClick = () => {
    if (onDoubleClick) {
      onDoubleClick(folder);
    }
  };

  const handleDelete = async () => {
    if (
      !confirm(
        `Are you sure you want to delete "${folder.name}"? All files and subfolders will also be deleted. This action cannot be undone.`,
      )
    ) {
      return;
    }

    try {
      const result = await deleteFolderAction(folder.id);
      if (result.success) {
        toast.success("Folder deleted successfully");
        router.refresh();
      } else {
        toast.error(result.error || "Failed to delete folder");
      }
    } catch {
      toast.error("Failed to delete folder");
    }
  };

  return (
    <ContextMenu>
      <Tooltip>
        <TooltipTrigger asChild>
          <ContextMenuTrigger asChild>
            <Card
              className={cn(
                "h-full cursor-pointer transition-all hover:shadow-md",
                selected && "ring-2 ring-primary",
              )}
              onClick={handleClick}
              onDoubleClick={handleDoubleClick}
              onMouseEnter={() => setIsHovered(true)}
              onMouseLeave={() => setIsHovered(false)}
            >
              <CardContent className="p-4">
                <div className="flex flex-col items-center gap-2">
                  {/* Folder icon */}
                  <FolderIcon className="h-12 w-12 text-yellow-500 fill-yellow-100" />

                  {/* Folder name */}
                  <div className="w-full text-center">
                    <p className="font-medium text-sm truncate" title={folder.name}>
                      {folder.name}
                    </p>
                    {folder.description && (
                      <p
                        className="text-xs text-muted-foreground truncate"
                        title={folder.description}
                      >
                        {folder.description}
                      </p>
                    )}
                  </div>
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

        {/* Tooltip with tags */}
        {folder.tags && folder.tags.length > 0 && (
          <TooltipContent side="bottom" className="max-w-xs">
            <div className="space-y-1">
              <p className="font-medium text-xs">Tags</p>
              <div className="flex flex-wrap gap-1">
                {folder.tags.map((tag) => (
                  <TagBadge key={tag.id} tag={tag} className="text-xs" />
                ))}
              </div>
            </div>
          </TooltipContent>
        )}
      </Tooltip>

      <ContextMenuContent className="w-48">
        <ContextMenuItem onClick={() => onRename?.(folder)}>
          <Pencil className="mr-2 h-4 w-4" />
          Rename
        </ContextMenuItem>
        <ContextMenuItem onClick={() => onMove?.(folder)}>
          <FolderInput className="mr-2 h-4 w-4" />
          Move to...
        </ContextMenuItem>

        <ContextMenuSeparator />

        <ContextMenuItem onClick={() => onManageTags?.(folder)}>
          <Tags className="mr-2 h-4 w-4" />
          Add tags...
        </ContextMenuItem>

        <ContextMenuSeparator />

        <ContextMenuItem onClick={() => onAIOrganize?.(folder)}>
          <Sparkles className="mr-2 h-4 w-4" />
          AI Organize Contents
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
