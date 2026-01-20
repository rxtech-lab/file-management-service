"use client";

import React, { createContext, useContext, useState, useCallback } from "react";
import {
  DndContext,
  DragOverlay,
  closestCenter,
  KeyboardSensor,
  PointerSensor,
  useSensor,
  useSensors,
  DragStartEvent,
  DragEndEvent,
  DragOverEvent,
  UniqueIdentifier,
} from "@dnd-kit/core";
import { sortableKeyboardCoordinates } from "@dnd-kit/sortable";
import { Folder, FileText } from "lucide-react";
import type { FileItem, Folder as FolderType } from "@/lib/api/types";

export type DragItem =
  | { type: "file"; item: FileItem }
  | { type: "folder"; item: FolderType };

interface DndProviderContextValue {
  activeItem: DragItem | null;
  overId: UniqueIdentifier | null;
}

const DndProviderContext = createContext<DndProviderContextValue>({
  activeItem: null,
  overId: null,
});

export function useDndContext() {
  return useContext(DndProviderContext);
}

interface DndProviderProps {
  children: React.ReactNode;
  onMoveFiles?: (
    fileIds: number[],
    targetFolderId: number | null,
  ) => Promise<void>;
  onMoveFolder?: (
    folderId: number,
    targetFolderId: number | null,
  ) => Promise<void>;
}

export function DndProvider({
  children,
  onMoveFiles,
  onMoveFolder,
}: DndProviderProps) {
  const [activeItem, setActiveItem] = useState<DragItem | null>(null);
  const [overId, setOverId] = useState<UniqueIdentifier | null>(null);

  const sensors = useSensors(
    useSensor(PointerSensor, {
      activationConstraint: {
        distance: 8,
      },
    }),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates,
    }),
  );

  const handleDragStart = useCallback((event: DragStartEvent) => {
    const { active } = event;
    const data = active.data.current;

    if (data?.type === "file" && data.item) {
      setActiveItem({ type: "file", item: data.item as FileItem });
    } else if (data?.type === "folder" && data.item) {
      setActiveItem({ type: "folder", item: data.item as FolderType });
    }
  }, []);

  const handleDragOver = useCallback((event: DragOverEvent) => {
    const { over } = event;
    setOverId(over?.id ?? null);
  }, []);

  const handleDragEnd = useCallback(
    async (event: DragEndEvent) => {
      const { active, over } = event;

      if (!over || !activeItem) {
        setActiveItem(null);
        setOverId(null);
        return;
      }

      const overData = over.data.current;

      // Check if dropping on a folder
      if (overData?.type === "folder-drop" || overData?.type === "root-drop") {
        const targetFolderId = overData.folderId as number | null;

        if (activeItem.type === "file") {
          // Move file to folder
          await onMoveFiles?.([activeItem.item.id], targetFolderId);
        } else if (activeItem.type === "folder") {
          // Don't allow dropping folder on itself or its children
          if (activeItem.item.id !== targetFolderId) {
            await onMoveFolder?.(activeItem.item.id, targetFolderId);
          }
        }
      }

      setActiveItem(null);
      setOverId(null);
    },
    [activeItem, onMoveFiles, onMoveFolder],
  );

  const handleDragCancel = useCallback(() => {
    setActiveItem(null);
    setOverId(null);
  }, []);

  return (
    <DndProviderContext.Provider value={{ activeItem, overId }}>
      <DndContext
        sensors={sensors}
        collisionDetection={closestCenter}
        onDragStart={handleDragStart}
        onDragOver={handleDragOver}
        onDragEnd={handleDragEnd}
        onDragCancel={handleDragCancel}
      >
        {children}
        <DragOverlay dropAnimation={null}>
          {activeItem && (
            <div className="flex items-center gap-2 rounded-lg border bg-background px-3 py-2 shadow-lg">
              {activeItem.type === "file" ? (
                <>
                  <FileText className="h-4 w-4 text-muted-foreground" />
                  <span className="text-sm font-medium">
                    {activeItem.item.title}
                  </span>
                </>
              ) : (
                <>
                  <Folder className="h-4 w-4 text-yellow-500" />
                  <span className="text-sm font-medium">
                    {activeItem.item.name}
                  </span>
                </>
              )}
            </div>
          )}
        </DragOverlay>
      </DndContext>
    </DndProviderContext.Provider>
  );
}
