"use client";

import { useDroppable } from "@dnd-kit/core";
import { cn } from "@/lib/utils";

interface DroppableFolderProps {
  folderId: number | null;
  children: React.ReactNode;
  className?: string;
  disabled?: boolean;
}

export function DroppableFolder({
  folderId,
  children,
  className,
  disabled = false,
}: DroppableFolderProps) {
  const { isOver, setNodeRef, active } = useDroppable({
    id: folderId !== null ? `drop-folder-${folderId}` : "drop-root",
    data: {
      type: folderId !== null ? "folder-drop" : "root-drop",
      folderId,
    },
    disabled,
  });

  // Prevent dropping folder on itself
  const isDroppingOnSelf =
    active?.data.current?.type === "folder" &&
    active?.data.current?.item?.id === folderId;

  const showDropIndicator = isOver && !isDroppingOnSelf && active !== null;

  return (
    <div
      ref={setNodeRef}
      className={cn(
        "transition-colors duration-150",
        showDropIndicator && "ring-2 ring-primary ring-offset-2 bg-primary/5",
        className,
      )}
    >
      {children}
    </div>
  );
}

interface DroppableFolderTreeItemProps {
  folderId: number;
  children: React.ReactNode;
  className?: string;
}

export function DroppableFolderTreeItem({
  folderId,
  children,
  className,
}: DroppableFolderTreeItemProps) {
  const { isOver, setNodeRef, active } = useDroppable({
    id: `tree-drop-${folderId}`,
    data: {
      type: "folder-drop",
      folderId,
    },
  });

  // Prevent dropping folder on itself
  const isDroppingOnSelf =
    active?.data.current?.type === "folder" &&
    active?.data.current?.item?.id === folderId;

  const showDropIndicator = isOver && !isDroppingOnSelf && active !== null;

  return (
    <div
      ref={setNodeRef}
      className={cn(
        "h-full transition-colors duration-150 rounded-xl",
        showDropIndicator && "bg-primary/10 ring-1 ring-primary",
        className,
      )}
    >
      {children}
    </div>
  );
}
