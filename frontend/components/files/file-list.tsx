"use client";

import { useRouter } from "next/navigation";
import {
  Folder as FolderIcon,
  Loader2,
  Pencil,
  FolderInput,
  Download,
  Tags,
  Info,
  Trash2,
} from "lucide-react";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
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
import { Checkbox } from "@/components/ui/checkbox";
import { toast } from "sonner";
import { FileTypeIcon, getFileTypeColor } from "./file-type-icon";
import { TagBadge } from "@/components/tags/tag-badge";
import { EmptyContextMenuWrapper } from "@/components/context-menus/empty-context-menu";
import {
  deleteFileAction,
  getFileDownloadAction,
} from "@/lib/actions/file-actions";
import { deleteFolderAction } from "@/lib/actions/folder-actions";
import type { FileItem, Folder } from "@/lib/api/types";
import { cn } from "@/lib/utils";
import { formatDate } from "@/lib/utils";

interface FileListProps {
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
  // Folder actions
  onFolderRename?: (folder: Folder) => void;
  onFolderMove?: (folder: Folder) => void;
  onFolderManageTags?: (folder: Folder) => void;
  // Empty space actions
  onNewFolder?: () => void;
  onUpload?: () => void;
  onSelectAll?: (selectAll: boolean) => void;
}

function formatFileSize(bytes?: number): string {
  if (!bytes) return "-";
  const units = ["B", "KB", "MB", "GB"];
  let size = bytes;
  let unitIndex = 0;
  while (size >= 1024 && unitIndex < units.length - 1) {
    size /= 1024;
    unitIndex++;
  }
  return `${size.toFixed(1)} ${units[unitIndex]}`;
}

