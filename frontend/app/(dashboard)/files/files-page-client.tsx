"use client";

import { useState, useRef, useCallback } from "react";
import { useRouter } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { LayoutGrid, List } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { SiteHeader } from "@/components/layout/site-header";
import { FileGrid } from "@/components/files/file-grid";
import { FileList } from "@/components/files/file-list";
import { FolderBreadcrumb } from "@/components/folders/folder-breadcrumb";
import { CreateFolderDialog } from "@/components/folders/create-folder-dialog";
import { UploadQueue } from "@/components/upload/upload-queue";
import { FileMetadataSheet } from "@/components/files/file-metadata-sheet";
import { triggerAIOrganize } from "@/components/ai/ai-organize-dialog";
import { DndProvider } from "@/components/dnd/dnd-provider";
import { DroppableFolder } from "@/components/dnd/droppable-folder";
import { useViewMode } from "@/hooks/use-view-mode";
import { useFileSelection } from "@/hooks/use-file-selection";
import { useUploadQueue } from "@/hooks/use-upload-queue";
import { moveFilesAction } from "@/lib/actions/file-actions";
import { moveFolderAction } from "@/lib/actions/folder-actions";
import type { FileItem, Folder } from "@/lib/api/types";

interface FilesPageClientProps {
  initialFiles: FileItem[];
  initialFolders: Folder[];
  currentFolder: Folder | null;
  ancestors?: Folder[];
  highlightFileId?: number;
}

