"use client";

import { useState, useCallback } from "react";
import type { FileItem, Folder } from "@/lib/api/types";

export function useFileSelection() {
  const [selectedFileIds, setSelectedFileIds] = useState<Set<number>>(
    new Set(),
  );
  const [selectedFolderIds, setSelectedFolderIds] = useState<Set<number>>(
    new Set(),
  );

  const selectFile = useCallback((file: FileItem, multi?: boolean) => {
    setSelectedFileIds((prev) => {
      const newSet = new Set(multi ? prev : []);
      if (prev.has(file.id) && multi) {
        newSet.delete(file.id);
      } else {
        newSet.add(file.id);
      }
      return newSet;
    });

    // Clear folder selection when selecting files (unless multi-select)
    if (!multi) {
      setSelectedFolderIds(new Set());
    }
  }, []);

  const selectFolder = useCallback((folder: Folder, multi?: boolean) => {
    setSelectedFolderIds((prev) => {
      const newSet = new Set(multi ? prev : []);
      if (prev.has(folder.id) && multi) {
        newSet.delete(folder.id);
      } else {
        newSet.add(folder.id);
      }
      return newSet;
    });

    // Clear file selection when selecting folders (unless multi-select)
    if (!multi) {
      setSelectedFileIds(new Set());
    }
  }, []);

  const clearSelection = useCallback(() => {
    setSelectedFileIds(new Set());
    setSelectedFolderIds(new Set());
  }, []);

  const selectAllFiles = useCallback((files: FileItem[]) => {
    setSelectedFileIds(new Set(files.map((f) => f.id)));
  }, []);

  const selectAllFolders = useCallback((folders: Folder[]) => {
    setSelectedFolderIds(new Set(folders.map((f) => f.id)));
  }, []);

  const selectAll = useCallback(
    (files: FileItem[], folders: Folder[]) => {
      selectAllFiles(files);
      selectAllFolders(folders);
    },
    [selectAllFiles, selectAllFolders],
  );

  const hasSelection = selectedFileIds.size > 0 || selectedFolderIds.size > 0;

  const selectionCount = selectedFileIds.size + selectedFolderIds.size;

  return {
    selectedFileIds,
    selectedFolderIds,
    selectFile,
    selectFolder,
    clearSelection,
    selectAllFiles,
    selectAllFolders,
    selectAll,
    hasSelection,
    selectionCount,
  };
}
