"use client";

import { FileCard } from "./file-card";
import { FolderCard } from "@/components/folders/folder-card";
import {
  DraggableFile,
  DraggableFolder,
} from "@/components/dnd/draggable-item";
import { DroppableFolderTreeItem } from "@/components/dnd/droppable-folder";
import { EmptyContextMenuWrapper } from "@/components/context-menus/empty-context-menu";
import type { FileItem, Folder } from "@/lib/api/types";

interface FileGridProps {
  files: FileItem[];
  folders: Folder[];
  selectedFileIds: Set<number>;
  selectedFolderIds: Set<number>;
  onSelectFile?: (file: FileItem, multi?: boolean) => void;
  onSelectFolder?: (folder: Folder, multi?: boolean) => void;
  onFileClick?: (file: FileItem) => void;
  onFolderDoubleClick?: (folder: Folder) => void;
  // File actions
  onFileRename?: (file: FileItem) => void;
  onFileMove?: (file: FileItem) => void;
  onFileManageTags?: (file: FileItem) => void;
  onFileViewMetadata?: (file: FileItem) => void;
  onFileAIOrganize?: (file: FileItem) => void;
  // Folder actions
  onFolderRename?: (folder: Folder) => void;
  onFolderMove?: (folder: Folder) => void;
  onFolderManageTags?: (folder: Folder) => void;
  onFolderAIOrganize?: (folder: Folder) => void;
  // Empty space actions
  onNewFolder?: () => void;
  onUpload?: () => void;
}

export function FileGrid({
  files,
  folders,
  selectedFileIds,
  selectedFolderIds,
  onSelectFile,
  onSelectFolder,
  onFileClick,
  onFolderDoubleClick,
  onFileRename,
  onFileMove,
  onFileManageTags,
  onFileViewMetadata,
  onFileAIOrganize,
  onFolderRename,
  onFolderMove,
  onFolderManageTags,
  onFolderAIOrganize,
  onNewFolder,
  onUpload,
}: FileGridProps) {
  const isEmpty = files.length === 0 && folders.length === 0;

  const content = isEmpty ? (
    <div className="h-full flex items-center justify-center">
      <div className="text-center text-muted-foreground">
        <p className="text-lg font-medium">This folder is empty</p>
        <p className="text-sm">
          Right-click to create a folder or upload files
        </p>
      </div>
    </div>
  ) : (
    <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-4">
      {/* Folders first */}
      {folders.map((folder) => (
        <DraggableFolder key={`folder-${folder.id}`} folder={folder}>
          <DroppableFolderTreeItem folderId={folder.id}>
            <FolderCard
              folder={folder}
              selected={selectedFolderIds.has(folder.id)}
              onSelect={onSelectFolder}
              onDoubleClick={onFolderDoubleClick}
              onRename={onFolderRename}
              onMove={onFolderMove}
              onManageTags={onFolderManageTags}
              onAIOrganize={onFolderAIOrganize}
            />
          </DroppableFolderTreeItem>
        </DraggableFolder>
      ))}

      {/* Then files */}
      {files.map((file) => (
        <DraggableFile key={`file-${file.id}`} file={file}>
          <FileCard
            file={file}
            selected={selectedFileIds.has(file.id)}
            onSelect={onSelectFile}
            onClick={onFileClick}
            onRename={onFileRename}
            onMove={onFileMove}
            onManageTags={onFileManageTags}
            onViewMetadata={onFileViewMetadata}
            onAIOrganize={onFileAIOrganize}
          />
        </DraggableFile>
      ))}
    </div>
  );

  return (
    <EmptyContextMenuWrapper
      className="flex-1 p-4 overflow-auto"
      onNewFolder={onNewFolder ?? (() => {})}
      onUpload={onUpload ?? (() => {})}
    >
      {content}
    </EmptyContextMenuWrapper>
  );
}