export function FileList({
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
  onFolderRename,
  onFolderMove,
  onFolderManageTags,
  onNewFolder,
  onUpload,
  onSelectAll,
}: FileListProps) {
  const router = useRouter();
  const allItemsCount = files.length + folders.length;
  const selectedCount = selectedFileIds.size + selectedFolderIds.size;
  const allSelected = allItemsCount > 0 && selectedCount === allItemsCount;
  const someSelected = selectedCount > 0 && selectedCount < allItemsCount;

  const isEmpty = files.length === 0 && folders.length === 0;

  const handleFileDownload = async (file: FileItem) => {
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

  const handleFileDelete = async (file: FileItem) => {
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

  const handleFolderDelete = async (folder: Folder) => {
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

  if (isEmpty) {
    return (
      <EmptyContextMenuWrapper
        className="flex-1 p-4 flex items-center justify-center"
        onNewFolder={onNewFolder ?? (() => {})}
        onUpload={onUpload ?? (() => {})}
      >
        <div className="text-center text-muted-foreground">
          <p className="text-lg font-medium">This folder is empty</p>
          <p className="text-sm">
            Right-click to create a folder or upload files
          </p>
        </div>
      </EmptyContextMenuWrapper>
    );
  }

  return (
    <EmptyContextMenuWrapper
      className="flex-1 overflow-auto"
      onNewFolder={onNewFolder ?? (() => {})}
      onUpload={onUpload ?? (() => {})}
    >
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead className="w-[40px]">
              <Checkbox
                checked={
                  allSelected ? true : someSelected ? "indeterminate" : false
                }
                onCheckedChange={(checked) => onSelectAll?.(!!checked)}
              />
            </TableHead>
            <TableHead>Name</TableHead>
            <TableHead className="w-[100px]">Size</TableHead>
            <TableHead className="w-[150px]">Modified</TableHead>
            <TableHead className="w-[200px]">Tags</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {/* Folders */}
          {folders.map((folder) => (
            <ContextMenu key={`folder-${folder.id}`}>
              <ContextMenuTrigger asChild>
                <TableRow
                  className={cn(
                    "cursor-pointer",
                    selectedFolderIds.has(folder.id) && "bg-muted",
                  )}
                  onClick={(e) =>
                    onSelectFolder?.(folder, e.metaKey || e.ctrlKey)
                  }
                  onDoubleClick={() => onFolderDoubleClick?.(folder)}
                >
                  <TableCell>
                    <Checkbox
                      checked={selectedFolderIds.has(folder.id)}
                      onClick={(e) => e.stopPropagation()}
                      onCheckedChange={() => onSelectFolder?.(folder, true)}
                    />
                  </TableCell>
                  <TableCell>
                    <div className="flex items-center gap-2">
                      <FolderIcon className="h-5 w-5 text-yellow-500 fill-yellow-100" />
                      <span className="font-medium">{folder.name}</span>
                    </div>
                  </TableCell>
                  <TableCell className="text-muted-foreground">-</TableCell>
                  <TableCell className="text-muted-foreground">
                    {formatDate(new Date(folder.updated_at))}
                  </TableCell>
                  <TableCell>
                    {folder.tags && folder.tags.length > 0 && (
                      <div className="flex flex-wrap gap-1">
                        {folder.tags.slice(0, 2).map((tag) => (
                          <TagBadge
                            key={tag.id}
                            tag={tag}
                            className="text-xs"
                          />
                        ))}
                        {folder.tags.length > 2 && (
                          <span className="text-xs text-muted-foreground">
                            +{folder.tags.length - 2}
                          </span>
                        )}
                      </div>
                    )}
                  </TableCell>
                </TableRow>
              </ContextMenuTrigger>

              <ContextMenuContent className="w-48">
                <ContextMenuItem onClick={() => onFolderRename?.(folder)}>
                  <Pencil className="mr-2 h-4 w-4" />
                  Rename
                </ContextMenuItem>
                <ContextMenuItem onClick={() => onFolderMove?.(folder)}>
                  <FolderInput className="mr-2 h-4 w-4" />
                  Move to...
                </ContextMenuItem>

                <ContextMenuSeparator />

                <ContextMenuItem onClick={() => onFolderManageTags?.(folder)}>
                  <Tags className="mr-2 h-4 w-4" />
                  Add tags...
                </ContextMenuItem>

                <ContextMenuSeparator />

                <ContextMenuItem
                  onClick={() => handleFolderDelete(folder)}
                  variant="destructive"
                >
                  <Trash2 className="mr-2 h-4 w-4" />
                  Delete
                </ContextMenuItem>
              </ContextMenuContent>
            </ContextMenu>
          ))}

          {/* Files */}
          {files.map((file) => {
            const isProcessing =
              file.processing_status === "pending" ||
              file.processing_status === "processing";

            return (
              <ContextMenu key={`file-${file.id}`}>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <ContextMenuTrigger asChild>
                      <TableRow
                        className={cn(
                          "cursor-pointer",
                          selectedFileIds.has(file.id) && "bg-muted",
                          isProcessing && "opacity-70",
                        )}
                        onClick={(e) => {
                          onSelectFile?.(file, e.metaKey || e.ctrlKey);
                          onFileClick?.(file);
                        }}
                      >
                        <TableCell>
                          <Checkbox
                            checked={selectedFileIds.has(file.id)}
                            onClick={(e) => e.stopPropagation()}
                            onCheckedChange={() => onSelectFile?.(file, true)}
                          />
                        </TableCell>
                        <TableCell>
                          <div className="flex items-center gap-2">
                            <div
                              className={cn(
                                "relative",
                                getFileTypeColor(
                                  file.file_type,
                                  file.mime_type,
                                ),
                              )}
                            >
                              <FileTypeIcon
                                fileType={file.file_type}
                                mimeType={file.mime_type}
                                className="h-5 w-5"
                              />
                              {isProcessing && (
                                <Loader2 className="h-3 w-3 absolute -bottom-1 -right-1 animate-spin" />
                              )}
                            </div>
                            <div>
                              <span className="font-medium">{file.title}</span>
                              {file.title !== file.original_filename && (
                                <span className="text-xs text-muted-foreground ml-2">
                                  ({file.original_filename})
                                </span>
                              )}
                            </div>
                          </div>
                        </TableCell>
                        <TableCell className="text-muted-foreground">
                          {formatFileSize(file.size)}
                        </TableCell>
                        <TableCell className="text-muted-foreground">
                          {formatDate(new Date(file.updated_at))}
                        </TableCell>
                        <TableCell>
                          {file.tags && file.tags.length > 0 && (
                            <div className="flex flex-wrap gap-1">
                              {file.tags.slice(0, 2).map((tag) => (
                                <TagBadge
                                  key={tag.id}
                                  tag={tag}
                                  className="text-xs"
                                />
                              ))}
                              {file.tags.length > 2 && (
                                <span className="text-xs text-muted-foreground">
                                  +{file.tags.length - 2}
                                </span>
                              )}
                            </div>
                          )}
                        </TableCell>
                      </TableRow>
                    </ContextMenuTrigger>
                  </TooltipTrigger>

                  {/* AI Summary tooltip */}
                  {file.summary && (
                    <TooltipContent side="bottom" className="max-w-xs">
                      <div className="space-y-1">
                        <p className="font-medium text-xs">AI Summary</p>
                        <p className="text-xs text-muted-foreground break-all">
                          {file.summary}
                        </p>
                      </div>
                    </TooltipContent>
                  )}
                </Tooltip>

                <ContextMenuContent className="w-48">
                  <ContextMenuItem onClick={() => onFileRename?.(file)}>
                    <Pencil className="mr-2 h-4 w-4" />
                    Rename
                  </ContextMenuItem>
                  <ContextMenuItem onClick={() => onFileMove?.(file)}>
                    <FolderInput className="mr-2 h-4 w-4" />
                    Move to...
                  </ContextMenuItem>
                  <ContextMenuItem onClick={() => handleFileDownload(file)}>
                    <Download className="mr-2 h-4 w-4" />
                    Download
                  </ContextMenuItem>

                  <ContextMenuSeparator />

                  <ContextMenuItem onClick={() => onFileManageTags?.(file)}>
                    <Tags className="mr-2 h-4 w-4" />
                    Add tags...
                  </ContextMenuItem>
                  <ContextMenuItem onClick={() => onFileViewMetadata?.(file)}>
                    <Info className="mr-2 h-4 w-4" />
                    View metadata
                  </ContextMenuItem>

                  <ContextMenuSeparator />

                  <ContextMenuItem
                    onClick={() => handleFileDelete(file)}
                    variant="destructive"
                  >
                    <Trash2 className="mr-2 h-4 w-4" />
                    Delete
                  </ContextMenuItem>
                </ContextMenuContent>
              </ContextMenu>
            );
          })}
        </TableBody>
      </Table>
    </EmptyContextMenuWrapper>
  );
}
