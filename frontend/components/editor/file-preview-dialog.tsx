"use client";

import { useEffect, useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { useFileManagementContext } from "@/lib/plugins/file-management-context";
import { getFileContentAction } from "@/lib/actions/file-content-actions";
import { FileWarning, Loader2 } from "lucide-react";

const MAX_FILE_SIZE = 4 * 1024 * 1024; // 4MB

export function FilePreviewDialog() {
  const { previewFile, previewComponent: PreviewComponent, closePreview } =
    useFileManagementContext();
  const [content, setContent] = useState<string>("");
  const [loading, setLoading] = useState(false);

  const open = previewFile !== null;

  const fileTooLarge = previewFile && previewFile.size && previewFile.size > MAX_FILE_SIZE;

  useEffect(() => {
    if (previewFile && open && !fileTooLarge) {
      setLoading(true);
      setContent("");
      getFileContentAction(previewFile.id).then((result) => {
        if (result.success && result.data) {
          setContent(result.data.content);
        }
        setLoading(false);
      });
    }
  }, [previewFile, open, fileTooLarge]);

  return (
    <Dialog
      open={open}
      onOpenChange={(o) => {
        if (!o) closePreview();
      }}
    >
      <DialogContent className="w-full max-w-[90vw] sm:max-w-[90vw] max-h-[90vh] flex flex-col">
        <DialogHeader>
          <DialogTitle>{previewFile?.title ?? "Preview"}</DialogTitle>
        </DialogHeader>

        <div className="h-[70vh] overflow-hidden">
          {fileTooLarge ? (
            <div className="flex flex-col items-center justify-center h-[400px] gap-3 text-muted-foreground">
              <FileWarning className="h-12 w-12" />
              <p className="text-lg font-medium">File too large to preview</p>
              <p className="text-sm">
                This file is {((previewFile?.size ?? 0) / 1024 / 1024).toFixed(1)} MB. Preview is limited to files under 4 MB.
              </p>
            </div>
          ) : loading ? (
            <div className="flex items-center justify-center h-[400px]">
              <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
            </div>
          ) : PreviewComponent && previewFile ? (
            <PreviewComponent
              file={previewFile}
              content={content}
            />
          ) : (
            <pre className="text-sm whitespace-pre-wrap p-4 border rounded-md max-h-[400px] overflow-auto">
              {content}
            </pre>
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}
