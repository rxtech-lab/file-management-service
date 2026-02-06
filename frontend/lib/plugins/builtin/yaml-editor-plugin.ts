import { FileCode2 } from "lucide-react";
import type { FileManagementPlugin, EditorActions } from "../types";
import type { FileItem } from "@/lib/api/types";
import { YamlPreviewContent } from "@/components/editor/yaml-preview-dialog";

function isYamlFile(file: FileItem): boolean {
  const filename = file.original_filename?.toLowerCase() ?? "";
  const mimeType = file.mime_type?.toLowerCase() ?? "";
  return (
    filename.endsWith(".yaml") ||
    filename.endsWith(".yml") ||
    mimeType.includes("yaml") ||
    mimeType.includes("x-yaml")
  );
}

export function createYamlEditorPlugin(): FileManagementPlugin {
  return {
    id: "yaml-editor",
    name: "YAML Editor",
    icon: FileCode2,
    priority: 50,

    canOpen: isYamlFile,

    open: (file: FileItem) => {
      // Default open: navigate to edit page in new tab
      const url = `/preview/${file.id}?plugin=yaml-editor`;
      window.open(url, "_blank");
    },

    isDefault: isYamlFile,

    onPreview: (file: FileItem) => {
      if (!isYamlFile(file)) return null;
      return YamlPreviewContent;
    },

    editor: {
      getEditUrl: (file: FileItem) => `/preview/${file.id}?plugin=yaml-editor`,

      getToolbarMenus: (_file: FileItem, actions: EditorActions) => [
        {
          id: "file",
          label: "File",
          items: [
            {
              id: "save-copy",
              title: "Save a copy",
              description: "Create a duplicate of this file",
              shortcut: "Ctrl+Shift+S",
              onAction: actions.saveACopy,
            },
            {
              id: "create-new",
              title: "Create new file",
              description: "Create a blank new file",
              shortcut: "Ctrl+Alt+N",
              onAction: actions.createNewFile,
            },
            {
              id: "download",
              title: "Download",
              description: "Download the original file",
              shortcut: "Ctrl+Shift+D",
              onAction: actions.downloadFile,
              separator: true,
            },
          ],
        },
      ],
    },
  };
}
