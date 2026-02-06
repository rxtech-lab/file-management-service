"use client";

import { useState, useCallback, useMemo } from "react";
import Link from "next/link";
import { ArrowLeft } from "lucide-react";
import { toast } from "sonner";
import { SidebarTrigger } from "@/components/ui/sidebar";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import { useFileManagementContext } from "@/lib/plugins/file-management-context";
import {
  updateFileAction,
  getFileDownloadAction,
  moveFilesAction,
} from "@/lib/actions/file-actions";
import { updateFileContentAction } from "@/lib/actions/file-content-actions";
import { useAutoSave } from "@/hooks/use-auto-save";
import { useKeyboardShortcuts } from "@/hooks/use-keyboard-shortcuts";
import { EditorToolbar } from "@/components/editor/editor-toolbar";
import { MonacoEditor } from "@/components/editor/monaco-editor";
import { MoveFileDialog } from "@/components/editor/move-file-dialog";
import type { FileItem } from "@/lib/api/types";
import type {
  ToolbarMenuGroup,
  ToolbarMenuItem,
  EditorActions,
} from "@/lib/plugins/types";

interface PreviewPageClientProps {
  file: FileItem;
  content: string;
  pluginId?: string;
}

export function PreviewPageClient({
  file,
  content: initialContent,
  pluginId,
}: PreviewPageClientProps) {
  const { plugins, getPreviewHandlers } = useFileManagementContext();
  const [content, setContent] = useState(initialContent);
  const [currentFile, setCurrentFile] = useState(file);
  const [moveDialogOpen, setMoveDialogOpen] = useState(false);

  // Find the plugin for this file
  const plugin = useMemo(() => {
    if (pluginId) {
      return plugins.find((p) => p.id === pluginId) ?? null;
    }
    return plugins.find((p) => p.canOpen(file) && p.editor) ?? null;
  }, [plugins, pluginId, file]);

  const hasEditor = !!plugin?.editor;

  // Resolve preview component (for non-editable files)
  const PreviewComponent = useMemo(() => {
    if (hasEditor) return null;
    if (pluginId) {
      const p = plugins.find((pp) => pp.id === pluginId);
      return p?.onPreview?.(file) ?? null;
    }
    const handlers = getPreviewHandlers(file);
    if (handlers.length > 0) {
      return handlers[0].onPreview?.(file) ?? null;
    }
    return null;
  }, [plugins, getPreviewHandlers, pluginId, file, hasEditor]);

  // Back link
  const backHref = currentFile.folder_id
    ? `/files/${currentFile.folder_id}`
    : "/files";

  // Auto-save
  const handleSave = useCallback(
    async (newContent: string) => {
      const result = await updateFileContentAction(currentFile.id, {
        content: newContent,
        mime_type: currentFile.mime_type || undefined,
      });
      if (!result.success) {
        throw new Error(result.error || "Failed to save");
      }
    },
    [currentFile.id, currentFile.mime_type],
  );

  const { saveStatus, triggerSave, forceSave, lastSavedAt, error } =
    useAutoSave({ onSave: handleSave, enabled: hasEditor });

  // Editor actions
  const editorActions: EditorActions = useMemo(
    () => ({
      save: async () => {
        await forceSave();
      },
      saveACopy: async () => {
        toast.info("Save a copy is not yet implemented");
      },
      createNewFile: async () => {
        toast.info("Create new file is not yet implemented");
      },
      moveFile: () => {
        setMoveDialogOpen(true);
      },
      downloadFile: async () => {
        const result = await getFileDownloadAction(currentFile.id);
        if (result.success && result.data) {
          window.open(result.data.download_url, "_blank");
        } else {
          toast.error("Failed to get download URL");
        }
      },
    }),
    [forceSave, currentFile.id],
  );

  // Toolbar menus
  const menus: ToolbarMenuGroup[] = useMemo(() => {
    if (plugin?.editor?.getToolbarMenus) {
      return plugin.editor.getToolbarMenus(currentFile, editorActions);
    }
    if (!hasEditor) return [];
    return [
      {
        id: "file",
        label: "File",
        items: [
          {
            id: "save-copy",
            title: "Save a copy",
            description: "Create a duplicate of this file",
            shortcut: "Ctrl+Shift+S",
            onAction: editorActions.saveACopy,
          },
          {
            id: "create-new",
            title: "Create new file",
            description: "Create a blank new file",
            shortcut: "Ctrl+Alt+N",
            onAction: editorActions.createNewFile,
          },
          {
            id: "download",
            title: "Download",
            description: "Download the original file",
            shortcut: "Ctrl+Shift+D",
            onAction: editorActions.downloadFile,
            separator: true,
          },
        ],
      },
    ];
  }, [plugin, currentFile, editorActions, hasEditor]);

  const allShortcutItems: ToolbarMenuItem[] = useMemo(
    () => menus.flatMap((m) => m.items),
    [menus],
  );

  useKeyboardShortcuts(allShortcutItems);

  const handleContentChange = useCallback(
    (newContent: string) => {
      setContent(newContent);
      triggerSave(newContent);
    },
    [triggerSave],
  );

  const handleTitleChange = useCallback(
    async (newTitle: string) => {
      const result = await updateFileAction(currentFile.id, {
        title: newTitle,
      });
      if (result.success && result.data) {
        setCurrentFile(result.data);
      } else {
        toast.error("Failed to rename file");
      }
    },
    [currentFile.id],
  );

  const handleMove = useCallback(
    async (folderId: number | null) => {
      const result = await moveFilesAction([currentFile.id], folderId);
      if (result.success) {
        setCurrentFile((prev) => ({ ...prev, folder_id: folderId }));
        toast.success("File moved successfully");
      } else {
        toast.error("Failed to move file");
      }
    },
    [currentFile.id],
  );

  const handleDownload = async () => {
    const result = await getFileDownloadAction(currentFile.id);
    if (result.success && result.data) {
      window.open(result.data.download_url, "_blank");
    } else {
      toast.error("Failed to get download URL");
    }
  };

  // Editor mode: full editing with toolbar
  if (hasEditor) {
    return (
      <div className="flex flex-col h-full">
        <EditorToolbar
          title={currentFile.title}
          onTitleChange={handleTitleChange}
          menus={menus}
          saveStatus={saveStatus}
          lastSavedAt={lastSavedAt}
          saveError={error}
          onMoveFile={() => setMoveDialogOpen(true)}
          extraContent={
            plugin?.editor?.renderToolbarExtras?.(currentFile) ?? undefined
          }
          leftContent={
            <div className="flex items-center gap-1 shrink-0 self-stretch">
              <SidebarTrigger />
              <Separator orientation="vertical" />
              <Button variant="ghost" size="icon" asChild className="h-8 w-8">
                <Link href={backHref}>
                  <ArrowLeft className="h-4 w-4" />
                </Link>
              </Button>
            </div>
          }
        />

        <div className="flex-1 overflow-hidden px-6 py-4">
          <MonacoEditor
            value={content}
            onChange={handleContentChange}
            language="yaml"
            className="h-full"
          />
        </div>

        <MoveFileDialog
          currentFolderId={currentFile.folder_id}
          open={moveDialogOpen}
          onOpenChange={setMoveDialogOpen}
          onMove={handleMove}
        />
      </div>
    );
  }

  // Preview mode: read-only with simple header
  return (
    <div className="flex flex-col h-full">
      <header className="flex h-14 shrink-0 items-center gap-2 border-b px-4">
        <SidebarTrigger className="-ml-1" />
        <Separator orientation="vertical" className="h-4" />

        <Button variant="ghost" size="icon" asChild className="h-8 w-8">
          <Link href={backHref}>
            <ArrowLeft className="h-4 w-4" />
          </Link>
        </Button>

        <h1 className="text-sm font-medium truncate flex-1">
          {currentFile.title}
        </h1>

        <div className="flex items-center gap-1">
          <Button variant="ghost" size="sm" onClick={handleDownload}>
            Download
          </Button>
        </div>
      </header>

      <div className="flex-1 overflow-auto p-6">
        <div className="mx-auto max-w-5xl h-full">
          {PreviewComponent ? (
            <PreviewComponent file={currentFile} content={content} />
          ) : (
            <pre className="text-sm whitespace-pre-wrap p-4 border rounded-md bg-background overflow-auto">
              {content}
            </pre>
          )}
        </div>
      </div>
    </div>
  );
}
