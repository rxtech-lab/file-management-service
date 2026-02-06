"use client";

import { useCallback } from "react";
import { useFileManagementContext } from "@/lib/plugins/file-management-context";
import type { FileItem } from "@/lib/api/types";
import type { FileOpenHandler, FileManagementPlugin } from "@/lib/plugins/types";

export interface UseFileActionReturn {
  /** Check if any plugin can open this file */
  canOpen: (file: FileItem) => boolean;

  /** Get all plugins that can handle this file */
  getOpenHandlers: (file: FileItem) => FileOpenHandler[];

  /** Handle double-click - opens edit page in new tab */
  handleDoubleClick: (file: FileItem) => void;

  /** Handle "Open with" menu item - opens with specific plugin */
  handleOpenWith: (file: FileItem, pluginId: string) => void;

  /** Get the default handler for a file (if any) */
  getDefaultHandler: (file: FileItem) => FileOpenHandler | null;

  /** Check if any plugin can preview this file */
  canPreview: (file: FileItem) => boolean;

  /** Get plugins that can preview this file */
  getPreviewHandlers: (file: FileItem) => FileManagementPlugin[];

  /** Open preview dialog using a specific plugin */
  handlePreviewWith: (file: FileItem, pluginId: string) => void;
}

export function useFileAction(): UseFileActionReturn {
  const {
    canOpen, getOpenHandlers, getDefaultHandler, plugins, callbacks,
    openPreviewWith, getPreviewHandlers, canPreview,
  } = useFileManagementContext();

  const handleDoubleClick = useCallback(
    (file: FileItem) => {
      const defaultHandler = getDefaultHandler(file);
      if (defaultHandler) {
        const plugin = defaultHandler.plugin;
        if (plugin.editor) {
          window.open(plugin.editor.getEditUrl(file), "_blank");
        } else {
          plugin.open(file);
        }
        callbacks?.onOpen?.(file, plugin.id);
      }
    },
    [getDefaultHandler, callbacks]
  );

  const handleOpenWith = useCallback(
    (file: FileItem, pluginId: string) => {
      const plugin = plugins.find((p) => p.id === pluginId);
      if (plugin && plugin.canOpen(file)) {
        if (plugin.editor) {
          window.open(plugin.editor.getEditUrl(file), "_blank");
        } else {
          plugin.open(file);
        }
        callbacks?.onOpen?.(file, pluginId);
      } else {
        console.warn(
          `Plugin "${pluginId}" cannot open this file or does not exist`
        );
      }
    },
    [plugins, callbacks]
  );

  const handlePreviewWith = useCallback(
    (file: FileItem, pluginId: string) => {
      openPreviewWith(file, pluginId);
    },
    [openPreviewWith]
  );

  return {
    canOpen,
    getOpenHandlers,
    handleDoubleClick,
    handleOpenWith,
    getDefaultHandler,
    canPreview,
    getPreviewHandlers,
    handlePreviewWith,
  };
}
