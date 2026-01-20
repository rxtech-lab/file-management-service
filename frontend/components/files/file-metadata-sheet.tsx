"use client";

import { useState, useEffect, useCallback } from "react";
import { useRouter } from "next/navigation";
import {
  Download,
  Trash2,
  Loader2,
  Check,
  X,
  Pencil,
} from "lucide-react";
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import { TagBadge } from "@/components/tags/tag-badge";
import { TagPicker } from "@/components/tags/tag-picker";
import { FileTypeIcon } from "./file-type-icon";
import { toast } from "sonner";
import {
  updateFileAction,
  deleteFileAction,
  getFileDownloadAction,
  addTagsToFileAction,
  removeTagsFromFileAction,
} from "@/lib/actions/file-actions";
import { formatDateTime } from "@/lib/utils";
import type { FileItem, Tag } from "@/lib/api/types";

interface FileMetadataSheetProps {
  file: FileItem | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
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

export function FileMetadataSheet({
  file,
  open,
  onOpenChange,
}: FileMetadataSheetProps) {
  const router = useRouter();
  const [isEditingTitle, setIsEditingTitle] = useState(false);
  const [isEditingSummary, setIsEditingSummary] = useState(false);
  const [isEditingTags, setIsEditingTags] = useState(false);
  const [title, setTitle] = useState("");
  const [summary, setSummary] = useState("");
  const [isSaving, setIsSaving] = useState(false);
  const [isDownloading, setIsDownloading] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);

  // Reset state when file changes
  useEffect(() => {
    if (file) {
      setTitle(file.title);
      setSummary(file.summary || "");
      setIsEditingTitle(false);
      setIsEditingSummary(false);
      setIsEditingTags(false);
    }
  }, [file]);

  const handleSaveTitle = useCallback(async () => {
    if (!file || !title.trim()) return;

    setIsSaving(true);
    try {
      const result = await updateFileAction(file.id, { title: title.trim() });
      if (result.success) {
        toast.success("Title updated");
        setIsEditingTitle(false);
        router.refresh();
      } else {
        toast.error(result.error || "Failed to update title");
      }
    } catch {
      toast.error("Failed to update title");
    } finally {
      setIsSaving(false);
    }
  }, [file, title, router]);

  const handleSaveSummary = useCallback(async () => {
    if (!file) return;

    setIsSaving(true);
    try {
      const result = await updateFileAction(file.id, {
        summary: summary.trim() || undefined,
      });
      if (result.success) {
        toast.success("Summary updated");
        setIsEditingSummary(false);
        router.refresh();
      } else {
        toast.error(result.error || "Failed to update summary");
      }
    } catch {
      toast.error("Failed to update summary");
    } finally {
      setIsSaving(false);
    }
  }, [file, summary, router]);

  const handleAddTag = useCallback(
    async (tag: Tag) => {
      if (!file) return;

      try {
        const result = await addTagsToFileAction(file.id, [tag.id]);
        if (result.success) {
          toast.success("Tag added");
          router.refresh();
        } else {
          toast.error(result.error || "Failed to add tag");
        }
      } catch {
        toast.error("Failed to add tag");
      }
    },
    [file, router],
  );

  const handleRemoveTag = useCallback(
    async (tagId: number) => {
      if (!file) return;

      try {
        const result = await removeTagsFromFileAction(file.id, [tagId]);
        if (result.success) {
          toast.success("Tag removed");
          router.refresh();
        } else {
          toast.error(result.error || "Failed to remove tag");
        }
      } catch {
        toast.error("Failed to remove tag");
      }
    },
    [file, router],
  );

  const handleDownload = useCallback(async () => {
    if (!file) return;

    setIsDownloading(true);
    try {
      const result = await getFileDownloadAction(file.id);
      if (result.success && result.data) {
        window.open(result.data.download_url, "_blank");
      } else {
        toast.error(result.error || "Failed to get download URL");
      }
    } catch {
      toast.error("Failed to download file");
    } finally {
      setIsDownloading(false);
    }
  }, [file]);

  const handleDelete = useCallback(async () => {
    if (!file) return;

    if (!confirm(`Are you sure you want to delete "${file.title}"?`)) {
      return;
    }

    setIsDeleting(true);
    try {
      const result = await deleteFileAction(file.id);
      if (result.success) {
        toast.success("File deleted");
        onOpenChange(false);
        router.refresh();
      } else {
        toast.error(result.error || "Failed to delete file");
      }
    } catch {
      toast.error("Failed to delete file");
    } finally {
      setIsDeleting(false);
    }
  }, [file, router, onOpenChange]);

  if (!file) return null;

  const isProcessing =
    file.processing_status === "pending" ||
    file.processing_status === "processing";

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent className="w-[400px]! max-w-[500px]! overflow-y-auto px-4">
        <SheetHeader>
          <SheetTitle className="flex items-center gap-2">
            <FileTypeIcon
              fileType={file.file_type}
              mimeType={file.mime_type}
              className="h-5 w-5"
            />
            File Details
          </SheetTitle>
        </SheetHeader>

