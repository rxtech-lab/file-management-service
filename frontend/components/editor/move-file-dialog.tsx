"use client";

import { useState, useEffect, useCallback } from "react";
import {
  ChevronRight,
  Folder as FolderIcon,
  FolderOpen,
} from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { getFolderTreeAction } from "@/lib/actions/folder-actions";
import { cn } from "@/lib/utils";
import type { FolderTree } from "@/lib/api/types";

interface MoveFileDialogProps {
  currentFolderId: number | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onMove: (folderId: number | null) => Promise<void>;
}

export function MoveFileDialog({
  currentFolderId,
  open,
  onOpenChange,
  onMove,
}: MoveFileDialogProps) {
  const [folders, setFolders] = useState<FolderTree[]>([]);
  const [selectedId, setSelectedId] = useState<number | null>(null);
  const [loading, setLoading] = useState(false);
  const [moving, setMoving] = useState(false);

  useEffect(() => {
    if (open) {
      setSelectedId(currentFolderId);
      setLoading(true);
      getFolderTreeAction().then((result) => {
        if (result.success && result.data) {
          setFolders(result.data);
        }
        setLoading(false);
      });
    }
  }, [open, currentFolderId]);

  const handleMove = useCallback(async () => {
    setMoving(true);
    try {
      await onMove(selectedId);
      onOpenChange(false);
    } finally {
      setMoving(false);
    }
  }, [selectedId, onMove, onOpenChange]);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[400px]">
        <DialogHeader>
          <DialogTitle>Move to folder</DialogTitle>
          <DialogDescription>
            Select a destination folder.
          </DialogDescription>
        </DialogHeader>

        <div className="max-h-[300px] overflow-y-auto border rounded-md p-2 space-y-1">
          {loading ? (
            <div className="text-sm text-muted-foreground p-2">
              Loading folders...
            </div>
          ) : (
            <>
              {/* Root option */}
              <button
                onClick={() => setSelectedId(null)}
                className={cn(
                  "flex items-center gap-2 w-full text-left text-sm px-2 py-1.5 rounded-sm hover:bg-accent transition-colors",
                  selectedId === null && "bg-accent font-medium",
                )}
              >
                <FolderOpen className="h-4 w-4 text-blue-500 shrink-0" />
                Root (All Files)
              </button>

              {/* Folder tree */}
              {folders.map((folder) => (
                <SelectableFolderItem
                  key={folder.id}
                  folder={folder}
                  selectedId={selectedId}
                  onSelect={setSelectedId}
                  level={0}
                />
              ))}
            </>
          )}
        </div>

        <DialogFooter>
          <Button
            variant="outline"
            size="sm"
            onClick={() => onOpenChange(false)}
          >
            Cancel
          </Button>
          <Button
            size="sm"
            onClick={handleMove}
            disabled={moving || selectedId === currentFolderId}
          >
            {moving ? "Moving..." : "Move here"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

function SelectableFolderItem({
  folder,
  selectedId,
  onSelect,
  level,
}: {
  folder: FolderTree;
  selectedId: number | null;
  onSelect: (id: number) => void;
  level: number;
}) {
  const [isOpen, setIsOpen] = useState(false);
  const hasChildren = folder.children && folder.children.length > 0;
  const isSelected = selectedId === folder.id;

  return (
    <div>
      <button
        onClick={() => onSelect(folder.id)}
        className={cn(
          "flex items-center gap-1.5 w-full text-left text-sm px-2 py-1.5 rounded-sm hover:bg-accent transition-colors",
          isSelected && "bg-accent font-medium",
        )}
        style={{ paddingLeft: `${(level + 1) * 12 + 8}px` }}
      >
        {hasChildren ? (
          <ChevronRight
            className={cn(
              "h-3.5 w-3.5 shrink-0 transition-transform cursor-pointer",
              isOpen && "rotate-90",
            )}
            onClick={(e) => {
              e.stopPropagation();
              setIsOpen(!isOpen);
            }}
          />
        ) : (
          <span className="w-3.5 shrink-0" />
        )}
        <FolderIcon className="h-4 w-4 text-yellow-500 shrink-0" />
        <span className="truncate">{folder.name}</span>
      </button>

      {hasChildren && isOpen && (
        <div className="space-y-1">
          {folder.children!.map((child) => (
            <SelectableFolderItem
              key={child.id}
              folder={child}
              selectedId={selectedId}
              onSelect={onSelect}
              level={level + 1}
            />
          ))}
        </div>
      )}
    </div>
  );
}
