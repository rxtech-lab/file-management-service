"use client";

import { FolderPlus, Upload } from "lucide-react";
import {
  ContextMenu,
  ContextMenuContent,
  ContextMenuItem,
  ContextMenuTrigger,
} from "@/components/ui/context-menu";

interface EmptyContextMenuWrapperProps {
  children: React.ReactNode;
  onNewFolder: () => void;
  onUpload: () => void;
  className?: string;
}

export function EmptyContextMenuWrapper({
  children,
  onNewFolder,
  onUpload,
  className,
}: EmptyContextMenuWrapperProps) {
  return (
    <ContextMenu>
      <ContextMenuTrigger asChild>
        <div className={className}>{children}</div>
      </ContextMenuTrigger>

      <ContextMenuContent className="w-48">
        <ContextMenuItem onClick={onNewFolder}>
          <FolderPlus className="mr-2 h-4 w-4" />
          New folder
        </ContextMenuItem>
        <ContextMenuItem onClick={onUpload}>
          <Upload className="mr-2 h-4 w-4" />
          Upload files
        </ContextMenuItem>
      </ContextMenuContent>
    </ContextMenu>
  );
}