        <div className="space-y-6 py-6">
          {/* Title */}
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <label className="text-sm font-medium text-muted-foreground">
                Title
              </label>
              {!isEditingTitle && (
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => setIsEditingTitle(true)}
                >
                  <Pencil className="h-3 w-3" />
                </Button>
              )}
            </div>
            {isEditingTitle ? (
              <div className="flex items-center gap-2">
                <Input
                  value={title}
                  onChange={(e) => setTitle(e.target.value)}
                  disabled={isSaving}
                  autoFocus
                />
                <Button
                  size="icon"
                  variant="ghost"
                  onClick={handleSaveTitle}
                  disabled={isSaving || !title.trim()}
                >
                  {isSaving ? (
                    <Loader2 className="h-4 w-4 animate-spin" />
                  ) : (
                    <Check className="h-4 w-4" />
                  )}
                </Button>
                <Button
                  size="icon"
                  variant="ghost"
                  onClick={() => {
                    setTitle(file.title);
                    setIsEditingTitle(false);
                  }}
                  disabled={isSaving}
                >
                  <X className="h-4 w-4" />
                </Button>
              </div>
            ) : (
              <p className="font-medium">{file.title}</p>
            )}
            {file.title !== file.original_filename && (
              <p className="text-xs text-muted-foreground">
                Original: {file.original_filename}
              </p>
            )}
          </div>

          <Separator />

          {/* File Info */}
          <div className="grid grid-cols-2 gap-4 text-sm">
            <div>
              <p className="text-muted-foreground">Type</p>
              <p className="font-medium capitalize">
                {file.file_type} ({file.mime_type || "unknown"})
              </p>
            </div>
            <div>
              <p className="text-muted-foreground">Size</p>
              <p className="font-medium">{formatFileSize(file.size)}</p>
            </div>
            <div>
              <p className="text-muted-foreground">Created</p>
              <p className="font-medium">{formatDateTime(file.created_at)}</p>
            </div>
            <div>
              <p className="text-muted-foreground">Modified</p>
              <p className="font-medium">{formatDateTime(file.updated_at)}</p>
            </div>
          </div>

          {/* Processing Status */}
          <div className="space-y-2">
            <p className="text-sm text-muted-foreground">Processing Status</p>
            <div className="flex items-center gap-2">
              {isProcessing ? (
                <>
                  <Loader2 className="h-4 w-4 animate-spin" />
                  <span className="text-sm capitalize">
                    {file.processing_status}...
                  </span>
                </>
              ) : file.processing_status === "completed" ? (
                <>
                  <Check className="h-4 w-4 text-green-500" />
                  <span className="text-sm text-green-600">Completed</span>
                </>
              ) : file.processing_status === "failed" ? (
                <>
                  <X className="h-4 w-4 text-red-500" />
                  <span className="text-sm text-red-600">Failed</span>
                </>
              ) : (
                <Badge variant="secondary" className="capitalize">
                  {file.processing_status}
                </Badge>
              )}
            </div>
            {file.processing_error && (
              <p className="text-xs text-red-500">{file.processing_error}</p>
            )}
          </div>

          <Separator />

          {/* Tags */}
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <label className="text-sm font-medium text-muted-foreground">
                Tags
              </label>
              {!isEditingTags && (
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => setIsEditingTags(true)}
                >
                  <Pencil className="h-3 w-3" />
                </Button>
              )}
            </div>
            <div className="flex flex-wrap gap-2">
              {file.tags && file.tags.length > 0 ? (
                file.tags.map((tag) => (
                  <TagBadge
                    key={tag.id}
                    tag={tag}
                    onRemove={
                      isEditingTags ? () => handleRemoveTag(tag.id) : undefined
                    }
                  />
                ))
              ) : (
                <span className="text-sm text-muted-foreground">No tags</span>
              )}
            </div>
            {isEditingTags && (
              <div className="pt-2">
                <TagPicker
                  selectedTagIds={file.tags?.map((t) => t.id) || []}
                  onSelect={handleAddTag}
                />
                <Button
                  variant="outline"
                  size="sm"
                  className="mt-2"
                  onClick={() => setIsEditingTags(false)}
                >
                  Done
                </Button>
              </div>
            )}
          </div>

          <Separator />

          {/* AI Summary */}
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <label className="text-sm font-medium text-muted-foreground">
                AI Summary
              </label>
              {!isEditingSummary && (
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => setIsEditingSummary(true)}
                >
                  <Pencil className="h-3 w-3" />
                </Button>
              )}
            </div>
            {isEditingSummary ? (
              <div className="space-y-2">
                <Textarea
                  value={summary}
                  onChange={(e) => setSummary(e.target.value)}
                  disabled={isSaving}
                  rows={4}
                  placeholder="Add a summary..."
                />
                <div className="flex gap-2">
                  <Button
                    size="sm"
                    onClick={handleSaveSummary}
                    disabled={isSaving}
                  >
                    {isSaving ? (
                      <Loader2 className="h-4 w-4 animate-spin mr-2" />
                    ) : null}
                    Save
                  </Button>
                  <Button
                    size="sm"
                    variant="outline"
                    onClick={() => {
                      setSummary(file.summary || "");
                      setIsEditingSummary(false);
                    }}
                    disabled={isSaving}
                  >
                    Cancel
                  </Button>
                </div>
              </div>
            ) : (
              <p className="text-sm break-all">
                {file.summary || (
                  <span className="text-muted-foreground">
                    {isProcessing ? "Processing..." : "No summary available"}
                  </span>
                )}
              </p>
            )}
          </div>

          <Separator />

          {/* Actions */}
          <div className="flex gap-2">
            <Button
              variant="outline"
              className="flex-1"
              onClick={handleDownload}
              disabled={isDownloading}
            >
              {isDownloading ? (
                <Loader2 className="h-4 w-4 animate-spin mr-2" />
              ) : (
                <Download className="h-4 w-4 mr-2" />
              )}
              Download
            </Button>
            <Button
              variant="destructive"
              onClick={handleDelete}
              disabled={isDeleting}
            >
              {isDeleting ? (
                <Loader2 className="h-4 w-4 animate-spin" />
              ) : (
                <Trash2 className="h-4 w-4" />
              )}
            </Button>
          </div>
        </div>
      </SheetContent>
    </Sheet>
  );
}
