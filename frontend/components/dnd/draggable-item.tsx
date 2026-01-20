"use client";

import { useDraggable } from "@dnd-kit/core";
import { CSS } from "@dnd-kit/utilities";
import { cn } from "@/lib/utils";
import type { FileItem, Folder } from "@/lib/api/types";

interface DraggableFileProps {
  file: FileItem;
  children: React.ReactNode;
  disabled?: boolean;
}

export function DraggableFile({
  file,
  children,
  disabled = false,
}: DraggableFileProps) {
  const { attributes, listeners, setNodeRef, transform, isDragging } =
    useDraggable({
      id: `file-${file.id}`,
      data: {
        type: "file",
        item: file,
      },
      disabled,
    });

  const style = transform
    ? {
        transform: CSS.Translate.toString(transform),
      }
    : undefined;

  return (
    <div
      ref={setNodeRef}
      style={style}
      className={cn("h-full rounded-xl", isDragging && "opacity-50")}
      {...listeners}
      {...attributes}
    >
      {children}
    </div>
  );
}

interface DraggableFolderProps {
  folder: Folder;
  children: React.ReactNode;
  disabled?: boolean;
}

export function DraggableFolder({
  folder,
  children,
  disabled = false,
}: DraggableFolderProps) {
  const { attributes, listeners, setNodeRef, transform, isDragging } =
    useDraggable({
      id: `folder-${folder.id}`,
      data: {
        type: "folder",
        item: folder,
      },
      disabled,
    });

  const style = transform
    ? {
        transform: CSS.Translate.toString(transform),
      }
    : undefined;

  return (
    <div
      ref={setNodeRef}
      style={style}
      className={cn("h-full rounded-xl", isDragging && "opacity-50")}
      {...listeners}
      {...attributes}
    >
      {children}
    </div>
  );
}