export function FilesPageClient({
  initialFiles,
  initialFolders,
  currentFolder,
  ancestors = [],
  highlightFileId,
}: FilesPageClientProps) {
  const router = useRouter();
  const { viewMode, setViewMode, isGrid } = useViewMode();
  const {
    selectedFileIds,
    selectedFolderIds,
    selectFile,
    selectFolder,
    clearSelection,
  } = useFileSelection();

  const uploadQueue = useUploadQueue(currentFolder?.id ?? null);

  // Auto-refresh the page every 10 seconds
  useQuery({
    queryKey: ["page-refresh", currentFolder?.id],
    queryFn: async () => {
      router.refresh();
      return Date.now();
    },
    refetchInterval: 10000, // 10 seconds
    refetchOnWindowFocus: true,
  });

  // Create folder dialog
  const [showCreateFolder, setShowCreateFolder] = useState(false);

  // Metadata sheet state
  const [metadataFile, setMetadataFile] = useState<FileItem | null>(null);

  // File input ref for upload
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleFolderDoubleClick = (folder: Folder) => {
    router.push(`/files/${folder.id}`);
  };

  const handleFileClick = (file: FileItem) => {
    // Empty handler for future use
  };

  const handleUploadClick = () => {
    fileInputRef.current?.click();
  };

  const handleFilesSelected = (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = e.target.files;
    if (files && files.length > 0) {
      uploadQueue.addFiles(Array.from(files));
    }
    // Reset input
    e.target.value = "";
  };

  const handleMoveFiles = useCallback(
    async (fileIds: number[], targetFolderId: number | null) => {
      try {
        const result = await moveFilesAction(fileIds, targetFolderId);
        if (result.success) {
          toast.success(`Moved ${fileIds.length} file(s)`);
          router.refresh();
        } else {
          toast.error(result.error || "Failed to move files");
        }
      } catch {
        toast.error("Failed to move files");
      }
    },
    [router],
  );

  const handleMoveFolder = useCallback(
    async (folderId: number, targetFolderId: number | null) => {
      try {
        const result = await moveFolderAction(folderId, targetFolderId);
        if (result.success) {
          toast.success("Folder moved");
          router.refresh();
        } else {
          toast.error(result.error || "Failed to move folder");
        }
      } catch {
        toast.error("Failed to move folder");
      }
    },
    [router],
  );

  // File action handlers
  const handleFileViewMetadata = (file: FileItem) => {
    setMetadataFile(file);
  };

  const handleFileRename = (file: FileItem) => {
    // Open metadata sheet for rename (can use the title edit feature)
    setMetadataFile(file);
  };

  const handleFileMove = (file: FileItem) => {
    // TODO: Open move dialog
    console.log("Move file:", file);
  };

  const handleFileManageTags = (file: FileItem) => {
    // Open metadata sheet for tag management
    setMetadataFile(file);
  };

  // Folder action handlers
  const handleFolderRename = (folder: Folder) => {
    // TODO: Open rename dialog
    console.log("Rename folder:", folder);
  };

  const handleFolderMove = (folder: Folder) => {
    // TODO: Open move dialog
    console.log("Move folder:", folder);
  };

  const handleFolderManageTags = (folder: Folder) => {
    // TODO: Open tag management
    console.log("Manage folder tags:", folder);
  };

  // AI Organize handlers
  const handleFileAIOrganize = (file: FileItem) => {
    triggerAIOrganize(file.id, file.title, () => {
      // Refresh the page when AI organization completes
      router.refresh();
    });
  };

  const handleFolderAIOrganize = (folder: Folder) => {
    // TODO: Implement folder AI organize (organize all files in folder)
    toast.info("AI Organize for folders coming soon");
  };

  // Empty space action handlers
  const handleNewFolder = () => {
    setShowCreateFolder(true);
  };

  return (
    <DndProvider onMoveFiles={handleMoveFiles} onMoveFolder={handleMoveFolder}>
      <SiteHeader onUploadClick={handleUploadClick} />

      <DroppableFolder
        folderId={currentFolder?.id ?? null}
        className="flex flex-col flex-1"
      >
        {/* Toolbar */}
        <div className="flex items-center justify-between px-4 py-2 border-b">
          <FolderBreadcrumb folder={currentFolder} ancestors={ancestors} />
          <div className="flex items-center gap-2">
            <Button
              variant={isGrid ? "secondary" : "ghost"}
              size="icon"
              onClick={() => setViewMode("grid")}
            >
              <LayoutGrid className="h-4 w-4" />
            </Button>
            <Button
              variant={!isGrid ? "secondary" : "ghost"}
              size="icon"
              onClick={() => setViewMode("list")}
            >
              <List className="h-4 w-4" />
            </Button>
          </div>
        </div>

        {/* Content */}
        {isGrid ? (
          <FileGrid
            files={initialFiles}
            folders={initialFolders}
            selectedFileIds={selectedFileIds}
            selectedFolderIds={selectedFolderIds}
            highlightFileId={highlightFileId}
            onSelectFile={selectFile}
            onSelectFolder={selectFolder}
            onFileClick={handleFileClick}
            onFolderDoubleClick={handleFolderDoubleClick}
            onFileRename={handleFileRename}
            onFileMove={handleFileMove}
            onFileManageTags={handleFileManageTags}
            onFileViewMetadata={handleFileViewMetadata}
            onFileAIOrganize={handleFileAIOrganize}
            onFolderRename={handleFolderRename}
            onFolderMove={handleFolderMove}
            onFolderManageTags={handleFolderManageTags}
            onFolderAIOrganize={handleFolderAIOrganize}
            onNewFolder={handleNewFolder}
            onUpload={handleUploadClick}
          />
        ) : (
          <FileList
            files={initialFiles}
            folders={initialFolders}
            selectedFileIds={selectedFileIds}
            selectedFolderIds={selectedFolderIds}
            highlightFileId={highlightFileId}
            onSelectFile={selectFile}
            onSelectFolder={selectFolder}
            onFileClick={handleFileClick}
            onFolderDoubleClick={handleFolderDoubleClick}
            onFileRename={handleFileRename}
            onFileMove={handleFileMove}
            onFileManageTags={handleFileManageTags}
            onFileViewMetadata={handleFileViewMetadata}
            onFileAIOrganize={handleFileAIOrganize}
            onFolderRename={handleFolderRename}
            onFolderMove={handleFolderMove}
            onFolderManageTags={handleFolderManageTags}
            onFolderAIOrganize={handleFolderAIOrganize}
            onNewFolder={handleNewFolder}
            onUpload={handleUploadClick}
            onSelectAll={(selectAll) => {
              if (!selectAll) {
                clearSelection();
              }
            }}
          />
        )}
      </DroppableFolder>

      {/* Hidden file input */}
      <input
        type="file"
        ref={fileInputRef}
        onChange={handleFilesSelected}
        multiple
        className="hidden"
      />

      {/* Upload Queue */}
      <UploadQueue
        items={uploadQueue.items}
        isExpanded={uploadQueue.isExpanded}
        onToggleExpanded={uploadQueue.toggleExpanded}
        onRemoveItem={uploadQueue.removeItem}
        onRetryItem={uploadQueue.retryItem}
        onClearCompleted={uploadQueue.clearCompleted}
      />

      <CreateFolderDialog
        open={showCreateFolder}
        onOpenChange={setShowCreateFolder}
        parentId={currentFolder?.id}
      />

      <FileMetadataSheet
        file={metadataFile}
        open={metadataFile !== null}
        onOpenChange={(open) => {
          if (!open) setMetadataFile(null);
        }}
      />
    </DndProvider>
  );
}
